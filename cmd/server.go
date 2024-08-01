package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var(
	redisClient *redis.Client
	upgrader = websocket.Upgrader{
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

func main(){
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

	log.Fatal(http.ListenAndServe(":"+port, nil))

}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"port": os.Getenv("PORT"),
	}
	json.NewEncoder(w).Encode(response)
}