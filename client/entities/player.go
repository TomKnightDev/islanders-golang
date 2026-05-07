package entities

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/f64"
)

var (
	mplusNormalFont font.Face
)

type Player struct {
	imageTile *ebiten.Image
	Id        uint16
	Username  string
	Position  f64.Vec2
	SendChan  chan resources.UpdateContents
	Cam       *camera.Camera
	WorldMap  *resources.WorldMap
}

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    12,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func NewPlayer(tilesImage *ebiten.Image, tile f64.Vec2) *Player {
	p := &Player{
		imageTile: tilesImage.SubImage(image.Rect(int(tile[0]), int(tile[1]), int(tile[0])+8, int(tile[0])+8)).(*ebiten.Image),
		SendChan:  make(chan resources.UpdateContents),
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

	// if p.Position[0]+x < 0 || p.Position[0]+x > 256 {
	// 	x = 0
	// }

	// if p.Position[1]+y < 0 || p.Position[1]+y > 256 {
	// 	y = 0
	// }

	if x != 0 || y != 0 {
		newX := p.Position[0] + x
		newY := p.Position[1] + y
		if p.canMoveTo(newX, newY) {
			p.Position[0] = newX
			p.Position[1] = newY
			p.SendChan <- resources.UpdateContents{
				Pos:  p.Position,
				Tile: f64.Vec2{0, 0},
			}
		}
	}

	p.Cam.SetPosition(p.Position[0]+float64(16)/2, p.Position[1]+float64(16)/2)

	// Zoom
	_, scrollAmount := ebiten.Wheel()
	if scrollAmount > 0 {
		p.Cam.Zoom(1.1)
	} else if scrollAmount < 0 {
		p.Cam.Zoom(0.9)
	}

	return nil
}

func (p *Player) canMoveTo(px, py float64) bool {
	if p.WorldMap == nil || len(p.WorldMap.Layers) == 0 {
		return true
	}
	w := p.WorldMap.Width
	data := p.WorldMap.Layers[0].Data
	// Check all four corners of the 8x8 player sprite
	for _, corner := range [4][2]float64{{px, py}, {px + 7, py}, {px, py + 7}, {px + 7, py + 7}} {
		tx, ty := int(corner[0]/8), int(corner[1]/8)
		if tx < 0 || ty < 0 || tx >= w || ty >= p.WorldMap.Height {
			return false
		}
		if !resources.IsPassable(data[ty*w+tx]) {
			return false
		}
	}
	return true
}

func (p *Player) Draw(screen *ebiten.Image) {
	// m := ebiten.GeoM{}

	// m.Translate(p.Position[0], p.Position[1])
	// m.Scale(settings.Scale, settings.Scale)

	// screen.DrawImage(p.imageTile, &ebiten.DrawImageOptions{
	// 	GeoM: m,
	// })

	// text.Draw(screen, p.Username, mplusNormalFont, int(p.Position[0]), int(p.Position[1]), color.White)

	// Draw the player
	opts := p.Cam.GetTranslation(&ebiten.DrawImageOptions{}, p.Position[0], p.Position[1])
	p.Cam.Surface.DrawImage(p.imageTile, opts)

	// Draw username above the sprite using the same screen-space position
	sx := opts.GeoM.Element(0, 2)
	sy := opts.GeoM.Element(1, 2)
	text.Draw(p.Cam.Surface, p.Username, mplusNormalFont, int(sx)-len(p.Username)*3, int(sy)-10, color.White)

	// Draw to screen and zoom
	p.Cam.Blit(screen)

}
