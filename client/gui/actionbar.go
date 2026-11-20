package gui

import (
	"bytes"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	actionBarWidth  = 480
	actionBarHeight = 38
	actionBarMargin = 64 // gap below the bar, from the bottom of the window
)

var (
	actionBarBack   = color.RGBA{R: 0x10, G: 0x1a, B: 0x14, A: 0xd8}
	actionBarFill   = color.RGBA{R: 0x3c, G: 0xb3, B: 0x4a, A: 0xff}
	actionBarBorder = color.RGBA{R: 0xd8, G: 0xe8, B: 0xdc, A: 0xff}
	actionBarFace   *text.GoTextFace
)

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	actionBarFace = &text.GoTextFace{Source: src, Size: 22}
}

// ActionBar shows the current action in the lower middle of the screen as a
// bar that fills green over the action's length.
type ActionBar struct {
	activity string
	duration time.Duration // 0: no set length, bar stays empty
	start    time.Time
}

func NewActionBar() *ActionBar {
	return &ActionBar{}
}

// Set starts showing an action that lasts seconds, of which elapsed have
// already passed on the server.
func (b *ActionBar) Set(activity string, seconds, elapsed float64) {
	b.activity = activity
	b.duration = time.Duration(seconds * float64(time.Second))
	b.start = time.Now().Add(-time.Duration(elapsed * float64(time.Second)))
}

func (b *ActionBar) Update() error {
	return nil
}

func (b *ActionBar) Draw(screen *ebiten.Image) {
	if b.activity == "" || b.activity == "Idle" {
		return
	}
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	x := float32(sw-actionBarWidth) / 2
	y := float32(sh - actionBarHeight - actionBarMargin)

	vector.FillRect(screen, x, y, actionBarWidth, actionBarHeight, actionBarBack, false)
	if b.duration > 0 {
		frac := float32(time.Since(b.start)) / float32(b.duration)
		frac = min(max(frac, 0), 1)
		vector.FillRect(screen, x, y, actionBarWidth*frac, actionBarHeight, actionBarFill, false)
	}
	vector.StrokeRect(screen, x, y, actionBarWidth, actionBarHeight, 1, actionBarBorder, false)

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x)+actionBarWidth/2, float64(y)+actionBarHeight/2)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, b.activity, actionBarFace, op)
}
