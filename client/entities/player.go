package entities

import (
	"image"
	"image/color"
	"log"

	"github.com/google/uuid"
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
	imageTile       *ebiten.Image
	Id              uuid.UUID
	Username        string
	Position        f64.Vec2
	SendChan        chan resources.UpdateContents
	Cam             *camera.Camera
	fireTime        int
	currentFireTime int
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
		imageTile:       tilesImage.SubImage(image.Rect(int(tile[0]), int(tile[1]), int(tile[0])+8, int(tile[0])+8)).(*ebiten.Image),
		SendChan:        make(chan resources.UpdateContents),
		fireTime:        100,
		currentFireTime: 0,
	}

	return p
}

func (p *Player) Update() error {
	x := 0.0
	y := 0.0

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		x -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		x += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		y -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		y += 1
	}

	// if p.Position[0]+x < 0 || p.Position[0]+x > 256 {
	// 	x = 0
	// }

	// if p.Position[1]+y < 0 || p.Position[1]+y > 256 {
	// 	y = 0
	// }

	p.Position[0] += x
	p.Position[1] += y

	if x != 0 || y != 0 {
		p.SendChan <- resources.UpdateContents{
			Pos:  p.Position,
			Tile: f64.Vec2{0, 0},
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

	// Fire direction
	fx := 0.0
	fy := 0.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		fx -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		fx += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		fy -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		fy += 1
	}

	// Fire projectile
	if p.currentFireTime >= p.fireTime && (fx != 0 || fy != 0) {
		p.FireProjectile(f64.Vec2{
			p.Position[0] + fx,
			p.Position[1] + fy,
		})
		p.currentFireTime = 0
	} else {
		p.currentFireTime++
	}

	return nil
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
	p.Cam.Surface.DrawImage(p.imageTile, p.Cam.GetTranslation(p.Position[0], p.Position[1]))

	// Draw to screen and zoom
	p.Cam.Blit(screen)

	text.Draw(screen, p.Username, mplusNormalFont, int(p.Cam.X+float64(16)/2), int(p.Cam.Y+float64(16)/2), color.White)

	// text.DrawWithOptions(screen, p.Username, mplusNormalFont, p.Cam.GetTranslation(p.Position[0]*p.Cam.Scale, p.Position[1]*p.Cam.Scale))

}

func (p *Player) FireProjectile(targetPos f64.Vec2) {
	// Get closest enemy pos

	NewProjectile(p.Position, targetPos)
}
