package app

import (
	"context"
	"encoding/json"
	"hexlet/Internal/handler" //docker-compose logs hexlet-project -f
	"hexlet/Internal/kafka"   //docker-compose up -d --build
	"hexlet/Internal/repository"
	"hexlet/Internal/service"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	kf "github.com/segmentio/kafka-go"
)

type App struct {
	Ctx       context.Context
	Repo      *repository.Repository
	Handler   *handler.App
	Scheduler *service.SchedulerService
	Counter   int
	Wg        sync.WaitGroup
	Cancel    context.CancelFunc
}
type Message struct {
	MessageID       int    `json:"message_id"`
	Timestamp       string `json:"timestamp"`
	ContentID       string `json:"content_id"`
	SocialAccountID string `json:"social_account_id"`
	UserID          string `json:"user_id"`
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
	_, cancel := context.WithCancel(context.Background())

	return &App{
		Ctx:       ctx,
		Repo:      repo,
		Handler:   handlerApp,
		Scheduler: scheduler,
		Cancel:    cancel,
		Counter:   1,
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

func (a *App) StartBackgroundWorker(msg kf.Message) {
	a.Wg.Add(1)
	go a.backgroundWorker(msg)
}

func (a *App) backgroundWorker(msg kf.Message) {
	defer a.Wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("Worker started")

	select {
	case <-a.Ctx.Done():
			log.Println("Worker stoped")
			return
	case <-ticker.C:
			a.proccesProcessing(msg)
	}
}

func (a *App) proccesProcessing(msg kf.Message) {
	key := string(msg.Key)
	value := string(msg.Value)
	topic := msg.Topic
	partition := msg.Partition
	offset := msg.Offset
	var msg1 Message
	err := json.Unmarshal([]byte(value), &msg1)
	if err != nil {
		log.Print(err)
		return
	}
	user_id, err1 := strconv.Atoi(msg1.UserID)
	if err1 != nil {
		log.Print(err1)
		return
	}
	platform, err2 := a.Repo.GetPlatformsByUserID(a.Ctx, "Telegram", user_id)
	if err2 != nil {
		log.Print(err2)
		return
	}
	message, err3 := a.Repo.GetTitleANDContent(a.Ctx, user_id)
	if err3 != nil {
		log.Print(err3)
		return
	}
	text:=message.Title+"\n"+message.Content
	for chatID,botToken:=range platform.APIConfig{
		SentToTelegram(chatID,botToken,text)
		_=chatID
		_=botToken
		log.Print("Cообщение отправлено")
	}
	err4:=a.Repo.MarkAsSent(a.Ctx,msg1.MessageID)
	if err4 != nil {
		log.Print(err4)
	}
	log.Print(platform, message)
	log.Printf("Получено сообщение:\n"+
		"  Топик: %s\n"+
		"  Партиция: %d\n"+
		"  Смещение: %d\n"+
		"  Ключ: %s\n"+
		"  Значение: %s\n"+
		"  Время: %v\n",
		topic, partition, offset, key, value, msg.Time)
}

func (a *App) WaitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutdown")
	a.Cancel()
	a.Wg.Wait()
	log.Println("Shutdown complete")
}
func SentToTelegram(chatID string,botToken string,text string) {
	//botToken := "8401889449:AAGp9DeX9LKN_TTmTLUprYyOfA_zNbIh9jo"
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Println("Ошибка создания бота:", err)
	}
	log.Printf("Авторизован как %s", bot.Self.UserName)
	//chatID := "@hxlt_autoposting"
	msg := tgbotapi.NewMessageToChannel(chatID, text)
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Ошибка отправки:", err)
	} else {
		log.Println("Сообщение успешно отправлено!")
	}
}
