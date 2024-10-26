package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	redis "github.com/probodevx/config"
	inrBalance "github.com/probodevx/controllers/inrbalance"
	"github.com/probodevx/controllers/orderbook"
	"github.com/probodevx/controllers/reset"
	"github.com/probodevx/controllers/stock"
	"github.com/probodevx/routes"
)

func main() {

	// connect db here

	app := fiber.New(fiber.Config{
		Immutable: true,
		Prefork:   false,
	})

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8000"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	// redisUsername := os.Getenv("REDIS_USERNAME")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Connect to Redis
	redis.ConnectRedis(redisHost, redisPort, redisPassword)
	if err := redis.CheckRedisConnection(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	// userManager := data.NewUserManager()
	routes.UserRoutes(app)
	//not working
	app.Post("/onramp/inr", inrBalance.AddUserBalance)

	app.Post("/symbol/create/:stockSymbol", stock.CreateStock)
	app.Post("/reset", reset.ResetAll)
	app.Get("/orderbook/:stockSymbol?", orderbook.GetOrderbookSymbol)
	app.Get("/balances/inr/:userId?", inrBalance.GetInrBalance)
	app.Get("/balances/stock/:userId?", stock.GetStockBalances)
	// // app.Post("/trade/mint", stock.MintStock)
	app.Post("/order/buy", orderbook.BuyOrder)
	app.Post("/order/sell", orderbook.SellOrder)
	app.Get("/balances/inr", orderbook.GetOrderbookSymbol)

	app.Listen(fmt.Sprintf(":%s", PORT))

}
