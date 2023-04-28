package storage

import (
	"context"

	"github.com/size12/gophkeeper/internal/entity"
)

// FileStorager interface for storage, which can storage files.
type FileStorager interface {
	GetRecord(ctx context.Context, recordID string) (entity.Record, error)
	CreateRecord(ctx context.Context, record entity.Record) (string, error)
	DeleteRecord(ctx context.Context, recordID string) error
}

// Storager interface for storage, which can storage only text data.
type Storager interface {
	CreateUser(credentials entity.UserCredentials) error
	LoginUser(credentials entity.UserCredentials) (entity.UserID, error)
	GetRecordsInfo(ctx context.Context) ([]entity.Record, error)
	FileStorager
}
