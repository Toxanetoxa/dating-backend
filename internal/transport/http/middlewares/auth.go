package middlewares

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	l    *slog.Logger
	user service.ProfileService
	auth service.AuthService
}

func NewAuthMiddleware(user service.ProfileService, auth service.AuthService, l *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		l:    l,
		user: user,
		auth: auth,
	}
}

func (a *AuthMiddleware) AuthEchoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// read token
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, Response{Success: false, Error: &Error{
				Code:    ErrCodeUnauthorized,
				Message: ErrTitleUnauthorized,
			}})
		}

		// get jwt
		token, err := jwt.Parse(getToken(authHeader), func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			sub, err := token.Claims.GetSubject()
			if err != nil {
				return nil, fmt.Errorf("invalid subject")
			}

			secret, err := a.auth.GetKeyBySessionID(c.Request().Context(), sub)
			if err != nil {
				if errors.Is(service.ErrUserTokenKeyNotFound, err) {
					return nil, fmt.Errorf("invalid token")
				}
				a.l.Error(err.Error())

				return nil, fmt.Errorf("could not get user token")
			}

			// hmacSampleSecret is a []byte containing your secret
			return []byte(secret), nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return c.JSON(http.StatusUnauthorized, Response{Success: false, Error: &Error{
					Code:    ErrCodeTokenExpired,
					Message: ErrTitleTokenExpired,
				}})
			}
			a.l.Debug(err.Error())

			return c.JSON(http.StatusUnauthorized, Response{Success: false, Error: &Error{
				Code:    ErrCodeInvalidAuthData,
				Message: ErrTitleInvalidAuthData,
			}})
		}

		userID, _ := token.Claims.GetIssuer()
		sessionID, _ := token.Claims.GetSubject()

		user, err := a.user.GetByID(c.Request().Context(), userID) // cratch todo remove it ?
		if err != nil && !errors.Is(service.ErrProfileNotCompleted, err) {

			return c.JSON(http.StatusInternalServerError, Response{Success: false, Error: &Error{
				Code:    ErrCodeInternal,
				Message: ErrTitleInternal,
			}})
		}

		if user.Status == entity.UserStatusInactive {
			return c.JSON(http.StatusForbidden, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeUserIsBlocked,
					Message: ErrTitleUserIsBlocked,
					Details: struct {
						UserID string `json:"userID"`
					}{user.ID},
				},
			})
		}

		c.Set(UserIDContextKey, userID)
		c.Set(UserSessionIDContextKey, sessionID)

		return next(c)
	}
}

func getToken(header string) string {
	// cut "Bearer " substring
	if len(header) >= 7 {
		return header[7:]
	} else {
		return header
	}
}
