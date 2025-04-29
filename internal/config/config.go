package config

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/AlenaMolokova/diploma/internal/constants"
)

type Config struct {
	RunAddr         string
	DatabaseURI     string
	AccrualAddr     string
	JWTSecret       string
	PollIntervalSec int
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		RunAddr:         ":8080",
		DatabaseURI:     "",
		AccrualAddr:     "http://localhost:8081",
		JWTSecret:       constants.DefaultJWTSecret,
		PollIntervalSec: constants.DefaultPollInterval,
	}

	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database URI")
	flag.StringVar(&cfg.AccrualAddr, "r", cfg.AccrualAddr, "accrual system address")
	flag.StringVar(&cfg.JWTSecret, "j", cfg.JWTSecret, "JWT secret")
	flag.IntVar(&cfg.PollIntervalSec, "p", cfg.PollIntervalSec, "poll interval in seconds")
	flag.Parse()

	log.Printf("DatabaseURI from flag -d: %q", cfg.DatabaseURI)
	
	if envDBURI := strings.TrimSpace(os.Getenv("DATABASE_URI")); envDBURI != "" {
		cfg.DatabaseURI = envDBURI
		log.Printf("DatabaseURI from env: %q", cfg.DatabaseURI)
	}
	
	if envRunAddr := strings.TrimSpace(os.Getenv("RUN_ADDRESS")); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
		log.Printf("RunAddr from env: %q", cfg.RunAddr)
	}
	
	if envAccrualAddr := strings.TrimSpace(os.Getenv("ACCRUAL_SYSTEM_ADDRESS")); envAccrualAddr != "" {
		cfg.AccrualAddr = envAccrualAddr
		log.Printf("AccrualAddr from env: %q", cfg.AccrualAddr)
	}
	
	if envJWTSecret := strings.TrimSpace(os.Getenv("JWT_SECRET")); envJWTSecret != "" {
		cfg.JWTSecret = envJWTSecret
		log.Printf("JWT secret from env: using custom value")
	}
	
	if envPollInterval := strings.TrimSpace(os.Getenv("POLL_INTERVAL")); envPollInterval != "" {
		if interval, err := strconv.Atoi(envPollInterval); err == nil && interval > 0 {
			cfg.PollIntervalSec = interval
			log.Printf("Poll interval from env: %d seconds", cfg.PollIntervalSec)
		}
	}

	if cfg.DatabaseURI == "" {
		log.Printf("Error: DATABASE_URI is empty")
		return nil, errors.New("DATABASE_URI is required")
	}

	log.Printf("Config loaded: RunAddr=%s, DatabaseURI=%s, AccrualAddr=%s, PollInterval=%ds",
		cfg.RunAddr, cfg.DatabaseURI, cfg.AccrualAddr, cfg.PollIntervalSec)
	return cfg, nil
}