package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type application struct {
	config config
	logger *slog.Logger
	wg     sync.WaitGroup
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %v", err)
		os.Exit(1)
	}

	var cfg config
	loadConfig(&cfg)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

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
