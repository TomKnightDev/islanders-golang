package entities

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/tomknightdev/socketio-game-test/client/settings"
	"golang.org/x/image/math/f64"
)

type NetworkPlayer struct {
	imageTile *ebiten.Image
	Id        uint16
	Username  string
	Position  f64.Vec2
	Tile      f64.Vec2
}

func NewNetworkPlayer(tilesImage *ebiten.Image, tile f64.Vec2) *NetworkPlayer {
	p := &NetworkPlayer{
		imageTile: tilesImage.SubImage(image.Rect(int(tile[0]), int(tile[1]), int(tile[0])+8, int(tile[1])+8)).(*ebiten.Image),
	}

	return p
}

func (p *NetworkPlayer) Update() error {

	return nil
}

func (p *NetworkPlayer) Draw(screen *ebiten.Image) {
	m := ebiten.GeoM{}

	m.Translate(p.Position[0], p.Position[1])
	m.Scale(settings.Scale, settings.Scale)

	screen.DrawImage(p.imageTile, &ebiten.DrawImageOptions{
		GeoM: m,
	})

	text.Draw(screen, p.Username, mplusNormalFont, int(p.Position[0]*settings.Scale), int(p.Position[1]*settings.Scale), color.White)
}
