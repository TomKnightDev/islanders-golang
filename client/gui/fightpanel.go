package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tomknightdev/islanders-golang/resources"
)

const (
	fightPanelWidth  = 460
	fightPanelHeight = 150
	fightPanelGap    = 12 // between the panel and the action bar below it
	fightSpriteScale = 7  // 8px sprites drawn at 56px
	fightBarWidth    = 150
	fightBarHeight   = 10

	lungeTime   = 350 * time.Millisecond  // attacker darts towards the target and back
	lungeDist   = 22.0                    // pixels
	flashTime   = 250 * time.Millisecond  // a hit target flashes red
	floatTime   = 900 * time.Millisecond  // the damage number rises and fades
	lingerAfter = 1500 * time.Millisecond // the panel stays up after the fight ends
)

var (
	fightPanelBack = color.RGBA{R: 0x10, G: 0x1a, B: 0x14, A: 0xe0}
	fightBarBack   = color.RGBA{R: 0x40, G: 0x10, B: 0x10, A: 0xff}
	fightBarFill   = color.RGBA{R: 0x4c, G: 0xc8, B: 0x50, A: 0xff}
	fightDealt     = color.RGBA{R: 0xff, G: 0xe0, B: 0x60, A: 0xff} // damage the player deals
	fightTaken     = color.RGBA{R: 0xff, G: 0x60, B: 0x50, A: 0xff} // damage the player takes
	fightNameFace  *text.GoTextFace
)

func init() {
	fightNameFace = &text.GoTextFace{Source: actionBarFace.Source, Size: 16}
}

// FightPanel shows the player against the enemy they're fighting, above
// the action bar: both sprites, names and health, with a lunge when either
// attacks, a red flash on a hit, and the damage floating up.
type FightPanel struct {
	chars        *ebiten.Image
	playerSprite *ebiten.Image
	playerName   string

	fight       resources.FightContents
	enemySprite *ebiten.Image
	active      bool
	endedAt     time.Time

	// The most recent attack, for the animation.
	attacker string
	hit      bool
	damage   int
	attackAt time.Time
}

func NewFightPanel(chars *ebiten.Image, playerSprite [2]int, playerName string) *FightPanel {
	return &FightPanel{
		chars:        chars,
		playerSprite: spriteAt(chars, playerSprite),
		playerName:   playerName,
	}
}

func spriteAt(chars *ebiten.Image, at [2]int) *ebiten.Image {
	return chars.SubImage(image.Rect(at[0], at[1], at[0]+8, at[1]+8)).(*ebiten.Image)
}

// Set applies a fight message from the server.
func (f *FightPanel) Set(c resources.FightContents) {
	if !c.Active {
		if f.active {
			f.active = false
			f.endedAt = time.Now()
		}
		return
	}
	if !f.active || c.EnemyName != f.fight.EnemyName || c.EnemySprite != f.fight.EnemySprite {
		f.enemySprite = spriteAt(f.chars, c.EnemySprite)
	}
	f.active = true
	f.fight = c
	if c.Attacker != "" {
		f.attacker, f.hit, f.damage, f.attackAt = c.Attacker, c.Hit, c.Damage, time.Now()
	}
}

func (f *FightPanel) Update() error { return nil }

