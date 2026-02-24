package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/youkassa"
	"github.com/toxanetoxa/dating-backend/pkg/pg"

	"github.com/google/uuid"
)

var (
	ErrBoostAlreadyExist = errors.New("boost already exist")
)

type Payment struct {
	l           *slog.Logger
	repo        PaymentsRepo
	userRepo    UserRepo
	YClient     youkassa.YClient
	push        PushService
	redirectURL string
	metrics     *entity.Metrics
}

const (
	ProductBoostID = 1
)

func NewPaymentService(l *slog.Logger, repo PaymentsRepo, ur UserRepo, yc youkassa.YClient, redirectURL string, p PushService, m *entity.Metrics) PaymentService {
	return &Payment{
		l:           l,
		repo:        repo,
		YClient:     yc,
		redirectURL: redirectURL,
		userRepo:    ur,
		push:        p,
		metrics:     m,
	}
}

// BuyBoost создание платежа на покупку буста для пользователя
func (p *Payment) BuyBoost(ctx context.Context, userID string) (confirmationUrl string, expired *time.Time, err error) {
	const (
		paymentDescription = "Покупка Boost в приложении Искра - Знакомства"
		paymentTLL         = time.Minute * 10
	)

	user, err := p.userRepo.GetByID(ctx, userID)
	if err != nil {
		return
	}

	// delete expired products
	err = p.deleteExpiredProducts(userID)
	if err != nil {
		return "", nil, err
	}

	// сначала надо проверить, нет ли уже у юзера активного буста
	userProduct, err := p.repo.GetUserProductByIDs(ctx, ProductBoostID, userID)
	if err != nil {
		if !errors.Is(err, pg.ErrEntityDoesntExist) {
			return
		}
	}

	if userProduct.ProductID == ProductBoostID {
		// буст уже куплен, вернём ошибку
		err = ErrBoostAlreadyExist
		return
	}

	// get price
	product, err := p.repo.GetProduct(ctx, ProductBoostID)
	if err != nil {
		return
	}

	// метрика для общего кол-ва попыток покупки
	p.metrics.BuyBoostTotal.Inc()

	// try to create payment in partner payment system
	extStatus, extID, url, err := p.YClient.CreatePayment(ctx, &youkassa.YPayment{
		Amount:      product.Price,
		Currency:    product.Currency,
		ReturnUrl:   p.redirectURL,
		Description: paymentDescription, // todo
		CustomerID:  userID,
		UserEmail:   user.Email,
		UserPhone:   user.Phone,
	})
	if err != nil {
		p.metrics.FailedPayments.Inc()

		p.l.Error("failed creating payment",
			slog.String("error", err.Error()))

		return
	}

	// create payment local
	localPaymentID := uuid.New().String()
	err = p.repo.CreatePayment(context.WithoutCancel(ctx), &entity.Payment{
		GeneralTechFields: entity.GeneralTechFields{
			ID:        localPaymentID,
			CreatedAt: time.Now(),
		},
		ExternalID:     extID,
		Status:         entity.PaymentStatusPending,
		ExternalStatus: extStatus,
		Price:          product.Price,
		Currency:       product.Currency,
		PaidPrice:      0,
		PaidCurrency:   "",
		Description:    paymentDescription + " для пользователя " + userID,
		UserID:         userID,
		ProductID:      ProductBoostID,
	})
	if err != nil {
		return
	}

	expiredTime := time.Now().Add(paymentTLL)

	return url, &expiredTime, nil
}

