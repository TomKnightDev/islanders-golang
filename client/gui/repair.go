package gui

import (
	"fmt"
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

// combineButton draws a Combine button, on the same line, the first time a
// material with more than one stack is listed. first tracks which kinds
// already have one.
func combineButton(inv resources.Inventory, s resources.Stack, bank bool, first map[resources.ItemKind]bool) {
	if resources.IsPiece(s.Kind) || first[s.Kind] || SendCombine == nil {
		return
	}
	first[s.Kind] = true
	stacks := 0
	for _, o := range inv.Stacks {
		if o.Kind == s.Kind {
			stacks++
		}
	}
	if stacks < 2 {
		return
	}
	imgui.SameLine()
	if imgui.Button(fmt.Sprintf("Combine##combine%s%v", s.Kind, bank)) {
		SendCombine(resources.CombineCommandContents{Item: s.Kind, Bank: bank})
	}
	if imgui.IsItemHovered() {
		imgui.SetTooltip(fmt.Sprintf("Merge all %d stacks into as few as possible, at their average quality (q%d).", stacks, inv.AverageQuality(s.Kind)))
	}
}
