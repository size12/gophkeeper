package storage

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/size12/gophkeeper/internal/entity"
)

// FileStorage keeps records on disk.
type FileStorage struct {
	directory string
}

// NewFileStorage returns new file storage.
func NewFileStorage(directory string) *FileStorage {
	err := os.Mkdir(directory, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatalln("Failed open directory for file storage")
		return nil
	}
	return &FileStorage{directory: directory}
}

// GetRecord reads file with record data.
func (storage *FileStorage) GetRecord(ctx context.Context, recordID string) (entity.Record, error) {
	metadata, ok := ctx.Value("recordMetadata").(string)
	if !ok {
		log.Println("Failed get record metadata from context in getting file record")
		return entity.Record{}, ErrUnknown
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
		Type:     entity.TypeFile,
		Data:     data,
	}

	return record, nil
}

// DeleteRecord deletes file with record data.
func (storage *FileStorage) DeleteRecord(_ context.Context, recordID string) error {
	filename := storage.directory + "/" + recordID
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}

	err := os.RemoveAll(filename)

	if err != nil {
		return ErrUnknown
	}

	return nil
}

// CreateRecord creates new file with record data.
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
