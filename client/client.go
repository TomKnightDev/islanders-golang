package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tomknightdev/socketio-game-test/client/entities"
	"github.com/tomknightdev/socketio-game-test/client/gui"
	"github.com/tomknightdev/socketio-game-test/messages"
	"golang.org/x/image/math/f64"
)

var (
	//go:embed resources/characters.png
	characters      []byte
	CharactersImage *ebiten.Image
	//go:embed resources/environments.png
	environments      []byte
	EnvironmentsImage *ebiten.Image
)

var addr string //= flag.String("addr", "localhost:8000", "http service address")

type Client struct {
	SendChan       chan f64.Vec2
	RecvChan       chan string
	NetworkPlayers []*entities.NetworkPlayer
	Player         *entities.Player
}

var client = Client{}

func init() {
	client.SendChan = make(chan f64.Vec2)
	client.RecvChan = make(chan string)

	img, err := png.Decode(bytes.NewReader(characters))
	if err != nil {
		log.Fatal(err)
	}
	CharactersImage = ebiten.NewImageFromImage(img)

	img, err = png.Decode(bytes.NewReader(environments))
	if err != nil {
		log.Fatal(err)
	}
	EnvironmentsImage = ebiten.NewImageFromImage(img)
}

func connectToServer(g *Game) error {
	fmt.Println("Client starting...")

	client.Player = entities.NewPlayer(CharactersImage)
	g.Player = client.Player

	client.Player.Username = g.playerName

	addr = g.serverAddr

	u := url.URL{Scheme: "ws", Host: addr, Path: "/connect"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// Send
	cm := messages.ConnectRequest{
		Username: client.Player.Username,
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
	client.Player.Id = cr.ClientId

	return nil
}

func gameLoop(g *Game) error {
	world := entities.NewWorld(EnvironmentsImage)
	g.Environment = append(g.Entities, world)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/game"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// Announce connection
	em := &messages.EntityMessage{
		EntityId:  client.Player.Id,
		EntityPos: f64.Vec2{-1, 0},
	}
	glm := &messages.GameLoopMessage{
		EntityMessages: []messages.EntityMessage{
			*em,
		},
	}
	c.WriteJSON(glm)

	// Receive
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				} else {
					fmt.Printf("error in reading message: %v %s", err, message)
				}
				break
			}

			var glm = &messages.GameLoopMessage{}
			if err = json.Unmarshal([]byte(message), glm); err != nil {
				fmt.Printf("unmarshal error:", err, glm, message)
			}

			for _, em := range glm.EntityMessages {
				// Update information about other players
				e := func(m messages.EntityMessage) *entities.NetworkPlayer {
					for _, e := range client.NetworkPlayers {
						if e.Id == m.EntityId {
							return e
						}
					}
					return nil
				}(em)

				// If doesn't exist, create it
				if e == nil {
					e = entities.NewNetworkPlayer(CharactersImage, em.EntityTile)
					e.Id = em.EntityId
					client.NetworkPlayers = append(client.NetworkPlayers, e)
					g.Entities = append(g.Entities, e)
				}

				e.Position = em.EntityPos
			}
		}
	}()

	// Send
	for {
		pos := <-client.Player.SendChan
		em := &messages.EntityMessage{
			EntityId:   client.Player.Id,
			EntityPos:  pos,
			EntityTile: f64.Vec2{0, 0},
		}
		glm := &messages.GameLoopMessage{
			EntityMessages: []messages.EntityMessage{
				*em,
			},
		}

		c.WriteJSON(glm)
	}
}

func chatLoop(g *Game) error {
	chat := gui.NewChat(screenWidth, screenHeight)
	g.Entities = append(g.Entities, chat)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/chat"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// Announce connection
	glm := &messages.ChatLoopMessage{
		ClientId: client.Player.Id,
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

			var chatMessage = &messages.ChatLoopMessage{}
			if err = json.Unmarshal([]byte(message), chatMessage); err != nil {
				fmt.Printf("unmarshal error:", err, chatMessage, message)
			}

			chat.RecvMessages = append(chat.RecvMessages, fmt.Sprint(chatMessage.ClientId, chatMessage.Message))
			fmt.Println(chatMessage.ClientId, chatMessage.Message)
		}
	}()

	// Send
	for {
		msg := <-chat.SendChan
		glm := &messages.ChatLoopMessage{
			ClientId: client.Player.Id,
			Message:  msg,
		}
		c.WriteJSON(glm)
	}
}
