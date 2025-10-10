package repository

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func InitDBConn(ctx context.Context) (dbpool *pgxpool.Pool, err error) {
	config := loadDBConfig()

	url := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("failed to parse pg config: %v", err)
		return
	}

	cfg.MaxConns = int32(25)
	cfg.MinConns = int32(5)
	cfg.HealthCheckPeriod = 1 * time.Minute
	cfg.MaxConnLifetime = 24 * time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	cfg.ConnConfig.DialFunc = (&net.Dialer{
		KeepAlive: cfg.HealthCheckPeriod,
		Timeout:   cfg.ConnConfig.ConnectTimeout,
	}).DialContext

	dbpool, err = pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect config: %v", err)
		return
	}

	if err = dbpool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
		return
	}

	log.Println("Database connection established")
	return
}

func loadDBConfig() DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", "postgres"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "valie"),
		Password: getEnv("DB_PASSWORD", "245524"),
		DBName:   getEnv("DB_NAME", "avtopostingdb"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}//host=localhost port=5432 user=valie password=245524 dbname=avtopostingdb sslmode=disable
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
