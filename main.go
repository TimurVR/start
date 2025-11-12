package main

import (
	"context"
	"fmt"
	"hexlet/Internal/app"
	storage "hexlet/Internal/storage"
	"log"
	"os"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	ctx := context.Background()
	
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Working directory:", wd)
	dbpool, err := storage.InitDBConn(ctx)
	if err != nil {
		log.Fatalf("failed to init DB connection: %v", err)
	}
	defer dbpool.Close()
	a := app.NewApp(ctx, dbpool)
	a.StartScheduler()
	log.Println("Kafka scheduler started")
	a.Routes(r)
	err = r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
/*        hexlet/Internal/app             coverage: 0.0% of statements
?       hexlet/Internal/domain  [no test files]
?       hexlet/Internal/dto     [no test files]
ok      hexlet/Internal/handler 0.305s  coverage: 84.0% of statements
ok      hexlet/Internal/repository      16.749s coverage: 85.7% of statements
?       hexlet/Internal/service [no test files]
        hexlet/Internal/storage         coverage: 0.0% of statements*/