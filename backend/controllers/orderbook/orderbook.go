package orderbook

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/data"
)

func GetOrderBook(c *fiber.Ctx) error {
	stockSymbol := c.Params("stockSymbol")
	if stockSymbol == "" {
		return c.JSON(data.ORDERBOOK)
	}
	newData, exists := data.ORDERBOOK[stockSymbol]

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("stock symbol not found")
	}
	return c.JSON(newData)
}

func SellOrder(c *fiber.Ctx) error {
	// body parse the data
	// check in stock_balances for that user
	// then check the symbol for that user
	// then check the yes or not quantity for that user
	// then deduct the quantity and lock the quantity
	// check for the symbol
	// check for price if not then create new
	// manage total
	// manage orders add user id and quantity in orders
	type inputFormat struct {
		UserId      string  `json:"userId"`
		StockSymbol string  `json:"stockSymbol"`
		Quantity    int     `json:"quantity"`
		Price       float32 `json:"price"`
		StockType   string  `json:"stockType"`
	}

	var inputData inputFormat
	err := c.BodyParser(&inputData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}

	userStockBalances, userExists := data.STOCK_BALANCES[inputData.UserId]
	if !userExists {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	stockBalance, stockExists := userStockBalances[inputData.StockSymbol]
	if !stockExists {
		return c.Status(fiber.StatusNotFound).SendString("Symbol not found")
	}

	getAvailableQuantity := func(stockType string) (int, error) {
		switch stockType {
		case "yes":
			return stockBalance.Yes.Quantity, nil
		case "no":
			return stockBalance.No.Quantity, nil
		default:
			return 0, fmt.Errorf("invalid stock type")
		}
	}

	availableQuantity, err := getAvailableQuantity(inputData.StockType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if availableQuantity < inputData.Quantity {
		return c.Status(fiber.StatusBadRequest).SendString("Insufficient balance")
	}

	switch inputData.StockType {
	case "yes":
		stockBalance.Yes.Quantity -= inputData.Quantity
	case "no":
		stockBalance.No.Quantity -= inputData.Quantity
	}
	data.STOCK_BALANCES[inputData.UserId][inputData.StockSymbol] = stockBalance

	updateOrderBook := func(orderBook data.OrderYesNo) data.OrderYesNo {
		strPrice := fmt.Sprintf("%.2f", inputData.Price)
		priceOption, exists := orderBook[strPrice]
		if !exists {
			priceOption = data.PriceOptions{
				Total:  0,
				Orders: make(data.Order),
			}
		}

		priceOption.Total += inputData.Quantity
		priceOption.Orders[inputData.UserId] += inputData.Quantity

		orderBook[strPrice] = priceOption
		return orderBook
	}

	stockData, exists := data.ORDERBOOK[inputData.StockSymbol]
	if !exists {
		stockData = data.OrderSymbol{
			Yes: make(data.OrderYesNo),
			No:  make(data.OrderYesNo),
		}
	}

	switch inputData.StockType {
	case "yes":
		stockData.Yes = updateOrderBook(stockData.Yes)
	case "no":
		stockData.No = updateOrderBook(stockData.No)
	}

	data.ORDERBOOK[inputData.StockSymbol] = stockData

	return c.SendString(fmt.Sprintf("Sell order placed for %v '%s' options at price %v.", inputData.Quantity, inputData.StockType, inputData.Price))
}

func BuyOrder(c *fiber.Ctx) error {
	type inputFormat struct {
		UserId      string  `json:"userId"`
		StockSymbol string  `json:"stockSymbol"`
		Quantity    int     `json:"quantity"`
		Price       float32 `json:"price"`
		StockType   string  `json:"stockType"`
	}
	var inputData inputFormat
	err := c.BodyParser(&inputData)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}

	return c.SendString("Buy order placed and trade executed")
}
