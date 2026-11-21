package entities

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	Health    int
	MaxHealth int
	Cam       *camera.Camera
	WorldMap  *resources.WorldMap
	flipX     bool

	clickTile  image.Point
	clickUntil time.Time

	// FollowCam keeps the camera on the character; dragging the map turns it
	// off until the player asks to follow again.
	FollowCam  bool
	pressing   bool
	dragging   bool
	pressAt    image.Point
	camAtPress f64.Vec2

	// path is the route the server says the character is walking, drawn only
	// on this player's screen.
	path []image.Point
}

// dragThreshold is how far, in pixels, the mouse must move before a press
// is a drag rather than a click.
const dragThreshold = 6

var (
	pathColour = color.RGBA{R: 0xff, G: 0xf0, B: 0xa0, A: 0xb0}
)

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
		FollowCam: true,
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

// SetPath sets the route the character is walking.
func (p *Player) SetPath(tiles [][2]int) {
	p.path = p.path[:0]
	for _, t := range tiles {
		p.path = append(p.path, image.Pt(t[0], t[1]))
	}
}

func (p *Player) Update() error {
	p.handleMouse()

	// Drop route tiles the character has reached.
	for len(p.path) > 0 {
		x, y := float64(p.path[0].X*resources.TileSize), float64(p.path[0].Y*resources.TileSize)
		if math.Abs(p.Position[0]-x) > 0.01 || math.Abs(p.Position[1]-y) > 0.01 {
			break
		}
		p.path = p.path[1:]
	}

	if p.FollowCam {
		p.Cam.SetPosition(p.Position[0]+float64(resources.TileSize)/2, p.Position[1]+float64(resources.TileSize)/2)
	}

	// Zoom
	_, scrollAmount := ebiten.Wheel()
	if scrollAmount > 0 && p.Cam.Scale < MaxZoom {
		p.Cam.Zoom(1.1)
	} else if scrollAmount < 0 && p.Cam.Scale > MinZoom {
		p.Cam.Zoom(0.9)
	}

	return nil
}

// handleMouse tells a click (walk there) from a drag (pan the map). Presses
// that start on a GUI window belong to the window.
func (p *Player) handleMouse() {
	cx, cy := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !imgui.CurrentIO().WantCaptureMouse() {
		p.pressing, p.dragging = true, false
		p.pressAt = image.Pt(cx, cy)
		p.camAtPress = f64.Vec2{p.Cam.X, p.Cam.Y}
	}
	if !p.pressing {
		return
	}
	dx, dy := cx-p.pressAt.X, cy-p.pressAt.Y
	if !p.dragging && dx*dx+dy*dy > dragThreshold*dragThreshold {
		p.dragging = true
		p.FollowCam = false
	}
	if p.dragging {
		p.Cam.SetPosition(p.camAtPress[0]-float64(dx)/p.Cam.Scale, p.camAtPress[1]-float64(dy)/p.Cam.Scale)
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		p.pressing = false
		if !p.dragging {
			p.walkToCursor()
		}
	}
}

func (p *Player) walkToCursor() {
	wx, wy := p.Cam.GetCursorCoords()
	target := image.Pt(int(math.Floor(wx/resources.TileSize)), int(math.Floor(wy/resources.TileSize)))
	select {
	case p.SendChan <- resources.MoveRequestContents{X: target.X, Y: target.Y}:
		p.clickTile = target
		p.clickUntil = time.Now().Add(time.Second)
	default: // network is backed up; drop the click rather than freeze the game
	}
}

// DrawPath draws the character's route from where they are to where they're
// going, on screen after the camera image.
func (p *Player) DrawPath(screen *ebiten.Image) {
	if len(p.path) == 0 || p.Cam == nil {
		return
	}
	half := float64(resources.TileSize) / 2
	x0, y0 := p.Cam.GetScreenCoords(p.Position[0]+half, p.Position[1]+half)
	for _, t := range p.path {
		x1, y1 := p.Cam.GetScreenCoords(float64(t.X*resources.TileSize)+half, float64(t.Y*resources.TileSize)+half)
		vector.StrokeLine(screen, float32(x0), float32(y0), float32(x1), float32(y1), 3, pathColour, true)
		x0, y0 = x1, y1
	}
	vector.FillCircle(screen, float32(x0), float32(y0), 6, pathColour, true)
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
	drawHealthBar(screen, p.Cam, p.Position, p.Health, p.MaxHealth)
}
