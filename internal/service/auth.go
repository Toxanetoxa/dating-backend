package service

import (
	"context"
	cr "crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"math/rand"
	"time"
	"unsafe"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service/dto"
	"github.com/toxanetoxa/dating-backend/internal/vkapi"
	"github.com/toxanetoxa/dating-backend/pkg/pg"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	emailHashSalt    = "Hackers1sl0hs"
	refreshTTL       = 0 // unlimited
	rdbCodePrefix    = "code_"
	rdbTokenPrefix   = "token_"
	rdbRefreshPrefix = "refresh_"
	authCodeLen      = 4
	fakeCode         = "1234"
)

// Test Accounts (email/code)
var testAccounts = map[string]string{
	"testgp@test.ru": "1234",
	"testrs@test.ru": "1234",
}

var (
	ErrInvalidHash             = errors.New("invalid hash")
	ErrCodeAlreadySent         = errors.New("code already sent")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrUserIsBlocked           = errors.New("user is blocked")
	ErrInvalidRefreshToken     = errors.New("invalid refresh token")
	ErrUserTokenKeyNotFound    = errors.New("key not found")
	ErrInvalidVkAuth           = errors.New("invalid vk authorization")
)

var src = rand.NewSource(time.Now().UnixNano()) // for random generators

// Auth сервис авторизации пользователей
type Auth struct {
	l           *slog.Logger
	rdb         *redis.Client
	email       EmailService
	users       ProfileService // todo тут лучше юзать repo
	tokenTTL    time.Duration
	authCodeTtl time.Duration
	useFakeCode bool
	metrics     *entity.Metrics
	vkClient    vkapi.Client
}

func NewAuthService(rdb *redis.Client, e EmailService, u ProfileService, l *slog.Logger, tokenTtl time.Duration, authCodeTtl time.Duration, fake bool, m *entity.Metrics, vkClient vkapi.Client) AuthService {
	return &Auth{
		rdb:         rdb,
		email:       e,
		users:       u,
		l:           l,
		tokenTTL:    tokenTtl,
		authCodeTtl: authCodeTtl,
		useFakeCode: fake,
		metrics:     m,
		vkClient:    vkClient,
	}
}

// RequestCode запрос одноразового кода
func (a *Auth) RequestCode(ctx context.Context, email, hash string) error {
	// проверка хэша
	if hash != a.hashPhone(email, emailHashSalt) {
		return ErrInvalidHash
	}

	// Проверить не был ли он ещё отправлен
	_, err := a.rdb.Get(ctx, rdbCodePrefix+email).Result()

	switch {
	case err == nil:
		return ErrCodeAlreadySent
	case errors.Is(err, redis.Nil):
		// следуем стандартному флоу
	default:
		a.l.Error(err.Error())
		return err
	}

	// сгенерировать код или подставить fake
	authCode := fakeCode

	if !a.useFakeCode {
		authCode = a.genCode()
	}

	// test account check
	testCode, ok := testAccounts[email]
	if ok {
		authCode = testCode
	}

	// отправка кода
	err = a.email.SendCode(ctx, email, authCode)
	if err != nil {
		// todo handle errors
		return err
	}

	// сохранить код
	a.rdb.Set(ctx, rdbCodePrefix+email, authCode, a.authCodeTtl)

	return nil
}

