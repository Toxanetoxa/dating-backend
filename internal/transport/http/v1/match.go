package v1

import (
	"github.com/toxanetoxa/dating-backend/internal/entity"
	"net/http"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type MatchHandlersManager struct {
	l *slog.Logger
	s service.MatchService
}

func NewMatchHandlersManager(l *slog.Logger, s service.MatchService) *MatchHandlersManager {
	return &MatchHandlersManager{
		l: l,
		s: s,
	}
}

// List handler for list of matches (without chat init)
func (m *MatchHandlersManager) List() echo.HandlerFunc {
	type responseElem struct {
		ID        string       `json:"ID"` // match id
		Profile   *entity.User `json:"profile"`
		CreatedAt time.Time    `json:"createdAt"`
	}
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		matches, err := m.s.GetByUserID(c.Request().Context(), userID)
		if err != nil {
			m.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		// process matches for response
		var data = make([]responseElem, 0)

		for i := range matches {
			if matches[i].ChatInit {
				continue // не возвращаем тех, с кем есть чат
			}

			profile := matches[i].User2
			if profile.ID == userID {
				profile = matches[i].User1
			}

			data = append(data, responseElem{
				ID:        matches[i].ID,
				Profile:   profile,
				CreatedAt: matches[i].CreatedAt,
			})
		}

		return c.JSON(http.StatusOK, Response{
			Success: true,
			Data:    data,
		})
	}
}

func setMatchRoutes(r *echo.Group, l *slog.Logger, s service.MatchService) {
	m := NewMatchHandlersManager(l, s)

	r.GET("", m.List())
}
