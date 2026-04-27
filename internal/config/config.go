package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env string

	API struct {
		Port int
	}

	DB struct {
		DatabaseURL string
		MaxConns    int32
	}

	Auth struct {
		JWTSecret string
	}

	Bot struct {
		TelegramToken string
		APIBaseURL    string
		HTTPTimeout   time.Duration
	}
}

func Load() (Config, error) {
	var cfg Config

	cfg.Env = getEnv("APP_ENV", "dev")

	port, err := strconv.Atoi(getEnv("API_PORT", "8080"))
	if err != nil || port <= 0 {
		return Config{}, errors.New("invalid API_PORT")
	}
	cfg.API.Port = port

	cfg.DB.DatabaseURL = os.Getenv("DATABASE_URL")
	if strings.TrimSpace(cfg.DB.DatabaseURL) == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	maxConns, err := strconv.Atoi(getEnv("DB_MAX_CONNS", "10"))
	if err != nil || maxConns <= 0 {
		return Config{}, errors.New("invalid DB_MAX_CONNS")
	}
	cfg.DB.MaxConns = int32(maxConns)

	cfg.Auth.JWTSecret = getEnv("JWT_SECRET", "dev_secret")

	cfg.Bot.TelegramToken = getEnv("TELEGRAM_BOT_TOKEN", "")
	cfg.Bot.APIBaseURL = getEnv("API_BASE_URL", "http://talkabout-api:8080")

	httpTimeout, err := time.ParseDuration(getEnv("BOT_HTTP_TIMEOUT", "10s"))
	if err != nil {
		return Config{}, errors.New("invalid BOT_HTTP_TIMEOUT")
	}
	cfg.Bot.HTTPTimeout = httpTimeout

	return cfg, nil
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
