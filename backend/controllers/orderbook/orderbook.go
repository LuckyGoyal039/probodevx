package orderbook

import (
	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/data"
)

func GetOrderBook(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return c.JSON(data.ORDERBOOK)
	}
	newData, exists := data.ORDERBOOK[userId]

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}
	return c.JSON(newData)
}

// func SellOrder(c *fiber.Ctx) error {
// 	type inputFormat struct {
// 		UserId      string  `json:"userId"`
// 		StockSymbol string  `json:"stockSymbol"`
// 		Quantity    int     `json:"quantity"`
// 		Price       float32 `json:"price"`
// 		StockType   string  `json:"stockType"`
// 	}
// 	var inputData inputFormat
// 	err := c.BodyParser(&inputData)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
// 	}
// 	stockData, exists := data.ORDERBOOK[inputData.StockSymbol]

// 	if !exists {
// 		data.ORDERBOOK[inputData.StockSymbol] = make(map[string]interface{})
// 		stockData = data.ORDERBOOK[inputData.StockSymbol]
// 	}
// 	var total float32
// 	total += float32(inputData.Quantity)
// 	var yes = make(map[string]interface{})
// 	var prices = make(map[string]interface{})
// 	var orders = make(map[string]interface{})
// 	orders[inputData.UserId]+=

// 	stoctype, exists := stockData[inputData.StockType]

// }

// func BuyOrder(c *fiber.Ctx) error {
// 	type inputFormat struct {
// 		UserId      string  `json:"userId"`
// 		StockSymbol string  `json:"stockSymbol"`
// 		Quantity    int     `json:"quantity"`
// 		Price       float32 `json:"price"`
// 		StockType   string  `json:"stockType"`
// 	}
// 	var inputData inputFormat
// 	err := c.BodyParser(&inputData)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
// 	}
// }
