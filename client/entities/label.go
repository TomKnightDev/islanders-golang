package entities

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

var (
	labelFace *text.GoTextFace
	placeFace *text.GoTextFace
)

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	labelFace = &text.GoTextFace{Source: src, Size: 16}
	placeFace = &text.GoTextFace{Source: src, Size: 26}
}

// Labeled is anything that draws a name above itself. Labels go straight onto
// the screen after the camera image, so they stay sharp at any zoom.
type Labeled interface {
	DrawLabel(screen *ebiten.Image)
}

// drawNameLabel centres name just above a sprite whose top-left is at pos.
func drawNameLabel(screen *ebiten.Image, cam *camera.Camera, name string, pos f64.Vec2) {
	if name == "" || cam == nil {
		return
	}
	x, y := cam.GetScreenCoords(pos[0]+resources.TileSize/2, pos[1])
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y-4)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignEnd
	// A one-pixel shadow keeps names readable over light tiles.
	op.ColorScale.ScaleWithColor(color.Black)
	op.GeoM.Translate(1, 1)
	text.Draw(screen, name, labelFace, op)
	op.GeoM.Translate(-1, -1)
	op.ColorScale.Reset()
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, name, labelFace, op)
}

// drawPlaceLabel writes a village or mine name over the map, centred just
// above the place.
func drawPlaceLabel(screen *ebiten.Image, cam *camera.Camera, p resources.Place) {
	x, y := cam.GetScreenCoords(float64(p.X*resources.TileSize+resources.TileSize/2), float64((p.Y-p.Radius)*resources.TileSize))
	b := screen.Bounds()
	if x < -200 || y < -50 || x > float64(b.Dx())+200 || y > float64(b.Dy())+50 {
		return // off screen
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x+2, y+2)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignEnd
	op.ColorScale.ScaleWithColor(color.RGBA{A: 0xb0})
	text.Draw(screen, p.Name, placeFace, op)
	op.GeoM.Translate(-2, -2)
	op.ColorScale.Reset()
	op.ColorScale.ScaleWithColor(color.RGBA{R: 0xff, G: 0xf1, B: 0xc8, A: 0xff})
	text.Draw(screen, p.Name, placeFace, op)
}

var (
	healthBack = color.RGBA{R: 0x40, G: 0x10, B: 0x10, A: 0xe0}
	healthFill = color.RGBA{R: 0x4c, G: 0xc8, B: 0x50, A: 0xff}
	healthLow  = color.RGBA{R: 0xe0, G: 0x50, B: 0x40, A: 0xff}
)

// drawHealthBar draws a small bar above a sprite whose top-left is at pos,
// only when it's hurt.
func drawHealthBar(screen *ebiten.Image, cam *camera.Camera, pos f64.Vec2, health, maxHealth int) {
	if cam == nil || maxHealth <= 0 || health >= maxHealth {
		return
	}
	const w, h = 40, 5
	x, y := cam.GetScreenCoords(pos[0]+resources.TileSize/2, pos[1])
	frac := float32(max(health, 0)) / float32(maxHealth)
	fill := healthFill
	if frac < 0.3 {
		fill = healthLow
	}
	left, top := float32(x)-w/2, float32(y)-30
	vector.FillRect(screen, left, top, w, h, healthBack, false)
	vector.FillRect(screen, left, top, w*frac, h, fill, false)
}
