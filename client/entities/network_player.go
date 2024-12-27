package entities

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
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
	p.Cam.Surface.DrawImage(p.imageTile, p.Cam.GetTranslation(p.Position[0], p.Position[1]))

	// Draw to screen and zoom
	p.Cam.Blit(screen)

	text.DrawWithOptions(screen, p.Username, mplusNormalFont, p.Cam.GetTranslation(p.Position[0], p.Position[1]))
}
