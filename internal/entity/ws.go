package entity

import (
	"encoding/base64"
	"encoding/json"
)

type WsEventType string

type WsMsgEvent struct {
	Type    WsEventType `json:"type"`
	Payload string      `json:"payload"`
}

func (w *WsMsgEvent) EncodePayload(data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payload := base64.StdEncoding.EncodeToString(jsonData)
	w.Payload = payload

	return nil
}

type WsMsgPayload struct {
	MatchID     string `json:"matchID"`
	UserID      string `json:"userID"`
	ClientMsgID string `json:"clientMsgID"`
	Text        string `json:"text"`
	ToUserID    string `json:"-"`
}

type WsMsgPayloadFromServer struct {
	ID          string `json:"ID"`
	MatchID     string `json:"matchID"`
	UserID      string `json:"userID"`
	ClientMsgID string `json:"clientMsgID"`
	Text        string `json:"text"`
}

const (
	WsEventTypeMsg WsEventType = "msg"
)
