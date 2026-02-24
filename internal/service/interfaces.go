package service

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/youkassa"
	"github.com/toxanetoxa/dating-backend/pkg/pg"
)

type (
	AuthService interface {
		RequestCode(ctx context.Context, phone, hash string) error
		LogIn(ctx context.Context, phone, code string) (accessToken string, refreshToken string, tokenExpire int, isNewUser bool, userID string, err error)
		LoginByVK(ctx context.Context, token string) (accessToken string, refreshToken string, tokenExpire int, isNewUser bool, userID string, err error)
		RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, tokenExpire int, err error)
		GetKeyBySessionID(ctx context.Context, userID string) (string, error)
	}

	TokenStorage interface {
		AddToken(ctx context.Context, userUUID string, secretKey string, ttl int) error
		GetTokenByUserID(ctx context.Context, userUUID string) (key string, err error)
	}

	CodesStorage interface {
		AddCode(ctx context.Context, phone, code string, ttl int) error
		GetCodeByPhone(ctx context.Context, phone string) (code string, err error)
	}

	SmsService interface {
		SendCode(ctx context.Context, phone string, ip string) (code string, err error)
	}

	EmailService interface {
		SendCode(ctx context.Context, to string, code string) error
	}

	PushService interface {
		SendNewMatchNotify(ctx context.Context, userID, matchID, matchUserID string) error
		SendNewLikeNotify(ctx context.Context, userID, likeUserID string) error
		SendNewMessageNotify(ctx context.Context, userID, matchID, fromUserID string) error
		SendNewInfoNotify(ctx context.Context, userID, title, description string) error
		SendPaymentSuccess(ctx context.Context, userID string, productID int) error
	}

	ProfileService interface {
		Create(ctx context.Context, email string) (user *entity.User, err error)
		CreateByVK(ctx context.Context, vkID, vkAccessToken string) (*entity.User, error)
		GetByEmail(ctx context.Context, email string) (user *entity.User, err error)
		GetByVkID(ctx context.Context, vkID string) (user *entity.User, err error)
		GetByID(ctx context.Context, id string) (user *entity.User, err error)
		UpdateProfile(ctx context.Context, profile UserUpdateProfileParams, userID string) error
		DeleteProfile(ctx context.Context, userID string) error
		UploadPhoto(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (id, url string, err error)
		SetMainPhoto(ctx context.Context, userID, photoID string) error
		DeletePhoto(ctx context.Context, photoID, userID string) (err error)
		GetPhotosByUserID(ctx context.Context, userID string) (photos []*entity.UserPhoto, err error)
		UpdateGeo(ctx context.Context, userID string, lat, long float64) error
		GetProfileByID(ctx context.Context, userID, targetID string) (*entity.User, error)
		GetDistanceToUser(ctx context.Context, userID string, toUserID string) (int, error)

		LogOut(ctx context.Context, sessionID string) error
		CreateSession(ctx context.Context, userID, sessionID, ip string) error
		DeleteSession(ctx context.Context, sessionID string) error
		UpdateDeviceID(ctx context.Context, deviceID, userID, sessionID string) error
	}

	UserRepo interface {
		Create(ctx context.Context, id, email string, withTx pg.Transaction) (tx pg.Transaction, err error)
		CreateByVK(ctx context.Context, id, vkID, vkAccessToken string) (err error)
		GetByEmail(ctx context.Context, email string) (user *entity.User, err error)
		GetByVkID(ctx context.Context, vkID string) (user *entity.User, err error)
		GetByID(ctx context.Context, userID string) (user *entity.User, err error)
		Update(ctx context.Context, user *entity.User) (err error)
		UpdateGeo(ctx context.Context, userID string, lat, long float64) error
		Deactivate(ctx context.Context, userID string) (err error)
		Activate(ctx context.Context, userID string) (err error)
		Delete(ctx context.Context, userID string) (err error)
		Find(ctx context.Context, ageFrom int, ageTo int, sex *entity.UserSex, radius int, lat, lng float64, userID string, limit int, withPhotos bool) (users []*entity.User, err error)
		Like(ctx context.Context, fromUserID, toUserID string) error
		Dislike(ctx context.Context, fromUserID, toUserID string) error
		CheckMatch(ctx context.Context, fromUserID, toUserID string) (match bool, err error)
		CheckHasMatch(ctx context.Context, fromUserID, toUserID string) (has bool, err error)
		GetLikes(ctx context.Context, userID string) (users []*entity.User, err error)
		ClearLikes(ctx context.Context, userID string) error // for debug and delete profile
		DeleteMatches(ctx context.Context, userID string) (err error)
		DeleteLike(ctx context.Context, userID, targetID string) (err error)
		GetDeviceIDs(ctx context.Context, userID string) (ids []string, err error)
		DeleteDeviceID(ctx context.Context, sessionID string) (err error)
		CreateSession(ctx context.Context, userID, sessionID, ip string) (err error)
		DeleteSession(ctx context.Context, sessionID string) (err error)
		SaveDeviceID(ctx context.Context, userID, sessionID, deviceID string) (err error)

		List(ctx context.Context, limit, offset int) (users []*entity.User, err error)
		Stats(ctx context.Context) (usersStats entity.UsersStats, err error)
	}

	UserPhotoRepo interface {
		Create(ctx context.Context, photo *entity.UserPhoto, withTx pg.Transaction) (tx pg.Transaction, err error)
		Delete(ctx context.Context, id, userID string) (err error)
		DeleteAllByUserID(ctx context.Context, userID string) (err error)
		SetMain(ctx context.Context, userID, photoId string) (err error)
		GetAllByUserID(ctx context.Context, userID string) (list []*entity.UserPhoto, err error)
		GetByID(ctx context.Context, id string) (photo *entity.UserPhoto, err error)
	}

	MatchRepo interface {
		Create(ctx context.Context, match *entity.Match) error
		DeleteByID(ctx context.Context, matchID string) error
		GetByID(ctx context.Context, id string) (match *entity.Match, err error)
		GetAllByUserID(ctx context.Context, userID string) (matches []*entity.Match, err error)
	}

	ChatRepo interface {
		GetChats(ctx context.Context, userID string) (chats []*entity.Chat, err error)
		GetMessagesByMatchID(ctx context.Context, userID, matchID string, limit, offset int) (messages []*entity.ChatMessage, err error)
		SaveMessage(ctx context.Context, message *entity.ChatMessage) error
		MarkAsRead(ctx context.Context, messageID string) error
		MarkAsDelivered(ctx context.Context, messageID string) error
	}

	PaymentsRepo interface {
		GetProduct(ctx context.Context, productID int) (*entity.Product, error)
		GetUserProductByIDs(ctx context.Context, productID int, userID string) (product *entity.UserProducts, err error)
		CreatePayment(ctx context.Context, payment *entity.Payment) error
		GetPaymentByID(ctx context.Context, paymentID string) (payment *entity.Payment, err error)
		GetPaymentByExternalID(ctx context.Context, externalPaymentID string) (payment *entity.Payment, err error)
		UpdatePaymentStatus(ctx context.Context, paymentID string, status entity.PaymentStatus, extStatus string) (err error)
		SetUserProduct(ctx context.Context, userID string, productID int, expire time.Time) (err error)
		GetPaymentsByUserID(ctx context.Context, userID string) (payments []*entity.Payment, err error)
		UpdatePayment(ctx context.Context, payment *entity.Payment) (err error)
		DeleteExpiredProductsByUserID(userID string) (err error)
		DeleteExpiredProducts(ctx context.Context) (err error)
	}

	AdminRepo interface {
		FindByLogin(ctx context.Context, login string) (*entity.Admin, error)
		GetByID(ctx context.Context, id string) (admin *entity.Admin, err error)
		SaveToken(ctx context.Context, adminToken *entity.AdminToken) error
		GetToken(ctx context.Context, adminID, tokenID string) (*entity.AdminToken, error)
		DeleteTokenByID(ctx context.Context, tokenID string) (err error)
	}

	FindService interface {
		Find(ctx context.Context, userID string, filter entity.Filter, limit int) ([]*entity.Couple, error)
		Like(ctx context.Context, userID, targetUserID string) (match bool, id string, err error)
		Dislike(ctx context.Context, userID, targetUserID string) error
		ClearLikes(ctx context.Context, userID string) error
	}

	LikeService interface {
		List(ctx context.Context, userID string) ([]*entity.Couple, error)
	}

	ChatService interface {
		ChatsList(ctx context.Context, userID string) (list []*entity.Chat, err error)
		MessagesList(ctx context.Context, userID string, chatID string, limit, offset int) (messages []*entity.ChatMessage, err error)
		SaveMessage(ctx context.Context, msg *entity.ChatMessage) error
		GetChatUsersByMatchID(ctx context.Context, matchID string) (users []*entity.User, err error)
		MarkAsDelivered(ctx context.Context, id string) error
		MarkAsRead(ctx context.Context, id string) error
		SendMessage(ctx context.Context, fromUserID, toUserID, matchID, text string) (string, error)
		GetWSEventBus() *chan *entity.WsMsgPayload
	}

	PaymentService interface {
		GetProductBoost(ctx context.Context) (*entity.Product, error)
		GetUserBoost(ctx context.Context, userID string) (*entity.UserProducts, error)
		BuyBoost(ctx context.Context, userID string) (confirmationUrl string, expired *time.Time, err error)
		PaymentCallback(ctx context.Context, event string, obj youkassa.NotificationObject) error
	}

	MatchService interface {
		GetByUserID(ctx context.Context, userID string) ([]*entity.Match, error)
	}

	// admin services

	AdminAuthService interface {
		Login(ctx context.Context, login, code string) (token string, err error)
		ValidateToken(ctx context.Context, token string) (*entity.Admin, string, error)
	}

	AdminProfileService interface {
		GetByID(ctx context.Context, id string) (*entity.Admin, error)
		LogOut(ctx context.Context, tokenID string) error
		UpdatePassword(ctx context.Context) error
	}

	AdminUsersService interface {
		List(ctx context.Context, limit, offset int) ([]*entity.User, error)
		FindByID(ctx context.Context, userID string) (*entity.User, error)
		FindByPhone(ctx context.Context, phone string) (*entity.User, error)
		Verify(ctx context.Context, userID string) error
		Block(ctx context.Context, userID string) error
		Unblock(ctx context.Context, userID string) error
		Stats(ctx context.Context) (usersStats entity.UsersStats, err error)
	}
)
