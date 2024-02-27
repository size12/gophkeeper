package storage

import (
	"context"
	"errors"

	"github.com/size12/gophkeeper/internal/entity"
)

// Storage struct which saves to DB and file storage.
type Storage struct {
	DBStorage   Storager
	FileStorage FileStorager
}

// NewStorage returns new storage.
func NewStorage(DBStorage Storager, fileStorage FileStorager) *Storage {
	return &Storage{
		DBStorage:   DBStorage,
		FileStorage: fileStorage,
	}
}

// CreateUser creates new user and saves to DB storage.
func (storage *Storage) CreateUser(credentials entity.UserCredentials) error {
	return storage.DBStorage.CreateUser(credentials)
}

// LoginUser check user login using DB storage.
func (storage *Storage) LoginUser(credentials entity.UserCredentials) (entity.UserID, error) {
	return storage.DBStorage.LoginUser(credentials)
}

// GetRecordsInfo gets all records from user from DB storage.
func (storage *Storage) GetRecordsInfo(ctx context.Context) ([]entity.Record, error) {
	return storage.DBStorage.GetRecordsInfo(ctx)
}

// CreateRecord creates record, saves to DB. If record type is file, saves to file storage too.
func (storage *Storage) CreateRecord(ctx context.Context, record entity.Record) (string, error) {
	data := record.Data

	if record.Type == entity.TypeFile {
		record.Data = nil
	}

	id, err := storage.DBStorage.CreateRecord(ctx, record)
	if err != nil {
		return "", err
	}

	if record.Type == entity.TypeFile {
		record.ID = id
		record.Data = data
		_, err = storage.FileStorage.CreateRecord(ctx, record)
		return "", err
	}

	return id, nil
}

// DeleteRecord deletes record from DB storage. If record type is file, deletes from file storage too.
func (storage *Storage) DeleteRecord(ctx context.Context, recordID string) error {
	err := storage.DBStorage.DeleteRecord(ctx, recordID)
	if err != nil {
		return err
	}

	err = storage.FileStorage.DeleteRecord(ctx, recordID)

	if !errors.Is(err, ErrNotFound) && err != nil {
		return ErrUnknown
	}

	return nil
}

// GetRecord gets record from DB or file storage.
func (storage *Storage) GetRecord(ctx context.Context, recordID string) (entity.Record, error) {
	record, err := storage.DBStorage.GetRecord(ctx, recordID)
	if err != nil {
		return record, err
	}

	if record.Type == entity.TypeFile {
		ctx = context.WithValue(ctx, "recordMetadata", record.Metadata)
		return storage.FileStorage.GetRecord(ctx, recordID)
	}

	return record, nil
}
