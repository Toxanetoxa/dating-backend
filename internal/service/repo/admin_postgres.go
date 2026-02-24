package repo

import (
	"context"
	"log/slog"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/pkg/pg"

	"gorm.io/gorm"
)

type AdminRepoPg struct {
	db *gorm.DB
	l  *slog.Logger
}

func NewAdminRepoPg(db *gorm.DB, l *slog.Logger) service.AdminRepo {
	return &AdminRepoPg{
		db: db,
		l:  l,
	}
}

func (a *AdminRepoPg) FindByLogin(ctx context.Context, login string) (admin *entity.Admin, err error) {
	defer pg.ProcessDbError(&err)

	err = a.db.WithContext(ctx).Model(&entity.Admin{}).Where("login = ?", login).First(&admin).Error

	return
}

func (a *AdminRepoPg) GetByID(ctx context.Context, id string) (admin *entity.Admin, err error) {
	defer pg.ProcessDbError(&err)

	err = a.db.WithContext(ctx).Model(&entity.Admin{}).Where("id = ?", id).First(&admin).Error

	return
}

func (a *AdminRepoPg) SaveToken(ctx context.Context, token *entity.AdminToken) (err error) {
	defer pg.ProcessDbError(&err)

	err = a.db.WithContext(ctx).Model(&entity.AdminToken{}).Create(&token).Error

	return
}

func (a *AdminRepoPg) GetToken(ctx context.Context, adminID, tokenID string) (token *entity.AdminToken, err error) {
	defer pg.ProcessDbError(&err)

	err = a.db.WithContext(ctx).Model(&entity.AdminToken{}).Where("id = $1 and admin_id = $2", tokenID, adminID).First(&token).Error

	return
}

func (a *AdminRepoPg) DeleteTokenByID(ctx context.Context, tokenID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = a.db.WithContext(ctx).Where("id = ?", tokenID).Delete(&entity.AdminToken{}).Error

	return
}
