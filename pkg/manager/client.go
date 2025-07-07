// Developer: Saif Hamdan

package manager

import (
	"fmt"
	"time"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/memory"
	"greenlync-api-gateway/pkg/shortuuid"

	"github.com/goccy/go-json"
	"github.com/gofiber/websocket/v2"
)

// ClientList is a map used to help manage a map of clientss
type ClientMap map[string]*Client

// ClientList is a map used to help manage a map of clientss
type ClientList []*Client

// Client is a websocket client, basically a frontend visitor
type Client struct {
	// Id
	Id string
	// account id
	ClientId int32
	// client sessionId
	SessionId string
	// StartedAt
	StartedAt time.Time
	// Ip Address
	IpAddress string
	// the websocket connection
	Conn *websocket.Conn
	// manager is the manager used to manage the client
	Hub *Hub
	// egress is used to avoid concurrent writes on the WebSocket
	Egress chan *model.Event
	// // same as Egress but for market feed and it has limit buffer
	Market chan []byte
	// trigger services that run on a thread to close if it set to false
	// Live chan bool
	// keywords are important for broadcasting when they are met we will send the message to the client
	Keywords map[string]struct{}
	// Shutdown
	Shutdown chan struct{}
	//
	Live      bool
	Publisher *Publisher
	Storage   *memory.Storage
}

func NewClient(conn *websocket.Conn, hub *Hub, sessionId string, clientId int32, ipAddress string, keywords map[string]struct{}, marketCapcity int) *Client {
	c := &Client{
		Id:        shortuuid.New(),
		ClientId:  clientId,
		SessionId: sessionId,
		IpAddress: ipAddress,
		StartedAt: time.Now(),
		Conn:      conn,
		Hub:       hub,
		Egress:    make(chan *model.Event),
		Shutdown:  make(chan struct{}),
		Keywords:  keywords,
		Live:      true,
		Storage:   memory.New(),
	}

	p := NewPublisher(c)

	c.Publisher = p

	return c
}

func (c *Client) PongHandler(pongMsg string) error {
	// c.Hub.Log.Logger.Info("client pong sessionId:", c.SessionId)
	return c.Conn.SetReadDeadline(time.Now().Add(PongWait))
}

func (c *Client) MatchKeywords(str ...string) bool {
	for i := range str {
		if _, ok := c.Keywords[str[i]]; ok {
			return true
		}
	}
	return false
}

func (c *Client) close() {
	if c.Live {
		c.Live = false
		// notify app to close all current subscribitons
		close(c.Shutdown)
		// close(c.Egress)
		// close(c.Market)
		// close connection
		c.Conn.Close()
	}
}

func (c *Client) ReadMessages() {
	defer func(c *Client) {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
		c.Hub.Log.Logger.Infof("Clearing read thread for: %s essionId", c.SessionId)
		c.Hub.Delete(c.SessionId)
	}(c)

	for {
		// ReadMessage is used to read the next message in queue in the connection
		messageType, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				c.Hub.Log.Logger.Error("error reading message: %v", err)
			}
			return // the error was normal due to closing connection
		}

		// for WS clients that doesn't have built in Ping/Pong mechanism
		str := string(data)
		if messageType == websocket.PingMessage || str == "9" {
			c.Egress <- &model.Event{
				Format: model.PongMessage,
			}
			continue
		}

		// parse data bytes to Event
		event := &model.Event{}
		err = json.Unmarshal(data, event)
		if err != nil {
			c.Egress <- c.Hub.ErrorHandler(ErrInvalidEventObject)
			continue
		}

		payload, err := json.Marshal(event.Payload)
		if err != nil {
			c.Egress <- c.Hub.ErrorHandler(ErrInvalidEventObject)
			continue
		}

		ctx := &Ctx{
			Client: c,
			Type:   event.Type,
			Data:   payload,
			Event:  event,
		}
		if _, ok := c.Hub.RouterMap[event.Type]; ok {
			err = c.Hub.RouterMap[event.Type](ctx)
			if err != nil {
				c.Hub.Log.Logger.Error(err)
				break
			}
		} else {
			c.Egress <- c.Hub.ErrorHandler(ErrMethoNotAllowed)
			continue
		}
	}
}

func (c *Client) WriteMessages() {
	ticker := time.NewTicker(PingInterval)
	if err := c.Conn.SetReadDeadline(time.Now().Add(PongWait)); err != nil {
		c.Hub.Log.Logger.Error(err)
		return
	}

	c.Conn.SetPongHandler(c.PongHandler)

	defer func(c *Client) {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
		c.Hub.Log.Logger.Infof("Clearing write thread for: %s session_id", c.SessionId)
		c.Hub.Delete(c.SessionId)
		// ticker.Stop()
	}(c)

	for {
		select {
		case <-c.Shutdown:
			return
		case event := <-c.Egress: // Recieve Data from Egress Channel
			if err := c.write(event); err != nil {
				return
			}
		case <-ticker.C: // check the Client is connected
			if len(c.Egress) == 0 {
				if err := c.write(&model.Event{Payload: "", Format: model.PingMessage}); err != nil {
					return
				}
			} else {
				if err := c.Conn.SetReadDeadline(time.Now().Add(PongWait)); err != nil {
					c.Hub.Log.Logger.Error(err)
					return
				}
			}
		}
	}
}

func (c *Client) Listen() {
	<-c.Shutdown
}

func (c *Client) write(event *model.Event) error {
	var err error
	switch event.Format {
	case "1": // TextMessage
		err = c.Conn.WriteMessage(websocket.TextMessage, []byte(event.Payload))
	case "2": // BinaryMessage
		err = c.Conn.WriteMessage(websocket.BinaryMessage, []byte(event.Payload))
	case model.PingMessage:
		err = c.Conn.WriteMessage(websocket.PingMessage, []byte(event.Payload))
	case model.PongMessage:
		err = c.Conn.WriteMessage(websocket.TextMessage, []byte(`10`))
	case "3": // JsonMessage
		event.SessionId = c.SessionId
		err = c.Conn.WriteJSON(event)
	}
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
			c.Hub.Log.Logger.Error("error reading message: %v", err)
		}
		return err // the error was normal due to closing connection
	}
	return nil
}
