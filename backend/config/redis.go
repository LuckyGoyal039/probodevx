package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func ConnectRedis(redisHost string, redisPort string, redisPassword string) *redis.Client {
	redisUrl := redisHost + ":" + redisPort
	Client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})

	if err := Client.Ping(context.TODO()).Err(); err != nil {
		panic("failed to start redis: " + err.Error())
	}

	fmt.Printf("connected to redis on port: %s\n", redisPort)
	return Client
}
