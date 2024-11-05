package apiStock

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/api/controllers/common"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/shared"
)

func GetStockBalances(c *fiber.Ctx) error {
	userId := c.Params("userId")
	event := shared.EventModel{
		UserId:      userId,
		Timestamp:   time.Now(),
		Data:        make(map[string]interface{}),
		EventType:   "get_stock_balance",
		ChannelName: "stock_balances",
	}
	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pubsub, err := common.SubscribeToResponse(redisClient, userId, ctx, "stock_balances")
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

	// if userId == "" {
	// 	return c.JSON(global.StockManager.GetAllStockBalances())
	// }
	// newData, exists := global.StockManager.GetStockBalances(userId)

	// if !exists {
	// 	return c.Status(fiber.StatusNotFound).SendString("User not found")
	// }
	return c.JSON(response)
}

func CreateStock(c *fiber.Ctx) error {

	stockSymbol := c.Params("stockSymbol")

	event := shared.EventModel{
		UserId:      "",
		EventType:   "create_symbol",
		Timestamp:   time.Now(),
		ChannelName: "",
		Data: map[string]interface{}{
			"stockSymbol": stockSymbol,
		},
	}
	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pubsub, err := common.SubscribeToResponse(redisClient, "", ctx, "")
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