package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"
	"log"
	"math"
	"net/url"
	"reflect"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/tomknightdev/islanders-golang/client/entities"
	"github.com/tomknightdev/islanders-golang/client/gui"
	"github.com/tomknightdev/islanders-golang/resources"
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
	NetworkPlayers map[uuid.UUID]*entities.NetworkPlayer
	Player         *entities.Player
}

var ChatWindow = &gui.Chat{}
var InfoWindow = &gui.Info{}
var ClientInstance = &Client{}

func init() {

	ClientInstance.NetworkPlayers = make(map[uuid.UUID]*entities.NetworkPlayer)

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

	ClientInstance.SendChan = make(chan *resources.Message)
	ClientInstance.RecvChan = make(chan *resources.Message)

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
				ClientInstance.SendChan <- resources.NewFailedToConnectMessage(messageContents)
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
		message := <-ClientInstance.SendChan
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

	clientId := messageContents["clientId"].(uuid.UUID)
	pos := messageContents["pos"].([]interface{})
	tile := messageContents["tile"].([]interface{})
	worldMap := resources.WorldMapWebSocketMessageConvert(messageContents["world"].(map[string]interface{}))

	// Create camera
	cam = camera.NewCamera(g.screenWidth, g.screenHeight, pos[0].(float64), pos[1].(float64), 0, 1)

	ClientInstance.Player = entities.NewPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
	ClientInstance.Player.Username = g.username
	ClientInstance.Player.Id = clientId
	ClientInstance.Player.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
	ClientInstance.Player.Cam = cam

	go func(client *Client) {
		for {
			message := <-client.Player.SendChan
			client.SendChan <- resources.NewUpdateMessage(client.Player.Id, message)
		}
	}(ClientInstance)

	g.Player = ClientInstance.Player

	// Now logged in, build world
	world := entities.NewWorld(EnvironmentsImage, *worldMap)
	world.Cam = cam
	g.Environment = append(g.Environment, world)

	// Add gui
	ChatWindow = gui.NewChat(g.screenWidth, g.screenHeight, g.renderMgr)
	g.Gui = append(g.Gui, ChatWindow)

	InfoWindow = gui.NewInfo(g.screenWidth, g.screenHeight, ClientInstance.Player, g.renderMgr)
	g.Gui = append(g.Gui, InfoWindow)

	// Messages from chat send channel will be forwarded to the client send channel
	go func(client *Client, chat *gui.Chat) {
		for {
			message := <-chat.SendChan
			client.SendChan <- resources.NewChatMessage(client.Player.Id, message)
		}
	}(ClientInstance, ChatWindow)

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

	networkClient, found := ClientInstance.NetworkPlayers[message.ClientId]

	if found {
		networkClient.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
		networkClient.Tile = f64.Vec2{tile[0].(float64), tile[1].(float64)}
		return
	}

	networkClient = entities.NewNetworkPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
	networkClient.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
	networkClient.Username = username
	networkClient.Cam = cam
	ClientInstance.NetworkPlayers[message.ClientId] = networkClient
	g.Entities[message.ClientId] = networkClient
}

func receiveEntityUpdateMessage(message *resources.Message, g *Game) {
	contents := message.Contents.([]interface{})

	for _, content := range contents {
		c := content.(map[string]interface{})

		entityId := c["entityId"].(uuid.UUID)
		pos := c["pos"].([]interface{})
		tile := c["tile"].([]interface{})

		np, found := ClientInstance.NetworkPlayers[entityId]

		if found {
			np.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
			np.Tile = f64.Vec2{tile[0].(float64), tile[1].(float64)}
			continue
		}

		np = entities.NewNetworkPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
		np.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
		ClientInstance.NetworkPlayers[entityId] = np
		g.Entities[entityId] = np
	}
}

func receiveChatMessage(message *resources.Message) {
	messageContents := message.Contents.(string)

	ChatWindow.RecvMessages = append(ChatWindow.RecvMessages, messageContents)
}

func removeNetworkPlayer(clientId uuid.UUID, g *Game) {
	delete(g.Entities, clientId)
	delete(ClientInstance.NetworkPlayers, clientId)
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

func (c *Client) GetClosestEnemyPos(fromPos f64.Vec2) f64.Vec2 {
	type e struct {
		ent      *entities.NetworkPlayer
		distance float64
	}

	closest := e{}

	for _, ent := range ClientInstance.NetworkPlayers {
		// Get distance
		dist := GetDistance(fromPos, ent.Position)

		if closest.ent.Username == "" {
			closest = e{
				ent,
				dist,
			}
		} else if dist < closest.distance {
			closest.ent = ent
			closest.distance = dist
		}

	}

	return closest.ent.Position
}

func GetDistance(a f64.Vec2, b f64.Vec2) float64 {
	xa := math.Abs(a[0] - b[0])
	xs := math.Pow(xa, 2)

	ya := math.Abs(a[1] - b[1])
	ys := math.Pow(ya, 2)
	return math.Sqrt(xs + ys)
}
