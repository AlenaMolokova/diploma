package config

import (
	"errors"
	"log"

	"github.com/AlenaMolokova/diploma/internal/constants"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddr         string `env:"RUN_ADDRESS" envDefault:":8080"`
	DatabaseURI     string `env:"DATABASE_URI"`
	AccrualAddr     string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8081"`
	JWTSecret       string `env:"JWT_SECRET" envDefault:"supersecretkey"`
	PollIntervalSec int    `env:"POLL_INTERVAL" envDefault:"5"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		JWTSecret:       constants.DefaultJWTSecret,
		PollIntervalSec: constants.DefaultPollInterval,
	}

	if err := env.Parse(cfg); err != nil {
		log.Printf("Failed to parse environment variables: %v", err)
		return nil, err
	}

	log.Printf("Config loaded: RunAddr=%s, DatabaseURI=%s, AccrualAddr=%s, PollInterval=%ds",
		cfg.RunAddr, cfg.DatabaseURI, cfg.AccrualAddr, cfg.PollIntervalSec)

	if cfg.DatabaseURI == "" {
		log.Printf("Error: DATABASE_URI is empty")
		return nil, errors.New("DATABASE_URI is required")
	}

	return cfg, nil
}
