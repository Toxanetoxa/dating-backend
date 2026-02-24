package repo

import (
	"context"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/pkg/pg"
	"log/slog"

	"gorm.io/gorm"
)

type PaymentsRepoPg struct {
	db *gorm.DB
	l  *slog.Logger
}

func NewPaymentsRepo(db *gorm.DB, l *slog.Logger) service.PaymentsRepo {
	return &PaymentsRepoPg{
		db: db,
		l:  l,
	}
}

func (p *PaymentsRepoPg) GetProduct(ctx context.Context, productID int) (product *entity.Product, err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", productID).First(&product).Error

	return
}

// GetUserProductByIDs инфа о купленном продукте для пользователя по id пользователя и id продукта
func (p *PaymentsRepoPg) GetUserProductByIDs(ctx context.Context, productID int, userID string) (product *entity.UserProducts, err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(&entity.UserProducts{}).Where("user_id = ? AND product_id = ?", userID, productID).First(&product).Error

	return
}

func (p *PaymentsRepoPg) CreatePayment(ctx context.Context, payment *entity.Payment) (err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(&entity.Payment{}).Create(payment).Error

	return
}

func (p *PaymentsRepoPg) GetPaymentByID(ctx context.Context, paymentID string) (payment *entity.Payment, err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(&entity.Payment{}).Where("id = ?", paymentID).First(&payment).Error

	return
}

func (p *PaymentsRepoPg) GetPaymentByExternalID(ctx context.Context, externalPaymentID string) (payment *entity.Payment, err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(&entity.Payment{}).Where("external_id = ?", externalPaymentID).First(&payment).Error

	return
}

func (p *PaymentsRepoPg) GetPaymentsByUserID(ctx context.Context, userID string) (payments []*entity.Payment, err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(&entity.Payment{}).Where("user_id = ?", userID).Find(&payments).Error

	return
}

func (p *PaymentsRepoPg) UpdatePaymentStatus(ctx context.Context, paymentID string, status entity.PaymentStatus, extStatus string) (err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(&entity.Payment{}).Where("id = ?", paymentID).Update("status", status).Update("external_status", extStatus).Error

	return
}

func (p *PaymentsRepoPg) SetUserProduct(ctx context.Context, UserID string, ProductID int, expire time.Time) (err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(&entity.UserProducts{}).Create(&entity.UserProducts{
		UserID:    UserID,
		ProductID: ProductID,
		Expire:    expire,
	}).Error

	return
}

func (p *PaymentsRepoPg) UpdatePayment(ctx context.Context, payment *entity.Payment) (err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.WithContext(ctx).Model(payment).Where("id = ?", payment.ID).Updates(payment).Error

	return err
}

func (p *PaymentsRepoPg) DeleteExpiredProductsByUserID(userID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.Model(&entity.UserProducts{}).Where("user_id = ? AND expire <= now()", userID).Delete(&entity.UserProducts{}).Error

	return
}

func (p *PaymentsRepoPg) DeleteExpiredProducts(ctx context.Context) (err error) {
	defer pg.ProcessDbError(&err)

	err = p.db.Model(&entity.UserProducts{}).WithContext(ctx).Where("expire <= now()").Delete(&entity.UserProducts{}).Error

	return
}
