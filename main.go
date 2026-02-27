package main

import (
	"context"
	"hexlet/Internal/app"
	"hexlet/Internal/config"
	storage "hexlet/Internal/storage"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

func main() {
	r := gin.Default()
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}
	dbpool, err := storage.InitDBConn(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init DB connection: %v", err)
	}
	defer dbpool.Close()
	a := app.NewApp(ctx, dbpool)
	a.StartScheduler()
	go func() {
		time.Sleep(30 * time.Second)
		readerConfig := kafka.ReaderConfig{
			Brokers:     []string{"kafka:9092"},
			Topic:       "publications.pending",
			GroupID:     "hexlet-publications-worker",
			MinBytes:    10e3,
			MaxBytes:    10e6,
			MaxWait:     5 * time.Second,
			StartOffset: kafka.FirstOffset,
		}
		reader := kafka.NewReader(readerConfig)
		defer reader.Close()
		log.Println("Kafka consumer started. Waiting for messages...")
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					log.Println("Kafka consumer stopped")
					return
				}
				log.Printf("Error reading from Kafka: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			log.Printf("Received Kafka message: %s", string(msg.Value))
			a.StartBackgroundWorker(msg)
		}
	}()

	log.Println("Kafka scheduler started")
	a.Routes(r)
	go func() {
		log.Println("HTTP server starting on :8080")
		err := r.Run(":8080")
		if err != nil {
			log.Fatal(err)
		}
	}()
	a.WaitForShutdown()
}
