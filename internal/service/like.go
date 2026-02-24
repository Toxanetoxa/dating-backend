package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/pkg/geopoint"
)

type Like struct {
	l    *slog.Logger
	repo UserRepo
}

func NewLikeService(l *slog.Logger, repo UserRepo) LikeService {
	return &Like{
		l:    l,
		repo: repo,
	}
}

// List список ?? todo
func (l *Like) List(ctx context.Context, userID string) ([]*entity.Couple, error) {
	user, err := l.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	users, err := l.repo.GetLikes(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	couples := make([]*entity.Couple, 0, len(users))

	for i := range users {
		couples = append(couples, &entity.Couple{
			Distance: int(math.Round(geopoint.Distance(user.Geolocation.Lat, user.Geolocation.Lng, users[i].Geolocation.Lat, users[i].Geolocation.Lng, "K"))),
			Profile:  users[i],
		})
	}

	return couples, nil
}

func (l *Like) DeleteAllByUserID(ctx context.Context, userID string) error {
	// todo
	return fmt.Errorf("not implemented")
}
