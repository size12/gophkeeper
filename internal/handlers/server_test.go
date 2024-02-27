package handlers

import (
	"context"
	"testing"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/handlers/mocks"
	"github.com/size12/gophkeeper/internal/storage"
	storagemocks "github.com/size12/gophkeeper/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewServerHandlers(t *testing.T) {
	store := storagemocks.NewStorager(t)
	auth := mocks.NewAuthenticator(t)
	handlers := NewServerHandlers(store, auth)
	assert.NotEmpty(t, handlers)
}

func TestServer_CreateUser(t *testing.T) {
	store := storagemocks.NewStorager(t)
	auth := mocks.NewAuthenticator(t)
	handlers := NewServerHandlers(store, auth)

	tc := []struct {
		name string
		mock func()
		arg  entity.UserCredentials
		want error
	}{
		{
			"Create user with good credentials",
			func() {
				store.On("CreateUser", entity.UserCredentials{
					Login:    "admin",
					Password: "749f09bade8aca755660eeb17792da880218d4fbdc4e25fbec279d7fe9f65d70",
				}).Return(nil).Once()
				store.On("LoginUser", entity.UserCredentials{
					Login:    "admin",
					Password: "749f09bade8aca755660eeb17792da880218d4fbdc4e25fbec279d7fe9f65d70",
				}).Return(entity.UserID("userID"), nil).Once()
				auth.On("CreateToken", entity.UserID("userID")).Return(entity.AuthToken("token"), nil).Once()
			},
			entity.UserCredentials{
				Login:    "admin",
				Password: "password",
			},
			nil,
		},
		{
			"Create user with bad credentials",
			func() {},
			entity.UserCredentials{
				Login:    "",
				Password: "",
			},
			ErrFieldIsEmpty,
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		_, err := handlers.CreateUser(test.arg)
		assert.Equal(t, test.want, err)

		store.AssertExpectations(t)
		auth.AssertExpectations(t)
	}
}

func TestServer_LoginUser(t *testing.T) {
	store := storagemocks.NewStorager(t)
	auth := mocks.NewAuthenticator(t)
	handlers := NewServerHandlers(store, auth)

	tc := []struct {
		name string
		mock func()
		arg  entity.UserCredentials
		want error
	}{
		{
			"Login user with good credentials",
			func() {
				store.On("LoginUser", entity.UserCredentials{
					Login:    "admin",
					Password: "749f09bade8aca755660eeb17792da880218d4fbdc4e25fbec279d7fe9f65d70",
				}).Return(entity.UserID("userID"), nil).Once()
				auth.On("CreateToken", entity.UserID("userID")).Return(entity.AuthToken("token"), nil).Once()
			},
			entity.UserCredentials{
				Login:    "admin",
				Password: "password",
			},
			nil,
		},
		{
			"Login user with bad credentials",
			func() {},
			entity.UserCredentials{
				Login:    "",
				Password: "",
			},
			ErrFieldIsEmpty,
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		_, err := handlers.LoginUser(test.arg)
		assert.Equal(t, test.want, err)

		store.AssertExpectations(t)
		auth.AssertExpectations(t)
	}
}

func TestServer_GetRecordsInfo(t *testing.T) {
	store := storagemocks.NewStorager(t)
	auth := mocks.NewAuthenticator(t)
	handlers := NewServerHandlers(store, auth)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get all records with valid context",
			func() {
				store.On("GetRecordsInfo", mock.AnythingOfType("*context.valueCtx")).Return([]entity.Record{}, nil).Once()
				auth.On("ValidateToken", entity.AuthToken("token")).Return(entity.UserID("userID"), nil).Once()
			},
			func() {
				ctx := context.WithValue(context.Background(), "authToken", entity.AuthToken("token"))
				_, err := handlers.GetRecordsInfo(ctx)
				assert.NoError(t, err)
			},
		},
		{
			"Get all records with not valid context",
			func() {},
			func() {
				ctx := context.Background()
				_, err := handlers.GetRecordsInfo(ctx)
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()

		store.AssertExpectations(t)
		auth.AssertExpectations(t)
	}
}

func TestServer_GetRecord(t *testing.T) {
	store := storagemocks.NewStorager(t)
	auth := mocks.NewAuthenticator(t)
	handlers := NewServerHandlers(store, auth)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get record with valid context",
			func() {
				store.On("GetRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").Return(entity.Record{}, nil).Once()
				auth.On("ValidateToken", entity.AuthToken("token")).Return(entity.UserID("userID"), nil).Once()
			},
			func() {
				ctx := context.WithValue(context.Background(), "authToken", entity.AuthToken("token"))
				_, err := handlers.GetRecord(ctx, "recordID")
				assert.NoError(t, err)
			},
		},
		{
			"Get record with not valid context",
			func() {},
			func() {
				ctx := context.Background()
				_, err := handlers.GetRecord(ctx, "recordID")
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()

		store.AssertExpectations(t)
		auth.AssertExpectations(t)
	}
}

func TestServer_CreateRecord(t *testing.T) {
	store := storagemocks.NewStorager(t)
	auth := mocks.NewAuthenticator(t)
	handlers := NewServerHandlers(store, auth)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Create record with valid context",
			func() {
				store.On("CreateRecord", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("entity.Record")).Return("", nil).Once()
				auth.On("ValidateToken", entity.AuthToken("token")).Return(entity.UserID("userID"), nil).Once()
			},
			func() {
				ctx := context.WithValue(context.Background(), "authToken", entity.AuthToken("token"))
				err := handlers.CreateRecord(ctx, entity.Record{})
				assert.NoError(t, err)
			},
		},
		{
			"Create record with not valid context",
			func() {},
			func() {
				ctx := context.Background()
				_, err := handlers.GetRecord(ctx, "recordID")
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()

		store.AssertExpectations(t)
		auth.AssertExpectations(t)
	}
}

func TestServer_DeleteRecord(t *testing.T) {
	store := storagemocks.NewStorager(t)
	auth := mocks.NewAuthenticator(t)
	handlers := NewServerHandlers(store, auth)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Delete record with valid context",
			func() {
				store.On("DeleteRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").Return(nil).Once()
				auth.On("ValidateToken", entity.AuthToken("token")).Return(entity.UserID("userID"), nil).Once()
			},
			func() {
				ctx := context.WithValue(context.Background(), "authToken", entity.AuthToken("token"))
				err := handlers.DeleteRecord(ctx, "recordID")
				assert.NoError(t, err)
			},
		},
		{
			"Create record with not valid context",
			func() {},
			func() {
				ctx := context.Background()
				err := handlers.DeleteRecord(ctx, "recordID")
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()

		store.AssertExpectations(t)
		auth.AssertExpectations(t)
	}
}
