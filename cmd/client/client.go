package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/gophkeeper/internal/client"
	"github.com/size12/gophkeeper/internal/config"
	"github.com/size12/gophkeeper/internal/handlers"
)

func main() {
	cfg := config.GetClientConfig()

	c := handlers.NewClientConn(cfg.ServerAddress)

	h := handlers.NewClient(c)

	tui := client.NewTUI(h)

	tui.Run()
}
