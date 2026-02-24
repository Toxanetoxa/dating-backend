package entity

import (
	"time"
)

type (
	PaymentStatus string

	Payment struct {
		GeneralTechFields
		ExternalID     string        // внешний идентификатор
		Status         PaymentStatus // статус платежа
		ExternalStatus string        // статус платежа на стороне партнёра
		Price          float64       // сумма платежа (цена)
		Currency       string        // валюта платежа
		PaidPrice      float64       // фактическая сумма оплаты (полученная от партнера)
		PaidCurrency   string        // фактическая валюта оплаты
		Description    string        // описание платежа (опционально)
		UserID         string        // id пользователя, покупающего продукт
		ProductID      int           // id покупаемого продукта
		UpdatedAt      time.Time     // последнее обновление данных платежа
	}
)

const (
	PaymentStatusNew      = "new"      // платеж создан на нашей стороне
	PaymentStatusPending  = "pending"  // платеж ожидает оплаты пользователем
	PaymentStatusError    = "error"    // ошибка при создании\обработке платежа на стороне партнера
	PaymentStatusPaid     = "paid"     // успешно оплачено
	PaymentStatusCanceled = "canceled" // платеж был отменен (со стороны партнера)
	PaymentStatusRefunded = "refunded" // произошёл возврат средств
)

func (p *Payment) TableName() string {
	return "payment"
}
