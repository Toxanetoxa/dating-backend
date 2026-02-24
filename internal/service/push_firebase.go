package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/toxanetoxa/dating-backend/internal/entity"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/api/option"
)

var (
	ErrCouldNotSendMessage = errors.New("could not send message")
	ErrInitFirebase        = errors.New("error initializing firebase app")
	ErrGettingFCMClient    = errors.New("error getting firebase Messaging client")
)

const (
	FCMResponseEntityNotFound = "Requested entity was not found."

	notifyTypeNewLike        = "newLike"
	notifyTypeNewMatch       = "newMatch"
	notifyTypeNewMessage     = "newMessage"
	notifyTypeNewInformation = "newInformation"
	notifyTypePaymentSuccess = "paymentSuccess"

	matchTitle                     = "У вас новая пара!"
	likeDescription                = "Посмотрите, взаимно ли это"
	successPaymentTitle            = "Покупка успешно оплачена"
	successPaymentBoostDescription = "Буст активирован"
)

type PushFirebase struct {
	l      *slog.Logger
	client *messaging.Client
	repo   UserRepo
	enable bool
}

func NewPushFirebaseService(l *slog.Logger, enable bool, jsonCredentials string, repo UserRepo) (PushService, error) {
	if !enable {
		// create fake push sender
		return &PushFirebase{
			l:      l,
			repo:   repo,
			enable: enable,
		}, nil
	}

	opt := option.WithCredentialsFile(jsonCredentials)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		l.Error("error initializing app: %v", err)

		return nil, errors.Join(ErrInitFirebase, err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		l.Error("error getting Messaging client: %v", err)

		return nil, errors.Join(ErrGettingFCMClient, err)
	}

	return &PushFirebase{
		l:      l,
		client: client,
		enable: enable,
		repo:   repo,
	}, nil
}

func (p *PushFirebase) send(ctx context.Context, userID string, data map[string]string) error {
	// fake send
	if !p.enable {
		p.l.Info("fake send push to user", "user_id", userID, "data", data)
		return nil
	}

	// get user's tokens
	tokens, err := p.repo.GetDeviceIDs(ctx, userID)
	if err != nil {
		p.l.Error("could not get user device ids from repo",
			slog.String("error", err.Error()))

		return errors.Join(ErrCouldNotSendMessage, err)
	}

	// for each device-id
	for i := range tokens {
		if tokens[i] == "" {
			p.l.Debug("empty token")
			continue
		}

		p.l.Debug("try send to device_id",
			slog.String("token", tokens[i]))
		// send to device:
		message := &messaging.Message{
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			Token: tokens[i],
		}

		response, err := p.client.Send(ctx, message)
		if err != nil {
			p.l.Error(err.Error(), "FCM response", response)

			//нужно удалить токен, если он вдруг стал невалидный
			if response == FCMResponseEntityNotFound {
				errDeleting := p.repo.DeleteDeviceID(context.TODO(), tokens[i])
				if errDeleting != nil {
					p.l.Error("deleting device id error",
						slog.String("error", errDeleting.Error()))
				}
			}
		}

		p.l.Debug("FCM response",
			slog.String("response", response))
	}

	return nil
}

func (p *PushFirebase) SendNewMatchNotify(ctx context.Context, userID, matchID, matchUserID string) error {

	user, err := p.repo.GetByID(ctx, matchUserID)
	if err != nil {
		return err
	}

	name := "Пользователь"
	if user.FirstName != nil {
		name = *user.FirstName
	}

	matchDescription := fmt.Sprintf("Между вами и пользователем %s пролетела Искра", name)

	data := map[string]string{
		"notifyType":  notifyTypeNewMatch,
		"matchID":     matchID,
		"title":       matchTitle,
		"description": matchDescription,
	}

	return p.send(ctx, userID, data)
}

func (p *PushFirebase) SendNewLikeNotify(ctx context.Context, userID, likeUserID string) error {

	user, err := p.repo.GetByID(ctx, likeUserID)
	if err != nil {
		return err
	}

	name := "Пользователь"
	if user.FirstName != nil {
		name = *user.FirstName
	}

	verb := "поставил"
	if user.Sex != nil && *user.Sex == entity.UserSexFemale {
		verb = "поставила"
	}

	likeTitle := fmt.Sprintf("%s %s вам лайк", name, verb)

	data := map[string]string{
		"notifyType":  notifyTypeNewLike,
		"userID":      likeUserID,
		"title":       likeTitle,
		"description": likeDescription,
	}

	return p.send(ctx, userID, data)
}

func (p *PushFirebase) SendNewMessageNotify(ctx context.Context, userID, matchID, fromUserID string) error {

	user, err := p.repo.GetByID(ctx, fromUserID)
	if err != nil {
		return err
	}

	name := "Пользователь"
	if user.FirstName != nil {
		name = *user.FirstName
	}

	verb := "написал"
	pronoun := "ему"
	if user.Sex != nil && *user.Sex == entity.UserSexFemale {
		verb = "написала"
		pronoun = "ей"
	}

	messageTitle := fmt.Sprintf("Вам %s %s", verb, name)

	messageDescription := fmt.Sprintf("Скорее ответьте %s", pronoun)

	profileJson, err := jsoniter.MarshalToString(user)
	if err != nil {

		return err
	}

	data := map[string]string{
		"notifyType":  notifyTypeNewMessage,
		"matchID":     matchID,
		"title":       messageTitle,
		"description": messageDescription,
		"profile":     profileJson,
	}

	return p.send(ctx, userID, data)
}

func (p *PushFirebase) SendNewInfoNotify(ctx context.Context, userID, title, description string) error {
	data := map[string]string{
		"notifyType":  notifyTypeNewInformation,
		"title":       title,
		"description": description,
	}

	return p.send(ctx, userID, data)
}

// SendPaymentSuccess успешная покупка, пока только буста
func (p *PushFirebase) SendPaymentSuccess(ctx context.Context, userID string, productID int) error {
	data := map[string]string{
		"notifyType":  notifyTypePaymentSuccess,
		"title":       successPaymentTitle,
		"description": successPaymentBoostDescription, // для разных продуктов может быть разный
		"productID":   fmt.Sprintf("%d", productID),
	}

	return p.send(ctx, userID, data)
}
