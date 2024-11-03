package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	redis "github.com/probodevx/config"
)

func main() {

	app := fiber.New(fiber.Config{
		Immutable: true,
		Prefork:   false,
	})
	// wsApp := fiber.New(fiber.Config{
	// 	Immutable: true,
	// 	Prefork:   false,
	// })

	PORT := os.Getenv("PORT")
	WSPORT := os.Getenv("WSPORT")
	if WSPORT == "" {
		WSPORT = "8080"
	}

	if PORT == "" {
		PORT = "8001"
	}

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
	redis.ConnectRedis(redisHost, redisPort, redisPassword)
	if err := redis.CheckRedisConnection(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	// main server apis

	app.Listen(fmt.Sprintf(":%s", PORT))

}
