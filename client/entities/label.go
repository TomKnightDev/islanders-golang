package entities

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

var labelFace *text.GoTextFace

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	labelFace = &text.GoTextFace{Source: src, Size: 16}
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
