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
	if err := waitForDB(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to wait for database: %v", err)
	}
	url := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	log.Printf("Connecting to database at: %s:%s", config.Host, config.Port)
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pg config: %v", err)
	}

	cfg.MaxConns = int32(25)
	cfg.MinConns = int32(5)
	cfg.HealthCheckPeriod = 1 * time.Minute
	cfg.MaxConnLifetime = 24 * time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 30 * time.Second
	cfg.ConnConfig.DialFunc = (&net.Dialer{
		KeepAlive: cfg.HealthCheckPeriod,
		Timeout:   cfg.ConnConfig.ConnectTimeout,
	}).DialContext
	for i := 0; i < 10; i++ {
		dbpool, err = pgxpool.ConnectConfig(ctx, cfg)
		if err != nil {
			log.Printf("Attempt %d: failed to connect to database: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err = dbpool.Ping(ctx); err != nil {
			log.Printf("Attempt %d: failed to ping database: %v", i+1, err)
			dbpool.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println("Database connection established")
		return dbpool, nil
	}

	return nil, fmt.Errorf("failed to connect to database after 10 attempts: %v", err)
}

func waitForDB(ctx context.Context, config DBConfig) error {
	url := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	log.Printf("Waiting for database to be ready...")

	for i := 0; i < 30; i++ {
		cfg, err := pgxpool.ParseConfig(url)
		if err != nil {
			return err
		}

		cfg.ConnConfig.ConnectTimeout = 5 * time.Second

		dbpool, err := pgxpool.ConnectConfig(ctx, cfg)
		if err != nil {
			log.Printf("Database not ready yet (attempt %d): %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := dbpool.Ping(ctx); err != nil {
			log.Printf("Database ping failed (attempt %d): %v", i+1, err)
			dbpool.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		dbpool.Close()
		log.Println("Database is ready!")
		return nil
	}

	return fmt.Errorf("database not ready after 30 attempts")
}

func loadDBConfig() DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", "postgres"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "2455"),
		DBName:   getEnv("DB_NAME", "avtopostingdb"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

/*
# Полная остановка и удаление volumes
docker-compose down -v

# Перезапуск
docker-compose up -d --build

# Проверим логи
docker-compose logs -f hexlet-project*/
