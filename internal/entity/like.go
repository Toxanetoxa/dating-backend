package entity

// Like ...
type Like struct {
	GeneralTechFields

	FromUserID string
	ToUserID   string
	FromUser   User `gorm:"foreignKey:FromUserID"`
	ToUser     User `gorm:"foreignKey:ToUserID"`
}

func (l *Like) TableName() string {
	return "user_like"
}
