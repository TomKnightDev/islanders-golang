package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	messages "github.com/tomknightdev/socketio-game-test/messages"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

type client struct {
	id       uint16
	username string
}

var player = client{}

func main() {
	err := connectToServer()
	if err != nil {
		log.Fatalf("failed to connect to server: %s", err)
	}

	if err := gameLoop(); err != nil {
		log.Fatalf("error in game loop: %s", err)
	}
}

func connectToServer() error {
	fmt.Println("Client starting...")

	var username string
	fmt.Print("Enter your username:")
	fmt.Scanln(&username)
	player.username = username

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/connect"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// Send
	cm := messages.ConnectRequest{
		Username: player.username,
	}
	c.WriteJSON(cm)

	// Receive
	_, message, err := c.ReadMessage()
	if err != nil {
		return err
	}
	var cr = &messages.ConnectResponse{}
	err = json.Unmarshal([]byte(message), cr)
	if err != nil {
		return err
	}

	log.Printf("%d", cr.ClientId)
	player.id = cr.ClientId

	return nil
}

func gameLoop() error {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/game"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// Announce connection
	glm := &messages.GameLoopMessage{
		ClientId: player.id,
		Message:  "connected",
	}
	c.WriteJSON(glm)

	// Receive
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Printf("error in reading message: %s", err)
			}

			var glm = &messages.GameLoopMessage{}
			if err = json.Unmarshal([]byte(message), glm); err != nil {
				fmt.Printf("unmarshal error:", err, glm, message)
			}

			fmt.Println(glm.ClientId, glm.Message)
		}
	}()

	// Send
	for {
		var message string
		fmt.Scanln(&message)
		glm := &messages.GameLoopMessage{
			ClientId: player.id,
			Message:  message,
		}
		c.WriteJSON(glm)
	}
}
