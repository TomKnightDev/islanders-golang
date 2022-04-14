package entities

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tomknightdev/socketio-game-test/client/settings"
	shared_resources "github.com/tomknightdev/socketio-game-test/resources"
)

type World struct {
	imageTile *ebiten.Image
	worldMap  shared_resources.WorldMap
}

func NewWorld(tilesImage *ebiten.Image, worldMap shared_resources.WorldMap) *World {
	w := &World{
		imageTile: tilesImage,
		worldMap:  worldMap,
	}

	return w
}

func (w *World) Update() error {
	return nil
}

func (w *World) Draw(screen *ebiten.Image) {
	// Tile coords
	sx := 6 * 8
	sy := 4 * 8

	for x := 0; x <= w.worldMap.Width; x++ {
		for y := 0; y <= w.worldMap.Height; y++ {
			m := ebiten.GeoM{}

			m.Translate(float64(x*8), float64(y*8))
			m.Scale(settings.Scale, settings.Scale)

			screen.DrawImage(w.imageTile.SubImage(image.Rect(sx, sy, sx+8, sy+8)).(*ebiten.Image),
				&ebiten.DrawImageOptions{
					GeoM: m,
				})
		}
	}
}
