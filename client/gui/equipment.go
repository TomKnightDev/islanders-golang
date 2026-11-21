package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Equipment shows what the character wears and wields, their combat stats,
// and the gear in the bank they can equip.
type Equipment struct {
	mgr   *renderer.Manager
	send  func(resources.EquipCommandContents)
	state resources.InventoryContents
}

func NewEquipment(mgr *renderer.Manager, send func(resources.EquipCommandContents)) *Equipment {
	return &Equipment{mgr: mgr, send: send}
}

func (e *Equipment) SetState(c resources.InventoryContents) { e.state = c }

func (e *Equipment) level(s resources.Skill) int {
	for _, st := range e.state.Skills {
		if st.Skill == s {
			return st.Level
		}
	}
	return 1
}

func (e *Equipment) Update() error {
	w, _ := screenSize()
	placeWindow("Equipment", imgui.Vec2{X: w - windowMargin, Y: inventoryTop - windowMargin}, imgui.Vec2{X: 1, Y: 1})
	imgui.BeginV("Equipment", nil, imgui.WindowFlagsAlwaysAutoResize)

	worn := map[resources.EquipSlot]resources.Stack{}
	for _, it := range e.state.Equipment {
		worn[it.Slot] = resources.Stack{Kind: it.Item, Count: it.Count}
	}
	for _, slot := range resources.EquipSlots {
		imgui.PushID(string(slot))
		imgui.Text(fmt.Sprintf("%-7s", resources.SlotNames[slot]))
		imgui.SameLine()
		if s, ok := worn[slot]; ok {
			name := resources.ItemNames[s.Kind]
			if slot == resources.SlotAmmo {
				name = fmt.Sprintf("%s x%d", name, s.Count)
			}
			imgui.Text(name)
			imgui.SameLine()
			if imgui.Button("Remove") {
				e.send(resources.EquipCommandContents{Op: resources.EquipOpUnequip, Slot: slot})
			}
		} else {
			dimmed("Empty")
		}
		imgui.PopID()
	}

	// Combat stats from levels and gear.
	b := resources.EquipmentBonuses(worn)
	style, strengthSkill := "Melee", resources.SkillStrength
	if b.Ranged {
		style, strengthSkill = "Ranged", resources.SkillRanged
	}
	imgui.Separator()
	imgui.Text(fmt.Sprintf("%s  ·  accuracy +%d  ·  max hit %d  ·  armour +%d",
		style, b.Accuracy, resources.MaxHit(e.level(strengthSkill), b.Strength), b.Defence))

	// Gear waiting in the bank.
	imgui.Separator()
	imgui.Text("Equip from the bank")
	any := false
	seen := map[resources.ItemKind]bool{}
	for _, s := range e.state.Bank.Stacks {
		def, ok := resources.Equipment[s.Kind]
		if !ok || seen[s.Kind] {
			continue
		}
		seen[s.Kind] = true
		any = true
		label := resources.ItemNames[s.Kind]
		if def.Slot == resources.SlotAmmo {
			label = fmt.Sprintf("%s x%d", label, e.state.Bank.Count(s.Kind))
		}
		if have := e.level(def.Skill); have < def.Level {
			button(fmt.Sprintf("%s (%s %d)##%s", label, resources.SkillNames[def.Skill], def.Level, s.Kind), false)
		} else if button(label+"##"+string(s.Kind), true) {
			e.send(resources.EquipCommandContents{Op: resources.EquipOpEquip, Item: s.Kind})
		}
		if imgui.IsItemHovered() {
			imgui.SetTooltip(gearTooltip(def))
		}
	}
	if !any {
		dimmed("No gear in the bank. Camps drop it.")
	}
	imgui.End()
	return nil
}

func gearTooltip(d resources.EquipDef) string {
	switch {
	case d.Slot == resources.SlotAmmo:
		return fmt.Sprintf("Ranged strength +%d", d.Strength)
	case d.Ranged:
		return fmt.Sprintf("Bow, two-handed. Accuracy +%d", d.Accuracy)
	case d.Slot == resources.SlotWeapon:
		return fmt.Sprintf("Accuracy +%d, strength +%d", d.Accuracy, d.Strength)
	default:
		return fmt.Sprintf("Armour +%d", d.Defence)
	}
}

func (e *Equipment) Draw(screen *ebiten.Image) {
	e.mgr.Draw(screen)
}
