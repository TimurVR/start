package config

import (
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfigMaster() (*Config, error) {
	cfg := &Config{
		DBHost:     getEnv("MASTER_HOST", ""),
		DBPort:     getEnv("MASTER_PORT", ""),
		DBUser:     getEnv("POSTGRES_USER", ""),
		DBPassword: getEnv("POSTGRES_PASSWORD", ""),
		DBName:     getEnv("POSTGRES_DB", ""),
	}
	return cfg, nil
}
func LoadConfigSlave() (*Config, error) {
	cfg := &Config{
		DBHost:     getEnv("SLAVE_HOST", ""),
		DBPort:     getEnv("SLAVE_PORT", ""),
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
