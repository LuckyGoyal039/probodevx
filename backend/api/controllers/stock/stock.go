package stock

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/engine/global"
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
