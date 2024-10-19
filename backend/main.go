package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	inrBalance "github.com/probodevx/controllers/inrbalance"
	"github.com/probodevx/controllers/orderbook"
	"github.com/probodevx/controllers/reset"
	"github.com/probodevx/controllers/stock"
	"github.com/probodevx/routes"
)

func main() {

	// connect db here

	app := fiber.New()

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8000"
	}

	routes.UserRoutes(app)
	app.Post("/symbol/create/:stockSymbol", stock.CreateStock)
	app.Post("/reset", reset.ResetAll)
	app.Get("/orderbook/:userId?", orderbook.GetOrderBook)
	app.Get("/balances/inr/:userId?", inrBalance.GetInrBalance)
	app.Get("/balances/stock", stock.GetStockBalances)
	app.Post("/onramp/inr", inrBalance.AddUserBalance)
	app.Post("/trade/mint", stock.MintStock)
	// app.Post("/order/sell", orderbook.SellOrder)
	// app.Post("/order/buy", orderbook.BuyOrder)
	// app.Get("/balances/inr", orderbook.GetOrderBook)
	// app.Use("/user", routes.UserRoutes)
	// app.Use("/user", routes.UserRoutes)

	app.Listen(fmt.Sprintf(":%s", PORT))

}
