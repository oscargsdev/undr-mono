package main

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

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %v", err)
		os.Exit(1)
	}

	var cfg config
	loadConfig(&cfg)

	fmt.Println("port: ", cfg.port)
	fmt.Println("env: ", cfg.env)
	fmt.Println("dsn: ", cfg.db.dsn)
	fmt.Println("max-open-conns: ", cfg.db.maxOpenConns)
	fmt.Println("max-idle-conns: ", cfg.db.maxIdleConns)
	fmt.Println("max-idle-time: ", cfg.db.maxIdleTime)

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
