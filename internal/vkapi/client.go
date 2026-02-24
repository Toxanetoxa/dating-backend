package vkapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	Client interface {
		UserInfo(ctx context.Context, accessToken string) (*ProfileInfo, error)
	}

	VKClient struct {
		c       *http.Client
		baseUrl string
		appId   string
	}

	ProfileInfo struct {
		ID        string `json:"user_id"`    // id пользователя
		FirstName string `json:"first_name"` // Имя пользователя
		LastName  string `json:"last_name"`  // Фамилия пользователя
		Sex       int    `json:"sex"`        // Пол. Возможные значения: 1 — женский,2 — мужской,0 — пол не указан
		Birthday  string `json:"birthday"`   // Дата рождения пользователя, возвращается в формате 29.07.1999
		Email     string `json:"email"`      // email если указан
		Phone     string `json:"phone"`      // номер телефона
		Verified  bool   `json:"verified"`   // статус верификации

		// есть avatar, но не используем
	}
)

const (
	httpClientTimeout = time.Minute
)

var (
	ErrBadApiResponse = errors.New("api response is not 200")
)

func NewVKAPIClient(baseUrl string, appId string) Client {
	// init http client
	c := &http.Client{
		Timeout: httpClientTimeout,
	}

	return &VKClient{
		c:       c,
		baseUrl: baseUrl,
		appId:   appId,
	}
}

// UserInfo получение информации о пользователе (user_info)
func (v *VKClient) UserInfo(_ context.Context, accessToken string) (*ProfileInfo, error) {
	// doc: https://id.vk.com/about/business/go/docs/ru/vkid/latest/vk-id/connection/api-integration/api-description

	var endpoint = v.baseUrl + "/oauth2/user_info"

	data := url.Values{}
	data.Set("client_id", v.appId)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := v.c.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrBadApiResponse
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	profileInfoResp := struct {
		User ProfileInfo `json:"user"`
	}{}

	err = json.Unmarshal(bodyBytes, &profileInfoResp)
	if err != nil {
		return nil, err
	}

	return &profileInfoResp.User, nil
}
