package config

import (
	"errors"
	"fmt"
	"os"

	env "github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	SignalPhoneNumber string `env:"SIGNAL_PHONE_NUMBER,required,notEmpty"`
	DatabasePath      string `env:"DATABASE_PATH,required,notEmpty"`
	SignalAPI         string `env:"SIGNAL_API,required,notEmpty"`
}

func Load() (Config, error) {
	var cfg Config
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse env: %w", err)
	}

	return cfg, nil
}
