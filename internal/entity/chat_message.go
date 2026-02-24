package entity

type ChatMessage struct {
	GeneralTechFields
	MatchID     string `json:"matchID"`
	UserID      string `json:"UserID"`
	Text        string `json:"text"`
	IsRead      bool   `json:"isRead"`
	IsDelivered bool   `json:"isDelivered"`
}

func (l *ChatMessage) TableName() string {
	return "message"
}
