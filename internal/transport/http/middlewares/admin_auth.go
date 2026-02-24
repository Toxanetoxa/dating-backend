package middlewares

import (
	"errors"
	"net/http"

	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type AdminAuthMiddleware struct {
	l    *slog.Logger
	auth service.AdminAuthService
}

func NewAdminAuthMiddleware(auth service.AdminAuthService, l *slog.Logger) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		l:    l,
		auth: auth,
	}
}

func (a *AdminAuthMiddleware) AdminAuthEchoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// read token
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, Response{Success: false, Error: &Error{
				Code:    ErrCodeUnauthorized,
				Message: ErrTitleUnauthorized,
			}})
		}

		user, tokenID, err := a.auth.ValidateToken(c.Request().Context(), getToken(authHeader))
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidToken):
				return c.JSON(http.StatusUnauthorized, Response{
					Success: false,
					Error: &Error{
						Code:    ErrCodeInvalidAuthData,
						Message: ErrTitleInvalidAuthData,
					},
				})
			default:
				a.l.Error(err.Error())
				return c.NoContent(http.StatusInternalServerError)
			}
		}

		c.Set(UserIDContextKey, user.ID)
		c.Set(UserSessionIDContextKey, tokenID)

		return next(c)
	}
}
