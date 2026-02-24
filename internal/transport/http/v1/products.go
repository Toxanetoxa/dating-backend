package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/service"
	. "github.com/toxanetoxa/dating-backend/internal/transport/http/dto"
	"log/slog"

	"github.com/labstack/echo/v4"
)

type ProductsHandlersManager struct {
	l *slog.Logger
	s service.PaymentService
}

func NewProductsHandlersManager(l *slog.Logger, s service.PaymentService) *ProductsHandlersManager {
	return &ProductsHandlersManager{
		l: l,
		s: s,
	}
}

// GetBoostInfo получить информацию о бусте (цена, срок действия, акция)
func (p *ProductsHandlersManager) GetBoostInfo() echo.HandlerFunc {
	type response struct {
		ID       string `json:"ID"`
		Price    string `json:"price"`
		OldPrice string `json:"oldPrice"`
		Currency string `json:"currency"`
		Discount string `json:"discount"`
		Validity int64  `json:"validity"`
		Name     string `json:"name"`
	}
	return func(c echo.Context) error {
		product, err := p.s.GetProductBoost(c.Request().Context())
		if err != nil {
			p.l.Error(err.Error())

			return c.JSON(http.StatusInternalServerError, Response{Success: false,
				Error: &Error{Code: ErrCodeInternal, Message: ErrTitleInternal}})
		}

		// calc discount
		discount := 100 - ((product.Price * 100) / product.OldPrice)

		// map response
		resp := response{
			ID:       fmt.Sprintf("%d", product.ID),
			Price:    strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", product.Price), "0"), "."),
			OldPrice: strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", product.OldPrice), "0"), "."),
			Currency: product.Currency,
			Discount: strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", discount), "0"), "."),
			Validity: product.Validity,
			Name:     product.Name,
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: resp})
	}
}

// BuyBoost покупка буста
func (p *ProductsHandlersManager) BuyBoost() echo.HandlerFunc {
	type response struct {
		Link    string     `json:"link"`
		Expired *time.Time `json:"expired"`
	}
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		link, exp, err := p.s.BuyBoost(c.Request().Context(), userID)
		if err != nil {
			if errors.Is(err, service.ErrBoostAlreadyExist) {
				// если буст уже есть
				return c.JSON(http.StatusOK, Response{
					Success: false,
					Error: &Error{
						Code:    ErrCodeBoostAlreadyExists,
						Message: ErrTitleBoostAlreadyExists,
					},
				})
			}

			// unexpected error
			return c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Error: &Error{
					Code:    ErrCodeInternal,
					Message: ErrTitleInternal,
				},
			})
		}

		return c.JSON(http.StatusOK, Response{
			Success: true,
			Data: response{
				Link:    link,
				Expired: exp,
			},
		})
	}
}

// ListMy список покупок пользователя (активированные услуги)
func (p *ProductsHandlersManager) ListMy() echo.HandlerFunc {
	type response struct {
		ProductID int       `json:"productID"`
		Expire    time.Time `json:"expire"`
	}
	return func(c echo.Context) error {
		userID, _ := c.Get(UserIDContextKey).(string)

		product, err := p.s.GetUserBoost(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Response{Success: false, Error: &Error{
				Code:    ErrCodeInternal,
				Message: ErrTitleInternal,
			},
			})
		}

		if product == nil {
			// буст не куплен

			return c.JSON(http.StatusOK, Response{Success: true, Data: []response{}})
		}

		return c.JSON(http.StatusOK, Response{Success: true, Data: []response{
			{
				ProductID: product.ProductID,
				Expire:    product.Expire,
			},
		}})
	}
}

func setProductsRoutes(g *echo.Group, s service.PaymentService, l *slog.Logger) {
	m := NewProductsHandlersManager(l, s)

	g.GET("/boost", m.GetBoostInfo())
	g.GET("/buy-boost", m.BuyBoost())
	g.GET("/my", m.ListMy())
}
