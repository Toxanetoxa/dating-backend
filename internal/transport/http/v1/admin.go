package v1

import (
	"errors"

	"net/http"
	"strconv"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"github.com/toxanetoxa/dating-backend/internal/transport/http/middlewares"
	"github.com/toxanetoxa/dating-backend/pkg/geopoint"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type AdminHandlersManager struct {
	l                   *slog.Logger
	auth                service.AdminAuthService
	us                  service.AdminUsersService
	adminProfileService service.AdminProfileService
}

func NewAdminRoutesManager(l *slog.Logger, a service.AdminAuthService, us service.AdminUsersService, aps service.AdminProfileService) *AdminHandlersManager {
	return &AdminHandlersManager{
		l:                   l,
		auth:                a,
		us:                  us,
		adminProfileService: aps,
	}
}

// Auth login for admins
func (a *AdminHandlersManager) Auth() echo.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	type response struct {
		Token string `json:"token"`
	}
	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			a.l.Debug("bind error: %v", err)

			return c.NoContent(http.StatusBadRequest)
		}

		token, err := a.auth.Login(c.Request().Context(), requestData.Login, requestData.Password)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidPassword):
				return c.JSON(http.StatusUnauthorized, Response{Success: false})
			case errors.Is(err, service.ErrAdminNotFound):
				return c.JSON(http.StatusUnauthorized, Response{Success: false})
			default:
				a.l.Error(err.Error())

				return c.JSON(http.StatusInternalServerError, Response{Success: false})
			}
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: response{Token: token}})
	}
}

// Logout login for admins
func (a *AdminHandlersManager) Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenID, _ := c.Get(UserSessionIDContextKey).(string)

		a.l.Info(tokenID)

		err := a.adminProfileService.LogOut(c.Request().Context(), tokenID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInternal,
					Message: ErrTitleInternal,
				},
			})
		}

		return c.JSON(http.StatusOK, Response{
			Success: true,
		})
	}
}

func (a *AdminHandlersManager) My() echo.HandlerFunc {
	type response struct {
		ID    string `json:"ID"`
		Login string `json:"login"`
	}
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		admin, err := a.adminProfileService.GetByID(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInternal,
					Message: ErrTitleInternal,
				},
			})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: response{ID: admin.ID, Login: admin.Login}})
	}
}

// Dashboard basic metrics for admins
func (a *AdminHandlersManager) Dashboard() echo.HandlerFunc {

	return func(c echo.Context) error {
		u, err := a.us.Stats(c.Request().Context())
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: entity.UsersStats{
			TotalUsers:    u.TotalUsers,
			UsersByDay:    u.UsersByDay,
			UsersByWeek:   u.UsersByWeek,
			UsersActive:   u.UsersActive,
			UsersNew:      u.UsersNew,
			UsersInActive: u.UsersInActive,
			SexStats:      u.SexStats}})
	}
}

// UserList список пользователей
func (a *AdminHandlersManager) UserList() echo.HandlerFunc {
	const (
		defaultLimit  = 20
		defaultOffset = 0
	)
	type userResponse struct {
		ID     string            `json:"ID"`
		Name   *string           `json:"name"`
		Email  string            `json:"phone"` // todo rename json
		Status entity.UserStatus `json:"status"`
		Photo  string            `json:"photo"`
	}
	return func(c echo.Context) error {
		limit := defaultLimit
		offset := defaultOffset

		if c.QueryParams().Has("limit") {
			limit, _ = strconv.Atoi(c.QueryParams().Get("limit"))
		}

		if c.QueryParams().Has("offset") {
			offset, _ = strconv.Atoi(c.QueryParams().Get("offset"))
		}

		list, err := a.us.List(c.Request().Context(), limit, offset)
		if err != nil {
			a.l.Error(err.Error())

			return c.NoContent(http.StatusInternalServerError)
		}

		result := make([]*userResponse, 0, len(list))

		for i := range list {
			el := &userResponse{
				ID:     list[i].ID,
				Name:   list[i].FirstName,
				Email:  list[i].Email,
				Status: list[i].Status,
			}

			if len(list[i].Photos) > 0 {
				el.Photo = list[i].Photos[0].URL
			}

			result = append(result, el)
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: result})
	}
}

