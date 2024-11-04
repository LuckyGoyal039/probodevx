package apiuser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/data"
)

func CreateNewUser(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).SendString("invalid userId")
	}

	event := data.UserEvent{
		UserId:    userId,
		EventType: "create_user",
		Timestamp: time.Now(),
	}

	eventJson, err := json.Marshal(event)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error creating event")
	}

	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Subscribe to channel
	responseChan := fmt.Sprintf("user_response_%s", userId)
	pubsub := redisClient.Subscribe(ctx, responseChan)
	defer pubsub.Close()

	// Push to queue
	if _, err := redisClient.LPush(ctx, "user_events", eventJson).Result(); err != nil {
		log.Printf("Error pushing to queue: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("error pushing to queue")
	}

	log.Println("Waiting for response message...")

	// Wait for response
	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		log.Printf("Error waiting for response: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("error waiting for response")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
		log.Printf("Error parsing response: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing response")
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}
