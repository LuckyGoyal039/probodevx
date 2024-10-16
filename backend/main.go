package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
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
	// app.Use("/user", routes.UserRoutes)
	// app.Use("/user", routes.UserRoutes)

	app.Listen(fmt.Sprintf(":%s", PORT))

}
