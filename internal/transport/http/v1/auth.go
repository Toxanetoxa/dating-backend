package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type AuthHandlersManager struct {
	l     *slog.Logger
	users service.ProfileService
	auth  service.AuthService
}

func NewAuthHandlersManager(u service.ProfileService, a service.AuthService, l *slog.Logger) *AuthHandlersManager {
	return &AuthHandlersManager{
		users: u,
		auth:  a,
		l:     l,
	}
}

// RequestCode запрос одноразового кода для входа\регистрации
func (a *AuthHandlersManager) RequestCode() echo.HandlerFunc {
	type request struct {
		Email string `json:"email" validate:"required,email"`
		Hash  string `json:"hash" validate:"required"`
	}

	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			a.l.Debug("can't bind request: %v", err)

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

		email := strings.ToLower(requestData.Email)

		err = a.auth.RequestCode(c.Request().Context(), email, requestData.Hash)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidHash):
				return c.JSON(http.StatusForbidden, Response{Success: false, Error: &Error{
					Code:    ErrCodeInvalidHash,
					Message: ErrTitleInvalidHash,
				}})

			case errors.Is(err, service.ErrCodeAlreadySent):
				return c.JSON(http.StatusTooManyRequests, Response{Success: false, Error: &Error{
					Code:    ErrCodeCodeAlreadySent,
					Message: ErrTitleCodeAlreadySent,
				}})

			default:
				a.l.Error(err.Error())

				return c.JSON(http.StatusInternalServerError, Response{Success: false, Error: &Error{
					Code:    ErrCodeInternal,
					Message: ErrTitleInternal,
				}})
			}
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

// Login вход по номеру и коду, с авторегистрацией, если пользователя небыло
func (a *AuthHandlersManager) Login() echo.HandlerFunc {
	type (
		request struct {
			Email string `json:"email" validate:"required,email"`
			Code  string `json:"code" validate:"required"`
		}
		response struct {
			Token     string `json:"token"`
			Refresh   string `json:"refreshToken"`
			Expire    int    `json:"expire"`
			IsNewUser bool   `json:"isNewUser"`
		}
	)

	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			a.l.Debug("can't bind request: %v", err)

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

		email := strings.ToLower(requestData.Email)

		accessToken, refreshToken, tokenExpire, isNewUser, userID, err := a.auth.LogIn(c.Request().Context(), email, requestData.Code)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidVerificationCode):
				return c.JSON(http.StatusUnauthorized, Response{Success: false, Error: &Error{
					Code:    ErrCodeInvalidVerificationCode,
					Message: ErrTitleInvalidVerificationCode,
				}})

			case errors.Is(err, service.ErrUserIsBlocked):
				return c.JSON(http.StatusForbidden, Response{Success: false,
					Error: &Error{
						Code:    ErrCodeUserIsBlocked,
						Message: ErrTitleUserIsBlocked,
						Details: struct {
							UserID string `json:"userID"`
						}{userID},
					}})

			default:
				a.l.Error(err.Error())

				return c.JSON(http.StatusInternalServerError, Response{Success: false})
			}
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: response{
			Token:     accessToken,
			Refresh:   refreshToken,
			Expire:    tokenExpire,
			IsNewUser: isNewUser,
		}})
	}
}

// RefreshToken обновление токенов
func (a *AuthHandlersManager) RefreshToken() echo.HandlerFunc {
	type (
		request struct {
			RefreshToken string `json:"refreshToken" validate:"required"`
		}
		response struct {
			Token   string `json:"token"`
			Refresh string `json:"refreshToken"`
			Expire  int    `json:"expire"`
		}
	)
	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			a.l.Debug("can't bind request: %v", err)

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

		accessToken, refreshToken, expire, err := a.auth.RefreshToken(c.Request().Context(), requestData.RefreshToken)
		if err != nil {
			// todo handle internal errors
			a.l.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: response{
			Token:   accessToken,
			Refresh: refreshToken,
			Expire:  expire,
		}})
	}
}

// AuthByVK вход через ВК
func (a *AuthHandlersManager) AuthByVK() echo.HandlerFunc {
	type request struct {
		AccessToken string `json:"accessToken" validate:"required"`
	}
	type response struct {
		Token     string `json:"token"`
		Refresh   string `json:"refreshToken"`
		Expire    int    `json:"expire"`
		IsNewUser bool   `json:"isNewUser"`
	}
	return func(c echo.Context) error {
		var requestData request
		err := c.Bind(&requestData)
		if err != nil {
			a.l.Debug("can't bind request: %v", err)

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInvalidData,
					Message: ErrTitleInvalidData,
				},
			})
		}

		// на входе получаем access token и вызываем юзкейс с ним для авторизации

		jwtAccessToken, refreshToken, tokenExpire, isNewUser, _, err := a.auth.LoginByVK(c.Request().Context(), requestData.AccessToken)
		if err != nil {
			a.l.Error(err.Error())

			// todo handle some different errors

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
			Data: response{
				Token:     jwtAccessToken,
				Refresh:   refreshToken,
				Expire:    tokenExpire,
				IsNewUser: isNewUser,
			},
		})
	}
}

func setAuthRoutes(r *echo.Group, p service.ProfileService, a service.AuthService, l *slog.Logger) {
	m := NewAuthHandlersManager(p, a, l)

	r.POST("/auth/request-code", m.RequestCode()) // rate limit middleware needed
	r.POST("/auth/login", m.Login())              // rate limit middleware needed
	r.POST("/auth/refresh-token", m.RefreshToken())
	r.POST("/auth/vk", m.AuthByVK())
}
