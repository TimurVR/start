package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"hexlet/Internal/domain"
	"hexlet/Internal/kafka"
	"hexlet/Internal/repository"
)

type SchedulerService struct {
	repo      *repository.Repository
	kafkaProd *kafka.Producer
	interval  time.Duration
	batchSize int
}

func NewSchedulerService(
	repo *repository.Repository,
	kafkaProd *kafka.Producer,
	interval time.Duration,
	batchSize int,
) *SchedulerService {
	return &SchedulerService{
		repo:      repo,
		kafkaProd: kafkaProd,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (s *SchedulerService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Printf("Scheduler started with interval: %v", s.interval)

	s.processScheduledPublications(ctx)

	for {
		select {
		case <-ticker.C:
			s.processScheduledPublications(ctx)
		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		}
	}
}

func (s *SchedulerService) processScheduledPublications(ctx context.Context) {
	log.Println("Checking for scheduled publications...")

	publications, err := s.repo.GetReadyForPublication(ctx, s.batchSize)
	if err != nil {
		log.Printf("Error getting scheduled publications: %v", err)
		return
	}

	if len(publications) == 0 {
		log.Println("No scheduled publications found")
		return
	}

	log.Printf("Found %d scheduled publications to process", len(publications))

	successfulCount := 0
	
	for _, pub := range publications {
		if err := s.processPublication(ctx, pub); err != nil {
			log.Printf("Error processing publication %d: %v", pub.ID_destination, err)
			continue
		}
		successfulCount++
	}

	log.Printf("Successfully processed %d publications", successfulCount)
}

func (s *SchedulerService) processPublication(ctx context.Context, pub domain.ScheduledPublication) error {
	event := domain.PublicationEvent{
		MessageID:       pub.ID_destination,
		Timestamp:       pub.Scheduled_for,
		ContentID:       strconv.Itoa(pub.ID_post),
		SocialAccountID: strconv.Itoa(pub.ID_platform),
		UserID:          strconv.Itoa(pub.ID_user),
	}
	if err := s.kafkaProd.SendPublicationEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to send event to kafka: %w", err)
	}
	if err := s.repo.MarkAsSentToKafkaInTx(ctx, pub.ID_destination); err != nil {
		return fmt.Errorf("failed to update publication status: %w", err)
	}
	log.Printf("Publication %d (Post %d -> Platform %d) sent to Kafka",
		pub.ID_destination, pub.ID_post, pub.ID_platform)
	return nil
}
