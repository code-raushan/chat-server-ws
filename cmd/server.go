package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

const (
	ConnectionCountKey           = "chat:connection-count"
	ConnectionCountUpdateChannel = "chat:connection-count-updated"
	NewMessageChannel            = "chat:new-message"
)

type Message struct {
	ChannelName string `json:"channelName"`
	Message     string `json:"message"`
	ID          string `json:"id"`
	// CreatedAt   string    `json:"createdAt"`
	Port string `json:"port"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redisURI := os.Getenv("REDIS_URI")

	if redisURI == "" {
		log.Fatal("Redis URI cannot be empty")
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: redisURI,
	})

	defer redisClient.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", HandleHealthCheck)
	http.HandleFunc("/ws", HandleWebsockets)

	log.Fatal(http.ListenAndServe(":"+port, nil))

}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"port":   os.Getenv("PORT"),
	}
	json.NewEncoder(w).Encode(response)
}

func HandleWebsockets(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Failed to upgrade HTTP connection to WS")
		return
	}

	ctx := context.Background()

	// Incrementing connection count

	count, err := redisClient.Incr(ctx, ConnectionCountKey).Result()
	if err != nil {
		log.Println("error publishing connection count: ", err)
	}

	// Publishing count to ConnnectionCountChannel

	err = redisClient.Publish(ctx, ConnectionCountUpdateChannel, count).Err()

	if err != nil {
		log.Println("Error publishing connection count update:", err)
	}

	// subscribing to the new message channel

	pubsub := redisClient.Subscribe(ctx, NewMessageChannel)
	defer pubsub.Close()

	// handling incoming messages
	go func() {
		for {
			_, mesg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				break
			}

			// publish new msg to the channel
			err = redisClient.Publish(ctx, NewMessageChannel, string(mesg)).Err()
			if err != nil {
				log.Println("error publishing msg to channel", err)
			}
		}
	}()

	for msg := range pubsub.Channel() {
		message := Message{
			ChannelName: NewMessageChannel,
			Message:     msg.Payload,
			ID:          uuid.New().String(),
			Port:        r.URL.Port(),
		}

		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("Error writing message to WebSocket:", err)
			break
		}
	}

	count, err = redisClient.Decr(ctx, ConnectionCountKey).Result()
	if err != nil {
		log.Println("Error decrementing connection count:", err)
	}

	// Publish connection count update
	err = redisClient.Publish(ctx, ConnectionCountUpdateChannel, count).Err()
	if err != nil {
		log.Println("Error publishing connection count update:", err)
	}

}
