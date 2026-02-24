package v1

import (
	"net/http"

	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type LikeHandlersManager struct {
	l *slog.Logger
	s service.LikeService
}

func NewLikeHandlersManager(l *slog.Logger, s service.LikeService) *LikeHandlersManager {
	return &LikeHandlersManager{
		l: l,
		s: s,
	}
}

func (l *LikeHandlersManager) List() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		couples, err := l.s.List(c.Request().Context(), userID)
		if err != nil {
			l.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: couples})
	}
}

func setLikeRoutes(r *echo.Group, l *slog.Logger, s service.LikeService) {
	manager := NewLikeHandlersManager(l, s)

	r.GET("", manager.List())
}
