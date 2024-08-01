package main

import (
	"fmt"
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
}