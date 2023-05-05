package handlers

import (
	"context"
	"log"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/storage"
	pb "github.com/size12/gophkeeper/protocols/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ClientConn describes client connection.
//
//go:generate mockery --name ClientConn
type ClientConn interface {
	Login(credentials entity.UserCredentials) (string, error)
	Register(credentials entity.UserCredentials) (string, error)
	GetRecordsInfo(token entity.AuthToken) ([]entity.Record, error)
	GetRecord(token entity.AuthToken, recordID string) (entity.Record, error)
	DeleteRecord(token entity.AuthToken, recordID string) error
	CreateRecord(token entity.AuthToken, record entity.Record) error
}

// ClientConnGPRC keeps connection with server. Uses gRPC.
type ClientConnGPRC struct {
	pb.GophkeeperClient
}

// NewClientConn connects to server and returning connection.
func NewClientConn(serverAddress string) *ClientConnGPRC {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	c := pb.NewGophkeeperClient(conn)
	return &ClientConnGPRC{GophkeeperClient: c}
}

// Login logins user by login and password.
func (conn *ClientConnGPRC) Login(credentials entity.UserCredentials) (string, error) {
	session, err := conn.GophkeeperClient.Login(context.Background(), &pb.UserCredentials{
		Login:    credentials.Login,
		Password: credentials.Password,
	})

	code := status.Code(err)

	if code == codes.Unauthenticated {
		return "", storage.ErrWrongCredentials
	}

	if code == codes.Internal {
		return "", storage.ErrUnknown
	}

	if code == codes.InvalidArgument {
		return "", ErrFieldIsEmpty
	}

	if err != nil {
		return "", err
	}

	return session.SessionToken, nil
}

// Register creates new user by login and password.
func (conn *ClientConnGPRC) Register(credentials entity.UserCredentials) (string, error) {
	session, err := conn.GophkeeperClient.Register(context.Background(), &pb.UserCredentials{
		Login:    credentials.Login,
		Password: credentials.Password,
	})

	code := status.Code(err)

	switch code {
	case codes.AlreadyExists:
		return "", storage.ErrLoginExists
	case codes.Internal:
		return "", storage.ErrUnknown
	case codes.InvalidArgument:
		return "", ErrFieldIsEmpty
	}

	return session.SessionToken, nil
}

// GetRecordsInfo gets all record.
func (conn *ClientConnGPRC) GetRecordsInfo(token entity.AuthToken) ([]entity.Record, error) {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authToken", string(token))

	gotRecords, err := conn.GophkeeperClient.GetRecordsInfo(ctx, &emptypb.Empty{})

	code := status.Code(err)

	switch code {
	case codes.Internal:
		return nil, storage.ErrUnknown
	case codes.Unauthenticated:
		return nil, storage.ErrUserUnauthorized
	}

	records := make([]entity.Record, 0, len(gotRecords.Records))

	for _, record := range gotRecords.Records {
		records = append(records, entity.Record{
			ID:       record.Id,
			Metadata: record.Metadata,
			Type:     entity.RecordType(record.Type),
		})
	}

	return records, nil
}

// GetRecord gets record from server by ID.
func (conn *ClientConnGPRC) GetRecord(token entity.AuthToken, recordID string) (entity.Record, error) {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authToken", string(token))

	gotRecord, err := conn.GophkeeperClient.GetRecord(ctx, &pb.RecordID{Id: recordID})

	record := entity.Record{}

	code := status.Code(err)

	switch code {
	case codes.Internal:
		return record, storage.ErrUnknown
	case codes.Unauthenticated:
		return record, storage.ErrUserUnauthorized
	case codes.NotFound:
		return record, storage.ErrNotFound
	}

	record = entity.Record{
		ID:       gotRecord.Id,
		Metadata: gotRecord.Metadata,
		Type:     entity.RecordType(gotRecord.Type),
		Data:     gotRecord.StoredData,
	}
	return record, nil
}

// DeleteRecord deletes record from server by ID.
func (conn *ClientConnGPRC) DeleteRecord(token entity.AuthToken, recordID string) error {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authToken", string(token))

	_, err := conn.GophkeeperClient.DeleteRecord(ctx, &pb.RecordID{Id: recordID})

	code := status.Code(err)

	switch code {
	case codes.Internal:
		return storage.ErrUnknown
	case codes.Unauthenticated:
		return storage.ErrUserUnauthorized
	case codes.NotFound:
		return storage.ErrNotFound
	}

	return nil
}

// CreateRecord creates record and saves to server.
func (conn *ClientConnGPRC) CreateRecord(token entity.AuthToken, record entity.Record) error {
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authToken", string(token))

	_, err := conn.GophkeeperClient.CreateRecord(ctx, &pb.Record{
		Type:       pb.MessageType(record.Type),
		Metadata:   record.Metadata,
		StoredData: record.Data,
	})

	code := status.Code(err)

	switch code {
	case codes.Internal:
		return storage.ErrUnknown
	case codes.Unauthenticated:
		return storage.ErrUserUnauthorized
	}

	return nil
}
