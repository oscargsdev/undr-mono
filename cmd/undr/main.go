package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultPort         = 4004
	DefaultEnv          = "devtest"
	DefaultDSN          = "postgrestest"
	DefaultMaxOpenConns = 250
	DefaultMaxIdleConns = 250
	DefaultMaxIdleTime  = 150
)

var ValidEnvs = []string{"dev", "staging", "prod"}

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

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

func loadConfig(cfg *config) {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		cfg.port = DefaultPort
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Printf("error loading server port: %v \n", err)
			os.Exit(1)
		}
		cfg.port = port
	}

	envStr := os.Getenv("ENV")
	if envStr == "" {
		cfg.env = DefaultEnv
	} else {
		if !slices.Contains(ValidEnvs, envStr) {
			fmt.Print("error loading environment: invalid value: ", envStr, "\n")
			os.Exit(1)
		}
		cfg.env = envStr
	}

	dsnStr := os.Getenv("DSN")
	if dsnStr == "" {
		cfg.db.dsn = DefaultDSN
	} else {
		cfg.db.dsn = os.Getenv("DSN")
	}

	maxOpenConnsStr := os.Getenv("MAX_OPEN_CONNS")
	if maxOpenConnsStr == "" {
		cfg.db.maxOpenConns = DefaultMaxOpenConns
	} else {
		maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
		if err != nil {
			fmt.Printf("error loading max open connections: %v \n", err)
			os.Exit(1)
		}
		cfg.db.maxOpenConns = maxOpenConns
	}

	maxIdleConnsStr := os.Getenv("MAX_IDLE_CONNS")
	if maxIdleConnsStr == "" {
		cfg.db.maxIdleConns = DefaultMaxIdleConns
	} else {
		maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)
		if err != nil {
			fmt.Printf("error loading max idle connections: %v \n", err)
			os.Exit(1)
		}
		cfg.db.maxIdleConns = maxIdleConns
	}

	maxIdleTimeStr := os.Getenv("MAX_IDLE_TIME")
	if maxIdleTimeStr == "" {
		cfg.db.maxIdleTime = time.Duration(DefaultMaxIdleTime) * time.Minute
	} else {
		maxIdleTime, err := strconv.Atoi(maxIdleTimeStr)
		if err != nil {
			fmt.Printf("error loading max idle time: %v \n", err)
			os.Exit(1)
		}
		cfg.db.maxIdleTime = time.Duration(maxIdleTime) * time.Minute
	}
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", "addr", srv.Addr)

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /health/live", live)

	return router
}

func live(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "alive and rocking")
}
