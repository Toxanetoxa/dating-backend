package v1

import (
	"errors"
	"net/http"

	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"github.com/toxanetoxa/dating-backend/pkg/pg"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ChatHandlersManager struct {
	l *slog.Logger
	s service.ChatService
}

func NewChatHandlersManager(l *slog.Logger, s service.ChatService) *ChatHandlersManager {
	return &ChatHandlersManager{
		l: l,
		s: s,
	}
}

// List список активных чатов
func (m *ChatHandlersManager) List() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		list, err := m.s.ChatsList(c.Request().Context(), userID)
		if err != nil {
			if errors.Is(err, pg.ErrEntityDoesntExist) {
				return c.JSON(http.StatusNotFound, Error{Code: ErrCodeResourceDoesntExist, Message: ErrTitleResourceDoesntExist})
			}
			m.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: list})
	}
}

// Messages список сообщений в чате
func (m *ChatHandlersManager) Messages() echo.HandlerFunc {
	type request struct {
		MatchID string `json:"matchID" validate:"required,uuid"`
		Limit   int    `json:"limit" validate:"required"`
		Offset  int    `json:"offset"`
	}
	return func(c echo.Context) error {
		var reqData request
		err := c.Bind(&reqData)
		if err != nil {
			m.l.Error(err.Error())
			return c.JSON(http.StatusBadRequest, Response{Success: false})
		}

		validate := validator.New()
		err = validate.Struct(reqData)
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

		list, err := m.s.MessagesList(c.Request().Context(), userID, reqData.MatchID, reqData.Limit, reqData.Offset)
		if err != nil {
			if errors.Is(err, pg.ErrEntityDoesntExist) {
				return c.JSON(http.StatusNotFound, Error{Code: ErrCodeResourceDoesntExist, Message: ErrTitleResourceDoesntExist})
			}

			m.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: list})
	}
}

func (m *ChatHandlersManager) SendMessage() echo.HandlerFunc {
	type request struct {
		MatchID string `json:"matchID"`
		Text    string `json:"text"`
	}
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		var reqData request
		err := c.Bind(&reqData)
		if err != nil {
			m.l.Error(err.Error())
			return c.JSON(http.StatusBadRequest, Response{Success: false})
		}

		_, err = m.s.SendMessage(c.Request().Context(), userID, "", reqData.MatchID, reqData.Text)
		if err != nil {
			if errors.Is(err, service.ErrMatchNotFound) {
				return c.JSON(http.StatusNotFound, Response{Success: false, Error: &Error{Code: ErrCodeResourceDoesntExist, Message: ErrTitleResourceDoesntExist}})
			}

			// todo handle some errors
			m.l.Error("could not send message", "err", err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false})
		}

		return c.JSON(http.StatusOK, Response{Success: true})
	}
}

func setChatRoutes(g *echo.Group, l *slog.Logger, s service.ChatService) {
	m := NewChatHandlersManager(l, s)
	g.GET("", m.List())
	g.POST("/messages", m.Messages())
	g.POST("/send-message", m.SendMessage())
}
