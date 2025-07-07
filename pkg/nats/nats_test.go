// Developer: Saif Hamdan
// Date: 18/7/2023

package nats

import (
	"testing"
	"greenlync-api-gateway/config"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestNats(t *testing.T) {
	testCases := []struct {
		NatsHost      string
		NatsPort      string
		CheckResponse func(t *testing.T, red *Nats, err error)
	}{
		{
			NatsHost: "nats://127.0.0.1",
			NatsPort: "4222",
			CheckResponse: func(t *testing.T, n *Nats, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, n)
			},
		},
		{
			NatsHost: "nats://127.0.0.1",
			NatsPort: "4223",
			CheckResponse: func(t *testing.T, n *Nats, err error) {
				require.EqualError(t, err, nats.ErrNoServers.Error())
				require.Nil(t, n)
			},
		},
	}

	for i := range testCases {
		nats, err := NewNatClient(&config.Config{
			Nats: config.Nats{
				Host: testCases[i].NatsHost,
				Port: testCases[i].NatsPort,
			},
		})
		testCases[i].CheckResponse(t, nats, err)
	}
}
