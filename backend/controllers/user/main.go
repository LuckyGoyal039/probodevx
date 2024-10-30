// user/handlers.go
package user

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	redis "github.com/probodevx/config"
	"github.com/probodevx/global"
)

func CreateNewUser(c *fiber.Ctx) error {

	//just like register
	//create a new user in db

	//after creating from in the db
	// push an event in the queue
	redisClient := redis.GetRedisClient()
	ctx := context.TODO()
	eventKey := "create_user"
	data := c.Params("userId")
	if _, err := redisClient.LPush(ctx, eventKey, data).Result(); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("error pushing queue")
	}

	//if success then wait for the pubsub to get the event(confirmation)
	// if get the confirmation the return
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
