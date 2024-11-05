package apiuser

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/api/controllers/common"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/shared"
)

func CreateNewUser(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).SendString("invalid userId")
	}

	event := shared.EventModel{
		UserId:      userId,
		EventType:   "create_user",
		Timestamp:   time.Now(),
		ChannelName: "",
		Data:        make(map[string]interface{}),
	}

	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubsub, err := common.SubscribeToResponse(redisClient, userId, ctx, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer pubsub.Close()

	if err := common.PushToQueue(redisClient, "main_queue", event, 10*time.Second); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	response, err := common.GetMessage(pubsub, ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if !response.Success {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": response.Error,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.Data)
}
