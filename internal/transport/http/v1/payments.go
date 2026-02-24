package v1

import (
	"net/http"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type PaymentsHandlersManager struct {
	l *slog.Logger
	s service.PaymentService
}

func NewPaymentsHandlersManager(l *slog.Logger, s service.PaymentService) *PaymentsHandlersManager {
	return &PaymentsHandlersManager{
		l: l,
		s: s,
	}
}

// PaymentCallback коллбэк (хук) для платежной системы (юкасса)
func (p *PaymentsHandlersManager) PaymentCallback() echo.HandlerFunc {
	type request struct {
		Type   string `json:"type"`
		Event  string `json:"event"`
		Object struct {
			Id     string `json:"id"`
			Status string `json:"status"`
			Amount struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"amount"`
			IncomeAmount struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"income_amount"`
			Description string `json:"description"`
			Recipient   struct {
				AccountId string `json:"account_id"`
				GatewayId string `json:"gateway_id"`
			} `json:"recipient"`
			PaymentMethod struct {
				Type          string `json:"type"`
				Id            string `json:"id"`
				Saved         bool   `json:"saved"`
				Title         string `json:"title"`
				AccountNumber string `json:"account_number"`
			} `json:"payment_method"`
			CapturedAt     time.Time `json:"captured_at"`
			CreatedAt      time.Time `json:"created_at"`
			Test           bool      `json:"test"`
			RefundedAmount struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"refunded_amount"`
			Paid       bool `json:"paid"`
			Refundable bool `json:"refundable"`
			Metadata   struct {
			} `json:"metadata"`
			MerchantCustomerId string `json:"merchant_customer_id"`
		} `json:"object"`
	}
	return func(c echo.Context) error {
		var reqData request
		err := c.Bind(&reqData)
		if err != nil {
			p.l.Debug("can't bind request: %v", err)

			return c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInvalidData,
					Message: ErrTitleInvalidData,
				},
			})
		}

		if reqData.Type == "notification" {

			err = p.s.PaymentCallback(c.Request().Context(), reqData.Event, reqData.Object)
			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}

		}

		return c.NoContent(http.StatusOK)
	}
}

func setPaymentsRoutes(g *echo.Group, s service.PaymentService, l *slog.Logger) {
	m := NewPaymentsHandlersManager(l, s)

	g.POST("/callback", m.PaymentCallback())
}
