package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Equipment shows what the character wears and wields, their combat stats,
// and the gear in the backpack they can equip. Gear in the bank is equipped
// from the bank window.
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
		worn[it.Slot] = resources.Stack{Kind: it.Item, Count: it.Count, Quality: it.Quality, Damage: it.Damage}
	}
	e.slots(worn, resources.EquipSlots)
	imgui.Separator()
	imgui.Text("Tool belt")
	e.slots(worn, resources.ToolSlots)

	// Combat stats from levels and gear.
	b := resources.EquipmentBonuses(worn)
	style := b.Style()
	_, strengthSkill := b.Skills()
	imgui.Separator()
	imgui.Text(fmt.Sprintf("%s  ·  accuracy +%d  ·  max hit %d  ·  armour +%d",
		style, b.Accuracy, resources.MaxHit(e.level(strengthSkill), b.Strength), b.Defence))

	imgui.Separator()
	imgui.Text("Equip from the backpack")
	if !equipButtons(e.state.Backpack, false, e.level, e.send) {
		dimmed("No gear in the backpack. Gear in the bank is equipped at a bank.")
	}
	imgui.End()
	return nil
}

// equipButtons offers a button per piece of gear in an inventory (arrows
// once per kind), greyed out where the player lacks the level. It reports
// whether there was any gear.
func equipButtons(inv resources.Inventory, fromBank bool, level func(resources.Skill) int, send func(resources.EquipCommandContents)) bool {
	any := false
	seenAmmo := map[resources.ItemKind]bool{}
	for i, s := range inv.Stacks {
		def, ok := resources.Equipment[s.Kind]
		if !ok || seenAmmo[s.Kind] {
			continue
		}
		any = true
		label := fmt.Sprintf("%s q%d", resources.ItemNames[s.Kind], s.Quality)
		cmd := resources.EquipCommandContents{Op: resources.EquipOpEquip, Item: s.Kind, ID: s.ID, FromBank: fromBank}
		if def.Slot == resources.SlotAmmo {
			seenAmmo[s.Kind] = true
			label = fmt.Sprintf("%s x%d q%d", resources.ItemNames[s.Kind], inv.Count(s.Kind), inv.AverageQuality(s.Kind))
			cmd.ID = 0
		}
		id := fmt.Sprintf("##equip%v%d", fromBank, i)
		if have := level(def.Skill); have < def.Level {
			button(fmt.Sprintf("%s (%s %d)%s", label, resources.SkillNames[def.Skill], def.Level, id), false)
		} else if button(label+id, true) {
			send(cmd)
		}
		if imgui.IsItemHovered() {
			imgui.SetTooltip(gearTooltip(def, s))
		}
	}
	return any
}

// slots draws a list of slots, each with what's in it and a Remove button.
func (e *Equipment) slots(worn map[resources.EquipSlot]resources.Stack, slots []resources.EquipSlot) {
	for _, slot := range slots {
		imgui.PushID(string(slot))
		imgui.Text(fmt.Sprintf("%-8s", resources.SlotNames[slot]))
		imgui.SameLine()
		if s, ok := worn[slot]; ok {
			name := fmt.Sprintf("%s q%d%s", resources.ItemNames[s.Kind], s.Quality, wornText(s))
			if slot == resources.SlotAmmo {
				name = fmt.Sprintf("%s x%d q%d", resources.ItemNames[s.Kind], s.Count, s.Quality)
			}
			imgui.Text(name)
			if imgui.IsItemHovered() {
				imgui.SetTooltip(gearTooltip(resources.Equipment[s.Kind], s))
			}
			imgui.SameLine()
			if imgui.Button("Remove") {
				e.send(resources.EquipCommandContents{Op: resources.EquipOpUnequip, Slot: slot})
			}
			repairButton(s, e.state.Skills, resources.RepairCommandContents{Slot: slot})
		} else if resources.ActionToolSlot(slot) {
			dimmed("Empty: that work goes at half speed")
		} else {
			dimmed("Empty")
		}
		imgui.PopID()
	}
}

// gearTooltip describes a piece at its quality: bonuses after quality and
// damage, or a tool's speed.
func gearTooltip(d resources.EquipDef, s resources.Stack) string {
	f := resources.GearFactor(resources.EffectiveQuality(s.Quality, s.Damage))
	scaled := func(v int) int { return int(float64(v)*f + 0.5) }
	d.Accuracy, d.Strength, d.Defence = scaled(d.Accuracy), scaled(d.Strength), scaled(d.Defence)
	switch {
	case resources.ActionToolSlot(d.Slot):
		return fmt.Sprintf("Tool, quality %d: each step takes %d%% of the usual time (200%% with no tool). Better tools also gather better materials.",
			s.Quality, int(resources.ToolSpeed(&s)*100+0.5))
	case d.Slot == resources.SlotAmmo:
		return fmt.Sprintf("Ranged strength +%d", d.Strength)
	case d.Drain > 0:
		return fmt.Sprintf("Necromancer staff, two-handed. Accuracy +%d, strength +%d, heals %d%% of damage", d.Accuracy, d.Strength, d.Drain)
	case d.Magic:
		return fmt.Sprintf("Mage staff, two-handed. Accuracy +%d, strength +%d", d.Accuracy, d.Strength)
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
