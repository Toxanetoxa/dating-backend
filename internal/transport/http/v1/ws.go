package v1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/toxanetoxa/dating-backend/internal/entity"
	"github.com/toxanetoxa/dating-backend/internal/service"
	"github.com/toxanetoxa/dating-backend/internal/transport/http/dto"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	writeWait  = time.Second * 10
	pongWait   = time.Second * 60
	pingPeriod = (pongWait * 9) / 10
)

// todo: Это всё надо рефакторить и разносить по пакетам. тестируем пока так

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type WsMsgEvent struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type WsMsgPayload struct {
	MatchID     string `json:"matchID"`
	UserID      string `json:"userID"`
	ClientMsgID string `json:"clientMsgID"`
	Text        string `json:"text"`
}

type WsMsgPayloadFromServer struct {
	ID          string `json:"ID"`
	MatchID     string `json:"matchID"`
	UserID      string `json:"userID"`
	ClientMsgID string `json:"clientMsgID"`
	Text        string `json:"text"`
}

type Hub struct {
	l           *slog.Logger
	register    chan *Client
	unregister  chan *Client
	bus         chan *WsMsgPayload
	externalBus *chan *entity.WsMsgPayload
	users       map[string]*Client
	chatService service.ChatService
}

func NewHub(l *slog.Logger, s service.ChatService) *Hub {
	return &Hub{
		l:           l,
		register:    make(chan *Client),
		users:       make(map[string]*Client),
		bus:         make(chan *WsMsgPayload),
		unregister:  make(chan *Client),
		chatService: s,
		externalBus: s.GetWSEventBus(),
	}
}

func (h *Hub) Run() {
	defer close(h.register)
	defer close(h.bus)

	for {
		select {
		case c := <-h.register:
			h.users[c.userID] = c
			h.l.Debug("client registered", "id", c.userID)

		case u := <-h.unregister:
			h.l.Debug("client unregistered", "id", u.userID)
			_, ok := h.users[u.userID]
			if ok {
				delete(h.users, u.userID)
			}
			close(u.send)

		case msg := <-h.bus:
			h.l.Debug("received msg", "msg", msg)
			// prepare text (проверку/подготовку сообщения нужно вынести в отдельный метод и занести в юзкейс)
			textToSave := strings.Replace(msg.Text, "\n", "", -1)
			textToSave = strings.Replace(textToSave, "\r", "", -1)

			// get match users
			chatUsers, err := h.chatService.GetChatUsersByMatchID(context.TODO(), msg.MatchID)
			if err != nil {
				h.l.Error(err.Error())
				continue
			}

			toUser := chatUsers[0].ID
			if toUser == msg.UserID {
				toUser = chatUsers[1].ID
			}

			// отправка сохранит сообщение в базе и отправит пуш
			messageID, err := h.chatService.SendMessage(context.TODO(), msg.UserID, toUser, msg.MatchID, textToSave)
			if err != nil {
				h.l.Error(err.Error())
				continue
			}

			msgToSend := WsMsgPayloadFromServer{
				MatchID:     msg.MatchID,
				UserID:      msg.UserID,
				ClientMsgID: msg.ClientMsgID,
				ID:          messageID,
				Text:        textToSave,
			}

			for i := range chatUsers {
				u, ok := h.users[chatUsers[i].ID]
				if ok {
					// user is online
					h.l.Debug("try send to online user", "user_id", msgToSend.UserID)
					jsonData, _ := json.Marshal(msgToSend)
					data := base64.StdEncoding.EncodeToString(jsonData)
					u.send <- &entity.WsMsgEvent{Type: entity.WsEventTypeMsg, Payload: data}
				}
			}

		case msg := <-*h.externalBus:
			h.l.Debug("received msg (ext bus)", "msg", msg)

			u, ok := h.users[msg.ToUserID]
			if ok {
				// user is online
				h.l.Debug("try send to online send", "to", msg.ToUserID)
				event := entity.WsMsgEvent{Type: entity.WsEventTypeMsg}
				err := event.EncodePayload(msg)
				if err != nil {
					h.l.Error(err.Error())
				}

				u.send <- &event
			}

		}
	}
}

type Client struct {
	l      *slog.Logger
	userID string
	hub    *Hub
	send   chan *entity.WsMsgEvent
	conn   *websocket.Conn
}

func NewClient(l *slog.Logger, h *Hub, c *websocket.Conn, userID string) *Client {
	return &Client{
		l:      l,
		hub:    h,
		conn:   c,
		send:   make(chan *entity.WsMsgEvent),
		userID: userID,
	}
}

func (c *Client) Writer() {
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		//c.hub.unregister <- c.userID
		c.conn.Close()
	}()

	for {
		select {
		case m, ok := <-c.send:
			if !ok {
				// The hub closed the channel ?
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.l.Debug("send to client")
			b, _ := json.Marshal(m)
			err := c.conn.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				c.l.Error("write error: %v", err)
			}

		case <-pingTicker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

func (c *Client) Reader() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// считываем сообщение от клиента
		mt, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.l.Error("read error: %v", err)
			}
			break
		}

		// пока обрабатываем только текстовые сообщения
		if mt != websocket.TextMessage {
			continue
		}

		// биндим сообщение к общей структуре события
		wsMsgEvent := entity.WsMsgEvent{}
		err = json.Unmarshal(msg, &wsMsgEvent)
		if err != nil {
			c.l.Error(err.Error())
			continue
		}

		// проверяем тип события. пока только сообщения чата
		switch wsMsgEvent.Type {
		case "msg":
			// приводим нагрузку события к структуре сообщения
			buf1, err := base64.StdEncoding.DecodeString(wsMsgEvent.Payload)
			if err != nil {
				c.l.Error(err.Error())
				continue
			}
			wsMsg := WsMsgPayload{}
			err = json.Unmarshal(buf1, &wsMsg)
			if err != nil {
				c.l.Error(err.Error())
				continue
			}
			wsMsg.UserID = c.userID
			// и отправляем сообщение в "хаб"
			c.hub.bus <- &wsMsg
		default:
			// нужно подумать как обрабатывать невалидные евенты (слать ли что в ответ ?)
			c.l.Error("invalid msg type")
			continue
		}
	}
}

// initWs инициализация соединения с клиентом по веб-сокету
func initWs(l *slog.Logger, hub *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			l.Error("upgrader error", "err", err.Error())
			return c.NoContent(http.StatusInternalServerError)
		}

		userID, _ := c.Get(dto.UserIDContextKey).(string)

		client := NewClient(l, hub, conn, userID)
		hub.register <- client

		go client.Writer()
		go client.Reader()

		return nil
	}
}

func setWsHandlers(g *echo.Group, l *slog.Logger, s service.ChatService) {
	hub := NewHub(l, s)
	go hub.Run()
	g.GET("", initWs(l, hub))
}
