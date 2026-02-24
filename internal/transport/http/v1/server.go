package v1

import (
	"net/http"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/internal/transport/http/middlewares"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	healthPath  = "/health"
	metricsPath = "/metrics"
)

func NewEchoServer(services *service.Services, l *slog.Logger, m *entity.Metrics, enableBodyDump bool) *echo.Echo {
	server := echo.New()

	// custom 404 error handler
	echo.NotFoundHandler = notFoundHandler

	r := server.Group("/api/v1")

	// init middlewares
	authMw := middlewares.NewAuthMiddleware(services.Profile, services.Auth, l)
	accessMw := middlewares.NewAccessMiddleware(l, m)
	bodyDumpMw := middlewares.NewBodyDumpMiddleware(l)

	if enableBodyDump {
		r.Use(bodyDumpMw.Handle())
	}

	// set global middlewares
	r.Use(accessMw.AccessEchoMiddleware())

	// test hello route
	r.GET("/", helloHandler)

	// technical routes
	r.GET(healthPath, healthHandler)
	r.GET(metricsPath, echo.WrapHandler(promhttp.Handler()))

	// auth
	setAuthRoutes(r, services.Profile, services.Auth, l)

	// profile
	profile := r.Group("/profile", authMw.AuthEchoMiddleware)
	setProfileRoutes(profile, services.Profile, l)

	// find
	find := r.Group("/find", authMw.AuthEchoMiddleware)
	setFindRoutes(find, l, services.Find)

	// like
	like := r.Group("/like", authMw.AuthEchoMiddleware)
	setLikeRoutes(like, l, services.Like)

	// matches
	match := r.Group("/match", authMw.AuthEchoMiddleware)
	setMatchRoutes(match, l, services.Match)

	// chat
	chat := r.Group("/chat", authMw.AuthEchoMiddleware)
	setChatRoutes(chat, l, services.Chat)

	// ws
	ws := r.Group("/ws", authMw.AuthEchoMiddleware)
	setWsHandlers(ws, l, services.Chat)

	// products
	products := r.Group("/products", authMw.AuthEchoMiddleware)
	setProductsRoutes(products, services.PaymentService, l)

	// payments (for callback)
	pays := r.Group("/payments")
	setPaymentsRoutes(pays, services.PaymentService, l)

	//admin
	admin := r.Group("/admin", middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	setAdminRoutes(admin, l, services.AdminAuth, services.AdminUsers, services.AdminProfile)

	return server
}
