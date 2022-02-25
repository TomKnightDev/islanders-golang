package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomknightdev/socketio-game-test/messages"
	"golang.org/x/image/math/f64"
)

var addr = flag.String("addr", GetOutboundIP().String()+":8285", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var Server = server{}

type server struct {
	clients map[uint16]*client
	enemies []*Entity
}
type client struct {
	id             uint16
	username       string
	chatConnection *websocket.Conn
	gameConnection *websocket.Conn
	position       f64.Vec2
	mu             sync.Mutex
}

func init() {
	Server.clients = make(map[uint16]*client)

	go serverLoop()
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func connect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("connect read:", err)
			break
		}

		var cm = &messages.ConnectRequest{}
		if err = json.Unmarshal(message, cm); err != nil {
			log.Print(err)
			continue
		}

		log.Printf("connect recv: %s", message)

		newClient := &client{
			id:       uint16(len(Server.clients)),
			username: cm.Username,
		}

		c.WriteJSON(&messages.ConnectResponse{
			ClientId: newClient.id,
		})

		Server.clients[newClient.id] = newClient

		// for id, client := range clients {
		// 	err = client.connection.WriteMessage(mt, []byte(fmt.Sprintf("%d: %s", id, message)))
		// 	if err != nil {
		// 		log.Println("write:", err)
		// 		break
		// 	}
		// }
	}
}

func gameLoop(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("game read:", err)
			break
		}

		var gameMessage = &messages.GameLoopMessage{}
		if err = json.Unmarshal(message, gameMessage); err != nil {
			log.Print(err)
			continue
		}

		for _, em := range gameMessage.EntityMessages {
			client := Server.clients[em.EntityId]

			if em.EntityPos[0] == -1 {
				client.gameConnection = c
			}

			// Update server side position of client
			client.position = em.EntityPos

			for _, c := range Server.clients {
				// This is the client sending the message
				if c.id == client.id {
					continue
				}

				err = c.gameConnection.WriteJSON(gameMessage)
				if err != nil {
					log.Println("game write:", err)
					break
				}
			}
		}
	}
}

func serverLoop() {
	for {
		time.Sleep(50 * time.Millisecond)

		if len(Server.enemies) < 10 {
			// Create enemy
			enemy := NewEntity(f64.Vec2{0, 6 * 8}, f64.Vec2{rand.Float64() * (256 - 100), rand.Float64() * (256 - 100)})
			Server.enemies = append(Server.enemies, enemy)
		}

		// Update client with enemy positions
		glm := messages.GameLoopMessage{}

		if len(Server.clients) > 0 {
			for _, e := range Server.enemies {
				e.Move(Server.clients[0].position)
				entityMessage := &messages.EntityMessage{
					EntityId:   e.id,
					EntityPos:  e.pos,
					EntityTile: e.tile,
				}
				glm.EntityMessages = append(glm.EntityMessages, *entityMessage)

			}
		}

		for _, c := range Server.clients {
			c.mu.Lock()
			err := c.gameConnection.WriteJSON(glm)
			if err != nil {
				log.Println("game write:", err)
				break
			}
			c.mu.Unlock()
		}
	}
}

func chatLoop(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("game read:", err)
			break
		}

		var chatMessage = &messages.ChatLoopMessage{}
		if err = json.Unmarshal(message, chatMessage); err != nil {
			log.Print(err)
			continue
		}

		log.Printf("game recv: %s", message)

		if chatMessage.Message == "connected" {
			cc := Server.clients[chatMessage.ClientId]
			if err != nil {
				fmt.Print(err)
				continue
			}
			cc.chatConnection = c
		}

		for _, client := range Server.clients {
			err = client.chatConnection.WriteJSON(chatMessage)
			if err != nil {
				log.Println("game write:", err)
				break
			}
		}
	}
}
