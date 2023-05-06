package handlers

import (
	"testing"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/handlers/mocks"
	"github.com/size12/gophkeeper/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewClientHandlers(t *testing.T) {
	conn := mocks.NewClientConn(t)
	handlers := NewClientHandlers(conn)
	assert.NotEmpty(t, handlers)
}

func TestClient_Register(t *testing.T) {
	conn := mocks.NewClientConn(t)
	handlers := NewClientHandlers(conn)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Register with good credentials",
			func() {
				conn.On("Register", entity.UserCredentials{
					Login:     "Login",
					Password:  "Password",
					MasterKey: []byte("hello"),
				}).Return("token", nil).Once()
			},
			func() {
				err := handlers.Register(entity.UserCredentials{
					Login:     "Login",
					Password:  "Password",
					MasterKey: []byte("hello"),
				})
				assert.NoError(t, err)
				assert.Equal(t, entity.AuthToken("token"), handlers.authToken)
				assert.Equal(t, []byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}, handlers.masterKey)
			},
		},
		{
			"Register with bad credentials",
			func() {},
			func() {
				handlers.authToken = ""
				err := handlers.Register(entity.UserCredentials{
					Login:     "",
					Password:  "",
					MasterKey: []byte("hello"),
				})
				assert.Equal(t, ErrFieldIsEmpty, err)
				assert.Empty(t, handlers.authToken)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		conn.AssertExpectations(t)
	}
}

func TestClient_Login(t *testing.T) {
	conn := mocks.NewClientConn(t)
	handlers := NewClientHandlers(conn)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Login with good credentials",
			func() {
				conn.On("Login", entity.UserCredentials{
					Login:     "Login",
					Password:  "Password",
					MasterKey: []byte("hello"),
				}).Return("token", nil).Once()
			},
			func() {
				err := handlers.Login(entity.UserCredentials{
					Login:     "Login",
					Password:  "Password",
					MasterKey: []byte("hello"),
				})
				assert.NoError(t, err)
				assert.Equal(t, entity.AuthToken("token"), handlers.authToken)
				assert.Equal(t, []byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}, handlers.masterKey)
			},
		},
		{
			"Login with bad credentials",
			func() {},
			func() {
				handlers.authToken = ""
				err := handlers.Login(entity.UserCredentials{
					Login:     "",
					Password:  "",
					MasterKey: []byte("hello"),
				})
				assert.Equal(t, ErrFieldIsEmpty, err)
				assert.Empty(t, handlers.authToken)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		conn.AssertExpectations(t)
	}
}

func TestClient_GetRecordsInfo(t *testing.T) {
	conn := mocks.NewClientConn(t)
	handlers := NewClientHandlers(conn)
	handlers.authToken = "token"

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get records info",
			func() {
				conn.On("GetRecordsInfo", entity.AuthToken("token")).Return([]entity.Record{}, nil).Once()
			},
			func() {
				records, err := handlers.GetRecordsInfo()
				assert.NoError(t, err)
				assert.Equal(t, []entity.Record{}, records)
			},
		},
		{
			"Get records info, but return error",
			func() {
				conn.On("GetRecordsInfo", entity.AuthToken("token")).Return([]entity.Record{}, storage.ErrUserUnauthorized).Once()
			},
			func() {
				records, err := handlers.GetRecordsInfo()
				assert.Equal(t, storage.ErrUserUnauthorized, err)
				assert.Equal(t, []entity.Record{}, records)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		conn.AssertExpectations(t)
	}
}

func TestClient_GetRecord(t *testing.T) {
	conn := mocks.NewClientConn(t)
	handlers := NewClientHandlers(conn)
	handlers.authToken = "token"
	handlers.masterKey = []byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get record",
			func() {
				conn.On("GetRecord", entity.AuthToken("token"), "1").Return(entity.Record{
					Data: []byte{0xcb, 0x1a, 0x6d, 0xb2, 0x12, 0xe2, 0x34, 0x9d, 0xf7, 0xe4, 0x2b, 0x9f, 0xa2, 0x9e, 0xd2, 0x12, 0x7, 0x2d, 0xa9, 0xff, 0xa, 0xd5, 0x88, 0x2b, 0x88, 0x6d, 0x61, 0x7, 0xf8, 0xd1, 0xc4, 0xf9, 0x17, 0xbc},
				}, nil).Once()
			},
			func() {
				record, err := handlers.GetRecord("1")
				assert.NoError(t, err)
				assert.Equal(t, entity.Record{
					Data: []byte("hello!"),
				}, record)
			},
		},
		{
			"Get record, but not found",
			func() {
				conn.On("GetRecord", entity.AuthToken("token"), "1").
					Return(entity.Record{}, storage.ErrNotFound).Once()
			},
			func() {
				record, err := handlers.GetRecord("1")
				assert.Equal(t, storage.ErrNotFound, err)
				assert.Empty(t, record)
			},
		},
		{
			"Get record, but unknown error",
			func() {
				conn.On("GetRecord", entity.AuthToken("token"), "1").
					Return(entity.Record{}, storage.ErrUnknown).Once()
			},
			func() {
				record, err := handlers.GetRecord("1")
				assert.Equal(t, storage.ErrUnknown, err)
				assert.Empty(t, record)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		conn.AssertExpectations(t)
	}
}

func TestClient_DeleteRecord(t *testing.T) {
	conn := mocks.NewClientConn(t)
	handlers := NewClientHandlers(conn)
	handlers.authToken = "token"
	handlers.masterKey = []byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Delete record",
			func() {
				conn.On("DeleteRecord", entity.AuthToken("token"), "1").Return(nil).Once()
			},
			func() {
				err := handlers.DeleteRecord("1")
				assert.NoError(t, err)
			},
		},
		{
			"Delete record, but will return error",
			func() {
				conn.On("DeleteRecord", entity.AuthToken("token"), "1").Return(storage.ErrNotFound).Once()
			},
			func() {
				err := handlers.DeleteRecord("1")
				assert.Equal(t, storage.ErrNotFound, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		conn.AssertExpectations(t)
	}
}

func TestClient_CreateRecord(t *testing.T) {
	conn := mocks.NewClientConn(t)
	handlers := NewClientHandlers(conn)
	handlers.authToken = "token"
	handlers.masterKey = []byte{0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55}

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Create record",
			func() {
				conn.On("CreateRecord", entity.AuthToken("token"), mock.AnythingOfType("entity.Record")).Return(nil).Once()
			},
			func() {
				err := handlers.CreateRecord(entity.Record{
					Data: []byte("hello!"),
				})
				assert.NoError(t, err)
			},
		},
		{
			"Create record, but user not authored",
			func() {
				conn.On("CreateRecord", entity.AuthToken("token"), mock.AnythingOfType("entity.Record")).Return(storage.ErrUserUnauthorized).Once()
			},
			func() {
				err := handlers.CreateRecord(entity.Record{
					Data: []byte("hello!"),
				})
				assert.Equal(t, storage.ErrUserUnauthorized, err)
			},
		},
		{
			"Create record, but return unknown error",
			func() {
				conn.On("CreateRecord", entity.AuthToken("token"), mock.AnythingOfType("entity.Record")).Return(storage.ErrUnknown).Once()
			},
			func() {
				err := handlers.CreateRecord(entity.Record{
					Data: []byte("hello!"),
				})
				assert.Equal(t, storage.ErrUnknown, err)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
		conn.AssertExpectations(t)
	}
}

func Test_GenerateRandom(t *testing.T) {
	bytes, err := generateRandom(12)
	assert.NoError(t, err)
	assert.NotEmpty(t, bytes)
	assert.Len(t, bytes, 12)
}
