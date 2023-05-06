package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClientConfig(t *testing.T) {
	cfg := GetClientConfig()
	assert.NotEmpty(t, cfg)
}

func TestGetServerConfig(t *testing.T) {
	cfg := GetServerConfig()
	assert.NotEmpty(t, cfg)
}
