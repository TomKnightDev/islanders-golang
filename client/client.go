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
	NetworkPlayers map[uint16]*entities.NetworkPlayer
	Player         *entities.Player
}

var ChatWindow = &gui.Chat{}
var InfoWindow = &gui.Info{}
var InventoryWindow = &gui.Inventory{}
var ActionBar = gui.NewActionBar()
var ActionsWindow = &gui.Actions{}
var SkillsWindow = &gui.Skills{}
var EquipmentWindow = &gui.Equipment{}
var world *entities.World
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
	client.RecvChan = make(chan *resources.Message, 1024)

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

	// Receive. Messages are queued for the game loop (see handleMessages) so
	// game state is only ever touched from Game.Update.
	go func() {
		for {
			message := &resources.Message{}
			err := conn.ReadJSON(message)
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			if message.MessageType == resources.FailedToConnectMessage {
				messageContents := message.Contents.(string)
				g.ConnectFailedMessage <- messageContents
				client.SendChan <- resources.NewFailedToConnectMessage(messageContents)
				log.Printf("failed to connect: %s", messageContents)
				return
			}
			client.RecvChan <- message
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

// handleMessages applies every queued server message. Called from Game.Update.
func handleMessages(g *Game) {
	for {
		select {
		case message := <-client.RecvChan:
			switch message.MessageType {
			case resources.ChooseClassMessage:
				showCharacterCreation(g)
			case resources.ConnectResponseMessage:
				handleConnectResponse(message, g)
				removeGui(&gui.MainMenu{}, g)
				removeGui(&gui.CharacterCreation{}, g)
			case resources.ChatMessage:
				receiveChatMessage(message)
			case resources.UpdateMessage:
				receiveUpdateMessage(message, g)
			case resources.ServerEntityUpdateMessage:
				receiveEntityUpdateMessage(message, g)
			case resources.InventoryMessage:
				receiveInventoryMessage(message)
			case resources.ObjectUpdateMessage:
				receiveObjectUpdateMessage(message)
			case resources.PathMessage:
				var path resources.PathContents
				if err := resources.DecodeContents(message.Contents, &path); err == nil && client.Player != nil {
					client.Player.SetPath(path.Tiles)
				}
			}
		default:
			return
		}
	}
}

func handleConnectResponse(message *resources.Message, g *Game) {
	// If successful, the we receive our server client id and spawn position
	messageContents := message.Contents.(map[string]interface{})

	clientId := messageContents["clientId"].(float64)
	pos := messageContents["pos"].([]interface{})
	tile := messageContents["tile"].([]interface{})
	worldMap := &resources.WorldMap{}
	if err := resources.DecodeContents(messageContents["world"], worldMap); err != nil {
		log.Printf("Failed to decode world map: %v", err)
		return
	}

	// Create camera
	cam = camera.NewCamera(g.screenWidth, g.screenHeight, pos[0].(float64), pos[1].(float64), 0, 1)
	cam.SetZoom(entities.StartZoom) // SetZoom also resizes the camera's image to match

	client.Player = entities.NewPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
	client.Player.Username = g.username
	client.Player.Id = uint16(clientId)
	client.Player.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
	client.Player.Cam = cam
	client.Player.WorldMap = worldMap

	go func(client *Client) {
		for {
			move := <-client.Player.SendChan
			client.SendChan <- resources.NewMoveRequestMessage(client.Player.Id, move)
		}
	}(client)

	g.Player = client.Player

	// Now logged in, build world
	world = entities.NewWorld(EnvironmentsImage, *worldMap)
	world.Cam = cam
	g.Environment = append(g.Environment, world)

	// Add gui. The action bar goes first so the imgui windows draw over it.
	g.Gui = append(g.Gui, ActionBar)

	ChatWindow = gui.NewChat(g.screenWidth, g.screenHeight, g.renderMgr)
	g.Gui = append(g.Gui, ChatWindow)

	InfoWindow = gui.NewInfo(g.screenWidth, g.screenHeight, client.Player, g.renderMgr)
	InfoWindow.SendChat = func(text string) {
		// Off the game loop: the send channel is unbuffered.
		go func() { client.SendChan <- resources.NewChatMessage(client.Player.Id, text) }()
	}
	g.Gui = append(g.Gui, InfoWindow)

	InventoryWindow = gui.NewInventory(g.renderMgr)
	g.Gui = append(g.Gui, InventoryWindow)

	EquipmentWindow = gui.NewEquipment(g.renderMgr, func(c resources.EquipCommandContents) {
		go func() { client.SendChan <- resources.NewEquipCommandMessage(c) }()
	})
	g.Gui = append(g.Gui, EquipmentWindow)
	g.Gui = append(g.Gui, gui.NewCameraControls(client.Player))

	// Action commands go through one queue so they reach the server in the
	// order the player made them.
	actionOut := make(chan resources.ActionCommandContents, 32)
	go func() {
		for c := range actionOut {
			client.SendChan <- resources.NewActionCommandMessage(c)
		}
	}()
	ActionsWindow = gui.NewActions(g.renderMgr, client.Player, worldMap.Places)
	ActionsWindow.Send = func(c resources.ActionCommandContents) {
		select {
		case actionOut <- c:
		default: // backed up; drop rather than freeze the game
		}
	}
	g.Gui = append(g.Gui, ActionsWindow)

	SkillsWindow = gui.NewSkills(g.renderMgr)
	g.Gui = append(g.Gui, SkillsWindow)

	g.Gui = append(g.Gui, gui.NewHover(worldMap, func() (int, int) {
		wx, wy := cam.GetCursorCoords()
		return int(math.Floor(wx / resources.TileSize)), int(math.Floor(wy / resources.TileSize))
	}))

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

	health, _ := messageContents["health"].(float64)
	maxHealth, _ := messageContents["maxHealth"].(float64)

	// The server owns movement, so our own position arrives like everyone else's.
	if client.Player != nil && message.ClientId == client.Player.Id {
		client.Player.SetPosition(f64.Vec2{pos[0].(float64), pos[1].(float64)})
		client.Player.Health, client.Player.MaxHealth = int(health), int(maxHealth)
		return
	}

	networkClient, found := client.NetworkPlayers[message.ClientId]

	if found {
		networkClient.SetPosition(f64.Vec2{pos[0].(float64), pos[1].(float64)})
		networkClient.Tile = f64.Vec2{tile[0].(float64), tile[1].(float64)}
		networkClient.Health, networkClient.MaxHealth = int(health), int(maxHealth)
		return
	}

	networkClient = entities.NewNetworkPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
	networkClient.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
	networkClient.Username = username
	networkClient.Cam = cam
	networkClient.Health, networkClient.MaxHealth = int(health), int(maxHealth)
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
		username, _ := c["username"].(string)
		health, _ := c["health"].(float64)
		maxHealth, _ := c["maxHealth"].(float64)
		if gone, _ := c["gone"].(bool); gone {
			removeNetworkPlayer(uint16(entityId), g) // e.g. a defeated enemy; it comes back when it respawns
			continue
		}

		np, found := client.NetworkPlayers[uint16(entityId)]

		if found {
			np.SetPosition(f64.Vec2{pos[0].(float64), pos[1].(float64)})
			np.Tile = f64.Vec2{tile[0].(float64), tile[1].(float64)}
			np.Health, np.MaxHealth = int(health), int(maxHealth)
			continue
		}

		np = entities.NewNetworkPlayer(CharactersImage, f64.Vec2{tile[0].(float64), tile[1].(float64)})
		np.Position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
		np.Username = username
		np.Cam = cam
		np.Health, np.MaxHealth = int(health), int(maxHealth)
		g.Entities[uint16(entityId)] = np
		client.NetworkPlayers[uint16(entityId)] = np
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

// removeGui removes the window of guiType's type, if it's open.
func removeGui(guiType Entity, g *Game) {
	gt := reflect.TypeOf(guiType)
	for i, e := range g.Gui {
		if gt == reflect.TypeOf(e) {
			g.Gui = append(g.Gui[:i], g.Gui[i+1:]...)
			return
		}
	}
}

// showCharacterCreation swaps the login menu for the class picker, for a
// name the server hasn't seen before.
func showCharacterCreation(g *Game) {
	for _, e := range g.Gui {
		switch w := e.(type) {
		case *gui.CharacterCreation:
			return // already open
		case *gui.MainMenu:
			w.Connected = true // hides the login window
		}
	}
	g.Gui = append(g.Gui, gui.NewCharacterCreation(g.renderMgr, CharactersImage, func(c resources.ClassID) {
		go func() {
			client.SendChan <- resources.NewCreateCharacterMessage(resources.CreateCharacterContents{Class: c})
		}()
	}))
}

func receiveInventoryMessage(message *resources.Message) {
	var contents resources.InventoryContents
	if err := resources.DecodeContents(message.Contents, &contents); err != nil {
		log.Printf("Bad inventory message: %v", err)
		return
	}
	InventoryWindow.Contents = contents
	ActionBar.Set(contents.Activity, contents.ActionSeconds, contents.ActionElapsed)
	ActionsWindow.SetState(contents)
	SkillsWindow.SetState(contents.Skills)
	EquipmentWindow.SetState(contents)
	if client.Player != nil {
		client.Player.Health, client.Player.MaxHealth = contents.Health, contents.MaxHealth
	}
}

func receiveObjectUpdateMessage(message *resources.Message) {
	var u resources.ObjectUpdateContents
	if err := resources.DecodeContents(message.Contents, &u); err != nil {
		log.Printf("Bad object update: %v", err)
		return
	}
	if world != nil {
		world.SetTile(u.X, u.Y, resources.LayerObject, u.Tile)
	}
}
