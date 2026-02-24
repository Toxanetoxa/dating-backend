package service

import (
	"context"
	"fmt"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"log/slog"
)

type AdminUsers struct {
	l    *slog.Logger
	repo UserRepo
}

func NewAdminUsers(l *slog.Logger, r UserRepo) AdminUsersService {
	return &AdminUsers{
		l:    l,
		repo: r,
	}
}

func (a *AdminUsers) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	return a.repo.List(ctx, limit, offset)
}

func (a *AdminUsers) FindByID(ctx context.Context, userID string) (*entity.User, error) {
	return a.repo.GetByID(ctx, userID)
}

func (a *AdminUsers) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	// todo: implement
	return nil, fmt.Errorf("not implemented")
}

func (a *AdminUsers) Verify(ctx context.Context, userID string) error {
	// todo: implement
	return fmt.Errorf("not implemented")
}

func (a *AdminUsers) Block(ctx context.Context, userID string) error {
	return a.repo.Deactivate(ctx, userID)
}

func (a *AdminUsers) Unblock(ctx context.Context, userID string) error {
	return a.repo.Activate(ctx, userID)
}

func (a *AdminUsers) Stats(ctx context.Context) (usersStats entity.UsersStats, err error) {
	return a.repo.Stats(ctx)
}
