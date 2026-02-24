package entity

// Dislike ...
type Dislike struct {
	GeneralTechFields

	FromUserID string
	ToUserID   string
}

func (l *Dislike) TableName() string {
	return "user_dislike"
}
