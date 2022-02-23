package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tomknightdev/socketio-game-test/messages"
)

var addr = flag.String("addr", GetOutboundIP().String()+":8285", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var clients = []*client{}

type client struct {
	id             uint16
	username       string
	chatConnection *websocket.Conn
	gameConnection *websocket.Conn
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

func main() {
	flag.Parse()
	log.SetFlags(0)

	http.HandleFunc("/connect", connect)
	http.HandleFunc("/chat", chatLoop)
	http.HandleFunc("/game", gameLoop)

	fmt.Printf(*addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
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
			id:       uint16(len(clients)),
			username: cm.Username,
		}

		c.WriteJSON(&messages.ConnectResponse{
			ClientId: newClient.id,
		})

		clients = append(clients, newClient)

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

		if gameMessage.ClientPos[0] == -1 {
			cc, err := getClient(gameMessage.ClientId)
			if err != nil {
				fmt.Print(err)
				continue
			}
			cc.gameConnection = c
		}

		// fmt.Println(gameMessage)

		for _, client := range clients {
			err = client.gameConnection.WriteJSON(gameMessage)
			if err != nil {
				log.Println("game write:", err)
				break
			}
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
			cc, err := getClient(chatMessage.ClientId)
			if err != nil {
				fmt.Print(err)
				continue
			}
			cc.chatConnection = c
		}

		for _, client := range clients {
			err = client.chatConnection.WriteJSON(chatMessage)
			if err != nil {
				log.Println("game write:", err)
				break
			}
		}
	}
}

func getClient(id uint16) (*client, error) {
	for _, c := range clients {
		if c.id == id {
			return c, nil
		}
	}

	return nil, fmt.Errorf("unable to find client with id %d", id)
}
