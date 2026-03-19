│   .env
│   .gitignore
│   coverage
│   docker-compose.yml
│   Dockerfile
│   go.mod
│   go.sum
│   main.go
│   README.md
│
├───.github
│   └───workflows
│           test.yml
│
├───db
│   └───migrations
│           001_init_schema.sql
│
├───Internal
│   ├───app
│   │       app.go // функция длоя иницилизации приложения+ кафка consumers
│   │
│   ├───config
│   │       .env
│   │       config.go // для загрузки кофига из env файла
│   │
│   ├───domain
│   │       post.go // структуры которые используются для кафки
│   │
│   ├───dto
│   │       request.go // структура для requsets
│   │       responce.go // струтура для responce
│   │
│   ├───handler
│   │       post_handler.go // API
│   │       unit_test.go // tets API
│   │
│   ├───kafka
│   │       config.go // Иницилизация кафки
│   │       producer.go // кафка producer
│   │
│   ├───repository
│   │       functional_test.go
│   │       post_repository.go // запросы к бд для API
│   │       schedule_repository.go // запросы к бд для kafka
│   │
│   ├───service
│   │       scheduler_service.go //планировщик публикаций
│   │
│   └───storage
│           storage.go // инилизация бд
│
└───scripts
        create-kafka-topics.bat