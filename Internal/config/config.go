package config

import (
	"os"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBUser:     getEnv("POSTGRES_USER", ""),
		DBPassword: getEnv("POSTGRES_PASSWORD", ""),
		DBName:     getEnv("POSTGRES_DB", ""),
	}

	return cfg, nil
}
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return defaultValue
	}
	return value
}
