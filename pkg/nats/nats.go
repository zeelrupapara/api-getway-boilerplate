// By Emran A. Hamdan, Lead Architect
package nats

import (
	"fmt"
	"time"
	"greenlync-api-gateway/config"

	"github.com/nats-io/nats.go"
)

type Nats struct {
	NC *nats.Conn
}

func natsErrHandler(nc *nats.Conn, sub *nats.Subscription, natsErr error) {
	// fmt.Printf("error: %v\n", natsErr)
	if natsErr == nats.ErrSlowConsumer {
		_, _, err := sub.Pending()
		if err != nil {
			return
		}
		return
		// Log error, notify operations...
	}
	// check for other errors
}

func NewNatClient(cfg *config.Config) (*Nats, error) {
	// Connect to nats server
	url := fmt.Sprint(cfg.Nats.Host + ":" + cfg.Nats.Port)
	fmt.Printf("Connecting to Nats  on %s \n", url)

	nc, err := nats.Connect(url, nats.Timeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	nc.SetErrorHandler(natsErrHandler)

	return &Nats{NC: nc}, nil
}
