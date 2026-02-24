package entity

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	RequestsTotal     prometheus.Counter
	Requests20x       prometheus.Counter
	Requests40x       prometheus.Counter
	Requests50x       prometheus.Counter
	LoginTotal        prometheus.Counter
	LogoutTotal       prometheus.Counter
	DeleteUserTotal   prometheus.Counter
	RegistrationTotal prometheus.Counter
	BuyBoostTotal     prometheus.Counter
	SuccessPayments   prometheus.Counter
	FailedPayments    prometheus.Counter
	CanceledPayments  prometheus.Counter
}

func NewMetrics() *Metrics {
	m := &Metrics{}

	m.RequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "number of rest api requests",
		},
	)

	m.Requests20x = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_20x",
			Help: "number of success requests",
		},
	)

	m.Requests40x = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_40x",
			Help: "number of bad requests",
		},
	)

	m.Requests50x = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_50x",
			Help: "number of internal server error requests",
		},
	)

	m.LoginTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "login_total",
			Help: "number of login requests",
		},
	)

	m.LogoutTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "logout_total",
			Help: "number of log out login requests",
		},
	)

	m.DeleteUserTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "user_delete_total",
			Help: "number of delete user requests",
		},
	)

	m.RegistrationTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "reg_total",
			Help: "number of registrations",
		},
	)

	m.BuyBoostTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "buy_boost_total",
			Help: "total boost purchases",
		},
	)

	m.SuccessPayments = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "payments_success",
			Help: "total success payments",
		},
	)

	m.FailedPayments = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "payments_failed",
			Help: "total failed to create payments",
		},
	)

	m.CanceledPayments = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "payments_canceled",
			Help: "total canceled payments",
		},
	)

	return m
}
