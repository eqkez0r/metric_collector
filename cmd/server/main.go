package main

import (
	"context"
	"github.com/Eqke/metric-collector/internal/config"
	httpserver "github.com/Eqke/metric-collector/internal/server"
	"github.com/Eqke/metric-collector/internal/storage/localstorage"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	storage := localstorage.New()
	settings, err := config.NewServerConfig()
	if err != nil {
		log.Fatal(err)
	}
	server := httpserver.New(ctx, settings, storage)
	server.Run()
}
