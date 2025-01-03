package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/probodevx/engine/shared"
	"github.com/redis/go-redis/v9"
)

// SubscribeToResponse subscribes to the user response channel and returns the channel.
func SubscribeToResponse(redisClient *redis.Client, userId string, ctx context.Context, channelName string) (*redis.PubSub, error) {
	// var responseChan string
	// if channelName == "" {
	// 	responseChan = fmt.Sprintf("user_response_%s", userId)
	// } else {
	// 	responseChan = channelName
	// }

	pubsub := redisClient.Subscribe(ctx, channelName)
	return pubsub, nil
}

// GetMessage waits for a message from the pubsub and returns the response.
func GetMessage(pubsub *redis.PubSub, ctx context.Context, userId string) (shared.ResponseModel, error) {
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			return shared.ResponseModel{}, fmt.Errorf("error waiting for response: %w", err)
		}

		var response shared.ResponseModel
		if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
			return shared.ResponseModel{}, fmt.Errorf("error parsing response: %w", err)
		}
		if userId == "" || response.User == userId {
			return response, nil
		}
	}
}

func PushToQueue(redisClient *redis.Client, queueName string, event interface{}, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Serialize the event to JSON
	eventJson, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error serializing event: %v", err)
		return fmt.Errorf("error serializing event")
	}

	// Push the event to the Redis queue
	if _, err := redisClient.LPush(ctx, queueName, eventJson).Result(); err != nil {
		log.Printf("Error pushing event to queue %s: %v", queueName, err)
		return fmt.Errorf("error pushing to queue")
	}

	return nil
}

func GetMapKeys(m interface{}) []string {
	// Use reflection to check if the input is a map
	val := reflect.ValueOf(m)
	if val.Kind() != reflect.Map {
		return nil
	}

	// Create a slice to store the keys
	keys := make([]string, 0, val.Len())

	// Iterate over the map and extract the keys
	for _, key := range val.MapKeys() {
		keys = append(keys, key.String())
	}

	return keys
}
