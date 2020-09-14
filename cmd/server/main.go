package main

import (
	"auto/internal/server"
	"auto/internal/storage"
	"errors"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("zap.NewDevelopment: %v", err)
	}
	defer logger.Sync()

	logger.Info("Application is starting")

	config, err := newConfig(logger)
	if err != nil {
		if errors.Is(err, helpErr) {
			logger.Info("-help invoked, exiting")
			os.Exit(0)
		}
		logger.Fatal("can not create config")
	}

	store, err := storage.New(logger, "/data/db")
	if err != nil {
		log.Fatalf("Error creating store: %v", err)
	}

	srv, err := server.New(logger, store, server.WithConfig(*config.http))
	if err != nil {
		logger.Fatal("server.New", zap.Error(err))
	}

	err = srv.Start()
	if err != nil {
		logger.Fatal("srv.Start", zap.Error(err))
	}
}
