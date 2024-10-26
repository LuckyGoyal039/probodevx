package orderbook

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/probodevx/data"
	"github.com/probodevx/global"
)

func GetOrderbookSymbol(c *fiber.Ctx) error {
	stockSymbol := utils.CopyString(c.Params("stockSymbol"))
	if stockSymbol == "" {
		return c.JSON(global.OrderBookManager.GetAllOrderBook())
	}
	newData, exists := global.OrderBookManager.GetOrderBook(stockSymbol)

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("stock symbol not found")
	}
	return c.JSON(newData)
}

func SellOrder(c *fiber.Ctx) error {
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
	//
	userStockBalances, userExists := global.StockManager.GetStockBalances(inputData.UserId)
	if !userExists {
		return c.Status(fiber.StatusNotFound).SendString("user not found")
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
	global.StockManager.UpdateStockBalanceSymbol(inputData.UserId, inputData.StockSymbol, stockBalance)
	// data.STOCK_BALANCES[inputData.UserId][inputData.StockSymbol] = stockBalance

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

		orderBook[strPrice] = priceOption
		return orderBook
	}

	stockData, exists := global.OrderBookManager.GetOrderBook(inputData.StockSymbol)
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

	global.OrderBookManager.UpdateOrderBookSymbol(inputData.StockSymbol, stockData)

	return c.SendString(fmt.Sprintf("Sell order placed for %v '%s' options at price %v.", inputData.Quantity, inputData.StockType, inputData.Price))
}

type inputFormat struct {
	UserId      string `json:"userId"`
	StockSymbol string `json:"stockSymbol"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
	StockType   string `json:"stockType"`
}

func BuyOrder(c *fiber.Ctx) error {

	var inputData inputFormat
	if err := c.BodyParser(&inputData); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input data")
	}
	ok, err := checkAndLockBalance(inputData.UserId, inputData.Price, inputData.Quantity)
	if !ok {
		if err.Error() == "User not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{"message": err.Error()})
	}

	// check it can place order or not
	canPlace := CheckCanPlaceOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType)

	switch canPlace {
	case "fulfill":
		PlaceFullFillOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
	case "partial":
		PlacePartialOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
	case "none":
		PlaceReverseBuyOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
	}
	// send this event to redis queue with symbol
	if orderbookData, exists := global.OrderBookManager.GetOrderBook(inputData.StockSymbol); exists {
		err := PushInQueue(inputData.StockSymbol, orderbookData)
		if err != nil {
			panic(fmt.Sprintf("error: %s", err))
		}
	}
	return c.JSON(fiber.Map{"message": "Buy order placed and trade executed"})
}
