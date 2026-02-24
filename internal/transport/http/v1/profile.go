package v1

import (
	"errors"
	"net/http"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"log/slog"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ProfileHandlersManager struct {
	l *slog.Logger
	s service.ProfileService
}

func NewUserHandlersManager(s service.ProfileService, l *slog.Logger) *ProfileHandlersManager {
	return &ProfileHandlersManager{
		l: l,
		s: s,
	}
}

// GetProfile получение данных своего профиля
func (u *ProfileHandlersManager) GetProfile() echo.HandlerFunc {
	type response struct {
		UserID   string             `json:"ID"`
		Name     *string            `json:"name"`
		Sex      *entity.UserSex    `json:"sex"`
		Birthday *entity.BirthDate  `json:"birthday"`
		City     *string            `json:"city"`
		About    *string            `json:"about"`
		Photos   []entity.UserPhoto `json:"photos"`
	}
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		user, err := u.s.GetByID(c.Request().Context(), userID)
		// todo нужно сюда вынести проверку на not complete
		if err != nil {
			if errors.Is(err, service.ErrProfileNotCompleted) {
				return c.JSON(http.StatusForbidden, Response{Success: false, Error: &Error{
					Code:    ErrCodeProfileContCompleted,
					Message: ErrTitleProfileNotCompleted,
				}})
			}
			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: response{
			UserID:   user.ID,
			Name:     user.FirstName,
			Sex:      user.Sex,
			Birthday: user.Birthday,
			City:     user.City,
			About:    user.About,
			Photos:   user.Photos,
		}})
	}
}

// UpdateProfile обновление данных своего профиля
func (u *ProfileHandlersManager) UpdateProfile() echo.HandlerFunc {
	type request struct {
		Name     string           `json:"name"`
		Sex      entity.UserSex   `json:"sex"`
		Birthday entity.BirthDate `json:"birthday"`
		City     string           `json:"city"`
		About    string           `json:"about"`
	}

	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		var reqData request
		err := c.Bind(&reqData)
		if err != nil {
			u.l.Debug("can't bind request: %v", err)

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInvalidData,
					Message: ErrTitleInvalidData,
				},
			})
		}

		err = u.s.UpdateProfile(c.Request().Context(), service.UserUpdateProfileParams{
			Name:     reqData.Name,
			Sex:      reqData.Sex,
			Birthday: reqData.Birthday,
			City:     reqData.City,
			About:    reqData.About,
		}, userID)

		if err != nil {
			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// DeleteUser удаление анкеты
func (u *ProfileHandlersManager) DeleteUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		err := u.s.DeleteProfile(c.Request().Context(), userID)
		if err != nil {
			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// UploadPhoto загрузка фото
func (u *ProfileHandlersManager) UploadPhoto() echo.HandlerFunc {
	type response struct {
		ID  string `json:"ID"`
		URL string `json:"URL"`
	}

	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		fileHeader, err := c.FormFile("photo")
		if err != nil {
			u.l.Error(err.Error())

			return c.JSON(http.StatusBadRequest, Response{Success: false})
		}

		id, url, err := u.s.UploadPhoto(c.Request().Context(), userID, fileHeader)
		if err != nil {
			if errors.Is(err, service.ErrInvalidFileType) {
				u.l.Debug(err.Error())

				return c.JSON(http.StatusBadRequest, Response{Success: false, Error: &Error{
					Code:    ErrCodeInvalidFileType,
					Message: ErrTitleInvalidFileType,
				}})
			}
			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: response{
			ID:  id,
			URL: url,
		}})
	}
}

func (u *ProfileHandlersManager) GetPhotos() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		photos, err := u.s.GetPhotosByUserID(c.Request().Context(), userID)
		if err != nil {
			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: photos})
	}
}

