package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	redis "github.com/probodevx/config"
	inrBalance "github.com/probodevx/controllers/inrbalance"
	"github.com/probodevx/controllers/orderbook"
	"github.com/probodevx/controllers/reset"
	"github.com/probodevx/controllers/stock"
	"github.com/probodevx/controllers/wss"
	"github.com/probodevx/routes"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	app := fiber.New(fiber.Config{
		Immutable: true,
		Prefork:   false,
	})
	wsApp := fiber.New(fiber.Config{
		Immutable: true,
		Prefork:   false,
	})

	PORT := os.Getenv("PORT")
	WSPORT := os.Getenv("WSPORT")
	if WSPORT == "" {
		WSPORT = "8080"
	}

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
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Connect to Redis
	redis.ConnectRedis(redisHost, redisPort, redisPassword)
	if err := redis.CheckRedisConnection(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	// main server apis
	routes.UserRoutes(app)
	app.Post("/onramp/inr", inrBalance.AddUserBalance)
	app.Post("/symbol/create/:stockSymbol", stock.CreateStock)
	app.Post("/reset", reset.ResetAll)
	app.Get("/orderbook/:stockSymbol?", orderbook.GetOrderbookSymbol)
	app.Get("/balances/inr/:userId?", inrBalance.GetInrBalance)
	app.Get("/balances/stock/:userId?", stock.GetStockBalances)
	app.Post("/order/buy", orderbook.BuyOrder)
	app.Post("/order/sell", orderbook.SellOrder)
	app.Get("/balances/inr", orderbook.GetOrderbookSymbol)
	// app.Post("/trade/mint", stock.MintStock)

	//wss api

	//middleware
	wsApp.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	wsApp.Get("/ws/:event", websocket.New(func(c *websocket.Conn) {
		wss.ConnectSocket(c)
	}))
	wsApp.Post("/broadcast/:event", wss.BroadCastMessage)
	// wsApp.Get()
	go func() {
		wsApp.Listen(fmt.Sprintf(":%s", WSPORT))
	}()
	go func() {
		app.Listen(fmt.Sprintf(":%s", PORT))
	}()
	wg.Wait()
}
