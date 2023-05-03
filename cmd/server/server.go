package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/gophkeeper/internal/config"
	"github.com/size12/gophkeeper/internal/handlers"
	"github.com/size12/gophkeeper/internal/storage"
)

func main() {
	cfg := config.GetServerConfig()

	db := storage.NewDBStorage(cfg.DBConnectionURL)
	db.MigrateUP()

	files := storage.NewFileStorage(cfg.FilesDirectory)

	serverStorage := storage.NewStorage(db, files)

	handlersAuth := handlers.NewAuthenticatorJWT([]byte("secret ewfwfw key"))
	serverHandlers := handlers.NewServerHandlers(serverStorage, handlersAuth)

	server := handlers.NewServerConn(serverHandlers)
	go server.Run(context.Background(), cfg.RunAddress)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigint

	server.Stop()
}
