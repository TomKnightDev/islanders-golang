package entities

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	camera "github.com/melonfunction/ebiten-camera"
	"golang.org/x/image/math/f64"
)

type NetworkPlayer struct {
	imageTile *ebiten.Image
	Id        uint16
	Username  string
	Position  f64.Vec2
	Tile      f64.Vec2
	Cam       *camera.Camera
	flipX     bool
}

func (p *NetworkPlayer) SetPosition(pos f64.Vec2) {
	if pos[0] < p.Position[0] {
		p.flipX = true
	} else if pos[0] > p.Position[0] {
		p.flipX = false
	}
	p.Position = pos
}

func NewNetworkPlayer(tilesImage *ebiten.Image, tile f64.Vec2) *NetworkPlayer {
	p := &NetworkPlayer{
		imageTile: tilesImage.SubImage(image.Rect(int(tile[0]), int(tile[1]), int(tile[0])+8, int(tile[1])+8)).(*ebiten.Image),
	}

	return p
}

// Update
func (p *NetworkPlayer) Update() error {
	return nil
}

func (p *NetworkPlayer) Draw(screen *ebiten.Image) {
	// m := ebiten.GeoM{}

	// m.Translate(p.Position[0], p.Position[1])
	// m.Scale(settings.Scale, settings.Scale)

	// screen.DrawImage(p.imageTile, &ebiten.DrawImageOptions{
	// 	GeoM: m,
	// })

	// text.Draw(screen, p.Username, mplusNormalFont, int(p.Position[0]*settings.Scale), int(p.Position[1]*settings.Scale), color.White)

	// Draw the player
	drawOpts := &ebiten.DrawImageOptions{}
	if p.flipX {
		drawOpts.GeoM.Scale(-1, 1)
		drawOpts.GeoM.Translate(8, 0)
	}
	opts := p.Cam.GetTranslation(drawOpts, p.Position[0], p.Position[1])
	p.Cam.Surface.DrawImage(p.imageTile, opts)
}

func (p *NetworkPlayer) DrawLabel(screen *ebiten.Image) {
	drawNameLabel(screen, p.Cam, p.Username, p.Position)
}
