package manager

import (
	"errors"
	model "greenlync-api-gateway/model/common/v1"

	"github.com/goccy/go-json"
)

const (
	DefaultMaxPendingMessages = 100
)

type Msg struct {
}

type Ctx struct {
	Client *Client
	Type   model.EventType
	Data   []byte
	// OriginalData []byte
	Event *model.Event // original event
}

func NewCtx(client *Client, ty model.EventType, data []byte, orgdata []byte) *Ctx {
	return &Ctx{
		Client: client,
		Type:   ty,
		Data:   data,
	}
}

func (c *Ctx) BodyParser(out interface{}) error {
	return json.Unmarshal(c.Data, out)
}

func (c *Ctx) SendEvent(e *model.Event) error {
	if c.Client != nil && c.Client.Live {
		e.Format = "3" // JsonMessage
		c.Client.Egress <- e
		return nil
	}
	return errors.New("connection is lost")
}

func (c *Ctx) WriteMessage(b []byte) error {
	if c.Client != nil && c.Client.Live {
		e := &model.Event{
			Type:    model.EventType_UserLogin,
			Payload: string(b),
			Format:  "1", // TextMessage
		}
		c.Client.Egress <- e
		return nil
	}
	return errors.New("connection is lost")
}

// write binary
func (c *Ctx) WriteBinary(b []byte) error {
	if c.Client != nil && c.Client.Live {
		e := &model.Event{
			Type:    model.EventType_UserLogin,
			Payload: string(b),
			Format:  "2", // BinaryMessage
		}
		c.Client.Egress <- e
		return nil
	}
	return errors.New("connection is lost")
}

func (c *Ctx) WriteTick(b []byte) error {
	if c.Client != nil && c.Client.Live {
		c.Client.Market <- b
		return nil
	}
	return errors.New("connection is lost")
}
