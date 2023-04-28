package handlers

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/storage"
)

// Authenticator is interface for user authenticating. Should can creates tokens, and gets userIDs from them.
type Authenticator interface {
	CreateToken(userID entity.UserID) (entity.AuthToken, error)
	ValidateToken(token entity.AuthToken) (entity.UserID, error)
}

// AuthenticatorJWT is authenticator which uses JWT.
type AuthenticatorJWT struct {
	secretKey []byte
}

// NewAuthenticatorJWT gets new AuthenticatorJWT.
func NewAuthenticatorJWT(secretKey []byte) *AuthenticatorJWT {
	return &AuthenticatorJWT{secretKey: secretKey}
}

// CreateToken implementation of Authenticator interface. Creates token, which stores userID.
func (auth *AuthenticatorJWT) CreateToken(userID entity.UserID) (entity.AuthToken, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(1 * time.Hour).Unix() // TODO: get this value from config.
	claims["userID"] = userID

	tokenString, err := token.SignedString(auth.secretKey)

	if err != nil {
		log.Println("Failed generate token for authentication:", err)
		return "", storage.ErrUnknown
	}

	return entity.AuthToken(tokenString), nil
}

// ValidateToken implementation of Authenticator interface. Validates token, returns userID.
func (auth *AuthenticatorJWT) ValidateToken(token entity.AuthToken) (entity.UserID, error) {
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(string(token), claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, storage.ErrUnknown
		}
		return auth.secretKey, nil
	})

	if err != nil {
		return "", storage.ErrUserUnauthorized
	}

	userID, ok := claims["userID"].(string)

	if !ok {
		return "", storage.ErrUserUnauthorized
	}

	return entity.UserID(userID), nil
}
