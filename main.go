package main

import (
	"context"
	"fmt"
	"hexlet/Internal/repository"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	ctx := context.Background()
	_ = ctx
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Working directory:", wd)
	dbpool, err := repository.InitDBConn(ctx)
	if err != nil {
		log.Fatalf("failed to init DB connection: %v", err)
	}
	defer dbpool.Close()
	a := repository.NewApp(ctx, dbpool)
	a.Routes(r)
	err = r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
