package entities

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/inkyblackness/imgui-go/v4"
	camera "github.com/melonfunction/ebiten-camera"
	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

type Player struct {
	imageTile *ebiten.Image
	Id        uint16
	Username  string
	Position  f64.Vec2
	SendChan  chan resources.MoveRequestContents
	Cam       *camera.Camera
	WorldMap  *resources.WorldMap
	flipX     bool

	clickTile  image.Point
	clickUntil time.Time
}

// Camera zoom limits and the zoom a session starts at. At 4x an 8px tile is
// 32px on screen.
const (
	StartZoom = 4.0
	MinZoom   = 1.0
	MaxZoom   = 10.0
)

// clickMarker outlines the tile the player last clicked.
var clickMarker = func() *ebiten.Image {
	img := ebiten.NewImage(resources.TileSize, resources.TileSize)
	c := color.RGBA{R: 0xff, G: 0xe0, B: 0x60, A: 0xff}
	for i := 0; i < resources.TileSize; i++ {
		img.Set(i, 0, c)
		img.Set(i, resources.TileSize-1, c)
		img.Set(0, i, c)
		img.Set(resources.TileSize-1, i, c)
	}
	return img
}()

func NewPlayer(tilesImage *ebiten.Image, tile f64.Vec2) *Player {
	p := &Player{
		imageTile: tilesImage.SubImage(image.Rect(int(tile[0]), int(tile[1]), int(tile[0])+8, int(tile[1])+8)).(*ebiten.Image),
		SendChan:  make(chan resources.MoveRequestContents, 8),
	}

	return p
}

// SetPosition applies a position from the server and faces the sprite the
// way it moved.
func (p *Player) SetPosition(pos f64.Vec2) {
	if pos[0] < p.Position[0] {
		p.flipX = true
	} else if pos[0] > p.Position[0] {
		p.flipX = false
	}
	p.Position = pos
}

func (p *Player) Update() error {
	// Click to walk. Clicks on a GUI window belong to the window.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !imgui.CurrentIO().WantCaptureMouse() {
		wx, wy := p.Cam.GetCursorCoords()
		target := image.Pt(int(math.Floor(wx/resources.TileSize)), int(math.Floor(wy/resources.TileSize)))
		select {
		case p.SendChan <- resources.MoveRequestContents{X: target.X, Y: target.Y}:
			p.clickTile = target
			p.clickUntil = time.Now().Add(time.Second)
		default: // network is backed up; drop the click rather than freeze the game
		}
	}

	p.Cam.SetPosition(p.Position[0]+float64(resources.TileSize)/2, p.Position[1]+float64(resources.TileSize)/2)

	// Zoom
	_, scrollAmount := ebiten.Wheel()
	if scrollAmount > 0 && p.Cam.Scale < MaxZoom {
		p.Cam.Zoom(1.1)
	} else if scrollAmount < 0 && p.Cam.Scale > MinZoom {
		p.Cam.Zoom(0.9)
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

	if time.Now().Before(p.clickUntil) {
		mx, my := float64(p.clickTile.X*resources.TileSize), float64(p.clickTile.Y*resources.TileSize)
		p.Cam.Surface.DrawImage(clickMarker, p.Cam.GetTranslation(&ebiten.DrawImageOptions{}, mx, my))
	}

	// Draw the player
	drawOpts := &ebiten.DrawImageOptions{}
	if p.flipX {
		drawOpts.GeoM.Scale(-1, 1)
		drawOpts.GeoM.Translate(8, 0)
	}
	opts := p.Cam.GetTranslation(drawOpts, p.Position[0], p.Position[1])
	p.Cam.Surface.DrawImage(p.imageTile, opts)
}

func (p *Player) DrawLabel(screen *ebiten.Image) {
	drawNameLabel(screen, p.Cam, p.Username, p.Position)

}
