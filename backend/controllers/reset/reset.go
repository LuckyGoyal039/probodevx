package reset

import (
	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/data"
)

func ResetAll(c *fiber.Ctx) error {
	data.INR_BALANCES = make(map[string]data.User)
	data.ORDERBOOK = make(map[string]data.OrderSymbol)
	data.STOCK_BALANCES = make(map[string]data.UserStockBalance)
	return c.SendString("reset successfully")
}
