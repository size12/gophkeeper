package handlers

import (
	"context"
	"testing"

	"github.com/size12/gophkeeper/internal/config"
	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/handlers/mocks"
	"github.com/size12/gophkeeper/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUser(t *testing.T) {
	serverCfg := config.GetServerConfig()
	handlers := mocks.NewServerHandlers(t)
	server := NewServerConn(handlers)
	server.Run(context.Background(), serverCfg.RunAddress)
	defer server.Stop()

	client := NewClientConn(serverCfg.RunAddress)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Create user",
			func() {
				handlers.On("CreateUser", entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				}).Return(entity.AuthToken("token"), nil).Once()
			},
			func() {
				token, err := client.Register(entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				})
				assert.NoError(t, err)
				assert.Equal(t, "token", token)
			},
		},
		{
			"Create user, but server will return error",
			func() {
				handlers.On("CreateUser", entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				}).Return(entity.AuthToken(""), storage.ErrLoginExists).Once()
			},
			func() {
				token, err := client.Register(entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				})
				assert.Equal(t, storage.ErrLoginExists, err)
				assert.Empty(t, token)
			},
		},
		{
			"Create user, but server will return unknown error",
			func() {
				handlers.On("CreateUser", entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				}).Return(entity.AuthToken(""), storage.ErrUnknown).Once()
			},
			func() {
				token, err := client.Register(entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				})
				assert.Equal(t, storage.ErrUnknown, err)
				assert.Empty(t, token)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		handlers.AssertExpectations(t)
	}
}

func TestLoginUser(t *testing.T) {
	serverCfg := config.GetServerConfig()
	client := NewClientConn(serverCfg.RunAddress)

	handlers := mocks.NewServerHandlers(t)

	server := NewServerConn(handlers)
	server.Run(context.Background(), serverCfg.RunAddress)
	defer server.Stop()

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Login user",
			func() {
				handlers.On("LoginUser", entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				}).Return(entity.AuthToken("token"), nil).Once()
			},
			func() {
				token, err := client.Login(entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				})
				assert.NoError(t, err)
				assert.Equal(t, "token", token)
			},
		},
		{
			"Login user, but server will return error",
			func() {
				handlers.On("LoginUser", entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				}).Return(entity.AuthToken(""), storage.ErrWrongCredentials).Once()
			},
			func() {
				token, err := client.Login(entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				})
				assert.Equal(t, storage.ErrWrongCredentials, err)
				assert.Empty(t, token)
			},
		},
		{
			"Create user, but server will return unknown error",
			func() {
				handlers.On("LoginUser", entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				}).Return(entity.AuthToken(""), storage.ErrUnknown).Once()
			},
			func() {
				token, err := client.Login(entity.UserCredentials{
					Login:    "Login",
					Password: "Password",
				})
				assert.Equal(t, storage.ErrUnknown, err)
				assert.Empty(t, token)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		handlers.AssertExpectations(t)
	}
}

