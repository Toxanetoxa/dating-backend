package repo

import (
	"context"
	"fmt"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/pkg/pg"
	"log/slog"

	"gorm.io/gorm"
)

type ChatRepoPg struct {
	db *gorm.DB
	l  *slog.Logger
}

func NewChatRepoPg(db *gorm.DB, l *slog.Logger) service.ChatRepo {
	return &ChatRepoPg{
		db: db,
		l:  l,
	}
}

// GetMessagesByMatchID массив сообщений из мэтча
func (c *ChatRepoPg) GetMessagesByMatchID(ctx context.Context, userID, matchID string, limit, offset int) (messages []*entity.ChatMessage, err error) {
	defer pg.ProcessDbError(&err)

	err = c.db.WithContext(ctx).Model(&entity.ChatMessage{}).Where("match_id = ?", matchID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&messages).Error
	if err != nil {
		return
	}

	// check // todo check user refactor
	var m *entity.Match
	err = c.db.WithContext(ctx).Model(&entity.Match{}).Where("id = ? AND (user_1_id = ? OR user_2_id = ?)", matchID, userID, userID).First(&m).Error

	return
}

// SaveMessage сохранить сообщение
func (c *ChatRepoPg) SaveMessage(ctx context.Context, msg *entity.ChatMessage) (err error) {
	defer pg.ProcessDbError(&err)

	// todo: добавить транзакцию !

	// check
	var m *entity.Match
	err = c.db.WithContext(ctx).Model(&entity.Match{}).Where("id = ? AND (user_1_id = ? OR user_2_id = ?)", msg.MatchID, msg.UserID, msg.UserID).First(&m).Error
	if err != nil {
		return
	}

	if !m.ChatInit {
		err = c.db.WithContext(ctx).Model(&entity.Match{}).Where("id = ?", m.ID).Update("chat_init", true).Error
		if err != nil {
			err = fmt.Errorf("could not update match: %w", err)
			return
		}
	}

	return c.db.WithContext(ctx).Create(&msg).Error
}

// MarkAsRead обновить статус прочитнного сообщения
func (c *ChatRepoPg) MarkAsRead(ctx context.Context, messageID string) (err error) {
	defer pg.ProcessDbError(&err)

	return c.db.WithContext(ctx).Model(&entity.ChatMessage{}).Where("id = ?", messageID).Update("is_read", true).Error
}

// MarkAsDelivered установить признак доставки сообщения
func (c *ChatRepoPg) MarkAsDelivered(ctx context.Context, messageID string) (err error) {
	defer pg.ProcessDbError(&err)

	return c.db.WithContext(ctx).Model(&entity.ChatMessage{}).Where("id = ?", messageID).Update("is_delivered", true).Error
}

// GetChats список чатов по id пользователя
func (c *ChatRepoPg) GetChats(ctx context.Context, userID string) (chats []*entity.Chat, err error) {
	// это нужно оптимизировать и уместить в 1 запрос
	var matches []*entity.Match

	err = c.db.WithContext(ctx).
		Model(&entity.Match{}).
		Preload("User1", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Photos", func(db *gorm.DB) *gorm.DB {
				return db.Order("is_main DESC")
			})
		}).
		Preload("User2", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Photos", func(db *gorm.DB) *gorm.DB {
				return db.Order("is_main DESC")
			})
		}).
		Joins("LEFT JOIN ( select message.match_id, MAX(message.created_at) as mm from message group by message.match_id) as m ON m.match_id = match.id").
		Where("match.user_1_id = ? OR match.user_2_id = ?", userID, userID).
		Where("match.chat_init").
		Order("m.mm DESC").
		Find(&matches).Error

	if err != nil {
		return
	}

	chats = make([]*entity.Chat, 0, len(matches))

	for i := range matches {
		var msg entity.ChatMessage

		profile := matches[i].User1
		if profile.ID == userID {
			profile = matches[i].User2
		}

		err = c.db.WithContext(ctx).Model(&entity.ChatMessage{}).Where("match_id = ?", matches[i].ID).Order("created_at DESC").First(&msg).Error

		if err != nil {
			return
		}

		chats = append(chats, &entity.Chat{
			GeneralTechFields: entity.GeneralTechFields{
				ID:        matches[i].ID,
				CreatedAt: matches[i].CreatedAt,
			},
			Profile:     profile,
			LastMessage: &msg,
		})
	}

	return
}
