package config

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultPort         = 4004
	DefaultEnv          = "devtest"
	DefaultDSN          = "postgrestest"
	DefaultMaxOpenConns = 250
	DefaultMaxIdleTime  = 150
)

var ValidEnvs = []string{"dev", "staging", "prod"}

type Config struct {
	Port int
	Env  string
	DB   struct {
		DSN          string
		MaxOpenConns int
		MaxIdleTime  time.Duration
	}
}

func LoadFromEnv(filenames ...string) Config {
	err := godotenv.Load(filenames...)
	if err != nil {
		fmt.Printf("error loading .env file: %v", err)
		os.Exit(1)
	}

	var cfg Config
	LoadConfig(&cfg)

	return cfg
}

func LoadConfig(cfg *Config) {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		cfg.Port = DefaultPort
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Printf("error loading server port: %v \n", err)
			os.Exit(1)
		}
		cfg.Port = port
	}

	envStr := os.Getenv("ENV")
	if envStr == "" {
		cfg.Env = DefaultEnv
	} else {
		if !slices.Contains(ValidEnvs, envStr) {
			fmt.Print("error loading environment: invalid value: ", envStr, "\n")
			os.Exit(1)
		}
		cfg.Env = envStr
	}

	dsnStr := os.Getenv("DSN")
	if dsnStr == "" {
		cfg.DB.DSN = DefaultDSN
	} else {
		cfg.DB.DSN = os.Getenv("DSN")
	}

	maxOpenConnsStr := os.Getenv("MAX_OPEN_CONNS")
	if maxOpenConnsStr == "" {
		cfg.DB.MaxOpenConns = DefaultMaxOpenConns
	} else {
		maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
		if err != nil {
			fmt.Printf("error loading max open connections: %v \n", err)
			os.Exit(1)
		}
		cfg.DB.MaxOpenConns = maxOpenConns
	}

	maxIdleTimeStr := os.Getenv("MAX_IDLE_TIME")
	if maxIdleTimeStr == "" {
		cfg.DB.MaxIdleTime = time.Duration(DefaultMaxIdleTime) * time.Minute
	} else {
		maxIdleTime, err := strconv.Atoi(maxIdleTimeStr)
		if err != nil {
			fmt.Printf("error loading max idle time: %v \n", err)
			os.Exit(1)
		}
		cfg.DB.MaxIdleTime = time.Duration(maxIdleTime) * time.Minute
	}
}
