package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	BcryptCost  int
	Port        string
}

// LoadConfig reads environment variables (loads .env if present) and returns a Config.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load() // ignore error — allow environment to provide values in production

	dbURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	bcryptCostStr := os.Getenv("BCRYPT_COST")
	port := os.Getenv("PORT")

	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if port == "" {
		port = "8080"
	}

	bcryptCost := 10
	if bcryptCostStr != "" {
		if v, err := strconv.Atoi(bcryptCostStr); err == nil && v > 0 {
			bcryptCost = v
		}
	}

	return &Config{
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		BcryptCost:  bcryptCost,
		Port:        port,
	}, nil
}
