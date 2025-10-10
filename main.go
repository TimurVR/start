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
	// Print the working directory
	fmt.Println("Working directory:", wd)
	dbpool, err := repository.InitDBConn(ctx)
	if err != nil {
		log.Fatalf("failed to init DB connection: %v", err)
	}
	defer dbpool.Close()
	err = r.Run(":8080")
}