func (f *FightPanel) Draw(screen *ebiten.Image) {
	if !f.active && time.Since(f.endedAt) > lingerAfter {
		return
	}
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	x := float32(sw-fightPanelWidth) / 2
	y := float32(sh - actionBarHeight - actionBarMargin - fightPanelGap - fightPanelHeight)
	vector.FillRect(screen, x, y, fightPanelWidth, fightPanelHeight, fightPanelBack, false)
	vector.StrokeRect(screen, x, y, fightPanelWidth, fightPanelHeight, 1, actionBarBorder, false)

	left := float64(x) + fightPanelWidth*0.22
	right := float64(x) + fightPanelWidth*0.78
	spriteY := float64(y) + 14
	since := time.Since(f.attackAt)

	// The attacker lunges towards the other side and back.
	lunge := 0.0
	if since < lungeTime {
		lunge = math.Sin(math.Pi*float64(since)/float64(lungeTime)) * lungeDist
	}
	playerX, enemyX := left, right
	switch f.attacker {
	case resources.AttackerPlayer:
		playerX += lunge
	case resources.AttackerEnemy:
		enemyX -= lunge
	}

	flashPlayer := f.hit && f.attacker == resources.AttackerEnemy && since < flashTime
	flashEnemy := f.hit && f.attacker == resources.AttackerPlayer && since < flashTime
	f.drawSprite(screen, f.playerSprite, playerX, spriteY, false, flashPlayer, false)
	f.drawSprite(screen, f.enemySprite, enemyX, spriteY, true, flashEnemy, f.fight.EnemyHealth == 0)

	centreText(screen, "VS", actionBarFace, float64(x)+fightPanelWidth/2, spriteY+28, color.White)

	nameY := spriteY + 8*fightSpriteScale + 16
	centreText(screen, f.playerName, fightNameFace, left, nameY, color.White)
	centreText(screen, f.fight.EnemyName, fightNameFace, right, nameY, color.White)
	barY := float32(nameY) + 14
	drawFightBar(screen, float32(left), barY, f.fight.PlayerHealth, f.fight.PlayerMaxHealth)
	drawFightBar(screen, float32(right), barY, f.fight.EnemyHealth, f.fight.EnemyMaxHealth)

	// The damage (or "Miss") floats up from whoever was attacked.
	if since < floatTime && f.attacker != "" {
		label, colour, at := "Miss", color.Color(color.White), right
		if f.attacker == resources.AttackerEnemy {
			at = left
		}
		if f.hit {
			label = fmt.Sprintf("-%d", f.damage)
			colour = fightDealt
			if f.attacker == resources.AttackerEnemy {
				colour = fightTaken
			}
		}
		progress := float64(since) / float64(floatTime)
		op := &text.DrawOptions{}
		op.GeoM.Translate(at, spriteY+10-progress*24)
		op.PrimaryAlign = text.AlignCenter
		op.ColorScale.ScaleWithColor(colour)
		op.ColorScale.ScaleAlpha(float32(1 - progress))
		text.Draw(screen, label, actionBarFace, op)
	}
}

// drawSprite draws a fighter centred on cx. Enemies face left, towards the
// player; a hit target flashes red; a defeated enemy is greyed out.
func (f *FightPanel) drawSprite(screen, sprite *ebiten.Image, cx, y float64, faceLeft, flash, defeated bool) {
	if sprite == nil {
		return
	}
	size := float64(8 * fightSpriteScale)
	op := &ebiten.DrawImageOptions{}
	if faceLeft {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(8, 0)
	}
	op.GeoM.Scale(fightSpriteScale, fightSpriteScale)
	op.GeoM.Translate(cx-size/2, y)
	switch {
	case defeated:
		op.ColorScale.Scale(0.4, 0.4, 0.4, 0.8)
	case flash:
		op.ColorScale.Scale(1.6, 0.4, 0.4, 1)
	}
	screen.DrawImage(sprite, op)
}

func drawFightBar(screen *ebiten.Image, cx, y float32, health, maxHealth int) {
	x := cx - fightBarWidth/2
	vector.FillRect(screen, x, y, fightBarWidth, fightBarHeight, fightBarBack, false)
	if maxHealth > 0 {
		vector.FillRect(screen, x, y, fightBarWidth*float32(max(health, 0))/float32(maxHealth), fightBarHeight, fightBarFill, false)
	}
	centreText(screen, fmt.Sprintf("%d / %d", max(health, 0), maxHealth), fightNameFace, float64(cx), float64(y+fightBarHeight)+12, color.White)
}

func centreText(screen *ebiten.Image, s string, face *text.GoTextFace, cx, cy float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(cx, cy)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(c)
	text.Draw(screen, s, face, op)
}
