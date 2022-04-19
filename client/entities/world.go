package entities

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	camera "github.com/melonfunction/ebiten-camera"
	shared_resources "github.com/tomknightdev/islanders-golang/resources"
)

type World struct {
	imageTile *ebiten.Image
	worldMap  shared_resources.WorldMap
	Cam       *camera.Camera
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
	// sx := 6 * 8
	// sy := 4 * 8

	w.Cam.Surface.Clear()

	tiles := [100][100]int{}
	for i, t := range w.worldMap.Layers[0].Data {
		tiles[i%100][i/100] = t
	}

	for y := 0; y < w.worldMap.Height; y++ {
		for x := 0; x < w.worldMap.Width; x++ {
			tileIndex := tiles[x][y]

			quotient := tileIndex / 16
			rem := tileIndex % 16
			tileXStart := (rem - 1) * 8
			tileYStart := quotient * 8
			tile := w.imageTile.SubImage(image.Rect(tileXStart, tileYStart, tileXStart+8, tileYStart+8))

			// Draw tiles
			w.Cam.Surface.DrawImage(tile.(*ebiten.Image), w.Cam.GetTranslation(float64(x*8), float64(y*8)))
		}
	}

	// Draw to screen and zoom
	w.Cam.Blit(screen)
}
