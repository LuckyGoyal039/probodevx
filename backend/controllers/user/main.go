package user

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/data"
)

func CreateNewUser(c *fiber.Ctx) error {
	// userId := uuid.New().String()
	userId := c.Params("userId")
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).SendString("invalid userId")
	}
	data.INR_BALANCES[userId] = data.User{
		Balance: 0,
		Locked:  0,
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": fmt.Sprintf("User %s created", userId),
	},
	)
}

func GetAllUsers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data": data.INR_BALANCES,
	})
}
