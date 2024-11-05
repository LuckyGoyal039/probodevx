package mint_api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/api/controllers/common"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/shared"
)

type mintInputs struct {
	UserId      string `json:"userId"`
	StockSymbol string `json:"stockSymbol"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
}

func MintStock(c *fiber.Ctx) error {
	var inputData mintInputs

	if err := c.BodyParser(&inputData); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid credentials")
	}
	event := shared.EventModel{
		UserId:      inputData.UserId,
		Timestamp:   time.Now(),
		EventType:   "trade_mint",
		ChannelName: "",
		Data: map[string]interface{}{
			"quantity":    inputData.Quantity,
			"price":       inputData.Price,
			"stockSymbol": inputData.StockSymbol,
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
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if !response.Success {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": response.Error,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.Data)

	// if err := updateBalance(inputData.Price, inputData.Quantity, inputData.UserId); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).SendString("insufficient balance")
	// }
	// stockData, exist := global.StockManager.GetStockBalances(inputData.UserId)
	// if !exist {
	// 	global.StockManager.AddNewUser(inputData.UserId)
	// }
	// global.StockManager.AddStockBalancesSymbol(inputData.StockSymbol)
	// data := stockData[inputData.StockSymbol]
	// data.No.Quantity = inputData.Quantity
	// data.Yes.Quantity = inputData.Quantity
	// global.StockManager.UpdateStockBalanceSymbol(inputData.UserId, inputData.StockSymbol, data)

	// global.OrderBookManager.AddOrderBookSymbol(inputData.StockSymbol)

	// currentBalance, exist := global.UserManager.GetUserBalance(inputData.UserId)
	// if !exist {
	// 	return c.Status(fiber.StatusBadRequest).SendString("user not found")
	// }
	// return c.Status(fiber.StatusOK).JSON(fiber.Map{
	// 	"message": fmt.Sprintf("Minted %v 'yes' and 'no' tokens for user %s, remaining balance is %v", inputData.Quantity, inputData.UserId, currentBalance),
	// })
}
