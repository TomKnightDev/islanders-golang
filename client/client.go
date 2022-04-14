package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"
	"log"
	"net/url"
	"reflect"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/tomknightdev/socketio-game-test/client/entities"
	"github.com/tomknightdev/socketio-game-test/client/gui"
	"github.com/tomknightdev/socketio-game-test/resources"
	"golang.org/x/image/math/f64"
)

var (
	//go:embed resources/characters.png
	characters      []byte
	CharactersImage *ebiten.Image
	//go:embed resources/environments.png
	environments      []byte
	EnvironmentsImage *ebiten.Image

	cam *camera.Camera
)

var addr string //= flag.String("addr", "localhost:8000", "http service address")

type Client struct {
	SendChan       chan *resources.Message
	RecvChan       chan *resources.Message
	NetworkPlayers map[uint16]*entities.NetworkPlayer
	Player         *entities.Player
}

var ChatWindow = &gui.Chat{}
var client = &Client{}

func init() {

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

	client.SendChan = make(chan *resources.Message)
	client.RecvChan = make(chan *resources.Message)

	addr = g.serverAddr

	u := url.URL{Scheme: "ws", Host: addr, Path: "/connect"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send connection request
	connectRequest := resources.NewConnectRequestMessage(resources.ConnectRequestContents{
		Username: g.username,
		Password: g.password,
	})

	if err = conn.WriteJSON(connectRequest); err != nil {
		return err
	}

	// Receive
	go func() {
		for {
			message := &resources.Message{}
			err := conn.ReadJSON(message)
			if err != nil {
				log.Println("Read error:", err)
			}

			switch message.MessageType {
			case resources.ConnectResponseMessage:
				handleConnectResponse(message, g)
				removeGui(&gui.MainMenu{}, g)
			case resources.FailedToConnectMessage:
				messageContents := message.Contents.(string)
				g.ConnectFailedMessage <- messageContents
				client.SendChan <- resources.NewFailedToConnectMessage(messageContents)
				log.Printf("failed to connect: %s", messageContents)
				return
			case resources.ChatMessage:
				receiveChatMessage(message)
			case resources.UpdateMessage:
				receiveUpdateMessage(message, g)
			case resources.ServerEntityUpdateMessage:
				receiveEntityUpdateMessage(message, g)
			}

		}
	}()

	// Send
	for {
		message := <-client.SendChan
		if message.MessageType == resources.FailedToConnectMessage {
			return fmt.Errorf("failed to connect: %v", message)
		}
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Failed to send message: %v - %v", message, err)
		}
	}
}

func handleConnectResponse(message *resources.Message, g *Game) {
	// If successful, the we receive our server client id and spawn position
	messageContents := message.Contents.(map[string]interface{})

	clientId := messageContents["clientId"].(float64)
	pos := messageContents["pos"].([]interface{})
	tile := messageContents["tile"].([]interface{})
	worldMap := resources.WorldMapWebSocketMessageConvert(messageContents["world"].(map[string]interface{}))

	client.Player = entities.NewPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
	client.Player.Username = g.username
	client.Player.Id = uint16(clientId)
	client.Player.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}

	go func(client *Client) {
		for {
			message := <-client.Player.SendChan
			client.SendChan <- resources.NewUpdateMessage(client.Player.Id, message)
		}
	}(client)

	g.Player = client.Player

	// Now logged in, build world
	world := entities.NewWorld(EnvironmentsImage, *worldMap)
	g.Environment = append(g.Environment, world)

	ChatWindow = gui.NewChat(g.screenWidth, g.screenHeight)
	g.Gui = append(g.Gui, ChatWindow)

	// Messages from chat send channel will be forwarded to the client send channel
	go func(client *Client, chat *gui.Chat) {
		for {
			message := <-chat.SendChan
			client.SendChan <- resources.NewChatMessage(client.Player.Id, message)
		}
	}(client, ChatWindow)
}

func receiveUpdateMessage(message *resources.Message, g *Game) {
	messageContents := message.Contents.(map[string]interface{})

	disconnected := messageContents["disconnected"].(bool)

	if disconnected {
		removeNetworkPlayer(message.ClientId, g)
		return
	}

	pos := messageContents["pos"].([]interface{})
	tile := messageContents["tile"].([]interface{})
	username := messageContents["username"].(string)

	networkClient, found := client.NetworkPlayers[message.ClientId]

	if found {
		networkClient.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
		networkClient.Tile = f64.Vec2{tile[0].(float64), tile[1].(float64)}
		return
	}

	networkClient = entities.NewNetworkPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
	networkClient.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
	networkClient.Username = username
	client.NetworkPlayers[message.ClientId] = networkClient
	g.Entities[message.ClientId] = networkClient
}

func receiveEntityUpdateMessage(message *resources.Message, g *Game) {
	contents := message.Contents.([]interface{})

	for _, content := range contents {
		c := content.(map[string]interface{})

		entityId := c["entityId"].(float64)
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

func receiveChatMessage(message *resources.Message) {
	messageContents := message.Contents.(string)

	ChatWindow.RecvMessages = append(ChatWindow.RecvMessages, messageContents)
}

func removeNetworkPlayer(clientId uint16, g *Game) {
	delete(g.Entities, clientId)
	delete(client.NetworkPlayers, clientId)
}

func removeGui(guiType Entity, g *Game) {
	gt := reflect.TypeOf(guiType)

	index := 0

	for i, g := range g.Gui {
		if gt == reflect.TypeOf(g) {
			index = i
		}
	}

	g.Gui = append(g.Gui[:index], g.Gui[index+1:]...)
}
