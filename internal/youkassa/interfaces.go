package youkassa

import "context"

// YClient клиент для запросов к юкассе
type YClient interface {
	CreatePayment(ctx context.Context, payment *YPayment) (status string, paymentID string, confirmationUrl string, err error)
	GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
}