// LogIn вход/регистрация пользователя по email и одноразовому коду
func (a *Auth) LogIn(ctx context.Context, email, code string) (accessToken string, refreshToken string, tokenExpire int, isNewUser bool, userID string, err error) {
	// Достать код
	verificationCode, err := a.rdb.Get(ctx, rdbCodePrefix+email).Result()
	switch {
	case errors.Is(err, redis.Nil):
		err = ErrInvalidVerificationCode
		return
	case err == nil:
		// следуем стандартному флоу
	default:
		a.l.Error(err.Error())
		return
	}

	// сверить код
	if code != verificationCode {
		err = ErrInvalidVerificationCode
		return
	}

	// если код был верный, надо его удалить
	err = a.rdb.Del(ctx, rdbCodePrefix+email).Err()
	if err != nil {
		a.l.Error("could not delete code: %v", err)
	}

	// достать пользователя. если нет, то создать нового
	user, err := a.users.GetByEmail(ctx, email)
	switch {
	case err == nil:
		// пользователь есть
		switch user.Status {
		case entity.UserStatusNew:
			isNewUser = true
		case entity.UserStatusInactive:
			err = ErrUserIsBlocked
			userID = user.ID

			return
		}
	case errors.Is(err, pg.ErrEntityDoesntExist):
		// нужно создать пользователя
		isNewUser = true
		user, err = a.users.Create(ctx, email)
		if err != nil {
			err = fmt.Errorf("failed to create new user: %w", err)
			return
		}
		// метрика на регистрацию нового пользователя
		a.metrics.RegistrationTotal.Inc()

	default:
		a.l.Error(err.Error())

		return
	}

	// создать сессию
	sessionID := uuid.New().String()
	err = a.users.CreateSession(ctx, user.ID, sessionID, "")
	if err != nil {
		a.l.Error("could not create session: %v", err)

		return
	}

	// сгенерить токены
	// для JWT
	secretKey := a.generateRandomKey()

	tokenExpire = int(a.tokenTTL) / 1000000000

	// Create a new token object
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": user.ID,
		"sub": sessionID,
		"exp": time.Now().Add(a.tokenTTL).Unix(),
	})
	// Sign and get the complete encoded token as a string using the secret
	accessToken, err = tokenObj.SignedString([]byte(secretKey))
	if err != nil {
		a.l.Error(err.Error())
		err = fmt.Errorf("falied to sign access token: %w", err)
		return
	}

	// refresh token
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": user.ID,
		"sub": sessionID,
		// "exp": time.Now().Add(refreshTTL).Unix(),
	})

	refreshSecret := a.generateRandomKey()

	refreshToken, err = refreshTokenObj.SignedString([]byte(refreshSecret))
	if err != nil {
		return
	}
	// положить токены
	err = a.rdb.Set(ctx, rdbTokenPrefix+sessionID, secretKey, a.tokenTTL).Err()
	if err != nil {
		return
	}

	// refresh
	err = a.rdb.Set(ctx, rdbRefreshPrefix+sessionID, refreshSecret, refreshTTL).Err()
	if err != nil {
		return
	}

	a.metrics.LoginTotal.Inc()

	return
}

func (a *Auth) LoginByVK(ctx context.Context, token string) (accessToken string, refreshToken string, tokenExpire int, isNewUser bool, userID string, err error) {
	vkProfile, err := a.vkClient.UserInfo(ctx, token)
	if err != nil {
		a.l.Error(err.Error())
		err = ErrInvalidVkAuth

		return
	}

	// если успех, конвертируем в наш формат

	vkProfileInfo, err := dto.VKAPIProfileInfoToVKProfileInfo(*vkProfile)
	if err != nil {
		a.l.Error("could not convert vk api response")

		return
	}

	user, err := a.users.GetByVkID(ctx, vkProfileInfo.ID)

	switch {
	case err == nil:
		// пользователь есть
		switch user.Status {
		case entity.UserStatusInactive:
			err = ErrUserIsBlocked
			userID = user.ID

			return
		}
	case errors.Is(err, pg.ErrEntityDoesntExist):
		// нужно создать пользователя
		isNewUser = true
		user, err = a.users.CreateByVK(ctx, vkProfileInfo.ID, token)
		if err != nil {
			err = fmt.Errorf("failed to create new user: %w", err)

			return
		}

		err = a.users.UpdateProfile(context.WithoutCancel(ctx), UserUpdateProfileParams{
			Name:     vkProfileInfo.FirstName,
			Sex:      vkProfileInfo.Sex,
			Birthday: vkProfileInfo.Birthday,
			About:    "",
			Email:    vkProfileInfo.Email,
			Phone:    vkProfileInfo.Phone,
		}, user.ID)
		if err != nil {
			a.l.Error("could not update profile (by vk)", "err", err.Error())

			return
		}
		// метрика на регистрацию нового пользователя
		a.metrics.RegistrationTotal.Inc()
	default:
		a.l.Error("could not get user from repo (db)", "err", err.Error())

		return
	}

	// создать сессию
	sessionID := uuid.New().String()
	err = a.users.CreateSession(ctx, user.ID, sessionID, "")
	if err != nil {
		a.l.Error("could not create session: %v", err)

		return
	}

	// сгенерить токены
	// для JWT
	secretKey := a.generateRandomKey()

	tokenExpire = int(a.tokenTTL) / 1000000000

	// Create a new token object
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": user.ID,
		"sub": sessionID,
		"exp": time.Now().Add(a.tokenTTL).Unix(),
	})
	// Sign and get the complete encoded token as a string using the secret
	accessToken, err = tokenObj.SignedString([]byte(secretKey))
	if err != nil {
		a.l.Error(err.Error())
		err = fmt.Errorf("falied to sign access token: %w", err)
		return
	}

	// refresh token
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": user.ID,
		"sub": sessionID,
		// "exp": time.Now().Add(refreshTTL).Unix(),
	})

	refreshSecret := a.generateRandomKey()

	refreshToken, err = refreshTokenObj.SignedString([]byte(refreshSecret))
	if err != nil {
		return
	}
	// положить токены
	err = a.rdb.Set(ctx, rdbTokenPrefix+sessionID, secretKey, a.tokenTTL).Err()
	if err != nil {
		return
	}

	// refresh
	err = a.rdb.Set(ctx, rdbRefreshPrefix+sessionID, refreshSecret, refreshTTL).Err()
	if err != nil {
		return
	}

	a.metrics.LoginTotal.Inc()

	return
}

