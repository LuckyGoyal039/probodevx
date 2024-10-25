package reset

import (
	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/data"
	"github.com/probodevx/global"
)

func ResetAll(c *fiber.Ctx) error {

	data.ResetAllManager(global.UserManager, global.StockManager, global.OrderBookManager)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "All data reset successfully",
	})
}
