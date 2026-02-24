package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log/slog"
	"net/http"
)

type BodyDumpMiddleware struct {
	l *slog.Logger
}

func NewBodyDumpMiddleware(l *slog.Logger) *BodyDumpMiddleware {
	return &BodyDumpMiddleware{l: l}
}

func (b *BodyDumpMiddleware) Handle() echo.MiddlewareFunc {
	return middleware.BodyDump(func(c echo.Context, reqBody []byte, respBody []byte) {
		if c.Request().Method == http.MethodPost {
			b.l.Info("request body dump", "body", string(reqBody), "headers", c.Request().Header)
		}
	})
}
