package entity

import "time"

type UserSession struct {
	ID        string
	UserID    string
	DeviceID  string
	Ip        string
	CreatedAt time.Time
}

// TableName имя таблицы для gorm
func (u *UserSession) TableName() string {
	return "user_session"
}
