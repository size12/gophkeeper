package storage

import (
	"context"
	"errors"

	"github.com/size12/gophkeeper/internal/entity"
)

type Storage struct {
	DBStorage   *DBStorage
	FileStorage *FileStorage
}

func NewStorage(DBStorage *DBStorage, fileStorage *FileStorage) *Storage {
	return &Storage{
		DBStorage:   DBStorage,
		FileStorage: fileStorage,
	}
}

func (storage *Storage) CreateUser(credentials entity.UserCredentials) error {
	return storage.DBStorage.CreateUser(credentials)
}

func (storage *Storage) LoginUser(credentials entity.UserCredentials) (entity.UserID, error) {
	return storage.DBStorage.LoginUser(credentials)
}

func (storage *Storage) GetRecordsInfo(ctx context.Context) ([]entity.Record, error) {
	return storage.DBStorage.GetRecordsInfo(ctx)
}

func (storage *Storage) CreateRecord(ctx context.Context, record entity.Record) (string, error) {
	data := record.Data

	if record.Type == "FILE" {
		record.Data = nil
	}

	id, err := storage.DBStorage.CreateRecord(ctx, record)

	if err != nil {
		return "", err
	}

	if record.Type == "FILE" {
		record.ID = id
		record.Data = data
		_, err = storage.FileStorage.CreateRecord(ctx, record)
		return "", err
	}

	return id, nil
}

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

func (storage *Storage) GetRecord(ctx context.Context, recordID string) (entity.Record, error) {
	record, err := storage.DBStorage.GetRecord(ctx, recordID)
	if err != nil {
		return record, err
	}

	if record.Type == "FILE" {
		ctx = context.WithValue(ctx, "recordMetadata", record.Metadata)
		return storage.FileStorage.GetRecord(ctx, recordID)
	}

	return record, nil
}
