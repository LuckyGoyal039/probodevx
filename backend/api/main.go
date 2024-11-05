package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	apiInrBalance "github.com/probodevx/api/controllers/inrbalance"
	apiuser "github.com/probodevx/api/controllers/user"
	redis "github.com/probodevx/config"
	"github.com/probodevx/controllers/mint"
	"github.com/probodevx/controllers/orderbook"
	"github.com/probodevx/controllers/reset"
	"github.com/probodevx/controllers/stock"
)

func main() {
	app := fiber.New(fiber.Config{
		Immutable: true,
		Prefork:   false,
	})

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Connect to Redis
	redis.ConnectRedis(redisHost, redisPort, redisPassword)
	if err := redis.CheckRedisConnection(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	API_PORT := os.Getenv("API_PORT")
	if API_PORT == "" {
		API_PORT = "8000"
	}

	// routes.UserRoutes(app)
	app.Post("/user/create/:userId", apiuser.CreateNewUser)
	// app.Get("/user/all", user.GetAllUsers)
	app.Get("/balances/inr/:userId?", apiInrBalance.GetInrBalance)
	app.Post("/onramp/inr", apiInrBalance.AddUserBalance)
	app.Post("/symbol/create/:stockSymbol", stock.CreateStock)
	app.Post("/reset", reset.ResetAll)
	app.Get("/orderbook/:stockSymbol?", orderbook.GetOrderbookSymbol)
	app.Get("/balances/stock/:userId?", stock.GetStockBalances)
	app.Post("/order/buy", orderbook.BuyOrder)
	app.Post("/order/sell", orderbook.SellOrder)
	// app.Get("/balances/inr", orderbook.GetOrderb
	app.Post("/trade/mint", mint.MintStock)
	app.Listen(fmt.Sprintf(":%s", API_PORT))
}
