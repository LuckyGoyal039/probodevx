package reset_api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	redis "github.com/probodevx/config"
)

func ResetAll(c *fiber.Ctx) error {
	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message := []byte("reset_all")

	if _, err := redisClient.LPush(ctx, "reset_all", message).Result(); err != nil {
		println("Error push in queue, %s", err)
		return c.Status(fiber.StatusInternalServerError).SendString("something went wrong")
	}

	pubsub := redisClient.Subscribe(ctx, "reset_all")
	defer pubsub.Close()

	if _, err := pubsub.ReceiveMessage(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error waiting for response")
	}

	// if ok := data.ResetAllManager(global.UserManager, global.StockManager, global.OrderBookManager); !ok {
	// 	return c.SendString("something went wrong")
	// }
	return c.SendString("reset successfully")
}
