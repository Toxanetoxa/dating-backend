package v1

import (
	"errors"
	"github.com/toxanetoxa/dating-backend/pkg/pg"
	"net/http"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FindHandlersManager struct {
	l *slog.Logger
	s service.FindService
}

func NewFindHandlersManager(l *slog.Logger, s service.FindService) *FindHandlersManager {
	return &FindHandlersManager{
		l: l,
		s: s,
	}
}

// Find пролучить набор пар для поиска
func (f *FindHandlersManager) Find() echo.HandlerFunc {
	type request struct {
		Limit  int `json:"limit" validate:"required"`
		Filter struct {
			Sex     *entity.UserSex `json:"sex" validate:"omitempty,oneof=male female "`
			Radius  int             `json:"radius"`
			AgeFrom int             `json:"ageFrom" validate:"gte=18"`
			AgeTo   int             `json:"ageTo" validate:"gte=18"`
		} `json:"filter"`
	}

	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			f.l.Debug("can't bind request: %v", err)

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInvalidData,
					Message: ErrTitleInvalidData,
				},
			})
		}

		validate := validator.New()
		err = validate.Struct(requestData)
		if _, ok := err.(validator.ValidationErrors); ok {
			failedFields := map[string]string{}
			for _, err := range err.(validator.ValidationErrors) {
				failedFields[err.Field()] = err.Tag()
			}
			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeValidation,
					Message: ErrTitleValidation,
				},
				Data: failedFields,
			})
		}

		userID, _ := c.Get(UserIDContextKey).(string)

		couples, err := f.s.Find(c.Request().Context(), userID, entity.Filter{
			Radius:  requestData.Filter.Radius,
			AgeFrom: requestData.Filter.AgeFrom,
			AgeTo:   requestData.Filter.AgeTo,
			Sex:     requestData.Filter.Sex,
		}, requestData.Limit)
		if err != nil {
			if errors.Is(service.ErrUserLocationNotSet, err) {
				return c.JSON(http.StatusForbidden, Response{Success: false, Error: &Error{
					Code:    ErrCodeUserGeolocationNotSet,
					Message: ErrTitleUserGeolocationNotSet,
				}})
			}
			f.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: couples})
	}
}

// Like поставить лайк пользователю
func (f *FindHandlersManager) Like() echo.HandlerFunc {
	type request struct {
		UserID uuid.UUID `json:"userID" validate:"required"`
	}
	type response struct {
		Match   bool   `json:"match"`
		MatchID string `json:"matchID,omitempty"`
	}

	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			f.l.Debug("can't bind request: %v", err)

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInvalidData,
					Message: ErrTitleInvalidData,
				},
			})
		}

		validate := validator.New()
		err = validate.Struct(requestData)
		if _, ok := err.(validator.ValidationErrors); ok {
			failedFields := map[string]string{}
			for _, err := range err.(validator.ValidationErrors) {
				failedFields[err.Field()] = err.Tag()
			}

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeValidation,
					Message: ErrTitleValidation,
				},
				Data: failedFields,
			})
		}

		userID, _ := c.Get(UserIDContextKey).(string)

		match, id, err := f.s.Like(c.Request().Context(), userID, requestData.UserID.String())
		if err != nil {
			switch {
			case errors.Is(err, service.ErrMatchAlreadyExist):
				return c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Error: &Error{
						Code:    ErrCodeMatchAlreadyExists,
						Message: ErrTitleMatchAlreadyExists,
					},
				})
			case errors.Is(err, service.ErrLikeAlreadyExist):
				return c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Error: &Error{
						Code:    ErrCodeResourceAlreadyExists,
						Message: ErrTitleResourceAlreadyExists,
					},
				})
			case errors.Is(err, pg.ErrUnclassified):
				return c.JSON(http.StatusNotFound, Response{
					Success: false,
					Error: &Error{
						Code:    ErrCodeResourceDoesntExist,
						Message: ErrTitleResourceDoesntExist,
					},
				})
			}

			f.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: response{Match: match, MatchID: id}})
	}
}

// Dislike отправить дизлайк пользователю
func (f *FindHandlersManager) Dislike() echo.HandlerFunc {
	type request struct {
		UserID uuid.UUID `json:"userID" validate:"required"`
	}

	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			f.l.Debug("can't bind request: %v", err)

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInvalidData,
					Message: ErrTitleInvalidData,
				},
			})
		}

		validate := validator.New()
		err = validate.Struct(requestData)
		if _, ok := err.(validator.ValidationErrors); ok {
			failedFields := map[string]string{}
			for _, err := range err.(validator.ValidationErrors) {
				failedFields[err.Field()] = err.Tag()
			}
			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeValidation,
					Message: ErrTitleValidation,
				},
				Data: failedFields,
			})
		}

		userID, _ := c.Get(UserIDContextKey).(string)

		err = f.s.Dislike(c.Request().Context(), userID, requestData.UserID.String())
		if err != nil {
			f.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// ClearLikes for debug
func (f *FindHandlersManager) ClearLikes() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		err := f.s.ClearLikes(c.Request().Context(), userID)
		if err != nil {
			f.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

func setFindRoutes(r *echo.Group, l *slog.Logger, s service.FindService) {
	manager := NewFindHandlersManager(l, s)

	r.POST("", manager.Find())
	r.POST("/like", manager.Like())
	r.POST("/dislike", manager.Dislike())
	r.GET("/clear-likes", manager.ClearLikes()) // for debug
}
