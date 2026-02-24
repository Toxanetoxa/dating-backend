package service

import (
	"context"
	"fmt"
	"log/slog"
)

// пока не используется. возможно, в будущем будет исопльзоваться.

type RDBTokenStorage struct {
	l *slog.Logger
}

func NewRDBTokenStorage(l *slog.Logger) TokenStorage {
	return &RDBTokenStorage{l: l}
}

// AddToken ...
func (t *RDBTokenStorage) AddToken(ctx context.Context, userUUID, key string, ttl int) error {
	// todo: implement me

	return fmt.Errorf("not implemented")
}

// GetTokenByUserID ...
func (t *RDBTokenStorage) GetTokenByUserID(ctx context.Context, userUUID string) (key string, err error) {
	// todo: implement me

	return "", fmt.Errorf("not implemented")
}
