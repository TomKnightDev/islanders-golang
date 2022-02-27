package main

import (
	"bytes"
	_ "embed"
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
	SendChan       chan *messages.Message
	RecvChan       chan *messages.Message
	NetworkPlayers map[uint16]*entities.NetworkPlayer
	Player         *entities.Player
}

var ChatWindow = &gui.Chat{}
var client = &Client{}

func init() {
	client.SendChan = make(chan *messages.Message)
	client.RecvChan = make(chan *messages.Message)
	client.NetworkPlayers = make(map[uint16]*entities.NetworkPlayer)

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

	addr = g.serverAddr

	u := url.URL{Scheme: "ws", Host: addr, Path: "/connect"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send connection request
	connectRequest := messages.NewConnectRequestMessage(messages.ConnectRequestContents{
		Username: g.username,
		Password: g.password,
	})

	if err = conn.WriteJSON(connectRequest); err != nil {
		return err
	}

	// Receive
	go func() {
		for {
			message := &messages.Message{}
			err := conn.ReadJSON(message)
			if err != nil {
				log.Println("Read error:", err)
				break
			}

			switch message.MessageType {
			case messages.ConnectResponseMessage:
				handleConnectResponse(message, g)
			case messages.FailedToConnectMessage:
				err := message.Contents.(error)
				log.Printf("Failed to connect to server: %v", err)
			case messages.ChatMessage:
				receiveChatMessage(message)
			case messages.UpdateMessage:
				receiveUpdateMessage(message, g)
			case messages.ServerEntityUpdateMessage:
				receiveEntityUpdateMessage(message, g)
			}
		}
	}()

	// Send
	for {
		message := <-client.SendChan
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Failed to send message: %v - %v", message, err)
		}
	}
}

func handleConnectResponse(message *messages.Message, g *Game) {
	// If successful, the we receive our server client id and spawn position
	messageContents := message.Contents.(map[string]interface{})

	clientId := messageContents["clientId"].(float64)
	pos := messageContents["pos"].([]interface{})
	tile := messageContents["tile"].([]interface{})

	client.Player = entities.NewPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
	client.Player.Username = g.username
	client.Player.Id = uint16(clientId)
	client.Player.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}

	go func(client *Client) {
		for {
			message := <-client.Player.SendChan
			client.SendChan <- messages.NewUpdateMessage(client.Player.Id, message)
		}
	}(client)

	g.Player = client.Player

	// Now logged in, build world
	world := entities.NewWorld(EnvironmentsImage)
	g.Environment = append(g.Environment, world)

	ChatWindow = gui.NewChat(g.screenWidth, g.screenHeight)
	g.Gui = append(g.Gui, ChatWindow)

	// Messages from chat send channel will be forwarded to the client send channel
	go func(client *Client, chat *gui.Chat) {
		for {
			message := <-chat.SendChan
			client.SendChan <- messages.NewChatMessage(client.Player.Id, message)
		}
	}(client, ChatWindow)
}

func receiveUpdateMessage(message *messages.Message, g *Game) {
	messageContents := message.Contents.(map[string]interface{})

	disconnected := messageContents["disconnected"].(bool)

	if disconnected {
		removeNetworkPlayer(message.ClientId, g)
		return
	}

	pos := messageContents["pos"].([]interface{})
	tile := messageContents["tile"].([]interface{})

	newtworkClient, found := client.NetworkPlayers[message.ClientId]

	if found {
		newtworkClient.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
		newtworkClient.Tile = f64.Vec2{tile[0].(float64), tile[1].(float64)}
		return
	}

	newtworkClient = entities.NewNetworkPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
	newtworkClient.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
	client.NetworkPlayers[message.ClientId] = newtworkClient
	g.Entities[message.ClientId] = newtworkClient
}

func receiveEntityUpdateMessage(message *messages.Message, g *Game) {
	contents := message.Contents.([]interface{})

	for _, content := range contents {
		c := content.(map[string]interface{})

		entityId := c["entityId"].(interface{}).(float64)
		pos := c["pos"].([]interface{})
		tile := c["tile"].([]interface{})

		np, found := client.NetworkPlayers[uint16(entityId)]

		if found {
			np.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
			np.Tile = f64.Vec2{tile[0].(float64), tile[1].(float64)}
			continue
		}

		np = entities.NewNetworkPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
		np.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
		client.NetworkPlayers[uint16(entityId)] = np
		g.Entities[uint16(entityId)] = np
	}
}

func receiveChatMessage(message *messages.Message) {
	messageContents := message.Contents.(string)

	ChatWindow.RecvMessages = append(ChatWindow.RecvMessages, messageContents)
}

func removeNetworkPlayer(clientId uint16, g *Game) {
	delete(g.Entities, clientId)
	delete(client.NetworkPlayers, clientId)
}
