package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/storage"
)

type ServerHandlers struct {
	Storage       storage.Storager
	Authenticator Authenticator
}

func NewServerHandlers(s storage.Storager, a Authenticator) *ServerHandlers {
	return &ServerHandlers{Storage: s, Authenticator: a}
}

func (handlers *ServerHandlers) LoginUser(credentials entity.UserCredentials) (entity.AuthToken, error) {
	if credentials.Login == "" || credentials.Password == "" {
		return "", ErrFieldIsEmpty
	}

	credentials.Password = credentials.Login + credentials.Password

	sha := sha256.New()
	sha.Write([]byte(credentials.Password))
	credentials.Password = hex.EncodeToString(sha.Sum(nil))

	userID, err := handlers.Storage.LoginUser(credentials)

	if err != nil {
		return "", err
	}

	authToken, err := handlers.Authenticator.CreateToken(userID)

	if err != nil {
		return "", storage.ErrUnknown
	}

	return authToken, nil
}

func (handlers *ServerHandlers) CreateUser(credentials entity.UserCredentials) (entity.AuthToken, error) {
	if credentials.Login == "" || credentials.Password == "" {
		return "", ErrFieldIsEmpty
	}

	passwordHash := credentials.Login + credentials.Password

	sha := sha256.New()
	sha.Write([]byte(passwordHash))
	passwordHash = hex.EncodeToString(sha.Sum(nil))

	err := handlers.Storage.CreateUser(entity.UserCredentials{
		Login:    credentials.Login,
		Password: passwordHash,
	})

	if err != nil {
		return "", err
	}

	return handlers.LoginUser(credentials)
}

func (handlers *ServerHandlers) GetRecordsInfo(ctx context.Context) ([]entity.Record, error) {
	token, ok := ctx.Value("authToken").(entity.AuthToken)

	if !ok {
		return nil, storage.ErrUserUnauthorized
	}

	userID, err := handlers.Authenticator.ValidateToken(token)

	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, "userID", userID)
	return handlers.Storage.GetRecordsInfo(ctx)
}

func (handlers *ServerHandlers) GetRecord(ctx context.Context, recordID string) (entity.Record, error) {
	token, ok := ctx.Value("authToken").(entity.AuthToken)

	if !ok {
		return entity.Record{}, storage.ErrUserUnauthorized
	}

	userID, err := handlers.Authenticator.ValidateToken(token)

	if err != nil {
		return entity.Record{}, err
	}

	ctx = context.WithValue(ctx, "userID", userID)
	return handlers.Storage.GetRecord(ctx, recordID)
}

func (handlers *ServerHandlers) CreateRecord(ctx context.Context, record entity.Record) error {
	token, ok := ctx.Value("authToken").(entity.AuthToken)

	if !ok {
		return storage.ErrUserUnauthorized
	}

	userID, err := handlers.Authenticator.ValidateToken(token)

	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "userID", userID)
	_, err = handlers.Storage.CreateRecord(ctx, record)
	return err
}

func (handlers *ServerHandlers) DeleteRecord(ctx context.Context, recordID string) error {
	token, ok := ctx.Value("authToken").(entity.AuthToken)

	if !ok {
		return storage.ErrUserUnauthorized
	}

	userID, err := handlers.Authenticator.ValidateToken(token)

	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "userID", userID)
	return handlers.Storage.DeleteRecord(ctx, recordID)
}
