package reset

import (
	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/engine/data"
	"github.com/probodevx/engine/global"
)

func ResetAll(c *fiber.Ctx) error {

	if ok := data.ResetAllManager(global.UserManager, global.StockManager, global.OrderBookManager); !ok {
		return c.SendString("something went wrong")
	}
	return c.SendString("reset successfully")
}
