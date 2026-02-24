package youkassa

import "time"

//const (
//	paymentStatusPending        = "pending"
//	paymentStatusSucceeded      = "succeeded"
//	paymentStatusCanceled       = "canceled"
//	paymentStatusWaitingCapture = "waiting_for_capture"
//
//	curRUB = "RUB"
//)

type (
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	}

	PaymentMethodData struct {
		Type string `json:"type"`
	}

	Confirmation struct {
		Type            string `json:"type"`
		ReturnUrl       string `json:"return_url"`
		ConfirmationUrl string `json:"confirmation_url"`
	}

	PaymentResponse struct {
		ID           string       `json:"id"`
		Status       string       `json:"status"`
		Paid         bool         `json:"paid"`
		Amount       Amount       `json:"amount"`
		CreatedAt    time.Time    `json:"created_at"`
		Description  string       `json:"description"`
		Confirmation Confirmation `json:"confirmation"`
	}

	PaymentRequest struct {
		Amount             Amount             `json:"amount"`                         // Сумма платежа, required
		PaymentMethodData  *PaymentMethodData `json:"payment_method_data,omitempty"`  // Данные для оплаты конкретным способом, optional
		Confirmation       *Confirmation      `json:"confirmation,omitempty"`         // Данные, необходимые для инициирования выбранного сценария подтверждения платежа пользователем, optional
		Capture            bool               `json:"capture,omitempty"`              // Автоматический прием поступившего платежа, optional
		Description        string             `json:"description,omitempty"`          // Описание транзакции (не более 128 символов), optional
		MerchantCustomerID string             `json:"merchant_customer_id,omitempty"` // Идентификатор покупателя в вашей системе, optional
		SavePaymentMethod  bool               `json:"save_payment_method,omitempty"`  // Сохранение платежных данных (с их помощью можно проводить повторные безакцептные списания ), optional
		Receipt            Receipt            `json:"receipt"`                        // Данные для формирования чека
	}

	Receipt struct {
		Customer Customer `json:"customer"` // Информация о пользователе
		Items    []Item   `json:"items"`    // Список товаров в заказе *
	}

	Customer struct {
		Email string `json:"email,omitempty"` // минимально необходимое поле для отправки чеков
		Phone string `json:"phone,omitempty"` // Телефон пользователя для отправки чека. Указывается в формате ITU-T E.164, например 79000000000. Обязательный параметр, если не передан email.
	}

	Item struct {
		Description string `json:"description"` // Название товара
		Amount      Amount `json:"amount"`      // цена товара
		VatCode     int    `json:"vat_code"`    // Ставка НДС (тег в 54 ФЗ — 1199)
		Quantity    string `json:"quantity"`    // Количество товара (тег в 54 ФЗ — 1023)
	}

	YPayment struct {
		// InternalID  string  // внутренний id платежа
		Amount      float64 // сумма платежа (цена)
		Currency    string  // валюта платежа
		ReturnUrl   string  // ...
		Description string
		CustomerID  string // id пользователя (покупателя) на нашей стороне
		UserEmail   string
		UserPhone   string
	}
)

type NotificationObject struct {
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
}
