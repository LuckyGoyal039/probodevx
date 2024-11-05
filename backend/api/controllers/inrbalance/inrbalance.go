package apiInrBalance

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/api/controllers/common"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/shared"
)

type UserBalanceEvent struct {
	UserId    string                 `json:"userId"`
	EventType string                 `json:"eventType"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func GetInrBalance(c *fiber.Ctx) error {
	userId := c.Params("userId")

	event := shared.EventModel{
		UserId:    userId,
		EventType: "get_balance",
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}

	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubsub, err := common.SubscribeToResponse(redisClient, userId, ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer pubsub.Close()

	if err := common.PushToQueue(redisClient, "main_queue", event, 10*time.Second); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error pushing to queue")
	}

	response, err := common.GetMessage(pubsub, ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error waiting for response")
	}
	return c.JSON(response)
}
func AddUserBalance(c *fiber.Ctx) error {
	type User struct {
		UserId string `json:"userId"`
		Amount int    `json:"amount"`
	}
	var inputs User

	if err := c.BodyParser(&inputs); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}

	event := shared.EventModel{
		UserId:    inputs.UserId,
		EventType: "add_balance",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"amount": inputs.Amount,
		},
	}

	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubsub, err := common.SubscribeToResponse(redisClient, inputs.UserId, ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer pubsub.Close()

	if err := common.PushToQueue(redisClient, "main_queue", event, 10*time.Second); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error pushing to queue")
	}

	response, err := common.GetMessage(pubsub, ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error waiting for response")
	}

	return c.JSON(response.Data)
}
