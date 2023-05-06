package handlers

import (
	"testing"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthenticatorJWT(t *testing.T) {
	auth := NewAuthenticatorJWT([]byte("secret key"))
	assert.NotEmpty(t, auth)
}

func TestAuthenticatorJWT(t *testing.T) {
	auth := NewAuthenticatorJWT([]byte("secret key"))

	userID := entity.UserID("user_id_12")

	token, err := auth.CreateToken(userID)
	assert.NoError(t, err)

	id, err := auth.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, id)
}
