package service

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"log/slog"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/pkg/pg"

	"github.com/google/uuid"
)

var (
	ErrMatchNotFound = errors.New("match not found")
	ErrNotAccess     = errors.New("not access to this chat")
	ErrEmptyMessage  = errors.New("empty message")
)

// Chat сервис чата
type Chat struct {
	l          *slog.Logger
	matchRepo  MatchRepo
	msgRepo    ChatRepo
	push       PushService
	WSEventBus chan *entity.WsMsgPayload
}

func NewChatService(l *slog.Logger, m MatchRepo, r ChatRepo, p PushService) ChatService {
	return &Chat{
		l:          l,
		matchRepo:  m,
		msgRepo:    r,
		push:       p,
		WSEventBus: make(chan *entity.WsMsgPayload, 10), // todo нужно неблокирующее решение !
	}
}

// ChatsList список чатов
func (c *Chat) ChatsList(ctx context.Context, userID string) ([]*entity.Chat, error) {
	return c.msgRepo.GetChats(ctx, userID)
}

// MessagesList список сообщений в чате (мэтче)
func (c *Chat) MessagesList(ctx context.Context, userID, matchID string, limit, offset int) (messages []*entity.ChatMessage, err error) {
	return c.msgRepo.GetMessagesByMatchID(ctx, userID, matchID, limit, offset)
}

// SaveMessage сохранить сообщение
func (c *Chat) SaveMessage(ctx context.Context, msg *entity.ChatMessage) error {
	return c.msgRepo.SaveMessage(ctx, msg)
}

// MarkAsRead отметить сообщение как прочитанное
func (c *Chat) MarkAsRead(ctx context.Context, id string) error {
	return c.msgRepo.MarkAsRead(ctx, id)
}

// MarkAsDelivered отметить сообщение как доставленное
func (c *Chat) MarkAsDelivered(ctx context.Context, id string) error {
	return c.msgRepo.MarkAsDelivered(ctx, id)
}

// GetChatUsersByMatchID получить участников чата (мэтча)
func (c *Chat) GetChatUsersByMatchID(ctx context.Context, matchID string) ([]*entity.User, error) {
	match, err := c.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("could not get match: %w", err)
	}

	u := []*entity.User{
		match.User1,
		match.User2,
	}

	return u, nil
}

func (c *Chat) GetWSEventBus() *chan *entity.WsMsgPayload {
	return &c.WSEventBus
}

// SendMessage Отправка сообщения в чат // todo обновить сигнатуру
func (c *Chat) SendMessage(ctx context.Context, fromUserID, toUserIDArg, matchID, text string) (string, error) {
	// проверка на пустое сообщение
	if len(text) == 0 {
		return "", ErrEmptyMessage
	}

	// достать мэтч
	match, err := c.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		if errors.Is(err, pg.ErrEntityDoesntExist) {
			return "", ErrMatchNotFound
		}

		return "", err
	}

	matchUsers := []string{match.User1ID, match.User2ID}

	// проверка что юзер есть в мэтче (что бы не залезть в чужой чат)
	if !slices.Contains(matchUsers, fromUserID) {
		return "", ErrNotAccess
	}

	_ = toUserIDArg // legacy

	toUserID := matchUsers[0]
	if fromUserID == toUserID {
		toUserID = matchUsers[1]
	}

	messageID := uuid.New().String()
	msg := entity.ChatMessage{
		GeneralTechFields: entity.GeneralTechFields{
			ID:        messageID,
			CreatedAt: time.Now(),
		},
		MatchID:     matchID,
		UserID:      fromUserID,
		Text:        text,
		IsRead:      true, // временное решение, пока нет события прочтения
		IsDelivered: true, // и это временное
	}

	err = c.msgRepo.SaveMessage(ctx, &msg)
	if err != nil {
		return "", fmt.Errorf("could not save message: %w", err)
	}

	// отправка пуш уведомления
	go func() {
		err := c.push.SendNewMessageNotify(context.WithoutCancel(ctx), toUserID, matchID, fromUserID)
		if err != nil {
			c.l.Error("could not send push",
				"error", err,
				"toUserID", toUserID,
				"messageID", messageID,
			)
		}
	}()

	// отправка события по веб-сокету
	c.WSEventBus <- &entity.WsMsgPayload{
		MatchID:     matchID,
		UserID:      fromUserID,
		ClientMsgID: "",
		Text:        text,
		ToUserID:    toUserID, // todo mb remove
	}

	// todo здесь заблокируется, когда буфер переполнится и пока кто-то не прочитает из канала

	return messageID, nil
}
