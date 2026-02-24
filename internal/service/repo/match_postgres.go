package repo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/pkg/pg"

	"gorm.io/gorm"
)

type MatchRepoPg struct {
	db *gorm.DB
	l  *slog.Logger
}

func NewMatchRepoPg(db *gorm.DB, l *slog.Logger) service.MatchRepo {
	return &MatchRepoPg{
		db: db,
		l:  l,
	}
}

func (m *MatchRepoPg) Create(ctx context.Context, match *entity.Match) (err error) {
	defer pg.ProcessDbError(&err)

	err = m.db.WithContext(ctx).Create(match).Error

	if err != nil {
		return
	}

	return
}

func (m *MatchRepoPg) DeleteByID(ctx context.Context, matchID string) error {
	// todo
	return fmt.Errorf("not implemented")
}

func (m *MatchRepoPg) GetByID(ctx context.Context, id string) (match *entity.Match, err error) {
	defer pg.ProcessDbError(&err)

	match = &entity.Match{}

	err = m.db.WithContext(ctx).
		Model(&entity.Match{}).
		Preload("User1", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Photos", func(db *gorm.DB) *gorm.DB {
				return db.Order("is_main DESC")
			})
		}).
		Preload("User2", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Photos", func(db *gorm.DB) *gorm.DB {
				return db.Order("is_main DESC")
			})
		}).
		Where("id = ?", id).First(&match).Error

	return
}

func (m *MatchRepoPg) GetAllByUserID(ctx context.Context, userID string) (matches []*entity.Match, err error) {
	defer pg.ProcessDbError(&err)

	err = m.db.WithContext(ctx).
		Model(&entity.Match{}).
		Preload("User1", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Photos", func(db *gorm.DB) *gorm.DB {
				return db.Order("is_main DESC")
			})
		}).
		Preload("User2", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Photos", func(db *gorm.DB) *gorm.DB {
				return db.Order("is_main DESC")
			})
		}).
		Where("user_1_id = ? OR user_2_id = ?", userID, userID).
		Where("chat_init = false").
		Order("created_at DESC").
		Find(&matches).Error

	return
}
