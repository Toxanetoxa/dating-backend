package entity

// Match ...
type Match struct {
	GeneralTechFields
	ChatInit    bool
	User1ID     string `gorm:"column:user_1_id"`
	User2ID     string `gorm:"column:user_2_id"`
	User1       *User  `gorm:"foreignKey:User1ID"`
	User2       *User  `gorm:"foreignKey:User2ID"`
	LastMessage *ChatMessage
}

// TableName имя таблицы для gorm
func (l *Match) TableName() string {
	return "match"
}
