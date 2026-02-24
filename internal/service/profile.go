package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"mime/multipart"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/pkg/geopoint"
	"github.com/toxanetoxa/dating-backend/pkg/pg"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slices"
)

var (
	ErrProfileNotCompleted = errors.New("profile not completed")
	ErrInvalidFileType     = errors.New("invalid file type")
	ErrNotMatch            = errors.New("user have not match")
)

type (
	UserRegistrationParams struct {
		Phone     string `validate:"required,e164"`
		Password  string `validate:"required"`
		FirstName string `validate:"required"`
	}

	UserAuthParams struct {
		Phone    string `validate:"required,e164"`
		Password string `validate:"required"`
	}

	UserUpdateProfileParams struct {
		Name     string
		Sex      entity.UserSex
		Birthday entity.BirthDate
		City     string
		About    string
		Email    string
		Phone    string
	}

	Profile struct {
		l            *slog.Logger
		repo         UserRepo
		photoRepo    UserPhotoRepo
		s3           *minio.Client
		s3BucketName string
		rdb          *redis.Client // todo: refactor
		filesBaseUrl string
		metrics      *entity.Metrics
	}
)

// NewProfileService init user service
func NewProfileService(l *slog.Logger, r UserRepo, phr UserPhotoRepo, s3client *minio.Client, bn string, a *redis.Client, f string, m *entity.Metrics) ProfileService {
	return &Profile{
		l:            l,
		repo:         r,
		s3:           s3client,
		photoRepo:    phr,
		s3BucketName: bn,
		rdb:          a,
		filesBaseUrl: f,
		metrics:      m,
	}
}

// Create создать пользователя
func (p *Profile) Create(ctx context.Context, email string) (*entity.User, error) {
	id := uuid.New()
	_, err := p.repo.Create(ctx, id.String(), email, nil)
	if err != nil {
		return nil, err
	}

	return &entity.User{
		GeneralTechFields: entity.GeneralTechFields{
			ID:        id.String(),
			CreatedAt: time.Now(),
		},
		Email:    email,
		Status:   entity.UserStatusNew,
		AuthType: entity.UserAuthTypeEmail,
	}, nil
}

func (p *Profile) CreateByVK(ctx context.Context, vkID, vkAccessToken string) (*entity.User, error) {
	id := uuid.New()
	err := p.repo.CreateByVK(ctx, id.String(), vkID, vkAccessToken)
	if err != nil {
		return nil, err
	}

	return &entity.User{
		GeneralTechFields: entity.GeneralTechFields{
			ID:        id.String(),
			CreatedAt: time.Now(),
		},
		Status:   entity.UserStatusNew,
		AuthType: entity.UserAuthTypeVK,
	}, nil
}

// GetByEmail получить пользователя по email
func (p *Profile) GetByEmail(ctx context.Context, email string) (user *entity.User, err error) {
	user, err = p.repo.GetByEmail(ctx, email)

	return
}

func (p *Profile) GetByVkID(ctx context.Context, vkID string) (user *entity.User, err error) {
	user, err = p.repo.GetByVkID(ctx, vkID)

	return
}

// GetByID получить пользователя по id
func (p *Profile) GetByID(ctx context.Context, id string) (user *entity.User, err error) {
	user, err = p.repo.GetByID(ctx, id)
	if user.Status == entity.UserStatusNew {
		return user, ErrProfileNotCompleted
	}

	return
}

