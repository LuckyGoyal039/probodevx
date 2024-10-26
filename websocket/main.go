package main

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var (
	// A map of event rooms with associated connections
	eventRooms = make(map[string]map[*websocket.Conn]bool)
	mutex      sync.Mutex // To protect concurrent access to eventRooms
)

func main() {
	app := fiber.New()

	// Middleware to check if the request is a WebSocket request
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket handler for subscribing to an event room
	app.Get("/ws/:event", websocket.New(func(c *websocket.Conn) {
		defer func() {
			// Remove connection from the event room on disconnect
			removeConnection(c.Params("event"), c)
			c.Close()
		}()

		event := c.Params("event")

		// Add the new connection to the event room
		addConnection(event, c)
		log.Printf("User connected to event room: %s\n", event)

		// Listen for messages from the client
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				break
			}
			log.Printf("Received from %s: %s\n", event, string(msg))
		}
	}))

	// Endpoint to broadcast a message to all users in an event room
	app.Post("/broadcast/:event", func(c *fiber.Ctx) error {
		event := c.Params("event")
		message := c.Body()

		// Broadcast the message to all connections in the event room
		broadcastMessage(event, message)
		return c.SendString("Message broadcasted to " + event + " room")
	})

	log.Fatal(app.Listen(":8080"))
}

// Add a WebSocket connection to an event room
func addConnection(event string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := eventRooms[event]; !exists {
		eventRooms[event] = make(map[*websocket.Conn]bool)
	}
	eventRooms[event][conn] = true
}

// Remove a WebSocket connection from an event room
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

// Broadcast a message to all WebSocket connections in an event room
func broadcastMessage(event string, message []byte) {
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
