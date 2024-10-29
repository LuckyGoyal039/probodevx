package orderbook

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
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
	type inputFormat struct {
		UserId      string `json:"userId"`
		StockSymbol string `json:"stockSymbol"`
		Quantity    int    `json:"quantity"`
		Price       int    `json:"price"`
		StockType   string `json:"stockType"`
	}

	var inputData inputFormat
	err := c.BodyParser(&inputData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid inputs")
	}
	ok := CheckAndLockStock(inputData.UserId, inputData.StockSymbol, inputData.StockType, inputData.Quantity)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Insufficient stock balance"})
	}

	canPlace := CheckBuyer(inputData.StockSymbol, inputData.StockType, inputData.Price, inputData.Quantity)
	switch canPlace {
	case "fullfill":
		FullFillSellOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		// UnLockBalance(inputData.UserId, inputData.Quantity)
	case "partial":
		PlacePartialSellOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		// updatedLockedAmount := inputData.Quantity - GetFullFillableQuantity(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType)
		// UnLockBalance(inputData.UserId, updatedLockedAmount)
	case "none":
		PlaceSellOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		// UnLockBalance(inputData.UserId, inputData.Quantity)
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Sell order placed for %v '%s' options at price %v.", inputData.Quantity, inputData.StockType, inputData.Price),
	})
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// check it can place order or not
	canPlace := CheckCanPlaceOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType)

	switch canPlace {
	case "fullfill":
		//check this out
		PlaceFullFillOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		UnLockBalance(inputData.UserId, inputData.Quantity, inputData.Price)
		// 	balance, _ := global.UserManager.GetUserBalance(inputData.UserId)
		// 	balance -= inputData.Price * inputData.Quantity
		// 	global.UserManager.UpdateUserInrBalance(inputData.UserId, balance)
		// case "partial":
		PlacePartialOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		updatedLockedAmount := inputData.Quantity - GetFullFillableQuantity(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType)
		UnLockBalance(inputData.UserId, updatedLockedAmount, inputData.Price)
		// balance, _ := global.UserManager.GetUserBalance(inputData.UserId)
		// balance -= inputData.Price * updatedLockedAmount
		// global.UserManager.UpdateUserInrBalance(inputData.UserId, balance)
	case "none":
		PlaceReverseBuyOrder(inputData.StockSymbol, inputData.Price, inputData.Quantity, inputData.StockType, inputData.UserId)
		// UnLockBalance(inputData.UserId, inputData.Quantity, inputData.Price)
	}
	if orderbookData, exists := global.OrderBookManager.GetOrderBook(inputData.StockSymbol); exists {
		err := PushInQueue(inputData.StockSymbol, orderbookData)
		if err != nil {
			panic(fmt.Sprintf("error: %s", err))
		}
	}
	return c.JSON(fiber.Map{"message": "Buy order placed and trade executed"})
}
