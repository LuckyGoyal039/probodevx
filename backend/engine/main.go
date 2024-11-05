// engine/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/controllers/start"
	// adjust this import path to match your projec
	// adjust this import path to match your project
)

func main() {
	app := fiber.New(fiber.Config{
		Immutable: true,
		Prefork:   false,
	})

	// Redis connection setup
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Connect to Redis
	redisClient := redis.ConnectRedis(redisHost, redisPort, redisPassword)
	if err := redis.CheckRedisConnection(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	// Create and start the processor
	ctx := context.Background()
	userProcessor := start.NewUserProcessor(redisClient)
	go func() {
		if err := userProcessor.StartProcessing(ctx); err != nil {
			log.Printf("Processor error: %v", err)
		}
	}()

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8002"
	}

	log.Printf("Engine starting on port %s", PORT)
	if err := app.Listen(fmt.Sprintf(":%s", PORT)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
