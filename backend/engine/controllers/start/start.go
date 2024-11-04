package start

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/probodevx/engine/controllers/user"
	"github.com/probodevx/engine/shared"
	"github.com/redis/go-redis/v9"
)

type LocalUserProcessor struct {
	shared.UserProcessor
}

func NewUserProcessor(redisClient *redis.Client) *LocalUserProcessor {
	if redisClient == nil {
		panic("Redis client cannot be nil")
	}
	return &LocalUserProcessor{
		UserProcessor: shared.UserProcessor{
			RedisClient: redisClient,
		},
	}
}

type EventHandler func(ctx context.Context, event shared.UserEvent) (interface{}, error)

var eventHandlers = map[string]EventHandler{
	"create_user": user.CreateNewUser,
	// "onramp_inr":    handleOrderBookEvent,
	// "create_symbol": handleUserEvent,
	// "reset":         handleUserEvent,
	// "orderbook":     handleUserEvent,
	// "inr_balance":   handleUserEvent,
	// "stock_balance": handleUserEvent,
	// "buy_order":     handleUserEvent,
	// "sell_order":    handleUserEvent,
	// "trade_mint":    handleUserEvent,
}

func (p *LocalUserProcessor) StartProcessing(ctx context.Context) error {
	log.Println("Starting event processing...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Blocking read from Redis queue
			result, err := p.RedisClient.BRPop(ctx, 5*time.Second, "user_events").Result()
			if err != nil {
				if err == redis.Nil {
					log.Println("No new events in the queue, continuing to wait...")
					continue
				}
				log.Printf("Error reading from queue: %v", err)
				continue
			}

			log.Println("Event read from queue, processing...")

			// Parse the event
			var event shared.UserEvent
			if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
				log.Printf("Error parsing event: %v", err)
				continue
			}

			// Look up the handler based on event type
			handler, exists := eventHandlers[event.EventType]
			if !exists {
				log.Printf("No handler found for event type: %s", event.EventType)
				continue
			}

			log.Printf("Processing %s event for user: %s", event.EventType, event.UserId)

			// Call the event handler
			response, err := handler(ctx, event)
			if err != nil {
				log.Printf("Error processing %s event: %v", event.EventType, err)
				continue
			}

			// Publish response
			responseChan := fmt.Sprintf("user_response_%s", event.UserId)
			responseJson, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshaling response: %v", err)
				continue
			}

			log.Printf("Publishing response to channel: %s", responseChan) // Add this log

			if err := p.RedisClient.Publish(ctx, responseChan, responseJson).Err(); err != nil {
				log.Printf("Error publishing response: %v", err)
				continue
			}

			log.Printf("Successfully processed %s event for user: %s", event.EventType, event.UserId)
		}
	}
}

// func (p *UserProcessor) processUserEvent(event UserEvent) map[string]interface{} {
// 	// Add your user creation logic here
// 	// This is where you'd interact with your database

// 	// Simulate processing time
// 	time.Sleep(time.Second)

// 	return map[string]interface{}{
// 		"status":  "success",
// 		"message": fmt.Sprintf("User %s created successfully", event.UserId),
// 		"userId":  event.UserId,
// 	}
// }
