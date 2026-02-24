package youkassa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

const httpClientTimeout = time.Minute

type Client struct {
	httpClient *http.Client
	baseUrl    string
	shopID     string
	secretKey  string
}

func NewYouKassaClient(baseUrl, shopID, secretKey string) YClient {
	// init http client
	h := &http.Client{
		Timeout: httpClientTimeout,
	}

	return &Client{
		httpClient: h,
		baseUrl:    baseUrl,
		shopID:     shopID,
		secretKey:  secretKey,
	}
}

func (c *Client) CreatePayment(_ context.Context, payment *YPayment) (string, string, string, error) {
	// help: https://yookassa.ru/developers/api?codeLang=bash#create_payment

	// prepare customer
	customer := Customer{}

	if len(payment.UserEmail) == 0 {
		customer.Phone = payment.UserPhone
	} else {
		customer.Email = payment.UserEmail
	}

	// create request
	data := PaymentRequest{
		Amount: Amount{
			Value:    fmt.Sprintf("%.2f", payment.Amount),
			Currency: payment.Currency,
		},
		Confirmation: &Confirmation{
			Type:      "redirect",
			ReturnUrl: payment.ReturnUrl,
		},
		Description:        payment.Description,
		MerchantCustomerID: payment.CustomerID,
		Capture:            true,
		//PaymentMethodData: &PaymentMethodData{
		//	Type: "bank_card",
		//},
		Receipt: Receipt{
			Customer: customer,
			Items: []Item{
				{
					Description: payment.Description,
					Amount: Amount{
						Value:    fmt.Sprintf("%.2f", payment.Amount),
						Currency: payment.Currency,
					},
					VatCode:  1,
					Quantity: "1",
				},
			},
		},
	}

	reqBody, err := json.Marshal(data)
	if err != nil {
		return "", "", "", err
	}

	req, err := http.NewRequest("POST", c.baseUrl+"payments", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", "", "", err
	}

	req.SetBasicAuth(c.shopID, c.secretKey)
	req.Header.Set("Idempotence-Key", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return "", "", "", err
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("response code is not 200: %s", bodyBytes)
	}

	payResp := PaymentResponse{}

	err = json.Unmarshal(bodyBytes, &payResp)
	if err != nil {
		return "", "", "", err
	}

	return payResp.Status, payResp.ID, payResp.Confirmation.ConfirmationUrl, nil
}

func (c *Client) GetPayment(_ context.Context, paymentID string) (*PaymentResponse, error) {
	reqUrl := c.baseUrl + "payments/" + paymentID
	r, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	r.SetBasicAuth(c.shopID, c.secretKey)

	resp, err := c.httpClient.Do(r)
	if err != nil {

		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, _ := io.ReadAll(resp.Body)

	payResp := PaymentResponse{}

	err = json.Unmarshal(bodyBytes, &payResp)
	if err != nil {
		return nil, err
	}

	return &payResp, nil
}