// PaymentCallback обработка уведомлений о статусах платежей от юкассы
func (p *Payment) PaymentCallback(ctx context.Context, event string, obj youkassa.NotificationObject) error {
	payment, err := p.repo.GetPaymentByExternalID(context.WithoutCancel(ctx), obj.Id)
	if err != nil {
		p.l.Error("could not find payment in repo",
			slog.String("error", err.Error()),
			slog.String("ext-id", obj.Id))

		return err
	}

	switch event {
	case "payment.waiting_for_capture":
		err = p.repo.UpdatePaymentStatus(ctx, payment.ID, entity.PaymentStatusPending, obj.Status)
		if err != nil {
			p.l.Error("could not update status for payment: %s, err: %s",
				slog.String("id", payment.ID),
				slog.String("error", err.Error()))
			return err
		}
	case "payment.succeeded":
		// успешная оплата, нужно активировать услугу и обновить статус
		product, err := p.repo.GetProduct(context.WithoutCancel(ctx), ProductBoostID)
		if err != nil {
			p.l.Error(err.Error())

			return err
		}

		ttlForBoost := time.Second * time.Duration(product.Validity)
		err = p.repo.SetUserProduct(context.WithoutCancel(ctx), payment.UserID, payment.ProductID, time.Now().Add(ttlForBoost))
		if err != nil {
			p.l.Error("could not save user product",
				slog.String("error", err.Error()),
				slog.String("user-id", payment.UserID))

			return err
		}

		p.metrics.SuccessPayments.Inc()

		// обновляем платёж (статус и данные об итоговой сумме)
		payment.ExternalStatus = obj.Status
		payment.Status = entity.PaymentStatusPaid
		amountValue, _ := strconv.ParseFloat(obj.Amount.Value, 64) // тут игнорируется ошибка, может и не распарсится
		payment.PaidPrice = amountValue
		payment.PaidCurrency = obj.Amount.Currency

		err = p.repo.UpdatePayment(context.WithoutCancel(ctx), payment)
		if err != nil {
			p.l.Error("could not update status for payment: %s, err: %s",
				slog.String("payment-id", payment.ID),
				slog.String("error", err.Error()))
			return err
		}

		// toMetric: здесь можно накинуть метрику, об успешной оплате (неуспешную тоже, ниже)
		err = p.push.SendPaymentSuccess(context.WithoutCancel(ctx), payment.UserID, ProductBoostID)
		if err != nil {
			p.l.Error("could not send push",
				slog.String("error", err.Error()),
				slog.String("user-id", payment.UserID))
		}

	case "payment.canceled":
		err = p.repo.UpdatePaymentStatus(ctx, payment.ID, entity.PaymentStatusCanceled, obj.Status)
		if err != nil {
			p.l.Error("could not update status for payment: %s, err: %s",
				slog.String("payment-id", payment.ID),
				slog.String("error", err.Error()))

			return err
		}

		// здесь возможна отправка уведомления о том что платёж был отменен внешней системой

		p.metrics.CanceledPayments.Inc()
	default:
		p.l.Error("could not parse external event (status)",
			slog.String("event", event))

		return fmt.Errorf("invalid status")
	}

	return nil
}

func (p *Payment) CheckPaymentStatus(_ context.Context) {
	// todo
}

// GetUserBoost возвращает данные по текущему бусту пользователя, либо nil если он не куплен
func (p *Payment) GetUserBoost(ctx context.Context, userID string) (*entity.UserProducts, error) {
	// delete expired products
	err := p.deleteExpiredProducts(userID)
	if err != nil {
		return nil, err
	}

	product, err := p.repo.GetUserProductByIDs(ctx, ProductBoostID, userID)
	if err != nil {
		if errors.Is(err, pg.ErrEntityDoesntExist) {
			// не был куплен

			return nil, nil
		}

		return nil, err
	}

	return &entity.UserProducts{
		UserID:    product.UserID,
		ProductID: ProductBoostID,
		Expire:    product.Expire,
	}, nil
}

// GetProductBoost возвращает данные по Бусту
func (p *Payment) GetProductBoost(ctx context.Context) (*entity.Product, error) {
	product, err := p.repo.GetProduct(ctx, ProductBoostID)
	if err != nil {
		return nil, err
	}

	return &entity.Product{
		ID:       product.ID,
		Price:    product.Price,
		OldPrice: product.OldPrice,
		Currency: product.Currency,
		Validity: product.Validity,
		Name:     product.Name,
	}, nil
}

func (p *Payment) deleteExpiredProducts(userID string) error {
	return p.repo.DeleteExpiredProductsByUserID(userID)
}
