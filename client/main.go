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
	playerName string
	serverAddr string
	connected  bool
	Entities   []Entity
}

func (g *Game) Update() error {
	for _, e := range g.Entities {
		e.Update()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, e := range g.Entities {
		e.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 800, 600
}

func main() {
	game := &Game{}

	mm := gui.NewMainMenu()

	go func() {
		game.serverAddr = <-mm.Connect
		game.playerName = <-mm.Connect
		err := connectToServer(game)
		if err != nil {
			log.Fatalf("failed to connect %s to server: %s", game.playerName, err)
		}
		game.connected = true
		go chatLoop(game)
	}()

	game.Entities = append(game.Entities, mm)

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Dungeon Crawl")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

}
