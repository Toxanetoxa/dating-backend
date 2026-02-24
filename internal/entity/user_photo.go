package entity

// UserPhoto ...
type UserPhoto struct {
	GeneralTechFields

	UserID string `json:"-"`
	URL    string `json:"URL"`
	IsMain bool   `json:"isMain"`
	Object string `json:"-"`
}

// TableName имя таблицы для gorm
func (u *UserPhoto) TableName() string {
	return "user_photo"
}