// BlockUser заблокировать пользователя
func (a *AdminHandlersManager) BlockUser() echo.HandlerFunc {
	type request struct {
		ID string `json:"ID"`
	}
	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			a.l.Debug("bind error: %v", err)

			return c.NoContent(http.StatusBadRequest)
		}

		err = a.us.Block(c.Request().Context(), requestData.ID)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.NoContent(http.StatusOK)
	}
}

// UnblockUser активировать\разблокировать пользователя
func (a *AdminHandlersManager) UnblockUser() echo.HandlerFunc {
	type request struct {
		ID string `json:"ID"`
	}
	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			a.l.Debug("bind error: %v", err)

			return c.NoContent(http.StatusBadRequest)
		}

		err = a.us.Unblock(c.Request().Context(), requestData.ID)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.NoContent(http.StatusOK)
	}
}

// VerifyUser ...
func (a *AdminHandlersManager) VerifyUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		// todo
		return nil
	}
}

func (a *AdminHandlersManager) UserInfo() echo.HandlerFunc {

	type userResponse struct {
		ID          string             `json:"ID"`
		Name        *string            `json:"name"`
		Email       string             `json:"phone"` // todo rename  json
		Status      entity.UserStatus  `json:"status"`
		Photos      []entity.UserPhoto `json:"photos"`
		UpdatedAt   time.Time          `json:"updatedAt"`
		CreatedAt   time.Time          `json:"createdAt"`
		Birthday    *entity.BirthDate  `json:"birthday"`
		Sex         *entity.UserSex    `json:"sex"`
		About       *string            `json:"about"`
		Geolocation geopoint.GeoPoint  `json:"geolocation"`
	}
	type request struct {
		ID string `json:"ID"`
	}
	return func(c echo.Context) error {
		elem, err := a.us.FindByID(c.Request().Context(), request{ID: c.Param("id")}.ID)
		if err != nil {
			a.l.Error(err.Error())
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, Response{Success: true, Data: &userResponse{
			ID:          elem.ID,
			Name:        elem.FirstName,
			Email:       elem.Email,
			Status:      elem.Status,
			Photos:      elem.Photos,
			UpdatedAt:   elem.UpdatedAt,
			CreatedAt:   elem.CreatedAt,
			Birthday:    elem.Birthday,
			Sex:         elem.Sex,
			About:       elem.About,
			Geolocation: elem.Geolocation,
		}})
	}
}
func setAdminRoutes(r *echo.Group, l *slog.Logger, a service.AdminAuthService, us service.AdminUsersService, aps service.AdminProfileService) {
	adminAuthMw := middlewares.NewAdminAuthMiddleware(a, l)

	m := NewAdminRoutesManager(l, a, us, aps)

	r.POST("/auth", m.Auth())
	r.POST("/logout", m.Logout(), adminAuthMw.AdminAuthEchoMiddleware)
	r.GET("/my", m.My(), adminAuthMw.AdminAuthEchoMiddleware)

	r.GET("/dashboard", m.Dashboard(), adminAuthMw.AdminAuthEchoMiddleware)

	r.GET("/user/:id", m.UserInfo(), adminAuthMw.AdminAuthEchoMiddleware)
	r.GET("/users", m.UserList(), adminAuthMw.AdminAuthEchoMiddleware)
	r.POST("/users/block", m.BlockUser(), adminAuthMw.AdminAuthEchoMiddleware)
	r.POST("/users/unblock", m.UnblockUser(), adminAuthMw.AdminAuthEchoMiddleware)
	r.POST("/users/verify", m.VerifyUser(), adminAuthMw.AdminAuthEchoMiddleware)
}
