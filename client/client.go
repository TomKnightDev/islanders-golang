package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/socketio-game-test/messages"
)

var addr = flag.String("addr", "localhost:8000", "http service address")

type client struct {
	id       uint16
	username string
}

var player = client{}

type Game struct {
	mgr          *renderer.Manager
	name         string
	connected    bool
	sendChan     chan string
	recvChan     chan string
	message      string
	recvMessages []string
}

func (g *Game) Update() error {
	g.mgr.Update(1.0 / 60.0)
	g.mgr.BeginFrame()
	{
		imgui.Text("Hello, world!")
		if !g.connected {
			imgui.InputText("Name", &g.name)
			if imgui.Button("Connect") {

				err := connectToServer(g)
				if err != nil {
					log.Fatalf("failed to connect to server: %s", err)
				}
				g.connected = true
				go gameLoop(g)
			}
		} else {
			imgui.InputText("Name", &g.message)
			if imgui.Button("Send") {
				g.sendChan <- g.message
			}

			for _, m := range g.recvMessages {
				imgui.Text(m)
			}
		}
	}
	g.mgr.EndFrame()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.mgr.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.mgr.SetDisplaySize(float32(800), float32(600))
	return 800, 600
}

func main() {
	mgr := renderer.New(nil)

	game := &Game{
		mgr:      mgr,
		sendChan: make(chan string),
		recvChan: make(chan string),
	}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Your game's title")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

}

func connectToServer(g *Game) error {
	fmt.Println("Client starting...")
	player.username = g.name

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

			g.recvMessages = append(g.recvMessages, fmt.Sprint(glm.ClientId, glm.Message))
			fmt.Println(glm.ClientId, glm.Message)
		}
	}()

	// Send
	for {
		msg := <-g.sendChan
		glm := &messages.GameLoopMessage{
			ClientId: player.id,
			Message:  msg,
		}
		c.WriteJSON(glm)
	}
}
