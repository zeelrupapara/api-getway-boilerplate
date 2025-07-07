package support

import (
	"fmt"
	"time"
)

type SupportType int32

const (
	SupportType_Zammad = 0
)

type Cfg struct {
	SupportType SupportType
	AccessToken string
	Url         string
}

type Tickets struct {
	Id        int
	Title     string
	Text      string
	Priority  string
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type InAppSupport interface {
	GetAllTickets() (any, error)
	GetTicket(ticketId int) (any, error)
	CreateTicket(t *map[string]interface{}) (any, error)
	UpdateTicket(ticketId int, t *map[string]interface{}) (any, error)
	GetTicketArticle(ticketId int) (any, error)
	CreateTicketArticle(t *map[string]interface{}) (any, error)
	GetAllGroups() (any, error)
}

func NewInAppSupport(cfg Cfg) (InAppSupport, error) {
	switch cfg.SupportType {
	case SupportType_Zammad:
		return NewZammadSupport(cfg.AccessToken, cfg.Url)
	default:
		return nil, fmt.Errorf("unsupported support type")
	}
}
