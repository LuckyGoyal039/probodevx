package inrBalance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	redis "github.com/probodevx/config"
	"github.com/probodevx/engine/global"
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
	jsonData, err := json.Marshal(inputs)
	if err != nil {
		println("unable to marshal data, %s", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	redisClient := redis.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := redisClient.LPush(ctx, "add_balance", jsonData).Result(); err != nil {
		println("error pushing in queue, %s", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	responseChan := fmt.Sprintf("user_response_%s", inputs.UserId)
	pubsub := redisClient.Subscribe(ctx, responseChan)
	defer pubsub.Close()

	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		println("error waiting for response, %s", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
		println("error parsing response, %s", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// userData, exists := global.UserManager.GetUser(inputs.UserId)

	// if !exists {
	// 	return c.Status(fiber.StatusNotFound).SendString("User not found")
	// }
	// totalBal := userData.Balance + inputs.Amount
	// if _, err := global.UserManager.UpdateUserInrBalance(inputs.UserId, totalBal); err != nil {
	// 	return c.SendString("invalid inputs")
	// }
	return c.JSON(fiber.Map{"message": response})
	// return c.JSON(fiber.Map{"message": fmt.Sprintf("Onramped %s with amount %v", inputs.UserId, inputs.Amount)})
}
