package gui

import (
	"fmt"

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
