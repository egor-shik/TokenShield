package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	ApiURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	api := os.Getenv("API_URL")
	if api == "" {
		return nil, errors.New("API_URL environment variable is required")
	}

	return &Config{
		Port:   port,
		ApiURL: api,
	}, nil
}
