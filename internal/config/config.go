package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort         = 4000
	defaultEnv          = "dev"
	defaultDSN          = "postgres://undr:pa55word@localhost:5431/undr-db?sslmode=disable"
	defaultMaxOpenConns = 25
	defaultMaxIdleTime  = 15
)

var validEnvs = []string{"dev", "staging", "prod"}

// Config contains the server and database configuration.
type Config struct {
	Port int
	Env  string
	DB   struct {
		DSN          string
		MaxOpenConns int
		MaxIdleTime  time.Duration
	}
}

// LoadFromEnv loads optional dotenv files and constructs a Config from
// environment variables and defaults.
func LoadFromEnv(filenames ...string) (*Config, error) {
	err := godotenv.Load(filenames...)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error while loading the environment file: %w", err)
		}
	}

	return loadConfig()
}

func loadConfig() (*Config, error) {
	var cfg Config

	portStr := os.Getenv("PORT")
	if portStr == "" {
		cfg.Port = defaultPort
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("parse PORT %q: %w", portStr, err)
		}
		cfg.Port = port
	}

	envStr := os.Getenv("ENV")
	if envStr == "" {
		cfg.Env = defaultEnv
	} else {
		if !slices.Contains(validEnvs, envStr) {
			return nil, fmt.Errorf("invalid ENV value %q, expected one from %v", envStr, validEnvs)
		}
		cfg.Env = envStr
	}

	dsnStr := os.Getenv("DSN")
	if dsnStr == "" {
		cfg.DB.DSN = defaultDSN
	} else {
		cfg.DB.DSN = dsnStr
	}

	maxOpenConnsStr := os.Getenv("MAX_OPEN_CONNS")
	if maxOpenConnsStr == "" {
		cfg.DB.MaxOpenConns = defaultMaxOpenConns
	} else {
		maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
		if err != nil {
			return nil, fmt.Errorf("parse MAX_OPEN_CONNS %q: %w", maxOpenConnsStr, err)
		}
		cfg.DB.MaxOpenConns = maxOpenConns
	}

	maxIdleTimeStr := os.Getenv("MAX_IDLE_TIME")
	if maxIdleTimeStr == "" {
		cfg.DB.MaxIdleTime = time.Duration(defaultMaxIdleTime) * time.Minute
	} else {
		maxIdleTime, err := strconv.Atoi(maxIdleTimeStr)
		if err != nil {
			return nil, fmt.Errorf("parse MAX_IDLE_TIME %q: %w", maxIdleTimeStr, err)
		}
		cfg.DB.MaxIdleTime = time.Duration(maxIdleTime) * time.Minute
	}

	return &cfg, nil
}
