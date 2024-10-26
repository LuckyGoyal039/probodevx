package wss

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	redis "github.com/probodevx/config"
)

var (
	eventRooms = make(map[string]map[*websocket.Conn]bool)
	mutex      sync.Mutex
)

func ConnectSocket(c *websocket.Conn) error {
	defer func() {
		removeConnection(c.Params("event"), c)
		c.Close()
	}()

	event := c.Params("event")
	addConnection(event, c)
	log.Printf("User connected to event room: %s", event)

	if err := c.WriteMessage(websocket.TextMessage, []byte("Subscribed to event: "+event)); err != nil {
		log.Println("Write error on subscription confirmation:", err)
		return err
	}
	// Redis client
	redisClient := redis.GetRedisClient()

	// Continuously listen for messages from the client and process them
	go func() {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				break
			}
			log.Printf("Received from %s: %s\n", event, string(msg))
		}
	}()

	// Continuously poll Redis and send messages to the WebSocket client
	for {
		log.Printf("Polling Redis for event: %s", event)
		msg, err := redisClient.RPop(context.TODO(), "orderbook:"+event).Result()
		if err != nil {
			if err.Error() == "redis: nil" {
				// No new messages in the Redis queue
				time.Sleep(500 * time.Millisecond) // Throttle polling
				continue
			}
			log.Println("Redis error:", err)
			break
		}

		// Send the Redis message back to the WebSocket client
		log.Printf("Sending to WebSocket client for event %s: %s", event, msg)
		err = c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}

	return nil
}

func removeConnection(event string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	if connections, exists := eventRooms[event]; exists {
		delete(connections, conn)
		if len(connections) == 0 {
			delete(eventRooms, event) // Clean up if no connections are left
		}
	}
}

func addConnection(event string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := eventRooms[event]; !exists {
		eventRooms[event] = make(map[*websocket.Conn]bool)
	}
	eventRooms[event][conn] = true
}

func handleBroadcast(event string, message []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	if connections, exists := eventRooms[event]; exists {
		for conn := range connections {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Write error:", err)
				conn.Close()
				delete(connections, conn)
			}
		}
	}
}

func BroadCastMessage(c *fiber.Ctx) error {
	event := c.Params("event")
	message := c.Body()

	// Broadcast the message to all connections in the event room
	handleBroadcast(event, message)
	return c.SendString("Message broadcasted to " + event + " room")
}
