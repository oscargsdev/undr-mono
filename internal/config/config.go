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
	DefaultPort         = 4000
	DefaultEnv          = "dev"
	DefaultDSN          = "postgres://undr:pa55word@localhost:5431/undr-db?sslmode=disable"
	DefaultMaxOpenConns = 25
	DefaultMaxIdleTime  = 15
)

var validEnvs = []string{"dev", "staging", "prod"}

type Config struct {
	Port int
	Env  string
	DB   struct {
		DSN          string
		MaxOpenConns int
		MaxIdleTime  time.Duration
	}
}

func LoadFromEnv(filenames ...string) (*Config, error) {
	err := godotenv.Load(filenames...)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error while loading the environment file: %w", err)
		}
	}

	return LoadConfig()
}

func LoadConfig() (*Config, error) {
	var cfg Config

	portStr := os.Getenv("PORT")
	if portStr == "" {
		cfg.Port = DefaultPort
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("parse PORT %q: %w", portStr, err)
		}
		cfg.Port = port
	}

	envStr := os.Getenv("ENV")
	if envStr == "" {
		cfg.Env = DefaultEnv
	} else {
		if !slices.Contains(validEnvs, envStr) {
			return nil, fmt.Errorf("invalid ENV value %q, expected one from %v", envStr, validEnvs)
		}
		cfg.Env = envStr
	}

	dsnStr := os.Getenv("DSN")
	if dsnStr == "" {
		cfg.DB.DSN = DefaultDSN
	} else {
		cfg.DB.DSN = dsnStr
	}

	maxOpenConnsStr := os.Getenv("MAX_OPEN_CONNS")
	if maxOpenConnsStr == "" {
		cfg.DB.MaxOpenConns = DefaultMaxOpenConns
	} else {
		maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
		if err != nil {
			return nil, fmt.Errorf("parse MAX_OPEN_CONNS %q: %w", maxOpenConnsStr, err)
		}
		cfg.DB.MaxOpenConns = maxOpenConns
	}

	maxIdleTimeStr := os.Getenv("MAX_IDLE_TIME")
	if maxIdleTimeStr == "" {
		cfg.DB.MaxIdleTime = time.Duration(DefaultMaxIdleTime) * time.Minute
	} else {
		maxIdleTime, err := strconv.Atoi(maxIdleTimeStr)
		if err != nil {
			return nil, fmt.Errorf("parse MAX_IDLE_TIME %q: %w", maxIdleTimeStr, err)
		}
		cfg.DB.MaxIdleTime = time.Duration(maxIdleTime) * time.Minute
	}

	return &cfg, nil
}
