package inrBalance

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/global"
)

func GetInrBalance(c *fiber.Ctx) error {

	userId := c.Params("userId")
	if userId == "" {
		return c.JSON(global.UserManager.INR_BALANCES)
	}
	newData, exists := global.UserManager.INR_BALANCES[userId]

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}
	return c.JSON(newData)
}

func AddUserBalance(c *fiber.Ctx) error {
	type User struct {
		UserId string  `json:"userId"`
		Amount float64 `json:"amount"`
	}
	var inputs User
	err := c.BodyParser(&inputs)
	if err != nil {
		return c.SendString("invalid inputs")
	}
	userData, exists := global.UserManager.INR_BALANCES[inputs.UserId]

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}
	userData.Balance += float32(inputs.Amount)
	global.UserManager.INR_BALANCES[inputs.UserId] = userData
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Onramped %s with amount %v", inputs.UserId, inputs.Amount)})
}
