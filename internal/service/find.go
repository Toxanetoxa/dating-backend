package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/pkg/geopoint"
	"github.com/toxanetoxa/dating-backend/pkg/pg"
	"log/slog"

	"github.com/google/uuid"
)

// Find - сервис для поиска пар, лайков, дизлайков
type Find struct {
	l         *slog.Logger
	userRepo  UserRepo
	matchRepo MatchRepo
	payRepo   PaymentsRepo
	push      PushService
	// todo
}

var (
	ErrUserLocationNotSet = errors.New("user location not set")
	ErrMatchAlreadyExist  = errors.New("match already exist")
	ErrLikeAlreadyExist   = errors.New("like already exist")
)

// NewFindService ...
func NewFindService(l *slog.Logger, u UserRepo, m MatchRepo, p PushService, pr PaymentsRepo) FindService {
	return &Find{
		l:         l,
		userRepo:  u,
		matchRepo: m,
		push:      p,
		payRepo:   pr,
	}
}

// Find подобрать пары для пользователя.
func (f *Find) Find(ctx context.Context, userID string, filter entity.Filter, limit int) ([]*entity.Couple, error) {

	_ = f.payRepo.DeleteExpiredProducts(context.Background()) // todo look here, вынести отсюда

	// todo Доделать

	user, err := f.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Geolocation.Lat == 0 && user.Geolocation.Lng == 0 {
		return nil, ErrUserLocationNotSet
	}

	withPhotosFilter := false

	// check user photo
	if len(user.Photos) > 0 {
		withPhotosFilter = true
	}

	users, err := f.userRepo.Find(ctx, filter.AgeFrom, filter.AgeTo, filter.Sex, filter.Radius, user.Geolocation.Lat, user.Geolocation.Lng, userID, limit, withPhotosFilter)
	if err != nil {
		f.l.Error(err.Error())
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

// Like установка лайка возвращает флаг, ид мэтча если был и ошибку
// здесь userID это тот, кто ставит лайк, targetUserID - кому он ставит лайк
func (f *Find) Like(ctx context.Context, userID string, targetUserID string) (bool, string, error) {
	has, err := f.userRepo.CheckHasMatch(ctx, targetUserID, userID)
	if err != nil {
		return false, "", err
	}

	if has {
		return true, "", ErrMatchAlreadyExist
	}

	match, err := f.userRepo.CheckMatch(ctx, targetUserID, userID)
	if err != nil {
		if errors.Is(err, pg.ErrEntityAlreadyExists) {

			return true, "", ErrMatchAlreadyExist
		}

		f.l.Error(err.Error())

		return false, "", err
	}

	newMatchID := uuid.New().String()

	if match {
		// создать мэтч
		newMatch := entity.Match{
			GeneralTechFields: entity.GeneralTechFields{ID: newMatchID},
			User1ID:           userID,
			User2ID:           targetUserID,
		}
		err = f.matchRepo.Create(ctx, &newMatch)
		if err != nil {

			return false, "", fmt.Errorf("could not create match: %w", err)
		}
		// удалить лайк
		err = f.userRepo.DeleteLike(ctx, targetUserID, userID) // todo: нужно это обернуть в транзакцию
		if err != nil {
			return false, "", fmt.Errorf("could not delete like: %w", err)
		}

		// отправка уведомления пользователю targetUserID
		err = f.push.SendNewMatchNotify(ctx, targetUserID, newMatchID, userID)
		if err != nil {
			f.l.Error(err.Error())
		}

		return true, newMatchID, nil
	}

	err = f.userRepo.Like(ctx, userID, targetUserID)
	if err != nil {
		if errors.Is(err, pg.ErrEntityAlreadyExists) {
			return false, "", ErrLikeAlreadyExist
		}
		f.l.Error(err.Error())

		return false, "", err
	}

	// отправить пуш о лайке
	err = f.push.SendNewLikeNotify(ctx, targetUserID, userID)
	if err != nil {
		f.l.Error(err.Error())
	}

	return false, "", nil
}

// Dislike установка дизлайка пользователю targetUserID
func (f *Find) Dislike(ctx context.Context, userID, targetUserID string) error {
	err := f.userRepo.Dislike(ctx, userID, targetUserID)
	if err != nil {
		f.l.Error(err.Error())
		return err
	}

	return nil
}

// ClearLikes удалить лайки и дизлайки, которые ставил юзер. нужно для тестов и отладки.
func (f *Find) ClearLikes(ctx context.Context, userID string) error {
	err := f.userRepo.ClearLikes(ctx, userID)
	if err != nil {
		f.l.Error(err.Error())
		return err
	}

	return nil
}

// RevertDislike отменить дизлайк (не в мвп)
func (f *Find) RevertDislike(ctx context.Context, userID, targetUserID string) error {
	// этот функционал в мвп пока отсутствует
	_ = ctx
	_ = userID
	_ = targetUserID
	return fmt.Errorf("not implemented")
}
