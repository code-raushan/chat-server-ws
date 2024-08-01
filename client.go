package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := "ws://localhost:8000/ws"
	log.Printf("connecting to %s", u)

	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// Goroutine to read messages from the server
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("Received: %s", message)
		}
	}()

	// Goroutine to read input from the console and send messages
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			message := scanner.Text()
			err := c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}()

	fmt.Println("Type your messages and press Enter to send. Press Ctrl+C to exit.")

	// Wait for interrupt signal to gracefully close the connection
	<-interrupt
	log.Println("interrupt received, closing connection")

	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
	<-done
}
