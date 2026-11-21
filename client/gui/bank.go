package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Bank is the window for using a bank: it opens when the character stops at
// one with the action list stopped, and shows the backpack beside the bank
// so items can move between them and gear can be equipped from either.
type Bank struct {
	mgr       *renderer.Manager
	sendBank  func(resources.BankCommandContents)
	sendEquip func(resources.EquipCommandContents)
	state     resources.InventoryContents
	open      bool
	shownAt   string // the bank the window was last opened for
}

func NewBank(mgr *renderer.Manager, sendBank func(resources.BankCommandContents), sendEquip func(resources.EquipCommandContents)) *Bank {
	return &Bank{mgr: mgr, sendBank: sendBank, sendEquip: sendEquip}
}

// SetState opens the window on arriving at a bank and closes it on leaving.
func (b *Bank) SetState(c resources.InventoryContents) {
	b.state = c
	at := ""
	if c.At == resources.BuildingBank {
		at = c.AtVillage
	}
	if at != b.shownAt {
		b.open = at != ""
		b.shownAt = at
	}
}

func (b *Bank) level(s resources.Skill) int {
	for _, st := range b.state.Skills {
		if st.Skill == s {
			return st.Level
		}
	}
	return 1
}

func (b *Bank) Update() error {
	if !b.open || b.shownAt == "" {
		return nil
	}
	w, h := screenSize()
	placeWindow("Bank", imgui.Vec2{X: w / 2, Y: h / 2}, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.BeginV("Bank", &b.open, imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text(fmt.Sprintf("%s bank  ·  Coins: %d", b.shownAt, b.state.Coins))
	imgui.Separator()

	imgui.BeginGroup()
	imgui.Text(fmt.Sprintf("Backpack  %d/%d", len(b.state.Backpack.Stacks), b.state.Backpack.Slots))
	if button("Deposit all", !b.state.Backpack.Empty()) {
		b.sendBank(resources.BankCommandContents{Op: resources.BankOpDepositAll})
	}
	b.items(b.state.Backpack, resources.BankOpDeposit, "Deposit")
	imgui.EndGroup()

	imgui.SameLineV(0, 40)

	imgui.BeginGroup()
	imgui.Text(fmt.Sprintf("Bank  %d/%d", len(b.state.Bank.Stacks), b.state.Bank.Slots))
	b.items(b.state.Bank, resources.BankOpWithdraw, "Take")
	imgui.EndGroup()

	imgui.Separator()
	imgui.Text("Equip from the bank")
	if !equipButtons(b.state.Bank, true, b.level, b.sendEquip) {
		dimmed("No gear in the bank.")
	}
	imgui.End()
	return nil
}

// items lists an inventory with buttons to move things out of it: gear and
// tools piece by piece, materials as 1 or all.
func (b *Bank) items(inv resources.Inventory, op, verb string) {
	if inv.Empty() {
		dimmed("  Empty")
		return
	}
	seen := map[resources.ItemKind]bool{}
	for i, s := range inv.Stacks {
		imgui.PushID(fmt.Sprintf("%s%d", op, i))
		if resources.IsPiece(s.Kind) {
			imgui.Text(fmt.Sprintf("%s q%d", resources.ItemNames[s.Kind], s.Quality))
			itemTooltip(s.Kind)
			imgui.SameLine()
			if imgui.Button(verb) {
				b.sendBank(resources.BankCommandContents{Op: op, ID: s.ID})
			}
		} else if !seen[s.Kind] {
			seen[s.Kind] = true
			n := inv.Count(s.Kind)
			imgui.Text(fmt.Sprintf("%s %d q%d", resources.ItemNames[s.Kind], n, inv.AverageQuality(s.Kind)))
			itemTooltip(s.Kind)
			imgui.SameLine()
			if imgui.Button(verb + " 1") {
				b.sendBank(resources.BankCommandContents{Op: op, Item: s.Kind, Count: 1})
			}
			imgui.SameLine()
			if imgui.Button(verb + " all") {
				b.sendBank(resources.BankCommandContents{Op: op, Item: s.Kind, Count: n})
			}
		}
		imgui.PopID()
	}
}

func (b *Bank) Draw(screen *ebiten.Image) {}