// UpdateProfile обновляет данные пользователя
func (p *Profile) UpdateProfile(ctx context.Context, profile UserUpdateProfileParams, userID string) error {
	// todo доделать ? или мб пока не надо
	// get the profile
	user, err := p.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pg.ErrEntityDoesntExist) {

			return err
		}

		p.l.Error(err.Error())

		return err
	}

	switch user.Status {
	case entity.UserStatusNew:
		user.Status = entity.UserStatusActive
		user.FirstName = &profile.Name
		user.Sex = &profile.Sex
		user.Birthday = &profile.Birthday
		user.City = &profile.City
		user.About = &profile.About

		if user.AuthType == entity.UserAuthTypeVK {
			user.Email = profile.Email
			user.Phone = profile.Phone
		}

		err = p.repo.Update(ctx, user)
		if err != nil {
			p.l.Error("could not update profile",
				slog.String("error", err.Error()))

			return err
		}

	case entity.UserStatusActive:
		user.Status = entity.UserStatusActive
		user.FirstName = &profile.Name
		user.City = &profile.City
		user.About = &profile.About
		err = p.repo.Update(ctx, user)
		if err != nil {
			p.l.Error("could not update profile",
				slog.String("error", err.Error()))

			return err
		}

	case entity.UserStatusInactive:

		return fmt.Errorf("update for inactive users not allowed")

	default:
		p.l.Error("invalid user status",
			slog.String("status", string(user.Status)),
			slog.String("user-id", userID))

		return fmt.Errorf("invalid user status")
	}

	return nil
}

// DeleteProfile меняет статус пользователя на удаленный (сейчас пока удаляет из базы)
func (p *Profile) DeleteProfile(ctx context.Context, userID string) error {
	// get photos
	photos, err := p.photoRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		p.l.Error(err.Error())

		return err
	}

	// delete from s3
	for i := range photos {
		err := p.s3.RemoveObject(ctx, p.s3BucketName, photos[i].Object, minio.RemoveObjectOptions{})
		if err != nil {
			p.l.Error(err.Error())

			return fmt.Errorf("could not delete photo from s3: %w", err)
		}

	}

	// delete from repo
	err = p.photoRepo.DeleteAllByUserID(ctx, userID)
	if err != nil {
		p.l.Error(err.Error())

		return err
	}

	// delete likes and dislikes
	err = p.repo.ClearLikes(ctx, userID)
	if err != nil {
		p.l.Error(err.Error())

		return fmt.Errorf("could not delete likes: %w", err)
	}

	err = p.repo.DeleteMatches(ctx, userID)
	if err != nil {
		p.l.Error(err.Error())

		return fmt.Errorf("could not delete matches: %w", err)
	}

	//log out (del tokens)
	err = p.rdb.Del(ctx, rdbTokenPrefix+userID).Err()
	if err != nil {
		p.l.Error("could not delete code: %v", err)
	}
	err = p.rdb.Del(ctx, rdbRefreshPrefix+userID).Err()
	if err != nil {
		p.l.Error("could not delete code: %v", err)
	}

	// delete profile
	err = p.repo.Delete(ctx, userID)
	if err != nil {
		p.l.Error(err.Error())

		return err
	}

	p.metrics.DeleteUserTotal.Inc()

	return err
}

