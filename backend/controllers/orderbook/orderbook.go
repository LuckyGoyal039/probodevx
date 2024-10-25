package orderbook

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/data"
	"github.com/probodevx/global"
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

func createReverseOrder(stockSymbol, stockType, userId string, price float64, quantity int) {
	reversePrice := 1000 - price
	reverseType := "no"
	if stockType == "no" {
		reverseType = "yes"
	}

	availableSymbol, exists := data.ORDERBOOK[stockSymbol]
	if !exists {
		availableSymbol = data.OrderSymbol{
			Yes: make(data.OrderYesNo),
			No:  make(data.OrderYesNo),
		}
		data.ORDERBOOK[stockSymbol] = availableSymbol
	}

	priceStr := strconv.FormatFloat(reversePrice, 'f', 2, 64)

	//check
	// less than or equal to input price
	var reverseOrders data.PriceOptions
	if reverseType == "yes" {
		reverseOrders, exists = availableSymbol.Yes[priceStr]
	} else {
		reverseOrders, exists = availableSymbol.No[priceStr]
	}

	if !exists {
		reverseOrders = data.PriceOptions{
			Total:  quantity,
			Orders: make(map[string]data.OrderOptions),
		}
	} else {
		reverseOrders.Total += quantity
	}

	reverseOrders.Orders[userId] = data.OrderOptions{
		Quantity: quantity,
		Reverse:  true,
	}

	if reverseType == "yes" {
		availableSymbol.Yes[priceStr] = reverseOrders
	} else {
		availableSymbol.No[priceStr] = reverseOrders
	}

	data.ORDERBOOK[stockSymbol] = availableSymbol
}

func checkAndLockBalance(userId string, price int, quantity int) (bool, error) {
	user, exists := global.UserManager.GetUser(userId)
	if !exists {
		return false, fmt.Errorf("User not found")
	}

	totalCost := price * quantity

	if user.Balance < totalCost {
		return false, fmt.Errorf("Insufficient balance")
	}
	leftBalance := user.Balance - totalCost
	lockedAmount := user.Locked + totalCost

	global.UserManager.UpdateUserInrBalance(userId, leftBalance)
	global.UserManager.UpdateUserInrLock(userId, lockedAmount)

	return true, nil
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
	// createReverseOrder(inputData.StockSymbol, inputData.StockType, inputData.UserId, inputData.Price, inputData.Quantity)
	// return c.JSON(fiber.Map{"message": "Buy order placed and reverse order created"})

	return c.JSON(fiber.Map{"message": "Buy order placed and trade executed"})
}
