package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Inventory shows what the character is carrying and what's in the bank.
// What it's doing is shown by the action bar instead.
type Inventory struct {
	mgr      *renderer.Manager
	Contents resources.InventoryContents
}

func NewInventory(mgr *renderer.Manager) *Inventory {
	return &Inventory{
		mgr:      mgr,
		Contents: resources.InventoryContents{Activity: "Idle"},
	}
}

func (inv *Inventory) Update() error {
	w, h := screenSize()
	placeWindow("Inventory", imgui.Vec2{X: w - windowMargin, Y: h - windowMargin}, imgui.Vec2{X: 1, Y: 1})
	imgui.BeginV("Inventory", nil, imgui.WindowFlagsAlwaysAutoResize)
	section("Backpack", inv.Contents.Backpack)
	imgui.Separator()
	section("Bank", inv.Contents.Bank)
	imgui.End()

	return nil
}

func section(title string, c resources.Inventory) {
	imgui.Text(fmt.Sprintf("%s  %d/%d slots", title, len(c.Stacks), c.Slots))
	if len(c.Stacks) == 0 {
		imgui.Text("  Empty")
		return
	}
	// One line per item kind, in the order first seen.
	var kinds []resources.ItemKind
	seen := map[resources.ItemKind]bool{}
	for _, s := range c.Stacks {
		if !seen[s.Kind] {
			seen[s.Kind] = true
			kinds = append(kinds, s.Kind)
		}
	}
	for _, k := range kinds {
		imgui.Text(fmt.Sprintf("  %s  %d", resources.ItemNames[k], c.Count(k)))
		if imgui.IsItemHovered() {
			imgui.BeginTooltip()
			imgui.PushTextWrapPosV(tooltipWidth)
			imgui.Text(resources.ItemNames[k])
			dimmed(resources.ItemDescriptions[k])
			imgui.PopTextWrapPos()
			imgui.EndTooltip()
		}
	}
}

func (inv *Inventory) Draw(screen *ebiten.Image) {
	inv.mgr.Draw(screen)
}
