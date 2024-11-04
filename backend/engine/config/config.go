package configEngine

import "github.com/redis/go-redis/v9"

type UserProcessor struct {
	redisClient *redis.Client
}

func NewUserProcessor(redisClient *redis.Client) *UserProcessor {
	if redisClient == nil {
		panic("Redis client cannot be nil")
	}
	return &UserProcessor{
		redisClient: redisClient,
	}
}
