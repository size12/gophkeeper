package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/storage"
)

// ServerHandlers interface for server handlers
//
//go:generate mockery --name ServerHandlers
type ServerHandlers interface {
	LoginUser(credentials entity.UserCredentials) (entity.AuthToken, error)
	CreateUser(credentials entity.UserCredentials) (entity.AuthToken, error)
	GetRecordsInfo(ctx context.Context) ([]entity.Record, error)
	GetRecord(ctx context.Context, recordID string) (entity.Record, error)
	CreateRecord(ctx context.Context, record entity.Record) error
	DeleteRecord(ctx context.Context, recordID string) error
}

// Server struct for server handlers.
type Server struct {
	Storage       storage.Storager
	Authenticator Authenticator
}

// NewServerHandlers returns server handlers based on storage and authenticator.
func NewServerHandlers(s storage.Storager, a Authenticator) *Server {
	return &Server{Storage: s, Authenticator: a}
}

// LoginUser logins user by login and password.
func (handlers *Server) LoginUser(credentials entity.UserCredentials) (entity.AuthToken, error) {
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
		log.Println("Failed create authToken:", err)
		return "", storage.ErrUnknown
	}

	return authToken, nil
}

// CreateUser creates new user by login and password.
func (handlers *Server) CreateUser(credentials entity.UserCredentials) (entity.AuthToken, error) {
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

// GetRecordsInfo gets all records from storage.
func (handlers *Server) GetRecordsInfo(ctx context.Context) ([]entity.Record, error) {
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

// GetRecord get record from storage by ID.
func (handlers *Server) GetRecord(ctx context.Context, recordID string) (entity.Record, error) {
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

// CreateRecord added record to storage.
func (handlers *Server) CreateRecord(ctx context.Context, record entity.Record) error {
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

// DeleteRecord deletes record from storage.
func (handlers *Server) DeleteRecord(ctx context.Context, recordID string) error {
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
