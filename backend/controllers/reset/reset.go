// reset/handlers.go
package reset

import (
	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/global"
)

func ResetAll(c *fiber.Ctx) error {
	// Phase 1: Lock all managers
	global.UserManager.Mu.Lock()
	defer global.UserManager.Mu.Unlock()

	global.OrderBookManager.Mu.Lock()
	defer global.OrderBookManager.Mu.Unlock()

	global.StockManager.Mu.Lock()
	defer global.StockManager.Mu.Unlock()

	// Phase 2: Reset all data
	global.UserManager.inrBalances = make(map[string]*User)
	global.OrderBookManager.orderBook = make(map[string]OrderSymbol)
	global.StockManager.stockBalances = make(map[string]UserStockBalance)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "All data reset successfully",
	})
}
