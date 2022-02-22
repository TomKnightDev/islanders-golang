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
	connected  bool
	entities   []Entity
}

func main() {
	game := &Game{}

	mm := gui.NewMainMenu()

	go func() {
		game.playerName = <-mm.Connect
		err := connectToServer(game)
		if err != nil {
			log.Fatalf("failed to connect %s to server: %s", game.playerName, err)
		}
		game.connected = true
		go gameLoop(game)
	}()

	game.entities = append(game.entities, mm)

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Dungeon Crawl")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

}

func RemoveEntity(game *Game, i int) {
	if i >= len(game.entities) {
		return
	}

	game.entities[i] = game.entities[len(game.entities)-1]
	game.entities = game.entities[:len(game.entities)-1]

	// return append(scenes[:s], scenes[s+1:]...)
}
