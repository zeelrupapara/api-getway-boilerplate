package support

import (
	"github.com/AlessandroSechi/zammad-go"
)

type ZammadSupport struct {
	client *zammad.Client
	InAppSupport
}

func NewZammadSupport(accessToken, url string) (InAppSupport, error) {
	// Create a client instance
	client, err := zammad.NewClient(&zammad.Client{
		Token: accessToken,
		OAuth: "",
		Url:   url,
	})
	if err != nil {
		return nil, err
	}

	return &ZammadSupport{
		client: client,
	}, nil
}

func (z *ZammadSupport) GetAllTickets() (any, error) {
	tickets, err := z.client.TicketList()
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (z *ZammadSupport) GetTicket(ticketId int) (any, error) {
	ticket, err := z.client.TicketShow(ticketId)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (z *ZammadSupport) CreateTicket(t *map[string]interface{}) (any, error) {
	newTicket, err := z.client.TicketCreate(t)
	if err != nil {
		return nil, err
	}

	return newTicket, nil
}

func (z *ZammadSupport) UpdateTicket(ticketId int, t *map[string]interface{}) (any, error) {
	newTicket, err := z.client.TicketUpdate(ticketId, t)
	if err != nil {
		return nil, err
	}

	return newTicket, nil
}

func (z *ZammadSupport) GetTicketArticle(ticketId int) (any, error) {
	ticket, err := z.client.TicketArticleByTicket(ticketId)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (z *ZammadSupport) CreateTicketArticle(t *map[string]interface{}) (any, error) {
	ticket, err := z.client.TicketArticleCreate(t)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}
