package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

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

	// TODO: Handle defaults
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Printf("error loading server port: %v", err)
		os.Exit(1)
	}
	cfg.port = port

	cfg.env = os.Getenv("ENV")

	cfg.db.dsn = os.Getenv("DSN")

	maxOpenConnsStr := os.Getenv("MAX_OPEN_CONNS")
	maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
	if err != nil {
		fmt.Printf("error loading max-open-conns: %v", err)
		os.Exit(1)
	}
	cfg.db.maxOpenConns = maxOpenConns

	maxIdleConnsStr := os.Getenv("MAX_IDLE_CONNS")
	maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)
	if err != nil {
		fmt.Printf("error loading max-idle-conns: %v", err)
		os.Exit(1)
	}
	cfg.db.maxIdleConns = maxIdleConns

	maxIdleTimeStr := os.Getenv("MAX_IDLE_TIME")
	maxIdleTime, err := strconv.Atoi(maxIdleTimeStr)
	if err != nil {
		fmt.Printf("error loading max-idle-time: %v", err)
		os.Exit(1)
	}
	cfg.db.maxIdleTime = time.Duration(maxIdleTime) * time.Minute

	fmt.Println("port: ", cfg.port)
	fmt.Println("env: ", cfg.env)
	fmt.Println("dsn: ", cfg.db.dsn)
	fmt.Println("max-open-conns: ", cfg.db.maxOpenConns)
	fmt.Println("max-idle-conns: ", cfg.db.maxIdleConns)
	fmt.Println("max-idle-time: ", cfg.db.maxIdleTime)
}
