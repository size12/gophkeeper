package storage

import (
	"context"
	"testing"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewStorage(t *testing.T) {
	db := mocks.NewStorager(t)
	file := mocks.NewFileStorager(t)
	storage := NewStorage(db, file)
	assert.NotEmpty(t, storage)
}

func TestStorage_CreateUser(t *testing.T) {
	db := mocks.NewStorager(t)
	file := mocks.NewFileStorager(t)
	storage := NewStorage(db, file)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Create user",
			func() {
				db.On("CreateUser", mock.AnythingOfType("entity.UserCredentials")).Return(nil)
			},
			func() {
				storage.CreateUser(entity.UserCredentials{
					Login:    "login",
					Password: "password",
				})
				db.AssertExpectations(t)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestStorage_GetRecordsInfo(t *testing.T) {
	db := mocks.NewStorager(t)
	file := mocks.NewFileStorager(t)
	storage := NewStorage(db, file)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get all records info",
			func() {
				db.On("GetRecordsInfo", context.Background()).Return([]entity.Record{}, nil)
			},
			func() {
				storage.GetRecordsInfo(context.Background())
				db.AssertExpectations(t)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestStorage_LoginUser(t *testing.T) {
	db := mocks.NewStorager(t)
	file := mocks.NewFileStorager(t)
	storage := NewStorage(db, file)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Login user",
			func() {
				db.On("LoginUser", mock.AnythingOfType("entity.UserCredentials")).Return(entity.UserID(""), nil)
			},
			func() {
				storage.LoginUser(entity.UserCredentials{
					Login:    "login",
					Password: "password",
				})
				db.AssertExpectations(t)
				file.AssertExpectations(t)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestStorage_CreateRecord(t *testing.T) {
	db := mocks.NewStorager(t)
	file := mocks.NewFileStorager(t)
	storage := NewStorage(db, file)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Create text record",
			func() {
				db.On("CreateRecord", context.Background(), mock.AnythingOfType("entity.Record")).Return("", nil)
			},
			func() {
				storage.CreateRecord(context.Background(), entity.Record{
					ID:       "",
					Metadata: "",
					Type:     entity.TypeText,
					Data:     nil,
				})
				db.AssertExpectations(t)
				file.AssertExpectations(t)
			},
		},
		{
			"Create file record",
			func() {
				db.On("CreateRecord", context.Background(), mock.AnythingOfType("entity.Record")).Return("", nil)
				file.On("CreateRecord", context.Background(), mock.AnythingOfType("entity.Record")).Return("", nil)
			},
			func() {
				storage.CreateRecord(context.Background(), entity.Record{
					ID:       "",
					Metadata: "",
					Type:     entity.TypeFile,
					Data:     nil,
				})
				db.AssertExpectations(t)
				file.AssertExpectations(t)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestStorage_GetRecord(t *testing.T) {
	db := mocks.NewStorager(t)
	file := mocks.NewFileStorager(t)
	storage := NewStorage(db, file)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get file record",
			func() {
				db.On("GetRecord", context.Background(), "").Return(entity.Record{Type: entity.TypeFile}, nil)
				file.On("GetRecord", mock.AnythingOfType("*context.valueCtx"), "").Return(entity.Record{}, nil)
			},
			func() {
				storage.GetRecord(context.Background(), "")
				db.AssertExpectations(t)
				file.AssertExpectations(t)
			},
		},
		{
			"Get text record",
			func() {
				db.On("GetRecord", context.Background(), "").Return(entity.Record{}, nil)
			},
			func() {
				storage.GetRecord(context.Background(), "")
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestStorage_DeleteRecord(t *testing.T) {
	db := mocks.NewStorager(t)
	file := mocks.NewFileStorager(t)
	storage := NewStorage(db, file)

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Delete file record",
			func() {
				db.On("DeleteRecord", context.Background(), "").Return(nil)
				file.On("DeleteRecord", context.Background(), "").Return(nil)
			},
			func() {
				storage.DeleteRecord(context.Background(), "")
				db.AssertExpectations(t)
				file.AssertExpectations(t)
			},
		},
		{
			"Delete text record",
			func() {
				db.On("DeleteRecord", context.Background(), "").Return(entity.Record{}, nil)
			},
			func() {
				storage.DeleteRecord(context.Background(), "")
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}
