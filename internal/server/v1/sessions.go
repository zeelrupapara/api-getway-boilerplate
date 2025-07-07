// Developer: zeelrupapara@gmail.com
// Description: Session management for GreenLync boilerplate
package v1

import (
	"fmt"
	"strings"
	"time"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/errors"
	"greenlync-api-gateway/utils"

	"github.com/gofiber/fiber/v2"
)

// Online Session struct
type OnlineSession struct {
	// Id of the session
	Id string `json:"id"`
	// User Id
	UserId string `json:"user_id"`
	// User id
	ClientId int32 `json:"client_id"`
	// UserType
	UserType model.UserType `json:"user_type"`
	// Client's session_id
	SessionId string `json:"session_id"`
	// StartedAt
	StartedAt time.Time `json:"started_at"`
	// Ip Address
	IpAddress string `json:"ip_address"`
	// Full name
	FullName string `json:"full_name"`
	// Channel
	Channel model.ChannelType `json:"channel"`
}

// @Id				GetAllSessions
// @Description	Get All Sessions
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	{array}		OnlineSession
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/system/sessions/active [get]
func (s *HttpServer) GetAllSessions(c *fiber.Ctx) error {
	activeSessions, err := s.getOnlineSessions(c)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, activeSessions)
}

// @Id				DeleteSession
// @Description	Delete Session
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		204
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Param			session_id	path	string	true	"Session ID"
// @Router			/api/v1/system/sessions/active/{session_id} [DELETE]
func (s *HttpServer) DeleteSession(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return s.App.HttpResponseBadQueryParams(c, fmt.Errorf("empty id"))
	}

	cfg, ok := s.OAuth2.GetActiveSessionById(id)
	if !ok {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("the id you provide doesn't exist or wrong"))
	}

	s.killClientSession(cfg, SessionDescionnectionReason_SessionKilled)

	return s.App.HttpResponseNoContent(c)
}

// @Id				DeleteAllSessions
// @Description	Delete All Sessions
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		204
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/system/sessions/active [DELETE]
func (s *HttpServer) DeleteAllSessions(c *fiber.Ctx) error {
	sessions := s.OAuth2.GetActiveSessions()
	s.OAuth2.LogoutAll()
	s.Hub.DeleteAll()

	// TODO: Implement session logout events
	// Simple logout completion for boilerplate
	for i := range sessions {
		s.Log.Logger.Infof("Logged out session: %s", sessions[i].SessionId)
	}

	return s.App.HttpResponseNoContent(c)
}

// @Id				GetAllSessionsHistroy
// @Description	Get All Sessions
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	{array}		model.Session{}
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/system/sessions/history [get]
func (s *HttpServer) GetAllSessionsHistroy(c *fiber.Ctx) error {
	sessions := []*model.Session{}
	err := s.DB.Where("finished_at IS NOT NULL").Find(&sessions).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseOK(c, sessions)
}

// @Id				DeleteAllSessionsHistroy
// @Description	Delete All Sessions History
// @Tags			System
// @Accept			json
// @Produce		json
// @Success		200	{array}		model.Session{}
// @Failure		500	{object}	http.HttpResponse
// @Security		BearerAuth
// @Router			/api/v1/system/sessions/history [DELETE]
func (s *HttpServer) DeleteAllSessionsHistroy(c *fiber.Ctx) error {
	err := s.DB.Where("finished_at IS NOT NULL").Delete(&model.Session{}).Error
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	return s.App.HttpResponseNoContent(c)
}

// get active sessions from oauth2
func (s *HttpServer) getOnlineSessions(c *fiber.Ctx) ([]*OnlineSession, error) {
	client, ok := utils.GetClient(c)
	if !ok {
		return nil, errors.ErrCouldNotParseClientCfg
	}
	userRole := strings.ToLower(client.Scope)
	
	// jana: Fetch accounts (Admin sees all, Dealer sees only assigned users)
	var accounts []*model.User
	query := s.DB.Select("id", "username", "role")
	// Simplified query for boilerplate - no complex role restrictions

	err := query.Find(&accounts).Error
	if err != nil {
		return nil, err
	}

	// Simplified for boilerplate
	accountMap := make(map[int32]*model.User)
	for _, acc := range accounts {
		accountMap[acc.Id] = acc
	}

	// Fetch active sessions
	sessions := s.OAuth2.GetActiveSessions()
	newSessions := []*OnlineSession{}

	for _, v := range sessions {
		acc, exists := accountMap[v.ClientId]

		// jana: Only allow Admin or assigned Dealer to see the session
		if userRole == "admin" || exists {
			ua := utils.UserAgentParser(v.UserAgent)
			fullName := "Unknown User"
			userId := ""
			userType := model.UserType_User

			if exists {
				fullName = acc.FirstName + " " + acc.LastName
				userId = acc.Username
				userType = model.UserType_User // Simplified for boilerplate
			} else {
				continue
			}

			newSessions = append(newSessions, &OnlineSession{
				Id:        v.Id,
				UserId:    userId,
				ClientId:  v.ClientId,
				UserType:  userType,
				SessionId: v.SessionId,
				IpAddress: v.IpAddress,
				StartedAt: v.StartedAt,
				FullName:  fullName,
				Channel:   ua.Channel,
			})
		}
	}

	return newSessions, nil
}
