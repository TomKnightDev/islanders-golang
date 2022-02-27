package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomknightdev/socketio-game-test/messages"
	"golang.org/x/image/math/f64"
)

var upgrader = websocket.Upgrader{} // use default options

var ServerInstance = &Server{}

type Server struct {
	clientsById       map[uint16]*client
	clientsByUsername map[string]*client
	enemies           []*Entity
}
type client struct {
	id       uint16
	username string
	password string
	position f64.Vec2
	tile     f64.Vec2
	mu       sync.Mutex
	conn     *websocket.Conn
}

func init() {
	ServerInstance.clientsById = make(map[uint16]*client)
	ServerInstance.clientsByUsername = make(map[string]*client)

	go serverLoop()
}

func connect(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Upgrade error:", err)
		return
	}
	defer conn.Close()

	message := &messages.Message{}
	var clientId uint16

	for {
		err := conn.ReadJSON(message)
		if err != nil {
			log.Println("Connect read error:", err)
			disconnectClient(clientId)
			break
		}

		// log.Printf("%s message recieved: %v", message.MessageType, message)

		switch message.MessageType {
		case messages.ConnectRequestMessage:
			clientId, err = connectClient(message, conn)
			if err != nil {
				conn.WriteJSON(messages.NewFailedToConnectMessage(err))
			}
		case messages.ChatMessage:
			handleChatMessage(message)
		case messages.UpdateMessage:
			handleUpdateMessage(message)
		default:
			log.Printf("Message type: %s not handled", message.MessageType)
		}
	}
}

func connectClient(message *messages.Message, conn *websocket.Conn) (uint16, error) {
	messageContents := message.Contents.(map[string]interface{})

	fmt.Printf("%v", messageContents)

	username := messageContents["username"].(string)
	password := messageContents["password"].(string)

	// Check to see if this c has connect previously
	c, found := ServerInstance.clientsByUsername[username]

	// If the client was found, check password
	if found {
		c.mu.Lock()
		if c.password != password {
			return 0, fmt.Errorf("incorrect password")
		}
		c.conn = conn

		// Send reponse to client
		conn.WriteJSON(messages.NewConnectResponseMessage(messages.ConnectResponseContents{
			ClientId: c.id,
			Pos:      c.position,
		}))
		c.mu.Unlock()

		// Update all other clients
		handleUpdateMessage(messages.NewUpdateMessage(c.id, messages.UpdateContents{
			Pos:  c.position,
			Tile: c.tile,
		}))
		return c.id, nil
	}

	// Client wasn't found so create it
	newClient := &client{
		id:       uint16(len(ServerInstance.clientsById)),
		username: username,
		password: password,
		conn:     conn,
	}

	// Add the client to server maps
	ServerInstance.clientsById[newClient.id] = newClient
	ServerInstance.clientsByUsername[newClient.username] = newClient

	// Send reponse to client
	conn.WriteJSON(messages.NewConnectResponseMessage(messages.ConnectResponseContents{
		ClientId: newClient.id,
		Pos:      f64.Vec2{1, 1},
		Tile:     f64.Vec2{0, 0},
	}))

	// Update all other clients
	sendUpdateToClients(messages.NewUpdateMessage(newClient.id, messages.UpdateContents{
		Pos:  f64.Vec2{1, 1},
		Tile: f64.Vec2{0, 0},
	}))

	return newClient.id, nil
}

func handleChatMessage(message *messages.Message) {
	sender := ServerInstance.clientsById[message.ClientId].username
	messageContents := message.Contents.(string)

	// Send the message to all clients
	m := messages.NewChatMessage(message.ClientId, fmt.Sprintf("%s: %s", sender, messageContents))
	for _, client := range ServerInstance.clientsById {
		if err := client.conn.WriteJSON(m); err != nil {
			log.Panicf("Failed to send message to: %s - %v", client.username, err)
		}
	}
}

func handleUpdateMessage(message *messages.Message) error {
	messageContents := message.Contents.(map[string]interface{})

	pos := messageContents["pos"].([]interface{})
	tile := messageContents["tile"].([]interface{})

	// Find the client to update
	client, found := ServerInstance.clientsById[message.ClientId]
	if !found {
		log.Printf("Client with id %d not found", message.ClientId)
	}

	client.mu.Lock()
	client.position = f64.Vec2{pos[0].(float64), pos[1].(float64)}
	client.tile = f64.Vec2{tile[0].(float64), tile[1].(float64)}
	client.mu.Unlock()

	sendUpdateToClients(message)

	return nil
}

func sendUpdateToClients(message *messages.Message) error {
	// Update all the other clients
	for _, c := range ServerInstance.clientsById {
		if c.id == message.ClientId || c.conn == nil {
			continue
		}

		c.mu.Lock()
		if err := c.conn.WriteJSON(message); err != nil {
			return err
		}
		c.mu.Unlock()
	}
	return nil
}

func serverLoop() {
	for {
		time.Sleep(50 * time.Millisecond)

		if len(ServerInstance.enemies) < 100 {
			// Create enemy
			enemy := NewEntity(f64.Vec2{0, 6 * 8}, f64.Vec2{rand.Float64() * (512 - 100), rand.Float64() * (512 - 100)})
			ServerInstance.enemies = append(ServerInstance.enemies, enemy)
		}

		// Update client with enemy positions
		contents := []messages.ServerEntityUpdateContents{}

		if len(ServerInstance.clientsById) > 0 {
			pos := ServerInstance.clientsById[0].position
			for _, e := range ServerInstance.enemies {
				e.Move(pos)
				contents = append(contents, messages.ServerEntityUpdateContents{
					EntityId: e.id,
					Pos:      e.pos,
					Tile:     e.tile,
				})
			}
		}

		message := messages.NewServerEntityUpdateMessage(contents)

		for _, c := range ServerInstance.clientsById {
			if c.conn == nil {
				continue
			}
			c.mu.Lock()
			err := c.conn.WriteJSON(message)
			if err != nil {
				log.Println("game write:", err)
				disconnectClient(c.id)
			}
			c.mu.Unlock()
		}
	}
}

func disconnectClient(clientId uint16) {
	client, found := ServerInstance.clientsById[clientId]

	if !found {
		log.Printf("Client %d not found", clientId)
	}

	client.mu.Lock()
	client.conn = nil
	client.mu.Unlock()

	// Update all the other clients
	message := messages.NewUpdateMessage(clientId, messages.UpdateContents{
		Disconnected: true,
	})

	for _, c := range ServerInstance.clientsById {
		if c.conn == nil {
			continue
		}
		c.mu.Lock()
		c.conn.WriteJSON(message)
		c.mu.Unlock()
	}
}
