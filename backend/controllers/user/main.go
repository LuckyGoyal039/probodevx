// user/handlers.go
package user

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/probodevx/global"
)

func CreateNewUser(c *fiber.Ctx) error {
	userId := utils.CopyString(c.Params("userId"))
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).SendString("invalid userId")
	}

	if err := global.UserManager.CreateUser(userId); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": fmt.Sprintf("User %s created", userId),
	})
}

func GetAllUsers(c *fiber.Ctx) error {
	users := global.UserManager.GetAllUsers()
	return c.JSON(fiber.Map{
		"data": users,
	})
}
