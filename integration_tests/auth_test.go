package integration_tests

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/stretchr/testify/assert"
)

func TestVKAuth(t *testing.T) {
	// example test (sunny)
	t.Run("sunny case", func(t *testing.T) {

		// проверяем что есть ошибка отсутствия авторизации
		err := Do(
			Get(basePath+"/profile/my"),
			Send().Headers("Content-Type").Add("application/json"),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
		assert.NoErrorf(t, err, "err must be nil")

		authRequestBody := `{
			"accessToken": "test-token"
		}`

		var token string
		// делаем авторизацию через вк и сохраняем полученный токен
		err = Do(
			Post(basePath+"/auth/vk"),
			Send().Headers("Content-Type").Add("application/json"),
			Send().Body().String(authRequestBody),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".success").Equal(true),
			Store().Response().Body().JSON().JQ(".data.token").In(&token),
		)
		assert.NoErrorf(t, err, "err must be nil")

		// проверить что авторизация прошла и получили профиль
		err = Do(
			Get(basePath+"/profile/my"),
			Send().Headers("Content-Type").Add("application/json"),
			Send().Headers("Authorization").Add("Bearer "+token),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".success").Equal(true),
			Expect().Body().JSON().JQ(".data.name").Equal("Тестер"),
		)
		assert.NoErrorf(t, err, "err must be nil")

	})

}
