package main

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/gophkeeper/internal/config"
	"github.com/size12/gophkeeper/internal/handlers"
	"github.com/size12/gophkeeper/internal/storage"
)

func main() {
	fmt.Println("Hello from server!")

	cfg := config.GetServerConfig()

	serverStorage := storage.NewDBStorage(cfg.DBConnectionURL)
	serverStorage.MigrateUP()

	handlersAuth := handlers.NewAuthenticatorJWT([]byte("secret ewfwfw key"))
	serverHandlers := handlers.NewServerHandlers(serverStorage, handlersAuth)

	handlers.NewServer(serverHandlers).Run(context.Background(), cfg.RunAddress)
	select {}
}
