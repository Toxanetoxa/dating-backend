package service

import (
	"context"
	"fmt"
	"log/slog"
)

type RDBCodesStorage struct {
	l *slog.Logger
}

func NewRDBCodesStorage(l *slog.Logger) CodesStorage {
	return &RDBCodesStorage{l: l}
}

// AddCode ...
func (c *RDBCodesStorage) AddCode(ctx context.Context, phone, code string, ttl int) error {
	// todo: implement me

	return fmt.Errorf("not implemented")
}

// GetCodeByPhone ...
func (c *RDBCodesStorage) GetCodeByPhone(ctx context.Context, phone string) (code string, err error) {
	// todo: implement me

	return "", fmt.Errorf("not implementd")
}
