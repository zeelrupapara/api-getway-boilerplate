package manager

import (
	"time"
	model "greenlync-api-gateway/model/common/v1"
)

type Publisher struct {
	c      *Client
	writer chan *model.Event
	// cleaner chan struct{}
}

func NewPublisher(c *Client) *Publisher {
	p := &Publisher{
		c:      c,
		writer: make(chan *model.Event, DefaultMaxPendingMessages),
		// cleaner: make(chan struct{}),
	}

	go p.Start()

	return p
}

func (p *Publisher) Start() {

	// go p.autoDrain()
	for {
		select {
		case <-p.c.Shutdown:
			return
		case msg := <-p.writer:
			p.c.Egress <- msg
		default:
			time.Sleep(time.Millisecond * 2)
			// continue // slow consumer
		}
	}
}

func (p *Publisher) SetMaxPendingMessages(pendingMessages int) {
	// p.writer = make(chan *Event, pendingMessages)
}

func (p *Publisher) Publish(event *model.Event) {
	// p.cleaner <- struct{}{}
	if len(p.writer) < cap(p.writer) {
		p.writer <- event
	}
}

// func (p *Publisher) autoDrain() {
// 	for {
// 		<-p.cleaner
// 		p.drain(25)
// 	}
// }

// func (p *Publisher) drain(percentage int) {
// 	if len(p.writer) == (cap(p.writer)) {
// 		for i := 0; i < ((len(p.writer)) - len(p.writer)*percentage/100); i++ {
// 			<-p.writer
// 		}
// 	} // drain when the the buffered channel goes above percentage of it's capcity
// }
