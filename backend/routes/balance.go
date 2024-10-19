package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/controllers/orderbook"
)

func Balances(app *fiber.App) {
	app.Get("/balances/inr/:userId", orderbook.GetOrderBook)
}
