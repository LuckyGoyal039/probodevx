package stock

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/global"
)

func GetStockBalances(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return c.JSON(global.StockManager.GetAllStockBalances())
	}
	newData, exists := global.StockManager.GetStockBalances(userId)

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}
	return c.JSON(newData)
}

func CreateStock(c *fiber.Ctx) error {
	stockSymbol := c.Params("stockSymbol")

	_, exists := global.OrderBookManager.GetOrderBook(stockSymbol)

	if exists {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": fmt.Sprintf("Symbol %s already exists", stockSymbol)})
	}

	global.OrderBookManager.AddOrderBookSymbol(stockSymbol)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": fmt.Sprintf("Symbol %s created", stockSymbol)})
}

// for k, userStocks := range data.STOCK_BALANCES {
// 	var val valMap
// 	stockMap, ok := userStocks.(map[string]valMap)
// 	if !ok {
// 		stockMap = make(map[string]valMap)
// 	}
// 	if _, exists := stockMap[stockSymbol]; !exists {
// 		stockMap[stockSymbol] = val
// 	}
// 	data.STOCK_BALANCES[k] = stockMap
// }
// return c.JSON(fiber.Map{
// 	"data": data.ORDERBOOK,
// })

// func MintStock(c *fiber.Ctx) error {
// 	type inputFormat struct {
// 		UserId      string  `json:"userId"`
// 		StockSymbol string  `json:"stockSymbol"`
// 		Quantity    float32 `json:"quantity"`
// 		Price       float32 `json:"price"`
// 	}

// 	var incomingData inputFormat
// 	err := c.BodyParser(&incomingData)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
// 	}

// 	stockData, exists := data.STOCK_BALANCES[incomingData.UserId].(map[string]interface{})
// 	if !exists {
// 		return c.Status(fiber.StatusBadRequest).SendString("User not found")
// 	}

// 	userData, exists := data.INR_BALANCES[incomingData.UserId]
// 	if !exists {
// 		return c.Status(fiber.StatusBadRequest).SendString("User not found")
// 	}

// 	currBalance := userData.Balance
// 	var requiredBalance = incomingData.Quantity * incomingData.Price * 2
// 	if requiredBalance > currBalance {
// 		return c.Status(fiber.StatusPaymentRequired).SendString("insufficient balance")
// 	}

// 	currBalance -= requiredBalance

// 	userData.Balance = currBalance
// 	data.INR_BALANCES[incomingData.UserId] = userData

// 	var mintData = make(map[string]interface{})
// 	mintData[incomingData.StockSymbol] = map[string]interface{}{
// 		"yes": map[string]interface{}{
// 			"quantity": incomingData.Quantity,
// 			"locked":   0,
// 		},
// 		"no": map[string]interface{}{
// 			"quantity": incomingData.Quantity,
// 			"locked":   0,
// 		},
// 	}
// 	var finalData = make(map[string]interface{})
// 	for k, v := range mintData {
// 		finalData[k] = v
// 	}
// 	for k, v := range stockData {
// 		finalData[k] = v
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": fmt.Sprintf("Minted %v 'yes' and 'no' tokens for user %s, remaining balance is %v", incomingData.Quantity, incomingData.UserId, currBalance)})
// }
