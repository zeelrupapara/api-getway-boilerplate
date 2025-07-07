// Developer: Saif Hamdan

package v1

import (
	"context"
	"fmt"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/cache"
	"greenlync-api-gateway/pkg/errors"
	"greenlync-api-gateway/pkg/manager"

	"github.com/gofiber/websocket/v2"
)

func (s *HttpServer) serveWS(c *websocket.Conn) {
	sessionId := c.Query("session_id")
	if len(sessionId) == 0 {
		c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("session_id %s", errors.RequiredParams)))
		c.Close()
		return
	}

	// check if the session valid
	ctx := context.Background()
	ok, err := s.OAuth2.VerifySessionId(ctx, sessionId)
	if err != nil {
		c.WriteJSON(s.App.WSResponseInternalServerErrorRequest(model.EventType_InternalError, err))
		c.Close()
		return
	}
	if !ok {
		c.WriteJSON(s.App.WSResponseUnauthorized(model.EventType_Unauthorized, errors.ErrInvalidSession))
		c.Close()
		return
	}

	// get the userInfo
	cfg, err := s.OAuth2.Inspect(ctx, cache.SessionsKey(sessionId))
	if err != nil {
		c.WriteJSON(s.App.WSResponseInternalServerErrorRequest(model.EventType_InternalError, err))
		c.Close()
		return
	}

	s.OAuth2.WSConnected(sessionId)

	// check if the sessionId is used
	client := &manager.Client{}
	if _, ok := s.Hub.Get(sessionId); ok {
		c.WriteJSON(s.App.WSResponseInternalServerErrorRequest(model.EventType_InternalError, errors.ErrSessionUsed))
		c.Close()
		return
	}
	rulesStr := s.Authz.Enforcer.GetFilteredNamedPolicy("p", 0, cfg.Scope)
	keywords := make(map[string]struct{})

	// [[role, resource, action]]
	for i := range rulesStr {
		if len(rulesStr[i]) >= 3 {
			keywords[fmt.Sprint(rulesStr[i][1], "_", rulesStr[i][2])] = struct{}{}
		} else {
			c.WriteJSON(s.App.WSResponseInternalServerErrorRequest(model.EventType_InternalError, errors.ErrInternalServerError))
			c.Close()
			return
		}
	}

	keywords[fmt.Sprint(cfg.ClientId)] = struct{}{}
	client = manager.NewClient(c, s.Hub, cfg.SessionId, cfg.ClientId, cfg.IpAddress, keywords, 100)

	// add new client to our Hub
	s.Hub.Store(client)
	client.Publisher.SetMaxPendingMessages(100)

	// start sending messages to the clint
	go client.WriteMessages()
	// start reciving messages from the client
	go client.ReadMessages()
	// TODO: Implement NATS client router for event-driven architecture
	// go s.NatsClientRouter(client)

	client.Listen()

	// Clean up on disconnect
	s.OAuth2.WSDisconnected(sessionId)
	// TODO: Implement market feed cleanup for boilerplate
	// TODO: Implement subscription cleanup for boilerplate
}

func (s *HttpServer) WSErrorHandler(err error) *model.Event {
	return s.App.WSResponseBadRequest(model.EventType_BadRequest, err)
}
