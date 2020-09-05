package main

import (
	"auto/internal/server"
	"auto/internal/storage"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
	"log"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("zap.NewDevelopment: %v", err)
	}
	defer logger.Sync()

	logger.Info("Application is starting")

	store, err := storage.New(logger, "db")
	if err != nil {
		log.Fatalf("Error creating store: %v", err)
	}

	srvCfg := server.EnvConfig{}
	if err := env.Parse(&srvCfg); err != nil {
		logger.Fatal("parsing server environment config", zap.Error(err))
	}

	srv, err := server.New(logger, store, server.WithEnvConfig(srvCfg))
	if err != nil {
		logger.Fatal("server.New", zap.Error(err))
	}

	err = srv.Start()
	if err != nil {
		logger.Fatal("srv.Start", zap.Error(err))
	}
}
