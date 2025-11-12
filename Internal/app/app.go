package app

import (
	"context"
	"hexlet/Internal/handler"
	"hexlet/Internal/kafka"
	"hexlet/Internal/repository"
	"hexlet/Internal/service"
	"os"
	"strings"
	"time"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

type App struct {
	Ctx       context.Context
	Repo      *repository.Repository
	Handler   *handler.App
	Scheduler *service.SchedulerService
}

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *App {
	repo := repository.NewRepository(dbpool)
	handlerApp := &handler.App{
		Ctx:  ctx,
		Repo: repo,
	}
	var scheduler *service.SchedulerService
	kafkaBrokers := getKafkaBrokers()
	
	if len(kafkaBrokers) > 0 {
		kafkaConfig := kafka.NewConfig(
			kafkaBrokers, 
			"publications.pending",
		)
		kafkaProducer := kafka.NewProducer(kafkaConfig)
		scheduler = service.NewSchedulerService(
			repo,
			kafkaProducer,
			1*time.Minute, 
			100,           
		)
	}

	return &App{
		Ctx:       ctx,
		Repo:      repo,
		Handler:   handlerApp,
		Scheduler: scheduler,
	}
}

func (a *App) Routes(r *gin.Engine) {
	a.Handler.Routes(r)
}

func (a *App) StartScheduler() {
	if a.Scheduler != nil {
		go a.Scheduler.Start(a.Ctx)
		log.Println("Kafka scheduler started")
	} else {
		log.Println("Kafka scheduler not configured")
	}
}

func getKafkaBrokers() []string {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		return []string{"localhost:29092"}
	}
	
	brokers := strings.Split(brokersEnv, ",")
	for i, broker := range brokers {
		brokers[i] = strings.TrimSpace(broker)
	}
	
	return brokers
}