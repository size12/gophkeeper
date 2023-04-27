package storage

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/size12/gophkeeper/internal/entity"
)

type FileStorage struct {
	directory string
}

func NewFileStorage(directory string) *FileStorage {
	err := os.Mkdir(directory, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatalln("Failed open directory for file storage")
		return nil
	}
	return &FileStorage{directory: directory}
}

func (storage *FileStorage) GetRecord(ctx context.Context, recordID string) (entity.Record, error) {
	metadata, ok := ctx.Value("recordMetadata").(string)
	if !ok {
		log.Println("Failed get userID from context in getting all records")
		return entity.Record{}, ErrUserUnauthorized
	}

	file, err := os.Open(storage.directory + "/" + recordID)

	if errors.Is(err, os.ErrNotExist) {
		return entity.Record{}, ErrNotFound
	}

	if err != nil {
		return entity.Record{}, ErrUnknown
	}

	data, err := io.ReadAll(file)

	if err != nil {
		return entity.Record{}, ErrUnknown
	}

	record := entity.Record{
		ID:       recordID,
		Metadata: metadata,
		Type:     "FILE",
		Data:     data,
	}

	return record, nil
}

func (storage *FileStorage) DeleteRecord(_ context.Context, recordID string) error {
	err := os.RemoveAll(storage.directory + "/" + recordID)
	if errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}
	if err != nil {
		return ErrUnknown
	}

	return nil
}

func (storage *FileStorage) CreateRecord(_ context.Context, record entity.Record) (string, error) {
	file, err := os.Create(storage.directory + "/" + record.ID)
	if err != nil {
		return "", ErrUnknown
	}

	_, err = file.Write(record.Data)
	if err != nil {
		return "", ErrUnknown
	}

	return record.ID, nil
}
