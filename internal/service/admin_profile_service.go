package service

import (
	"context"
	"log/slog"

	"github.com/toxanetoxa/dating-backend/internal/entity"
)

type AdminProfile struct {
	l    *slog.Logger
	repo AdminRepo
}

func NewAdminProfileService(l *slog.Logger, r AdminRepo) AdminProfileService {
	return &AdminProfile{
		l:    l,
		repo: r,
	}
}

// GetByID получить админа по его id
func (a *AdminProfile) GetByID(ctx context.Context, id string) (*entity.Admin, error) {
	return a.repo.GetByID(ctx, id)
}

// LogOut удалит токен в бд
func (a *AdminProfile) LogOut(ctx context.Context, tokenID string) error {
	return a.repo.DeleteTokenByID(ctx, tokenID)
}

// UpdatePassword ...
func (a *AdminProfile) UpdatePassword(ctx context.Context) error {
	return nil
}
