package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/size12/gophkeeper/internal/entity"
	"github.com/size12/gophkeeper/internal/storage"
	pb "github.com/size12/gophkeeper/protocols/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ServerConn keeps server endpoints alive.
type ServerConn struct {
	pb.UnimplementedGophkeeperServer
	Handlers *Server
}

// NewServerConn returns new server connection.
func NewServerConn(h *Server) *ServerConn {
	return &ServerConn{
		Handlers: h,
	}
}

// Run runs server listener.
func (server *ServerConn) Run(ctx context.Context, runAddress string) {
	listen, err := net.Listen("tcp", runAddress)
	if err != nil {
		log.Fatal(err)
	}

	sgrpc := grpc.NewServer()
	pb.RegisterGophkeeperServer(sgrpc, server)

	go func() {
		fmt.Println("Сервер gRPC начал работу")
		if err := sgrpc.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()
	// TODO: gracefully shutdown.
}

// Register process register endpoint.
func (server *ServerConn) Register(_ context.Context, credentials *pb.UserCredentials) (*pb.Session, error) {
	token, err := server.Handlers.CreateUser(entity.UserCredentials{
		Login:    credentials.Login,
		Password: credentials.Password,
	})

	if errors.Is(err, ErrFieldIsEmpty) {
		return nil, status.Errorf(codes.InvalidArgument, "Login or password is empty.")
	}

	if errors.Is(err, storage.ErrLoginExists) {
		return nil, status.Errorf(codes.AlreadyExists, "Login already exists.")
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error.")
	}

	return &pb.Session{SessionToken: string(token)}, nil
}

// Login process login endpoint.
func (server *ServerConn) Login(_ context.Context, credentials *pb.UserCredentials) (*pb.Session, error) {
	token, err := server.Handlers.LoginUser(entity.UserCredentials{
		Login:    credentials.Login,
		Password: credentials.Password,
	})

	if errors.Is(err, ErrFieldIsEmpty) {
		return nil, status.Errorf(codes.InvalidArgument, "Login or password is empty.")
	}

	if errors.Is(err, storage.ErrWrongCredentials) {
		return nil, status.Errorf(codes.Unauthenticated, "Wrong login or password.")
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error.")
	}

	return &pb.Session{SessionToken: string(token)}, nil
}

// GetRecordsInfo process get all records endpoint.
func (server *ServerConn) GetRecordsInfo(ctx context.Context, _ *emptypb.Empty) (*pb.RecordsList, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("authToken")) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Didn't send metadata for authentication.")
	}

	token := entity.AuthToken(md.Get("authToken")[0])
	ctx = context.WithValue(ctx, "authToken", token)

	records, err := server.Handlers.GetRecordsInfo(ctx)

	if errors.Is(err, storage.ErrUserUnauthorized) {
		return nil, status.Errorf(codes.Unauthenticated, "Bad authentication token.")
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error.")
	}

	recordsList := make([]*pb.Record, 0, len(records))

	for _, record := range records {
		recordsList = append(recordsList, &pb.Record{
			Id:       record.ID,
			Metadata: record.Metadata,
			Type:     record.Type,
		})
	}

	return &pb.RecordsList{Records: recordsList}, nil
}

// GetRecord process get record endpoint.
func (server *ServerConn) GetRecord(ctx context.Context, recordID *pb.RecordID) (*pb.Record, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("authToken")) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Didn't send metadata for authentication.")
	}

	token := entity.AuthToken(md.Get("authToken")[0])
	ctx = context.WithValue(ctx, "authToken", token)

	record, err := server.Handlers.GetRecord(ctx, recordID.Id)

	if errors.Is(err, storage.ErrUserUnauthorized) {
		return nil, status.Errorf(codes.Unauthenticated, "Bad authentication token.")
	}

	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "Not found record with such id.")
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error.")
	}

	return &pb.Record{
		Id:         record.ID,
		Type:       record.Type,
		Metadata:   record.Metadata,
		StoredData: record.Data,
	}, nil
}

// CreateRecord process create record endpoint.
func (server *ServerConn) CreateRecord(ctx context.Context, record *pb.Record) (*emptypb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("authToken")) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Didn't send metadata for authentication.")
	}

	token := entity.AuthToken(md.Get("authToken")[0])
	ctx = context.WithValue(ctx, "authToken", token)

	err := server.Handlers.CreateRecord(ctx, entity.Record{
		Metadata: record.Metadata,
		Type:     record.Type,
		Data:     record.StoredData,
	})

	if errors.Is(err, storage.ErrUserUnauthorized) {
		return &emptypb.Empty{}, status.Errorf(codes.Unauthenticated, "Bad authentication token.")
	}

	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "Internal server error.")
	}

	return &emptypb.Empty{}, nil
}

// DeleteRecord process delete record endpoint.
func (server *ServerConn) DeleteRecord(ctx context.Context, recordID *pb.RecordID) (*emptypb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get("authToken")) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Didn't send metadata for authentication.")
	}

	token := entity.AuthToken(md.Get("authToken")[0])
	ctx = context.WithValue(ctx, "authToken", token)

	err := server.Handlers.DeleteRecord(ctx, recordID.Id)

	if errors.Is(err, storage.ErrUserUnauthorized) {
		return &emptypb.Empty{}, status.Errorf(codes.Unauthenticated, "Bad authentication token.")
	}

	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "Not found record with such id.")
	}

	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "Internal server error.")
	}

	return &emptypb.Empty{}, nil
}
