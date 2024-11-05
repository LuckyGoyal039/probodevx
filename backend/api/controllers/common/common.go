package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/probodevx/engine/shared"
	"github.com/redis/go-redis/v9"
)

// SubscribeToResponse subscribes to the user response channel and returns the channel.
func SubscribeToResponse(redisClient *redis.Client, userId string, ctx context.Context, channelName string) (*redis.PubSub, error) {
	var responseChan string
	if channelName == "" {
		responseChan = fmt.Sprintf("user_response_%s", userId)
	} else {
		responseChan = channelName
	}

	pubsub := redisClient.Subscribe(ctx, responseChan)
	return pubsub, nil
}

// GetMessage waits for a message from the pubsub and returns the response.
func GetMessage(pubsub *redis.PubSub, ctx context.Context) (shared.ResponseModel, error) {
	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		return shared.ResponseModel{}, fmt.Errorf("error waiting for response: %w", err)
	}

	var response shared.ResponseModel
	if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
		return shared.ResponseModel{}, fmt.Errorf("error parsing response: %w", err)
	}
	return response, nil
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
