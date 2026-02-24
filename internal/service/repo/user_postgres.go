package repo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/pkg/geopoint"
	"github.com/toxanetoxa/dating-backend/pkg/pg"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepoPg struct {
	db *gorm.DB
	l  *slog.Logger
}

func NewUserRepo(db *gorm.DB, l *slog.Logger) service.UserRepo {
	return &UserRepoPg{
		db: db,
		l:  l,
	}
}

func (u *UserRepoPg) Create(ctx context.Context, id, email string, _ pg.Transaction) (tx pg.Transaction, err error) {
	defer pg.ProcessDbError(&err)

	user := entity.User{
		GeneralTechFields: entity.GeneralTechFields{
			ID:        id,
			CreatedAt: time.Now(),
		},
		Email:    email,
		Status:   entity.UserStatusNew,
		AuthType: entity.UserAuthTypeEmail,
	}
	result := u.db.WithContext(ctx).Create(&user)

	return nil, result.Error
}

func (u *UserRepoPg) CreateByVK(ctx context.Context, id, vkID, vkAccessToken string) (err error) {
	defer pg.ProcessDbError(&err)

	user := entity.User{
		GeneralTechFields: entity.GeneralTechFields{
			ID:        id,
			CreatedAt: time.Now(),
		},
		Status:      entity.UserStatusNew,
		AuthType:    entity.UserAuthTypeVK,
		VkAuthToken: vkAccessToken,
		VkID:        vkID,
	}

	return u.db.WithContext(ctx).Create(&user).Error
}

func (u *UserRepoPg) GetByEmail(ctx context.Context, email string) (user *entity.User, err error) {
	defer pg.ProcessDbError(&err)

	user = &entity.User{}
	err = u.db.WithContext(ctx).Where("email = ?", email).First(user).Error

	return
}

func (u *UserRepoPg) GetByVkID(ctx context.Context, vkID string) (user *entity.User, err error) {
	defer pg.ProcessDbError(&err)

	user = &entity.User{}
	err = u.db.WithContext(ctx).Where("vk_id = ?", vkID).First(user).Error

	return
}

func (u *UserRepoPg) GetByID(ctx context.Context, userID string) (user *entity.User, err error) {
	defer pg.ProcessDbError(&err)

	user = &entity.User{
		GeneralTechFields: entity.GeneralTechFields{
			ID: userID,
		},
	}
	err = u.db.WithContext(ctx).Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("is_main DESC")
	}).First(user).Error

	return
}

func (u *UserRepoPg) Update(ctx context.Context, user *entity.User) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Save(user).Error

	return
}

func (u *UserRepoPg) Deactivate(ctx context.Context, userID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.Debug().WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("status", entity.UserStatusInactive).Error

	return
}

func (u *UserRepoPg) Activate(ctx context.Context, userID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("status", entity.UserStatusActive).Error

	return
}

func (u *UserRepoPg) Delete(ctx context.Context, userID string) (err error) {
	defer pg.ProcessDbError(&err)

	user := entity.User{
		GeneralTechFields: entity.GeneralTechFields{
			ID: userID,
		},
	}

	err = u.db.WithContext(ctx).Delete(&user).Error

	return
}

func (u *UserRepoPg) UpdateGeo(ctx context.Context, userID string, lat, long float64) (err error) {
	defer pg.ProcessDbError(&err)

	newGeo := geopoint.GeoPoint{
		Lng: long,
		Lat: lat,
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("geolocation", newGeo).Error

	return
}

func (u *UserRepoPg) Find(ctx context.Context, ageFrom int, ageTo int, sex *entity.UserSex, radius int, lat, lng float64, userID string, limit int, withPhotos bool) (users []*entity.User, err error) {
	defer pg.ProcessDbError(&err)

	if limit > 100 || limit <= 0 {
		return nil, fmt.Errorf("invalid limit. max 100, min 1")
	}
	likeSub := u.db.WithContext(ctx).Model(&entity.Like{}).Select("to_user_id").Where("from_user_id = ?", userID)
	dislikeSub := u.db.WithContext(ctx).Model(&entity.Dislike{}).Select("to_user_id").Where("from_user_id = ?", userID)
	matchSub := u.db.WithContext(ctx).Model(&entity.Match{}).Select("user_1_id").Where("user_2_id = ?", userID)
	match2Sub := u.db.WithContext(ctx).Model(&entity.Match{}).Select("user_2_id").Where("user_1_id = ?", userID)

	tx := u.db.WithContext(ctx).
		Select(fmt.Sprintf("*, ST_Distance(geolocation, 'SRID=4326;POINT(%f %f)'::geometry) as act_dist, "+
			"(SELECT COUNT(*) FROM user_product WHERE user_id = users.id and product_id = 1) as boost", lng, lat)).
		Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_main DESC")
		}).
		Where("date_part('year', age(birthday))::int between ? and ?", ageFrom, ageTo).
		Where("id <> ?", userID).
		Where("id not in (?)", likeSub).
		Where("id not in (?)", dislikeSub).
		Where("id not in (?)", matchSub).
		Where("id not in (?)", match2Sub)

	if radius > 0 {
		tx.Where(fmt.Sprintf("ST_DWithin(geolocation, 'POINT(%f %f)', %d.0)", lng, lat, radius*1000))
	}

	if sex != nil {
		tx.Where("sex = ?", sex)
	}

	tx.Where("status = ?", entity.UserStatusActive)

	if !withPhotos {
		tx.Where("(SELECT COUNT(*) FROM user_photo WHERE user_id = users.id) = 0")
	}

	tx.Order("boost DESC, act_dist ASC")

	tx.Limit(limit)

	err = tx.Find(&users).Error

	return
}

