package repo

import (
	"context"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/pkg/pg"
	"log/slog"

	"gorm.io/gorm"
)

type UserPhotoRepoPg struct {
	db *gorm.DB
	l  *slog.Logger
}

func NewUserPhotoRepo(db *gorm.DB, l *slog.Logger) service.UserPhotoRepo {
	return &UserPhotoRepoPg{
		db: db,
		l:  l,
	}
}

func (u *UserPhotoRepoPg) Create(ctx context.Context, photo *entity.UserPhoto, _ pg.Transaction) (tx pg.Transaction, err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Create(photo).Error

	return
}

func (u *UserPhotoRepoPg) Delete(ctx context.Context, id, userID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Where("id = ? and user_id = ?", id, userID).Delete(&entity.UserPhoto{}).Error

	return
}

func (u *UserPhotoRepoPg) DeleteAllByUserID(ctx context.Context, userID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.UserPhoto{}).Error

	return
}

func (u *UserPhotoRepoPg) SetMain(ctx context.Context, userID, photoID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Model(&entity.UserPhoto{}).Where("user_id = ?", userID).Update("is_main", false).Error
	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.UserPhoto{}).Where("id = ? and user_id = ?", photoID, userID).Update("is_main", true).Error

	return
}

func (u *UserPhotoRepoPg) GetAllByUserID(ctx context.Context, userID string) (list []*entity.UserPhoto, err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_main DESC").Find(&list).Error

	return
}

func (u *UserPhotoRepoPg) GetByID(ctx context.Context, id string) (photo *entity.UserPhoto, err error) {
	defer pg.ProcessDbError(&err)

	photo = &entity.UserPhoto{
		GeneralTechFields: entity.GeneralTechFields{
			ID: id,
		},
	}

	err = u.db.WithContext(ctx).First(photo).Error

	return
}
