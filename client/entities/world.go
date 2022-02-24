package entities

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	imageTile *ebiten.Image
}

func NewWorld(tilesImage *ebiten.Image) *World {
	w := &World{
		imageTile: tilesImage,
	}

	return w
}

func (w *World) Update() error {
	return nil
}

func (w *World) Draw(screen *ebiten.Image) {

	for x := 0; x <= 32; x++ {
		for y := 0; y <= 32; y++ {
			m := ebiten.GeoM{}

			m.Translate(float64(x*8), float64(y*8))
			m.Scale(4, 4)

			sx := 6 * 8
			sy := 4 * 8

			screen.DrawImage(w.imageTile.SubImage(image.Rect(sx, sy, sx+8, sy+8)).(*ebiten.Image),
				&ebiten.DrawImageOptions{
					GeoM: m,
				})
		}
	}
}
