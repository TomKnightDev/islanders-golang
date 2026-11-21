package gui

import (
	"image"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

const (
	classSpriteSize = 64 // an 8px sprite drawn at 8x
	classTextureID  = 100
)

// CharacterCreation lets a new player pick their class. It opens when the
// server says the name is new, and closes once the character exists.
type CharacterCreation struct {
	mgr    *renderer.Manager
	choose func(resources.ClassID)
	chosen bool
}

// NewCharacterCreation registers each class's sprite with imgui so the
// window can show it.
func NewCharacterCreation(mgr *renderer.Manager, characters *ebiten.Image, choose func(resources.ClassID)) *CharacterCreation {
	for i, c := range resources.Classes {
		x, y := c.Sprite[0], c.Sprite[1]
		mgr.Cache.SetTexture(imgui.TextureID(classTextureID+i), characters.SubImage(image.Rect(x, y, x+8, y+8)).(*ebiten.Image))
	}
	return &CharacterCreation{mgr: mgr, choose: choose}
}

func (cc *CharacterCreation) Update() error {
	w, h := screenSize()
	imgui.SetNextWindowPosV(imgui.Vec2{X: w / 2, Y: h / 2}, imgui.ConditionAlways, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.BeginV("Create your character", nil,
		imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoMove|imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text("Choose a class. You can train any skill later; this is where you start.")
	imgui.Spacing()
	for i, c := range resources.Classes {
		imgui.PushIDInt(i)
		imgui.Image(imgui.TextureID(classTextureID+i), imgui.Vec2{X: classSpriteSize, Y: classSpriteSize})
		imgui.SameLine()
		imgui.BeginGroup()
		imgui.Text(c.Name)
		imgui.PushTextWrapPosV(imgui.CursorPosX() + 520)
		dimmed(c.Desc)
		imgui.PopTextWrapPos()
		if button("Choose "+c.Name, !cc.chosen) {
			cc.chosen = true
			cc.choose(c.ID)
		}
		imgui.EndGroup()
		imgui.PopID()
		imgui.Spacing()
	}
	if cc.chosen {
		dimmed("Creating your character...")
	}
	imgui.End()
	return nil
}

func (cc *CharacterCreation) Draw(screen *ebiten.Image) {
	cc.mgr.Draw(screen)
}
