package mint

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/global"
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

	if err := updateBalance(inputData.Price, inputData.Quantity, inputData.UserId); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("insufficient balance")
	}
	global.StockManager.AddStockBalancesSymbol(inputData.StockSymbol)
	stockData, exist := global.StockManager.GetStockBalances(inputData.UserId)
	if !exist {
		return c.Status(fiber.StatusBadRequest).SendString("user not found")
	}
	data := stockData[inputData.StockSymbol]
	data.No.Quantity = inputData.Quantity
	data.Yes.Quantity = inputData.Quantity
	global.StockManager.UpdateStockBalanceSymbol(inputData.UserId, inputData.StockSymbol, data)

	global.OrderBookManager.AddOrderBookSymbol(inputData.StockSymbol)

	currentBalance, exist := global.UserManager.GetUserBalance(inputData.UserId)
	if !exist {
		return c.Status(fiber.StatusBadRequest).SendString("user not found")
	}
	return c.Status(fiber.StatusCreated).SendString(fmt.Sprintf("Minted %v 'yes' and 'no' tokens for user %s, remaining balance is %v", inputData.Quantity, inputData.UserId, currentBalance))
}
