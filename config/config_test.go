package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {

	cfg := NewConfig()

	require.NotEmpty(t, cfg)
	require.NotEmpty(t, cfg)
}
