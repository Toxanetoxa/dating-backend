package dto

type (
	// Response базовая структура ответа
	Response struct {
		Success bool        `json:"success"`
		Error   *Error      `json:"error,omitempty"`
		Data    interface{} `json:"data,omitempty"`
	}

	// Error стандартная структура ошибки
	Error struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	}
)

const (
	UserIDContextKey        string = "UserID"
	UserSessionIDContextKey string = "SessionID"
)
