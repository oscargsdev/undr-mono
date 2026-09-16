package main

import (
	"log/slog"
	"os"
	"sync"

	"github.com/oscargsdev/undr-mono/internal/config"
	"github.com/oscargsdev/undr-mono/internal/database"
)

type application struct {
	config *config.Config
	logger *slog.Logger
	wg     sync.WaitGroup
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("error while loading config from env", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("configuration loaded")

	db, err := database.OpenDB(cfg)
	if err != nil {
		logger.Error("error while opening database connection", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
	}

	err = app.serve()
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}
}