func (u *UserRepoPg) Like(ctx context.Context, fromUserID, toUserID string) (err error) {
	defer pg.ProcessDbError(&err)

	like := entity.Like{
		GeneralTechFields: entity.GeneralTechFields{
			ID: uuid.New().String(),
		},
		FromUserID: fromUserID,
		ToUserID:   toUserID,
	}

	return u.db.WithContext(ctx).Create(&like).Error
}

func (u *UserRepoPg) Dislike(ctx context.Context, fromUserID, toUserID string) (err error) {
	defer pg.ProcessDbError(&err)

	dislike := entity.Dislike{
		GeneralTechFields: entity.GeneralTechFields{
			ID: uuid.New().String(),
		},
		FromUserID: fromUserID,
		ToUserID:   toUserID,
	}

	return u.db.WithContext(ctx).Create(&dislike).Error
}

// CheckHasMatch првоерка создан ли мэтч
func (u *UserRepoPg) CheckHasMatch(ctx context.Context, fromUserID, toUserID string) (has bool, err error) {
	defer pg.ProcessDbError(&err)

	var exist bool

	err = u.db.WithContext(ctx).Model(&entity.Match{}).Select("count(*) > 0").
		Where("user_1_id = ? and user_2_id = ?", fromUserID, toUserID).
		Or("user_2_id = ? and user_1_id = ?", fromUserID, toUserID).
		Find(&exist).Error

	if err != nil {
		return false, err
	}

	if exist {
		return true, nil
	} else {
		return false, nil
	}

}

// CheckMatch проверяет есть ли лайк от пользователя, которому мы хотим поставить лайк
func (u *UserRepoPg) CheckMatch(ctx context.Context, fromUserID, toUserID string) (match bool, err error) {
	defer pg.ProcessDbError(&err)

	var (
		exists      bool
		existsMatch bool
	)

	// сначала проверить нет ли уже мэтча
	err = u.db.WithContext(ctx).Model(&entity.Match{}).Select("count(*) > 0").
		Where("user_1_id = ? and user_2_id = ?", fromUserID, toUserID).
		Or("user_2_id = ? and user_1_id = ?", fromUserID, toUserID).
		Find(&existsMatch).Error
	// todo !!! ???

	if existsMatch {
		//err = gorm.ErrDuplicatedKey

		return true, nil
	}

	// проверка на взаимный лайк
	err = u.db.WithContext(ctx).Model(&entity.Like{}).
		Select("count(*) > 0").
		Where("from_user_id = ? and to_user_id = ?", fromUserID, toUserID).
		//Or("to_user_id = ? and from_user_id = ?", fromUserID, toUserID).
		Find(&exists).
		Error

	return exists, err
}

func (u *UserRepoPg) GetLikes(ctx context.Context, userID string) (users []*entity.User, err error) {
	defer pg.ProcessDbError(&err)

	likes := []entity.Like{}

	dis := u.db.WithContext(ctx).Raw("select to_user_id from user_dislike where from_user_id = ?", userID)

	err = u.db.WithContext(ctx).Preload("FromUser", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Photos", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_main DESC")
		})
	}).
		Where("to_user_id = ?", userID).
		Where("from_user_id not in (?)", dis).
		Find(&likes).Error

	for i := range likes {
		users = append(users, &likes[i].FromUser)
	}

	return
}

