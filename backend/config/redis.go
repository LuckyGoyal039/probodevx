package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func ConnectRedis(redisHost string, redisPort string, redisPassword string) *redis.Client {
	redisUrl := redisHost + ":" + redisPort

	Redis = redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})

	ctx := context.Background()
	_, err := Redis.Ping(ctx).Result()
	if err != nil {
		panic("failed to connect to redis: " + err.Error())
	}

	fmt.Printf("connected to redis on port: %s\n", redisPort)
	return Redis
}

func GetRedisClient() *redis.Client {
	return Redis
}

func CheckRedisConnection() error {
	if Redis == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	ctx := context.Background()
	_, err := Redis.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis connection error: %v", err)
	}

	return nil
}
