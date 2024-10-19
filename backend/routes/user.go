package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/controllers/user"
)

func UserRoutes(app *fiber.App) {
	userGroup := app.Group("/user")
	userGroup.Post("/create/:userId", user.CreateNewUser)
	userGroup.Get("/all", user.GetAllUsers)
}
