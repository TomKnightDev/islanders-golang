package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tomknightdev/socketio-game-test/client/gui"
	"github.com/tomknightdev/socketio-game-test/messages"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

type client struct {
	id       uint16
	username string
}

var player = client{}

func (g *Game) Update() error {
	for _, e := range g.entities {
		e.Update()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, e := range g.entities {
		e.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 800, 600
}

func connectToServer(g *Game) error {
	fmt.Println("Client starting...")
	player.username = g.playerName

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

func gameLoop(g *Game) error {
	chat := gui.NewChat()
	g.entities = append(g.entities, chat)

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

			chat.RecvMessages = append(chat.RecvMessages, fmt.Sprint(glm.ClientId, glm.Message))
			fmt.Println(glm.ClientId, glm.Message)
		}
	}()

	// Send
	for {
		msg := <-chat.SendChan
		glm := &messages.GameLoopMessage{
			ClientId: player.id,
			Message:  msg,
		}
		c.WriteJSON(glm)
	}
}
