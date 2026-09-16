package main

import (
	"log/slog"
	"os"
	"sync"

	"github.com/oscargsdev/undr-mono/internal/config"
	"github.com/oscargsdev/undr-mono/internal/database"
)

type application struct {
	config config.Config
	logger *slog.Logger
	wg     sync.WaitGroup
}

func main() {
	cfg := config.LoadFromEnv()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := database.OpenDB(cfg)
	if err != nil {
		logger.Error(err.Error())
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
