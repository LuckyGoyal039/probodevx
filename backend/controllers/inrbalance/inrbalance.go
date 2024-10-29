package inrBalance

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/probodevx/global"
)

func GetInrBalance(c *fiber.Ctx) error {

	userId := c.Params("userId")
	if userId == "" {
		return c.JSON(global.UserManager.GetAllUsers())
	}
	newData, exists := global.UserManager.GetUser(userId)

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}
	return c.JSON(newData)
}

func AddUserBalance(c *fiber.Ctx) error {
	type User struct {
		UserId string `json:"userId"`
		Amount int    `json:"amount"`
	}
	var inputs User
	err := c.BodyParser(&inputs)
	if err != nil {
		return c.SendString("invalid inputs")
	}
	userData, exists := global.UserManager.GetUser(inputs.UserId)

	if !exists {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}
	totalBal := userData.Balance + inputs.Amount
	if _, err := global.UserManager.UpdateUserInrBalance(inputs.UserId, totalBal); err != nil {
		return c.SendString("invalid inputs")
	}
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Onramped %s with amount %v", inputs.UserId, inputs.Amount)})
}