// RefreshToken обновление токена авторизации по рефреш-токену
func (a *Auth) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, tokenExpire int, err error) {
	// расшифровать jwt
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidRefreshToken //fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		sub, err := token.Claims.GetSubject()
		if err != nil {
			return nil, ErrInvalidRefreshToken
		}

		// достать ключ рефреш токена для пользователя
		secret, err := a.rdb.Get(ctx, rdbRefreshPrefix+sub).Result()
		if err != nil {
			return nil, ErrInvalidRefreshToken
		}

		return []byte(secret), nil
	})

	if err != nil {
		return
	}

	// удалить старый рефреш

	// сначала получить user id
	uid, err := token.Claims.GetIssuer()
	if err != nil {
		a.l.Error(err.Error())
		err = ErrInvalidRefreshToken
		return
	}

	// достаём session ID
	sub, err := token.Claims.GetSubject()
	if err != nil {
		err = ErrInvalidRefreshToken
		return
	}

	// теперь удаляем, используя session id
	err = a.rdb.Del(ctx, rdbRefreshPrefix+sub).Err()
	if err != nil {
		a.l.Error(err.Error())
		return
	}

	// сгенерить и сохранить новые токены
	// для JWT
	secretKey := a.generateRandomKey()

	tokenExpire = int(a.tokenTTL) / 1000000000

	// Create a new token object
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": uid,
		"sub": sub,
		"exp": time.Now().Add(a.tokenTTL).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	newAccessToken, err = tokenObj.SignedString([]byte(secretKey))
	if err != nil {
		a.l.Error(err.Error())
		err = fmt.Errorf("falied to sign access token: %w", err)
		return
	}

	// refresh token
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": uid,
		"sub": sub,
		// "exp": time.Now().Add(refreshTTL).Unix(),
	})

	refreshSecret := a.generateRandomKey()

	newRefreshToken, err = refreshTokenObj.SignedString([]byte(refreshSecret))
	if err != nil {
		return
	}
	// положить токены
	err = a.rdb.Set(ctx, rdbTokenPrefix+sub, secretKey, a.tokenTTL).Err()
	if err != nil {
		return
	}

	// refresh
	err = a.rdb.Set(ctx, rdbRefreshPrefix+sub, refreshSecret, refreshTTL).Err()
	if err != nil {
		return
	}

	return
}

// LogOut разлогиниться - удалить текущую сессию и токены текущей сессии пользователя TODO: useless method (not used, but may be need it)
func (a *Auth) LogOut(ctx context.Context, sessionID string) error {
	err := a.rdb.Del(ctx, rdbTokenPrefix+sessionID).Err()
	if err != nil {
		return fmt.Errorf("could not delete token: %w", err)
	}

	err = a.rdb.Del(context.WithoutCancel(ctx), rdbRefreshPrefix+sessionID).Err()
	if err != nil {
		return fmt.Errorf("could not delete refresh-token: %w", err)
	}

	err = a.users.DeleteSession(context.WithoutCancel(ctx), sessionID)
	if err != nil {
		return fmt.Errorf("could not delte session: %w", err)
	}

	a.metrics.LogoutTotal.Inc()

	return nil
}

// GetKeyBySessionID получить секретный ключ токена для сессии пользователя
func (a *Auth) GetKeyBySessionID(ctx context.Context, sessionID string) (string, error) {
	key, err := a.rdb.Get(ctx, rdbTokenPrefix+sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrUserTokenKeyNotFound
		}

		return "", fmt.Errorf("could not get key from rdb: %w", err)
	}

	return key, nil
}

// hashPhone sha256 sum
func (a *Auth) hashPhone(phone, salt string) string {
	h := sha256.New()
	h.Write([]byte(phone + salt))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// genCode генерирует одноразовый код - 4 цифры
func (a *Auth) genCode() string {
	res := ""
	for i := 1; i <= authCodeLen; i++ {
		n, err := cr.Int(cr.Reader, big.NewInt(9))
		if err != nil {
			a.l.Error("can't generate code: %v", err)

			return ""
		}

		res += n.Text(10)
	}

	return res
}

// generateRandomKey генерирует рандомный и сильный ключ для токенов
func (a *Auth) generateRandomKey() string {
	const (
		n             = 10                                                     // key length
		letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // dictionary
		letterIdxBits = 6                                                      // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits                                     // # of letter indices fitting in 63 bits
	)

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
