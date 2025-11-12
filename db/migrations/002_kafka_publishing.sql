-- Поля для Kafka
ALTER TABLE post_destinations 
ADD COLUMN kafka_event_sent BOOLEAN DEFAULT FALSE;

ALTER TABLE post_destinations 
ADD COLUMN kafka_sent_at TIMESTAMP WITH TIME ZONE;

-- Индекс 
CREATE INDEX idx_post_destinations_kafka_ready 
ON post_destinations(scheduled_for, status, kafka_event_sent);

-- Комментарии 
COMMENT ON COLUMN post_destinations.kafka_event_sent IS 'Флаг отправки события в Kafka';
COMMENT ON COLUMN post_destinations.kafka_sent_at IS 'Время отправки события в Kafka';