func TestGetRecordsInfo(t *testing.T) {
	serverCfg := config.GetServerConfig()
	client := NewClientConn(serverCfg.RunAddress)

	handlers := mocks.NewServerHandlers(t)

	server := NewServerConn(handlers)
	server.Run(context.Background(), serverCfg.RunAddress)
	defer server.Stop()

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get all records",
			func() {
				handlers.On("GetRecordsInfo", mock.AnythingOfType("*context.valueCtx")).
					Return([]entity.Record{}, nil).Once()
			},
			func() {
				_, err := client.GetRecordsInfo("token")
				assert.NoError(t, err)
			},
		},
		{
			"Get all records, but server will return error",
			func() {
				handlers.On("GetRecordsInfo", mock.AnythingOfType("*context.valueCtx")).
					Return([]entity.Record{}, storage.ErrUserUnauthorized).Once()
			},
			func() {
				_, err := client.GetRecordsInfo("token")
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
		{
			"Get all records, but server will return unknown error",
			func() {
				handlers.On("GetRecordsInfo", mock.AnythingOfType("*context.valueCtx")).
					Return([]entity.Record{}, storage.ErrUnknown).Once()
			},
			func() {
				_, err := client.GetRecordsInfo("token")
				assert.Equal(t, storage.ErrUnknown, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		handlers.AssertExpectations(t)
	}
}

func TestGetRecord(t *testing.T) {
	serverCfg := config.GetServerConfig()
	client := NewClientConn(serverCfg.RunAddress)

	handlers := mocks.NewServerHandlers(t)

	server := NewServerConn(handlers)
	server.Run(context.Background(), serverCfg.RunAddress)
	defer server.Stop()

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get record",
			func() {
				handlers.On("GetRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").
					Return(entity.Record{}, nil).Once()
			},
			func() {
				_, err := client.GetRecord("token", "recordID")
				assert.NoError(t, err)
			},
		},
		{
			"Get record, but not authenticated.",
			func() {
				handlers.On("GetRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").
					Return(entity.Record{}, storage.ErrUserUnauthorized).Once()
			},
			func() {
				_, err := client.GetRecord("token", "recordID")
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
		{
			"Get record, but wrong ID.",
			func() {
				handlers.On("GetRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").
					Return(entity.Record{}, storage.ErrNotFound).Once()
			},
			func() {
				_, err := client.GetRecord("token", "recordID")
				assert.Equal(t, storage.ErrNotFound, err)
			},
		},
		{
			"Get record, but unknown error.",
			func() {
				handlers.On("GetRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").
					Return(entity.Record{}, storage.ErrUnknown).Once()
			},
			func() {
				_, err := client.GetRecord("token", "recordID")
				assert.Equal(t, storage.ErrUnknown, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		handlers.AssertExpectations(t)
	}
}

func TestCreateRecord(t *testing.T) {
	serverCfg := config.GetServerConfig()
	client := NewClientConn(serverCfg.RunAddress)

	handlers := mocks.NewServerHandlers(t)

	server := NewServerConn(handlers)
	server.Run(context.Background(), serverCfg.RunAddress)
	defer server.Stop()

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Create record.",
			func() {
				handlers.On("CreateRecord", mock.AnythingOfType("*context.valueCtx"), entity.Record{}).Return(nil).Once()
			},
			func() {
				err := client.CreateRecord("token", entity.Record{})
				assert.NoError(t, err)
			},
		},
		{
			"Create record, but not authenticated.",
			func() {
				handlers.On("CreateRecord", mock.AnythingOfType("*context.valueCtx"), entity.Record{}).Return(storage.ErrUserUnauthorized).Once()
			},
			func() {
				err := client.CreateRecord("token", entity.Record{})
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
		{
			"Create record, but unknown error.",
			func() {
				handlers.On("CreateRecord", mock.AnythingOfType("*context.valueCtx"), entity.Record{}).Return(storage.ErrUnknown).Once()
			},
			func() {
				err := client.CreateRecord("token", entity.Record{})
				assert.Equal(t, storage.ErrUnknown, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		handlers.AssertExpectations(t)
	}
}

func TestDeleteRecord(t *testing.T) {
	serverCfg := config.GetServerConfig()
	client := NewClientConn(serverCfg.RunAddress)

	handlers := mocks.NewServerHandlers(t)

	server := NewServerConn(handlers)
	server.Run(context.Background(), serverCfg.RunAddress)
	defer server.Stop()

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Delete record.",
			func() {
				handlers.On("DeleteRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").Return(nil).Once()
			},
			func() {
				err := client.DeleteRecord("token", "recordID")
				assert.NoError(t, err)
			},
		},
		{
			"Delete record, but not authenticated.",
			func() {
				handlers.On("DeleteRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").Return(storage.ErrUserUnauthorized).Once()
			},
			func() {
				err := client.DeleteRecord("token", "recordID")
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
		{
			"Delete record, but not found.",
			func() {
				handlers.On("DeleteRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").Return(storage.ErrNotFound).Once()
			},
			func() {
				err := client.DeleteRecord("token", "recordID")
				assert.Equal(t, storage.ErrNotFound, err)
			},
		},
		{
			"Delete record, but unknown error.",
			func() {
				handlers.On("DeleteRecord", mock.AnythingOfType("*context.valueCtx"), "recordID").Return(storage.ErrUnknown).Once()
			},
			func() {
				err := client.DeleteRecord("token", "recordID")
				assert.Equal(t, storage.ErrUnknown, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		handlers.AssertExpectations(t)
	}
}
