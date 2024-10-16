package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/probodevx/data"
)

func CreateNewUser(c *fiber.Ctx) error {
	userId := uuid.New().String()
	data.INR_BALANCES[userId] = data.User{
		Balance: 0,
		Locked:  0,
	}
	return c.JSON(fiber.Map{
		"msg":  "User created successfully",
		userId: data.INR_BALANCES[userId],
	})
}

func GetAllUsers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data": data.INR_BALANCES,
	})
}
