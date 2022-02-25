package entities

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tomknightdev/socketio-game-test/messages"
	"golang.org/x/image/math/f64"
)

type Player struct {
	imageTile *ebiten.Image
	Id        uint16
	Username  string
	Position  f64.Vec2
	SendChan  chan messages.UpdateContents
}

func NewPlayer(tilesImage *ebiten.Image) *Player {
	p := &Player{
		imageTile: tilesImage.SubImage(image.Rect(0, 0, 8, 8)).(*ebiten.Image),
		SendChan:  make(chan messages.UpdateContents),
	}

	return p
}

func (p *Player) Update() error {
	x := 0.0
	y := 0.0

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		x -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		x += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		y -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		y += 1
	}

	if p.Position[0]+x < 0 || p.Position[0]+x > 256 {
		x = 0
	}

	if p.Position[1]+y < 0 || p.Position[1]+y > 256 {
		y = 0
	}

	p.Position[0] += x
	p.Position[1] += y

	if x != 0 || y != 0 {
		p.SendChan <- messages.UpdateContents{
			Pos:  p.Position,
			Tile: f64.Vec2{0, 0},
		}
	}

	return nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	m := ebiten.GeoM{}

	m.Translate(p.Position[0], p.Position[1])
	m.Scale(2, 2)

	screen.DrawImage(p.imageTile, &ebiten.DrawImageOptions{
		GeoM: m,
	})
}
