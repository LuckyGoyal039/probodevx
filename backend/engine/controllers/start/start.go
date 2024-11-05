package start

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	inrBalance "github.com/probodevx/engine/controllers/inrbalance"
	"github.com/probodevx/engine/controllers/mint"
	"github.com/probodevx/engine/controllers/reset"
	"github.com/probodevx/engine/controllers/stock"
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

type EventHandler func(ctx context.Context, event shared.EventModel) (interface{}, error)

var eventHandlers = map[string]EventHandler{
	"create_user":       user.CreateNewUser,
	"get_balance":       inrBalance.GetInrBalance,
	"onramp_inr":        inrBalance.AddUserBalance,
	"create_symbol":     stock.CreateStock,
	"get_stock_balance": stock.GetStockBalances,
	"reset":             reset.ResetAll,
	"trade_mint":        mint.MintStock,
	// "buy_order":     handleUserEvent,
	// "sell_order":    handleUserEvent,
	// "orderbook":     handleUserEvent,
	// "inr_balance":   handleUserEvent,
}

func (p *LocalUserProcessor) StartProcessing(ctx context.Context) error {
	log.Println("Starting event processing...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Blocking read from Redis queue
			result, err := p.RedisClient.BRPop(ctx, 5*time.Second, "main_queue").Result()
			if err != nil {
				if err == redis.Nil {
					log.Println("No new events in the queue, continuing to wait...")
					continue
				}
				log.Printf("Error reading from queue: %v", err)
				p.publishErrorResponse(ctx, "", "Error reading from queue", "")
				continue
			}

			log.Println("Event read from queue, processing...")

			// Parse the event
			var event shared.EventModel
			if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
				log.Printf("Error parsing event: %v", err)
				p.publishErrorResponse(ctx, "", "Error parsing event", event.ChannelName)
				continue
			}

			// Look up the handler based on event type
			handler, exists := eventHandlers[event.EventType]
			if !exists {
				errMsg := fmt.Sprintf("No handler found for event type: %s", event.EventType)
				log.Println(errMsg)
				p.publishErrorResponse(ctx, event.UserId, errMsg, event.ChannelName)
				continue
			}

			log.Printf("Processing %s event for user: %s", event.EventType, event.UserId)

			// Call the event handler
			responseData, err := handler(ctx, event)
			if err != nil {
				errMsg := fmt.Sprintf("Error processing %s event: %v", event.EventType, err)
				log.Println(errMsg)
				p.publishErrorResponse(ctx, event.UserId, errMsg, event.ChannelName)
				continue
			}

			// Publish successful response
			response := shared.ResponseModel{
				Success: true,
				Data:    responseData,
			}
			p.publishResponse(ctx, event.UserId, response, event.ChannelName)
			log.Printf("Successfully processed %s event for user: %s", event.EventType, event.UserId)
		}
	}
}

func (p *LocalUserProcessor) publishErrorResponse(ctx context.Context, userId, errorMessage string, channelName string) {
	var responseChan string
	if channelName == "" {
		responseChan = fmt.Sprintf("user_response_%s", userId)
	} else {
		responseChan = channelName
	}
	response := shared.ResponseModel{
		Success: false,
		Error:   errorMessage,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling error response: %v", err)
		return
	}
	if err := p.RedisClient.Publish(ctx, responseChan, responseJson).Err(); err != nil {
		log.Printf("Error publishing error response: %v", err)
	}
}

func (p *LocalUserProcessor) publishResponse(ctx context.Context, userId string, response shared.ResponseModel, channelName string) {
	var responseChan string
	if channelName == "" {
		responseChan = fmt.Sprintf("user_response_%s", userId)
	} else {
		responseChan = channelName
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}
	if err := p.RedisClient.Publish(ctx, responseChan, responseJson).Err(); err != nil {
		log.Printf("Error publishing response: %v", err)
	}
}
