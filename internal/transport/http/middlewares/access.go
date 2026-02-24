package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/toxanetoxa/dating-backend/internal/entity"
	"log/slog"
)

type AccessMiddleware struct {
	l       *slog.Logger
	metrics *entity.Metrics
}

func NewAccessMiddleware(l *slog.Logger, m *entity.Metrics) *AccessMiddleware {
	return &AccessMiddleware{l: l, metrics: m}
}

func (m *AccessMiddleware) AccessEchoMiddleware() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogMethod:    true,
		LogError:     true,
		LogUserAgent: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			if values.URI == "/api/v1/metrics" {
				// skip metrics

				return nil
			}
			m.metrics.RequestsTotal.Inc()
			switch {
			case values.Status >= 200 && values.Status < 300:
				m.metrics.Requests20x.Inc()
			case values.Status >= 400 && values.Status < 500:
				m.metrics.Requests40x.Inc()
			case values.Status >= 500:
				m.metrics.Requests50x.Inc()
			}

			m.l.Info("request",
				"URI", values.URI,
				"method", values.Method,
				"status", values.Status,
				"latency", values.Latency.Milliseconds(),
				"user_agent", values.UserAgent,
				"error", values.Error)

			return nil
		},
	})
}
