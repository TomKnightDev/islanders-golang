package gui

import (
	"fmt"
	"slices"
	"sort"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// SendRepair sends a repair command; the client sets it once connected.
var SendRepair func(resources.RepairCommandContents)

// skillLevel is a level from the server's skill list.
func skillLevel(skills []resources.SkillState, s resources.Skill) int {
	for _, st := range skills {
		if st.Skill == s {
			return st.Level
		}
	}
	return 1
}

// wornText is " (40% worn)" for a worn piece, or "".
func wornText(s resources.Stack) string {
	if s.Damage == 0 {
		return ""
	}
	return fmt.Sprintf(" (%d%% worn)", s.Damage)
}

// repairButton draws a Repair button after a worn piece, on the same line.
// It repairs whatever cmd names: an equipment slot or a piece ID.
func repairButton(s resources.Stack, skills []resources.SkillState, cmd resources.RepairCommandContents) {
	if s.Damage == 0 || SendRepair == nil {
		return
	}
	imgui.SameLine()
	if imgui.Button(fmt.Sprintf("Repair##repair%s%d", cmd.Slot, cmd.ID)) {
		SendRepair(cmd)
	}
	if imgui.IsItemHovered() {
		loss := resources.RepairLoss(s.Damage, skillLevel(skills, resources.SkillRepair))
		imgui.SetTooltip(fmt.Sprintf("Mend it back to 0%% wear, for %d quality (q%d → q%d). Trains Repair.", loss, s.Quality, max(1, s.Quality-loss)))
	}
}

// SendCombine sends a combine command; the client sets it once connected.
var SendCombine func(resources.CombineCommandContents)

// slotOrder is the order to list an inventory's slots in: grouped by item
// kind in the order kinds first appear, best quality first within a kind.
func slotOrder(inv resources.Inventory) []int {
	var kinds []resources.ItemKind
	byKind := map[resources.ItemKind][]int{}
	for i, s := range inv.Stacks {
		if _, ok := byKind[s.Kind]; !ok {
			kinds = append(kinds, s.Kind)
		}
		byKind[s.Kind] = append(byKind[s.Kind], i)
	}
	var order []int
	for _, k := range kinds {
		idx := byKind[k]
		sort.SliceStable(idx, func(a, b int) bool { return inv.Stacks[idx[a]].Quality > inv.Stacks[idx[b]].Quality })
		order = append(order, idx...)
	}
	return order
}

// stackLabel reads "Copper ore x3 q23" for a material, or "Iron sword q40
// (20% worn)" for a piece.
func stackLabel(s resources.Stack) string {
	if resources.IsPiece(s.Kind) {
		return fmt.Sprintf("%s q%d%s", resources.ItemNames[s.Kind], s.Quality, wornText(s))
	}
	return fmt.Sprintf("%s x%d q%d", resources.ItemNames[s.Kind], s.Count, s.Quality)
}

// stackPicker is the checkboxes next to an inventory's material stacks and
// the Combine selected button beneath them. Stacks are picked by kind and
// quality, which is what tells them apart, so a pick survives the list
// changing.
type stackPicker struct {
	picked map[resources.Stack]bool // Kind and Quality only
}

func pickKey(s resources.Stack) resources.Stack {
	return resources.Stack{Kind: s.Kind, Quality: s.Quality}
}

// checkbox draws a stack's checkbox at the start of its line; pieces have
// none.
func (sp *stackPicker) checkbox(s resources.Stack, id string) {
	if resources.IsPiece(s.Kind) {
		imgui.Dummy(imgui.Vec2{X: imgui.FrameHeight()})
		imgui.SameLine()
		return
	}
	if sp.picked == nil {
		sp.picked = map[resources.Stack]bool{}
	}
	on := sp.picked[pickKey(s)]
	if imgui.Checkbox("##pick"+id, &on) {
		sp.picked[pickKey(s)] = on
	}
	if imgui.IsItemHovered() {
		imgui.SetTooltip("Pick stacks of one material, then Combine selected.")
	}
	imgui.SameLine()
}

// combineButton draws Combine selected for the picked stacks still in inv.
// It's enabled for two or more stacks of one material.
func (sp *stackPicker) combineButton(inv resources.Inventory, bank bool) {
	var kind resources.ItemKind
	var qualities []int
	stacks, total, sum, mixed := 0, 0, 0, false
	present := map[resources.Stack]bool{}
	for _, s := range inv.Stacks {
		key := pickKey(s)
		present[key] = true
		if !sp.picked[key] {
			continue
		}
		if kind != "" && s.Kind != kind {
			mixed = true
		}
		kind = s.Kind
		if !slices.Contains(qualities, s.Quality) {
			qualities = append(qualities, s.Quality)
		}
		stacks++
		total += s.Count
		sum += s.Quality * s.Count
	}
	for key := range sp.picked {
		if !present[key] {
			delete(sp.picked, key) // gone from the inventory
		}
	}
	if stacks == 0 {
		return
	}
	label := fmt.Sprintf("Combine selected (%d)##combine%v", stacks, bank)
	if button(label, !mixed && stacks >= 2 && SendCombine != nil) {
		SendCombine(resources.CombineCommandContents{Item: kind, Qualities: qualities, Bank: bank})
		sp.picked = map[resources.Stack]bool{}
	}
	if imgui.IsItemHovered() {
		switch {
		case mixed:
			imgui.SetTooltip("Only stacks of the same material can be combined.")
		case stacks < 2:
			imgui.SetTooltip("Pick at least two stacks.")
		default:
			imgui.SetTooltip(fmt.Sprintf("Merge %d %s into as few slots as possible, at q%d.", total, resources.ItemUnits[kind], sum/total))
		}
	}
	imgui.SameLine()
	if imgui.Button(fmt.Sprintf("Clear##clear%v", bank)) {
		sp.picked = map[resources.Stack]bool{}
	}
}