// ClearLikes удалит лайки и дизлайки для юзера. нужно для тестов и отладки.
func (u *UserRepoPg) ClearLikes(ctx context.Context, userID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Where("from_user_id = ?", userID).Delete(&entity.Like{}).Error
	if err != nil {

		return
	}

	err = u.db.WithContext(ctx).Where("to_user_id = ?", userID).Delete(&entity.Like{}).Error
	if err != nil {

		return
	}

	err = u.db.WithContext(ctx).Where("from_user_id = ?", userID).Delete(&entity.Dislike{}).Error
	if err != nil {

		return
	}

	err = u.db.WithContext(ctx).Where("to_user_id = ?", userID).Delete(&entity.Dislike{}).Error

	return
}

// DeleteMatches удалить мэтчи (нужно для удаления профиля)
func (u *UserRepoPg) DeleteMatches(ctx context.Context, userID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.ChatMessage{}).Error
	if err != nil {
		return err
	}

	err = u.db.WithContext(ctx).Where("user_1_id = $1 OR user_2_id = $1", userID).Delete(&entity.Match{}).Error

	return
}

func (u *UserRepoPg) DeleteLike(ctx context.Context, userID, targetID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Where("from_user_id = ? and to_user_id = ?", userID, targetID).Delete(&entity.Like{}).Error

	return
}

func (u *UserRepoPg) List(ctx context.Context, limit, offset int) (users []*entity.User, err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Model(&entity.User{}).Preload("Photos", func(db *gorm.DB) *gorm.DB {
		return db.Order("is_main DESC")
	}).Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error

	return
}

func (u *UserRepoPg) Stats(ctx context.Context) (usersStats entity.UsersStats, err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("status = ?", entity.UserStatusNew).Count(&usersStats.UsersNew).Error
	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("status = ?", entity.UserStatusActive).Count(&usersStats.UsersActive).Error
	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("status = ?", entity.UserStatusInactive).Count(&usersStats.UsersInActive).Error
	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Count(&usersStats.TotalUsers).Error
	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("sex = ?", entity.UserSexMale).Count(&usersStats.SexStats.Male).Error

	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("sex = ?", entity.UserSexFemale).Count(&usersStats.SexStats.Female).Error
	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("sex IS NULL").Count(&usersStats.SexStats.Undefined).Error
	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("created_at > now() - interval '1 day'").Count(&usersStats.UsersByDay).Error
	if err != nil {
		return
	}

	err = u.db.WithContext(ctx).Model(&entity.User{}).Where("created_at > now() - interval '7 day' ").Count(&usersStats.UsersByWeek).Error

	return
}

func (u *UserRepoPg) CreateSession(ctx context.Context, userID, sessionID, ip string) (err error) {
	defer pg.ProcessDbError(&err)

	session := entity.UserSession{
		ID:     sessionID,
		UserID: userID,
		Ip:     ip,
	}

	err = u.db.WithContext(ctx).Model((*entity.UserSession)(nil)).Create(&session).Error

	return
}

func (u *UserRepoPg) DeleteSession(ctx context.Context, sessionID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.Debug().WithContext(ctx).Where("id = ?", sessionID).Delete(&entity.UserSession{}).Error

	return
}

func (u *UserRepoPg) SaveDeviceID(ctx context.Context, userID, sessionID, deviceID string) (err error) {
	defer pg.ProcessDbError(&err)

	// подумать ещё над этим надо и обернуть в транзацию
	err = u.db.Model((*entity.UserSession)(nil)).Where("device_id = ?", deviceID).Update("device_id", "").Error
	if err != nil {
		u.l.Error(err.Error())
	}

	err = u.db.WithContext(ctx).Model((*entity.UserSession)(nil)).Where("id = ? and user_id = ?", sessionID, userID).Update("device_id", deviceID).Error

	return
}

func (u *UserRepoPg) GetDeviceIDs(ctx context.Context, userID string) (ids []string, err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Model((*entity.UserSession)(nil)).Select("device_id").Where("user_id = ?", userID).Find(&ids).Error

	return
}

func (u *UserRepoPg) DeleteDeviceID(ctx context.Context, sessionID string) (err error) {
	defer pg.ProcessDbError(&err)

	err = u.db.WithContext(ctx).Model((*entity.UserSession)(nil)).Where("id = ?", sessionID).Update("device_id", "").Error

	return
}