// DeletePhoto удаление фото
func (u *ProfileHandlersManager) DeletePhoto() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		photoID := c.Param("photo_id")

		err := u.s.DeletePhoto(c.Request().Context(), photoID, userID)
		if err != nil {
			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// SetMainPhoto установить фото как основное (аватарка)
func (u *ProfileHandlersManager) SetMainPhoto() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		photoID := c.Param("photo_id")

		err := u.s.SetMainPhoto(c.Request().Context(), userID, photoID)

		if err != nil {
			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// UpdateGeo обновление геопозиции пользователя
func (u *ProfileHandlersManager) UpdateGeo() echo.HandlerFunc {
	type request struct {
		Lat  float64 `json:"lat"`
		Long float64 `json:"long"`
	}

	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		var reqData request
		err := c.Bind(&reqData)
		if err != nil {
			u.l.Debug("can't bind request: %v", err)

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInvalidData,
					Message: ErrTitleInvalidData,
				},
			})
		}

		err = u.s.UpdateGeo(c.Request().Context(), userID, reqData.Lat, reqData.Long)
		if err != nil {
			// todo math and handle ?
			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// UpdateDeviceID добавить или обновить device id
func (u *ProfileHandlersManager) UpdateDeviceID() echo.HandlerFunc {
	type request struct {
		DeviceID string `json:"deviceID"`
	}
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)
		sessionID, _ := c.Get(UserSessionIDContextKey).(string)

		var reqData request
		err := c.Bind(&reqData)
		if err != nil {
			u.l.Debug("can't bind request: %v", err)

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInvalidData,
					Message: ErrTitleInvalidData,
				},
			})
		}

		err = u.s.UpdateDeviceID(c.Request().Context(), reqData.DeviceID, userID, sessionID)
		if err != nil {
			u.l.Error("could not update device id: %v", err)

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// GetUserByID получить профиль пользователя по uuid
func (u *ProfileHandlersManager) GetUserByID() echo.HandlerFunc {
	type response struct {
		entity.User
		Distance int `json:"distance"`
	}
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		targetID := c.Param("id")

		user, err := u.s.GetProfileByID(c.Request().Context(), userID, targetID)
		if err != nil {
			// check 404
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, Response{Success: false, Error: &Error{
					Code:    ErrCodeResourceDoesntExist,
					Message: ErrTitleResourceDoesntExist,
				}})
			}

			// match check
			if errors.Is(err, service.ErrNotMatch) {

				return c.JSON(http.StatusForbidden, Response{Success: false, Error: &Error{
					Code:    ErrCodeNotMatch,
					Message: ErrTitleNotMatch,
				}})
			}

			u.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		//get distance
		distance, err := u.s.GetDistanceToUser(c.Request().Context(), userID, targetID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Response{Success: false, Error: &Error{Code: 0, Message: "wer"}})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: response{
			Distance: distance,
			User:     *user,
		}})
	}
}

// Logout выход со всех устройств
func (u *ProfileHandlersManager) Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID, _ := c.Get(UserSessionIDContextKey).(string)

		err := u.s.LogOut(c.Request().Context(), sessionID)
		if err != nil {
			// todo handle errors
			u.l.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// setProfileRoutes устанавливает хэндлеры
func setProfileRoutes(r *echo.Group, s service.ProfileService, l *slog.Logger) {
	m := NewUserHandlersManager(s, l)

	r.GET("/my", m.GetProfile())
	r.POST("/my", m.UpdateProfile())
	r.DELETE("/my", m.DeleteUser())
	r.GET("/my/photos", m.GetPhotos())
	r.POST("/my/photos", m.UploadPhoto())
	r.DELETE("/my/photos/:photo_id", m.DeletePhoto())
	r.GET("/my/photos/:photo_id/set-main", m.SetMainPhoto())
	r.POST("/my/geo", m.UpdateGeo())
	r.GET("/:id", m.GetUserByID())
	r.POST("/set-device-id", m.UpdateDeviceID())
	r.POST("/logout", m.Logout())
}
