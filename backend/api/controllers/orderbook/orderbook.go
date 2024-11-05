package api_orderbook

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/api/controllers/common"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/shared"
)

func GetOrderbookSymbol(c *fiber.Ctx) error {
	stockSymbol := c.Params("stockSymbol")
	event := shared.EventModel{
		UserId:      "",
		Timestamp:   time.Now(),
		EventType:   "orderbook",
		ChannelName: "get_orderbook",
		Data: map[string]interface{}{
			"stockSymbol": stockSymbol,
		},
	}
	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubsub, err := common.SubscribeToResponse(redisClient, "", ctx, "get_orderbook")
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
	return c.JSON(response.Data)
}

type inputFormat struct {
	UserId      string `json:"userId"`
	StockSymbol string `json:"stockSymbol"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
	StockType   string `json:"stockType"`
}

func SellOrder(c *fiber.Ctx) error {

	var inputData inputFormat
	err := c.BodyParser(&inputData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}
	event := shared.EventModel{
		UserId:      inputData.UserId,
		Timestamp:   time.Now(),
		EventType:   "sell_order",
		ChannelName: "",
		Data: map[string]interface{}{
			"userId":      inputData.UserId,
			"stockSymbol": inputData.StockSymbol,
			"quantity":    inputData.Quantity,
			"price":       inputData.Price,
			"stockType":   inputData.StockType,
		},
	}

	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubsub, err := common.SubscribeToResponse(redisClient, inputData.UserId, ctx, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer pubsub.Close()

	if err := common.PushToQueue(redisClient, "main_queue", event, 10*time.Second); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	response, err := common.GetMessage(pubsub, ctx)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if !response.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": response.Error,
		})
	}

	return c.JSON(response.Data)
}

func BuyOrder(c *fiber.Ctx) error {

	var inputData inputFormat
	err := c.BodyParser(&inputData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}
	event := shared.EventModel{
		UserId:      inputData.UserId,
		Timestamp:   time.Now(),
		EventType:   "buy_order",
		ChannelName: "",
		Data: map[string]interface{}{
			"userId":      inputData.UserId,
			"stockSymbol": inputData.StockSymbol,
			"quantity":    inputData.Quantity,
			"price":       inputData.Price,
			"stockType":   inputData.StockType,
		},
	}

	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubsub, err := common.SubscribeToResponse(redisClient, inputData.UserId, ctx, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer pubsub.Close()

	if err := common.PushToQueue(redisClient, "main_queue", event, 10*time.Second); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	response, err := common.GetMessage(pubsub, ctx)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if !response.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": response.Error,
		})
	}

	return c.JSON(response.Data)
}

func CancelOrder(c *fiber.Ctx) error {
	var inputData inputFormat
	err := c.BodyParser(&inputData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}
	event := shared.EventModel{
		UserId:      inputData.UserId,
		Timestamp:   time.Now(),
		EventType:   "cancel_order",
		ChannelName: "",
		Data: map[string]interface{}{
			"userId":      inputData.UserId,
			"stockSymbol": inputData.StockSymbol,
			"quantity":    inputData.Quantity,
			"price":       inputData.Price,
			"stockType":   inputData.StockType,
		},
	}

	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubsub, err := common.SubscribeToResponse(redisClient, inputData.UserId, ctx, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer pubsub.Close()

	if err := common.PushToQueue(redisClient, "main_queue", event, 10*time.Second); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	response, err := common.GetMessage(pubsub, ctx)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if !response.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": response.Error,
		})
	}

	return c.JSON(response.Data)
}
