package entities

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/math/f64"
)

type NetworkPlayer struct {
	imageTile *ebiten.Image
	Id        uint16
	Username  string
	Position  f64.Vec2
}

func NewNetworkPlayer(tilesImage *ebiten.Image) *NetworkPlayer {
	p := &NetworkPlayer{
		imageTile: tilesImage.SubImage(image.Rect(8, 0, 16, 8)).(*ebiten.Image),
	}

	return p
}

func (p *NetworkPlayer) Update() error {

	return nil
}

func (p *NetworkPlayer) Draw(screen *ebiten.Image) {
	m := ebiten.GeoM{}

	m.Translate(p.Position[0], p.Position[1])
	m.Scale(4, 4)

	screen.DrawImage(p.imageTile, &ebiten.DrawImageOptions{
		GeoM: m,
	})
}
