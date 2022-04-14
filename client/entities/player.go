package entities

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/tomknightdev/socketio-game-test/client/settings"
	"github.com/tomknightdev/socketio-game-test/resources"
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

	if p.Position[0]+x < 0 || p.Position[0]+x > 256 {
		x = 0
	}

	if p.Position[1]+y < 0 || p.Position[1]+y > 256 {
		y = 0
	}

	p.Position[0] += x
	p.Position[1] += y

	if x != 0 || y != 0 {
		p.SendChan <- resources.UpdateContents{
			Pos:  p.Position,
			Tile: f64.Vec2{0, 0},
		}
	}

	return nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	m := ebiten.GeoM{}

	m.Translate(p.Position[0], p.Position[1])
	m.Scale(settings.Scale, settings.Scale)

	screen.DrawImage(p.imageTile, &ebiten.DrawImageOptions{
		GeoM: m,
	})

	text.Draw(screen, p.Username, mplusNormalFont, int(p.Position[0]*settings.Scale), int(p.Position[1]*settings.Scale), color.White)

}
