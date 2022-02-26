package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/tomknightdev/socketio-game-test/client/gui"
)

type Entity interface {
	Update() error
	Draw(*ebiten.Image)
}

type Game struct {
	username     string
	password     string
	serverAddr   string
	Gui          []Entity
	Environment  []Entity
	Entities     []Entity
	Player       Entity
	screenWidth  int
	screenHeight int
}

func (g *Game) Update() error {
	if g.Player != nil {
		if err := g.Player.Update(); err != nil {
			log.Print(err)
		}
	}
	for _, gui := range g.Gui {
		gui.Update()
	}
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
	for _, gui := range g.Gui {
		gui.Draw(screen)
	}
	if g.Player != nil {
		g.Player.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (sc, sh int) {
	return g.screenWidth, g.screenHeight
}

func main() {
	game := &Game{
		screenWidth:  800,
		screenHeight: 600,
	}

	mm := gui.NewMainMenu(game.screenWidth, game.screenHeight)

	go func() {
		game.serverAddr = <-mm.Connect
		game.username = <-mm.Connect
		game.password = <-mm.Connect
		err := connectToServer(game)
		if err != nil {
			log.Fatalf("Failed to connect %s to server: %s", game.username, err)
		}
		// go chatLoop(game)
		// go gameLoop(game)
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
