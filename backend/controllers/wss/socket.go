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

// Room represents a group of connections subscribed to the same event
type Room struct {
	connections map[*websocket.Conn]bool
	mutex       sync.RWMutex
}

// RoomManager manages all active rooms
type RoomManager struct {
	rooms map[string]*Room
	mutex sync.RWMutex
}

// Create a global room manager
var manager = &RoomManager{
	rooms: make(map[string]*Room),
}

// Join adds a connection to a room
func (rm *RoomManager) Join(event string, conn *websocket.Conn) {
	rm.mutex.Lock()
	if rm.rooms[event] == nil {
		rm.rooms[event] = &Room{
			connections: make(map[*websocket.Conn]bool),
		}
	}
	rm.mutex.Unlock()

	room := rm.rooms[event]
	room.mutex.Lock()
	room.connections[conn] = true
	room.mutex.Unlock()

	log.Printf("Client joined room %s. Total clients in room: %d", event, len(room.connections))
}

// Leave removes a connection from a room
func (rm *RoomManager) Leave(event string, conn *websocket.Conn) {
	rm.mutex.RLock()
	room := rm.rooms[event]
	rm.mutex.RUnlock()

	if room != nil {
		room.mutex.Lock()
		delete(room.connections, conn)
		clientCount := len(room.connections)

		// If room is empty, remove it
		if clientCount == 0 {
			rm.mutex.Lock()
			delete(rm.rooms, event)
			rm.mutex.Unlock()
		}
		room.mutex.Unlock()

		log.Printf("Client left room %s. Remaining clients in room: %d", event, clientCount)
	}
}

// Broadcast sends a message to all connections in a room
func (rm *RoomManager) Broadcast(event string, message []byte) {
	rm.mutex.RLock()
	room := rm.rooms[event]
	rm.mutex.RUnlock()

	if room != nil {
		room.mutex.RLock()
		for conn := range room.connections {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error broadcasting to connection: %v", err)
				// Handle disconnected clients in a separate goroutine to avoid deadlock
				go func(c *websocket.Conn) {
					rm.Leave(event, c)
					c.Close()
				}(conn)
			}
		}
		room.mutex.RUnlock()
	}
}

// WebSocketHandler handles incoming WebSocket connections
func WebSocketHandler(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	return websocket.New(func(c *websocket.Conn) {
		// Get event parameter from the URL
		event := c.Params("event")
		if event == "" {
			log.Println("No event specified")
			return
		}

		// Join the room
		manager.Join(event, c)

		defer func() {
			manager.Leave(event, c)
			c.Close()
		}()

		log.Printf("User connected to event room: %s", event)

		// Send confirmation message
		if err := c.WriteMessage(websocket.TextMessage, []byte("Subscribed to event: "+event)); err != nil {
			log.Println("Write error on subscription confirmation:", err)
			return
		}

		// Redis client
		redisClient := redis.GetRedisClient()

		// Handle incoming messages from the client
		go func() {
			for {
				messageType, msg, err := c.ReadMessage()
				if err != nil {
					log.Printf("Read error: %v", err)
					return
				}

				if messageType == websocket.TextMessage {
					log.Printf("Received from %s: %s\n", event, string(msg))
					// Optionally broadcast the message to all clients in the room
					manager.Broadcast(event, msg)
				}
			}
		}()

		// Poll Redis and broadcast messages to all clients in the room
		for {
			log.Printf("Polling Redis for event: %s", event)
			msg, err := redisClient.RPop(context.TODO(), "orderbook:"+event).Result()
			if err != nil {
				if err.Error() == "redis: nil" {
					time.Sleep(500 * time.Millisecond) // Throttle polling
					continue
				}
				log.Println("Redis error:", err)
				break
			}

			// Broadcast the Redis message to all clients in the room
			log.Printf("Broadcasting to room %s: %s", event, msg)
			manager.Broadcast(event, []byte(msg))
		}
	})(c)
}