// UploadPhoto загрузить фото в хранилище и бд
func (p *Profile) UploadPhoto(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (id, url string, err error) {
	var validFileTypes = []string{"image/jpeg", "image/png", "image/webp"}

	buf, _ := fileHeader.Open()
	mime, err := mimetype.DetectReader(buf)
	if err != nil {
		err = ErrInvalidFileType
		p.l.Debug(err.Error())

		return
	}

	if !slices.Contains(validFileTypes, mime.String()) {
		err = ErrInvalidFileType
		p.l.Debug("not supported file",
			slog.String("mime", mime.String()))

		return
	}

	filename := fmt.Sprintf("photo_%d%s", time.Now().UnixMilli(), mime.Extension())

	_, _ = buf.Seek(0, 0)

	info, err := p.s3.PutObject(ctx, p.s3BucketName, filename, buf, fileHeader.Size, minio.PutObjectOptions{ContentType: mime.String()})
	if err != nil {
		p.l.Error(err.Error())

		return
	}

	photoID := uuid.New()
	id = photoID.String()

	_, err = p.photoRepo.Create(ctx, &entity.UserPhoto{
		GeneralTechFields: entity.GeneralTechFields{
			ID: photoID.String(),
		},
		UserID: userID,
		URL:    p.filesBaseUrl + info.Key,
		IsMain: false,
		Object: info.Key,
	}, nil)
	if err != nil {
		return
	}

	url = p.filesBaseUrl + info.Key

	return
}

// DeletePhoto удалить фотку из базы и хранилища
func (p *Profile) DeletePhoto(ctx context.Context, photoID, userID string) error {

	// todo get by user for check
	photo, err := p.photoRepo.GetByID(ctx, photoID)
	if err != nil {
		return err
	}

	err = p.s3.RemoveObject(ctx, p.s3BucketName, photo.Object, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	err = p.photoRepo.Delete(ctx, photoID, userID)
	if err != nil {
		return err
	}

	return err
}

// SetMainPhoto сделать фото основным
func (p *Profile) SetMainPhoto(ctx context.Context, userID, photoID string) error {
	err := p.photoRepo.SetMain(ctx, userID, photoID)
	if err != nil {
		return err
	}

	return err
}

// UpdateGeo обновить текущую геолокацию
func (p *Profile) UpdateGeo(ctx context.Context, userID string, lat, long float64) error {
	err := p.repo.UpdateGeo(ctx, userID, lat, long)
	if err != nil {
		return err
	}

	return nil
}

// GetPhotosByUserID возвращает список фоток из базы
func (p *Profile) GetPhotosByUserID(ctx context.Context, userID string) (photos []*entity.UserPhoto, err error) {
	photos, err = p.photoRepo.GetAllByUserID(ctx, userID)
	return
}

// GetProfileByID получить профиль пользователя (только того, с кем есть мэтч)
func (p *Profile) GetProfileByID(ctx context.Context, userID, targetID string) (*entity.User, error) {
	match, err := p.repo.CheckMatch(ctx, userID, targetID)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrNotMatch
	}

	user, err := p.repo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetDistanceToUser расчитать расстояние до пользователя
func (p *Profile) GetDistanceToUser(ctx context.Context, userID string, toUserID string) (int, error) {
	user, err := p.repo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}

	toUser, err := p.repo.GetByID(ctx, toUserID)
	if err != nil {
		return 0, err
	}

	dist := int(math.Round(geopoint.Distance(user.Geolocation.Lat, user.Geolocation.Lng, toUser.Geolocation.Lat, toUser.Geolocation.Lng, "K")))

	return dist, nil
}

// CreateSession создать новую сессию авторизации пользователя
func (p *Profile) CreateSession(ctx context.Context, userID, sessionID, ip string) error {
	return p.repo.CreateSession(ctx, userID, sessionID, ip)
}

// DeleteSession удалить сессию авторизации пользователя
func (p *Profile) DeleteSession(ctx context.Context, sessionID string) error {
	return p.repo.DeleteSession(ctx, sessionID)
}

// UpdateDeviceID добавить или обновить device id для сессии пользователя
func (p *Profile) UpdateDeviceID(ctx context.Context, deviceID, userID, sessionID string) error {
	return p.repo.SaveDeviceID(ctx, userID, sessionID, deviceID)
}

// LogOut разлогиниться - удалить текущую сессию и токены текущей сессии пользователя
func (p *Profile) LogOut(ctx context.Context, sessionID string) error {
	err := p.rdb.Del(ctx, rdbTokenPrefix+sessionID).Err()
	if err != nil {
		return fmt.Errorf("coul not delete token: %w", err)
	}

	err = p.rdb.Del(context.WithoutCancel(ctx), rdbRefreshPrefix+sessionID).Err()
	if err != nil {
		return fmt.Errorf("could not delete refresh-token: %w", err)
	}

	err = p.DeleteSession(context.WithoutCancel(ctx), sessionID)
	if err != nil {
		return err
	}

	p.metrics.LogoutTotal.Inc()

	return nil
}
