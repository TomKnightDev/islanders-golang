package main

import (
	"log"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/tomknightdev/socketio-game-test/client/gui"
)

type Entity interface {
	Update() error
	Draw(*ebiten.Image)
}

type Game struct {
	username             string
	password             string
	serverAddr           string
	Gui                  []Entity
	Environment          []Entity
	Entities             map[uint16]Entity
	Player               Entity
	screenWidth          int
	screenHeight         int
	ConnectFailedMessage chan string
	renderMgr            *renderer.Manager
}

func (g *Game) Update() error {
	if g.Player != nil {
		if err := g.Player.Update(); err != nil {
			log.Print(err)
		}
	}

	g.renderMgr.Update(1.0 / 60.0)
	g.renderMgr.BeginFrame()
	{
		for _, gui := range g.Gui {
			gui.Update()
		}
	}
	g.renderMgr.EndFrame()

	for _, e := range g.Entities {
		e.Update()
	}
	for _, e := range g.Environment {
		e.Update()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, e := range g.Environment {
		e.Draw(screen)
	}
	for _, e := range g.Entities {
		e.Draw(screen)
	}
	if g.Player != nil {
		g.Player.Draw(screen)
	}
	for _, gui := range g.Gui {
		gui.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (sc, sh int) {
	return g.screenWidth, g.screenHeight
}

func main() {
	game := &Game{
		screenWidth:          1024,
		screenHeight:         768,
		Entities:             make(map[uint16]Entity),
		ConnectFailedMessage: make(chan string),
		renderMgr:            renderer.New(nil),
	}

	game.renderMgr.SetDisplaySize(float32(game.screenWidth), float32(game.screenHeight))

	mm := gui.NewMainMenu(game.screenWidth, game.screenHeight, game.renderMgr)

	go func() {
		for {
			m := <-game.ConnectFailedMessage
			mm.FailedToConnect = append(mm.FailedToConnect, m)
		}
	}()

	go func() {
		for {
			game.serverAddr = <-mm.Connect
			game.username = <-mm.Connect
			game.password = <-mm.Connect
			err := connectToServer(game)
			if err != nil {
				log.Printf("Failed to connect %s to server: %s", game.username, err)
			} else {
				return
			}
		}
	}()

	game.Gui = append(game.Gui, mm)

	ebiten.SetWindowSize(game.screenWidth, game.screenHeight)
	ebiten.SetWindowTitle("Dungeon Crawl")
	ebiten.SetWindowResizable(true)
	ebiten.NewImage(512, 512)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
