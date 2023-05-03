package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/gophkeeper/internal/config"
	"github.com/size12/gophkeeper/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewDBStorage(t *testing.T) {
	cfg := config.GetServerConfig()
	assert.NotPanics(t, func() {
		NewDBStorage(cfg.DBConnectionURL)
	})
}

func TestDBStorage_CreateUser(t *testing.T) {
	cfg := config.GetServerConfig()
	storage := NewDBStorage(cfg.DBConnectionURL)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	storage.DB = db

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Create user with good credentials (doesn't exists)",
			func() {
				mock.ExpectQuery(`SELECT COUNT(*) FROM users WHERE login = $1`).WithArgs("my_login").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec(`INSERT INTO users (login, password) VALUES ($1, $2)`).WithArgs("my_login", "my_password").WillReturnResult(sqlmock.NewResult(0, 1))
			},
			func() {
				err := storage.CreateUser(entity.UserCredentials{
					Login:    "my_login",
					Password: "my_password",
				})
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Create user with good credentials (doesn't exists), but DB will return error",
			func() {
				mock.ExpectQuery(`SELECT COUNT(*) FROM users WHERE login = $1`).WithArgs("my_login").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec(`INSERT INTO users (login, password) VALUES ($1, $2)`).WithArgs("my_login", "my_password").WillReturnError(errors.New("some DB error"))
			},
			func() {
				err := storage.CreateUser(entity.UserCredentials{
					Login:    "my_login",
					Password: "my_password",
				})
				assert.Equal(t, ErrUnknown, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Create user with good credentials (already exists)",
			func() {
				mock.ExpectQuery(`SELECT COUNT(*) FROM users WHERE login = $1`).WithArgs("my_login").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			func() {
				err := storage.CreateUser(entity.UserCredentials{
					Login:    "my_login",
					Password: "my_password",
				})
				assert.Equal(t, ErrLoginExists, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestDBStorage_LoginUser(t *testing.T) {
	cfg := config.GetServerConfig()
	storage := NewDBStorage(cfg.DBConnectionURL)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	storage.DB = db

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Login user with good credentials",
			func() {
				mock.ExpectQuery(`SELECT user_id FROM users WHERE login = $1 AND password = $2`).WithArgs("my_login", "my_password").WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("6584c88d-1bb4-4686-83be-925abb24fc20"))
			},
			func() {
				userID, err := storage.LoginUser(entity.UserCredentials{
					Login:    "my_login",
					Password: "my_password",
				})
				assert.NoError(t, err)
				assert.Equal(t, entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"), userID)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Login user with good credentials, but DB will return error",
			func() {
				mock.ExpectQuery(`SELECT user_id FROM users WHERE login = $1 AND password = $2`).WithArgs("my_login", "my_password").WillReturnError(errors.New("some DB error"))
			},
			func() {
				userID, err := storage.LoginUser(entity.UserCredentials{
					Login:    "my_login",
					Password: "my_password",
				})
				assert.Equal(t, ErrUnknown, err)
				assert.Equal(t, entity.UserID(""), userID)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Login user with bad credentials",
			func() {
				mock.ExpectQuery(`SELECT user_id FROM users WHERE login = $1 AND password = $2`).WithArgs("my_login", "my_password").WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
			},
			func() {
				userID, err := storage.LoginUser(entity.UserCredentials{
					Login:    "my_login",
					Password: "my_password",
				})
				assert.Equal(t, ErrWrongCredentials, err)
				assert.Equal(t, entity.UserID(""), userID)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestDBStorage_GetRecordsInfo(t *testing.T) {
	cfg := config.GetServerConfig()
	storage := NewDBStorage(cfg.DBConnectionURL)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	storage.DB = db

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get all info from unauthorized user",
			func() {},
			func() {
				records, err := storage.GetRecordsInfo(context.Background())
				assert.Equal(t, ErrUserUnauthorized, err)
				assert.Empty(t, records)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Get all info from authorized user",
			func() {
				mock.ExpectQuery("SELECT record_id, record_type, metadata FROM users_data WHERE user_id = $1").WithArgs("6584c88d-1bb4-4686-83be-925abb24fc20").
					WillReturnRows(sqlmock.NewRows([]string{"record_id", "record_type", "metadata"}).AddRow("1", entity.TypeLoginAndPassword, "login and password").AddRow("2", entity.TypeText, "custom text"))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				records, err := storage.GetRecordsInfo(ctx)
				assert.NoError(t, err)

				assert.Equal(t, []entity.Record{
					{
						ID:       "1",
						Type:     entity.TypeLoginAndPassword,
						Metadata: "login and password",
					},
					{
						ID:       "2",
						Type:     entity.TypeText,
						Metadata: "custom text",
					},
				}, records)

				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Get all info from authorized user, but DB will return error",
			func() {
				mock.ExpectQuery("SELECT record_id, record_type, metadata FROM users_data WHERE user_id = $1").WithArgs("6584c88d-1bb4-4686-83be-925abb24fc20").WillReturnError(errors.New("some DB error"))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				records, err := storage.GetRecordsInfo(ctx)
				assert.Equal(t, ErrUnknown, err)
				assert.Empty(t, records)
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestDBStorage_CreateRecord(t *testing.T) {
	cfg := config.GetServerConfig()
	storage := NewDBStorage(cfg.DBConnectionURL)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	storage.DB = db

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Create record with unauthorized user",
			func() {},
			func() {
				recordID, err := storage.CreateRecord(context.Background(), entity.Record{})
				assert.Equal(t, ErrUserUnauthorized, err)
				assert.Empty(t, recordID)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Create record with authorized user",
			func() {
				mock.ExpectQuery("INSERT INTO users_data (user_id, record_type, metadata, encoded_data) VALUES ($1, $2, $3, $4) RETURNING record_id").
					WithArgs("6584c88d-1bb4-4686-83be-925abb24fc20", entity.TypeText, "my text", hex.EncodeToString([]byte("hello!"))).
					WillReturnRows(sqlmock.NewRows([]string{"record_id"}).AddRow("1"))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				recordID, err := storage.CreateRecord(ctx, entity.Record{
					Metadata: "my text",
					Type:     entity.TypeText,
					Data:     []byte("hello!"),
				})
				assert.NoError(t, err)
				assert.Equal(t, "1", recordID)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Create record with authorized user, but DB will return error",
			func() {
				mock.ExpectQuery("INSERT INTO users_data (user_id, record_type, metadata, encoded_data) VALUES ($1, $2, $3, $4) RETURNING record_id").
					WithArgs("6584c88d-1bb4-4686-83be-925abb24fc20", entity.TypeText, "my text", hex.EncodeToString([]byte("hello!"))).
					WillReturnError(errors.New("some DB error"))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				recordID, err := storage.CreateRecord(ctx, entity.Record{
					Metadata: "my text",
					Type:     entity.TypeText,
					Data:     []byte("hello!"),
				})
				assert.Equal(t, ErrUnknown, err)
				assert.Empty(t, recordID)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestDBStorage_GetRecord(t *testing.T) {
	cfg := config.GetServerConfig()
	storage := NewDBStorage(cfg.DBConnectionURL)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	storage.DB = db

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Get record with unauthorized user",
			func() {},
			func() {
				record, err := storage.GetRecord(context.Background(), "1")
				assert.Equal(t, ErrUserUnauthorized, err)
				assert.Empty(t, record)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Get record with authorized user",
			func() {
				mock.ExpectQuery("SELECT record_id, record_type, metadata, encoded_data FROM users_data WHERE record_id = $1 AND user_id = $2").
					WithArgs("1", "6584c88d-1bb4-4686-83be-925abb24fc20").
					WillReturnRows(sqlmock.NewRows([]string{"record_id", "record_type", "metadata", "encoded_data"}).
						AddRow("1", entity.TypeText, "my text", hex.EncodeToString([]byte("hello!"))))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				record, err := storage.GetRecord(ctx, "1")
				assert.NoError(t, err)
				assert.Equal(t, entity.Record{
					ID:       "1",
					Metadata: "my text",
					Type:     entity.TypeText,
					Data:     []byte("hello!"),
				}, record)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Get non existed record with authorized user",
			func() {
				mock.ExpectQuery("SELECT record_id, record_type, metadata, encoded_data FROM users_data WHERE record_id = $1 AND user_id = $2").
					WithArgs("1", "6584c88d-1bb4-4686-83be-925abb24fc20").
					WillReturnRows(sqlmock.NewRows([]string{"record_id", "record_type", "metadata", "encoded_data"}))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				record, err := storage.GetRecord(ctx, "1")
				assert.Equal(t, ErrNotFound, err)
				assert.Empty(t, record)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Get record with authorized user, but DB will return error",
			func() {
				mock.ExpectQuery("SELECT record_id, record_type, metadata, encoded_data FROM users_data WHERE record_id = $1 AND user_id = $2").
					WithArgs("1", "6584c88d-1bb4-4686-83be-925abb24fc20").
					WillReturnError(errors.New("some DB error"))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				record, err := storage.GetRecord(ctx, "1")
				assert.Equal(t, ErrUnknown, err)
				assert.Empty(t, record)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}

func TestDBStorage_DeleteRecord(t *testing.T) {
	cfg := config.GetServerConfig()
	storage := NewDBStorage(cfg.DBConnectionURL)
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	storage.DB = db

	tc := []struct {
		name  string
		mock  func()
		valid func()
	}{
		{
			"Delete record with unauthorized user",
			func() {},
			func() {
				err := storage.DeleteRecord(context.Background(), "1")
				assert.Equal(t, ErrUserUnauthorized, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Delete record with authorized user",
			func() {
				mock.ExpectExec("DELETE FROM users_data WHERE record_id = $1 AND user_id = $2").
					WithArgs("1", "6584c88d-1bb4-4686-83be-925abb24fc20").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				err := storage.DeleteRecord(ctx, "1")
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Delete record with authorized user, but DB will return error",
			func() {
				mock.ExpectExec("DELETE FROM users_data WHERE record_id = $1 AND user_id = $2").
					WithArgs("1", "6584c88d-1bb4-4686-83be-925abb24fc20").
					WillReturnError(errors.New("some DB error"))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				err := storage.DeleteRecord(ctx, "1")
				assert.Equal(t, ErrUnknown, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			"Delete non existed record with authorized user",
			func() {
				mock.ExpectExec("DELETE FROM users_data WHERE record_id = $1 AND user_id = $2").
					WithArgs("1", "6584c88d-1bb4-4686-83be-925abb24fc20").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			func() {
				ctx := context.WithValue(context.Background(), "userID", entity.UserID("6584c88d-1bb4-4686-83be-925abb24fc20"))
				err := storage.DeleteRecord(ctx, "1")
				assert.Equal(t, ErrNotFound, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}

	for _, test := range tc {
		t.Log(test.name)
		test.mock()
		test.valid()
	}
}
