package v1

import (
	"net/http"

	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"

	"github.com/labstack/echo/v4"
)

// notFoundHandler для стандартной ошибки 404
func notFoundHandler(c echo.Context) error {
	return c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &Error{
			Code:    ErrCodePathNotFound,
			Message: ErrTitlePathNotFound,
		},
	})
}

// helloHandler тестовый хэндер для корня api
func helloHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{Success: true, Data: "hello dating"})
}

// healthHandler проверка здоровья сервиса
func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, struct {
		Alive bool `json:"alive"`
	}{true})
}
