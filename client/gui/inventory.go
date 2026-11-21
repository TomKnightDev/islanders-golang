package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Inventory shows the character's coins and what they're carrying. The
// bank is only seen in the bank window, standing at a bank. What the
// character is doing is shown by the action bar instead.
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
	imgui.Text(fmt.Sprintf("Coins: %d", inv.Contents.Coins))
	imgui.Separator()
	section("Backpack", inv.Contents.Backpack)
	inventoryTop = imgui.WindowPos().Y
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
		if resources.IsPiece(k) {
			// Gear and tools are listed piece by piece, each with its quality.
			for _, s := range c.Stacks {
				if s.Kind == k {
					imgui.Text(fmt.Sprintf("  %s  q%d", resources.ItemNames[k], s.Quality))
					itemTooltip(k)
				}
			}
			continue
		}
		imgui.Text(fmt.Sprintf("  %s  %d  q%d", resources.ItemNames[k], c.Count(k), c.AverageQuality(k)))
		itemTooltip(k)
	}
}

// itemTooltip describes an item when the line just drawn is hovered.
func itemTooltip(k resources.ItemKind) {
	if !imgui.IsItemHovered() {
		return
	}
	imgui.BeginTooltip()
	imgui.PushTextWrapPosV(tooltipWidth)
	imgui.Text(resources.ItemNames[k])
	dimmed(resources.ItemDescriptions[k])
	imgui.PopTextWrapPos()
	imgui.EndTooltip()
}

func (inv *Inventory) Draw(screen *ebiten.Image) {
	inv.mgr.Draw(screen)
}
