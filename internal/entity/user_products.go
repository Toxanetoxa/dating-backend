package entity

import (
	"time"
)

type UserProducts struct {
	UserID    string
	ProductID int
	Expire    time.Time
}

func (u *UserProducts) TableName() string {
	return "user_product"
}
