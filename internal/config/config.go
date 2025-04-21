package config

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"
)

type Config struct {
	RunAddr     string
	DatabaseURI string
	AccrualAddr string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		RunAddr:     ":8080",
		DatabaseURI: "",
		AccrualAddr: "http://localhost:8081",
	}

	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database URI")
	flag.StringVar(&cfg.AccrualAddr, "r", cfg.AccrualAddr, "accrual system address")
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

	if cfg.DatabaseURI == "" {
		log.Printf("Error: DATABASE_URI is empty")
		return nil, errors.New("DATABASE_URI is required")
	}

	log.Printf("Config loaded: RunAddr=%s, DatabaseURI=%s, AccrualAddr=%s", cfg.RunAddr, cfg.DatabaseURI, cfg.AccrualAddr)
	return cfg, nil
}