@echo off
echo Waiting for Kafka to start (this may take up to 60 seconds)...

:check_kafka
docker-compose ps kafka | find "Up" > nul
if %errorlevel% equ 0 (
    echo Kafka is running!
    goto create_topic
) else (
    echo Kafka is not ready yet, waiting 10 seconds...
    timeout /t 10 /nobreak
    goto check_kafka
)

:create_topic
echo Creating Kafka topic...
docker-compose exec kafka kafka-topics --create --topic publications.pending --bootstrap-server kafka:9092 --partitions 3 --replication-factor 1 --if-not-exists

if %errorlevel% equ 0 (
    echo Kafka topic 'publications.pending' created successfully!
) else (
    echo Failed to create Kafka topic. It might already exist.
)

echo Listing all topics:
docker-compose exec kafka kafka-topics --list --bootstrap-server kafka:9092

pause