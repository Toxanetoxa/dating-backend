package entity

type Chat struct {
	GeneralTechFields
	Profile     *User        `json:"profile"`
	LastMessage *ChatMessage `json:"lastMessage"`
}
