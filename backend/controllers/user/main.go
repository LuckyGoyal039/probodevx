// user/handlers.go
package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/data"
	"github.com/probodevx/engine/global"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Push to queue
	if _, err := redisClient.LPush(ctx, "user_events", eventJson).Result(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error pushing to queue")
	}

	// Subscribe to response channel
	responseChan := fmt.Sprintf("user_response_%s", userId)
	pubsub := redisClient.Subscribe(ctx, responseChan)
	defer pubsub.Close()

	// Wait for response with timeout
	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error waiting for response")
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing response")
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func GetAllUsers(c *fiber.Ctx) error {
	users := global.UserManager.GetAllUsers()
	return c.JSON(fiber.Map{
		"data": users,
	})
}
