// shared/user_processor.go
package shared

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type UserEvent struct {
	UserId    string    `json:"userId"`
	EventType string    `json:"eventType"`
	Timestamp time.Time `json:"timestamp"`
}

type UserProcessor struct {
	RedisClient *redis.Client
}
