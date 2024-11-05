// shared/user_processor.go
package shared

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type EventModel struct {
	UserId    string    `json:"userId"`
	EventType string    `json:"eventType"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]interface{}
}

type ResponseModel struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
type UserProcessor struct {
	RedisClient *redis.Client
}
