package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unsafe"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/pkg/pg"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenTTL = time.Hour * 48
	salt     = "!%:R_adf45674567"
)

var (
	ErrAdminNotFound   = errors.New("admin not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidToken    = errors.New("invalid token")
)

type AdminAuth struct {
	l    *slog.Logger
	repo AdminRepo
}

func NewAdminAuthService(l *slog.Logger, r AdminRepo) AdminAuthService {
	return &AdminAuth{
		l:    l,
		repo: r,
	}
}

func (a *AdminAuth) Login(ctx context.Context, login, password string) (string, error) {
	admin, err := a.repo.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pg.ErrEntityDoesntExist) {
			return "", ErrAdminNotFound
		}

		a.l.Error(err.Error())

		return "", err
	}

	if !validatePassword(password, admin.PasswordHash) {
		return "", ErrInvalidPassword
	}

	secret := a.generateRandomKey()
	expTime := time.Now().Add(tokenTTL)
	tokenID := uuid.New().String()

	// Create a new token object
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": tokenID,
		"iss": admin.ID,
		"exp": expTime.Unix(),
	})
	// сохранить секрет токена
	adminToken := &entity.AdminToken{
		GeneralTechFields: entity.GeneralTechFields{ID: tokenID},
		AdminID:           admin.ID,
		Secret:            secret,
		Expire:            expTime,
	}
	err = a.repo.SaveToken(ctx, adminToken)
	if err != nil {
		return "", err
	}
	// Sign and get the complete encoded token as a string using the secret
	accessToken, err := tokenObj.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (a *AdminAuth) ValidateToken(ctx context.Context, token string) (*entity.Admin, string, error) {
	jwToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		tid, err := token.Claims.GetSubject()
		if err != nil {
			return nil, ErrInvalidToken
		}

		uid, err := token.Claims.GetIssuer()
		if err != nil {
			return nil, ErrInvalidToken
		}

		userToken, err := a.repo.GetToken(ctx, uid, tid)
		if err != nil {
			if errors.Is(err, pg.ErrEntityDoesntExist) {
				return nil, ErrInvalidToken
			}
			a.l.Error(err.Error())

			return nil, err
		}

		return []byte(userToken.Secret), nil
	})

	if err != nil {
		return nil, "", ErrInvalidToken
	}

	adminID, err := jwToken.Claims.GetIssuer()
	if err != nil {
		return nil, "", err
	}

	tokenID, err := jwToken.Claims.GetSubject()
	if err != nil {
		return nil, "", err
	}

	admin, err := a.repo.GetByID(ctx, adminID)
	if err != nil {
		return nil, "", err
	}

	return admin, tokenID, nil
}

func validatePassword(password, hash string) bool {
	return hashPassword(password) == hash
}

func hashPassword(pass string) string {
	h := sha256.New()
	h.Write([]byte(pass + salt))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (a *AdminAuth) generateRandomKey() string {
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
