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
	Open     bool // shown; the menu bar toggles it
	picker   stackPicker
	mgr      *renderer.Manager
	Contents resources.InventoryContents
}

func NewInventory(mgr *renderer.Manager) *Inventory {
	return &Inventory{
		Open:     true,
		mgr:      mgr,
		Contents: resources.InventoryContents{Activity: "Idle"},
	}
}

func (inv *Inventory) Update() error {
	if !inv.Open {
		_, h := screenSize()
		inventoryTop = h - windowMargin
		return nil
	}
	w, h := screenSize()
	placeWindow("Inventory", imgui.Vec2{X: w - windowMargin, Y: h - windowMargin}, imgui.Vec2{X: 1, Y: 1})
	imgui.BeginV("Inventory", &inv.Open, imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text(fmt.Sprintf("Coins: %d", inv.Contents.Coins))
	imgui.Separator()
	section("Backpack", inv.Contents.Backpack, inv.Contents.Skills, &inv.picker)
	inventoryTop = imgui.WindowPos().Y
	imgui.End()

	return nil
}

func section(title string, c resources.Inventory, skills []resources.SkillState, picker *stackPicker) {
	imgui.Text(fmt.Sprintf("%s  %d/%d slots", title, len(c.Stacks), c.Slots))
	if len(c.Stacks) == 0 {
		imgui.Text("  Empty")
		return
	}
	// A line per slot: each material stack has one quality.
	for _, i := range slotOrder(c) {
		s := c.Stacks[i]
		picker.checkbox(s, fmt.Sprint(i))
		imgui.Text(stackLabel(s))
		itemTooltip(s.Kind)
		repairButton(s, skills, resources.RepairCommandContents{ID: s.ID})
	}
	picker.combineButton(c, false)
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
