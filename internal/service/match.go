package service

import (
	"context"
	"log/slog"

	"github.com/toxanetoxa/dating-backend/internal/entity"
)

type Match struct {
	l    *slog.Logger
	repo MatchRepo
}

func NewMatchService(l *slog.Logger, r MatchRepo) MatchService {
	return &Match{
		l:    l,
		repo: r,
	}
}

// GetByUserID список мэтчей (пар) пользователя
func (m *Match) GetByUserID(ctx context.Context, userID string) ([]*entity.Match, error) {
	return m.repo.GetAllByUserID(ctx, userID)
}
