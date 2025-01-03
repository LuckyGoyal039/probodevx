package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	redis "github.com/probodevx/config"
	// "github.com/probodevx/controllers/wss"
)

func main() {
	wsApp := fiber.New(fiber.Config{
		Immutable: true,
		Prefork:   false,
	})

	WSPORT := os.Getenv("WSPORT")
	if WSPORT == "" {
		WSPORT = "8080"
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

	wsApp.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	wsApp.Get("/ws/:event", WebSocketHandler)

	wsApp.Listen(fmt.Sprintf(":%s", WSPORT))

}
