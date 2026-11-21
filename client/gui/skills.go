package gui

import (
	"fmt"
	"image/color"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Skills lists the player's skills with their level and progress towards
// the next one.
type Skills struct {
	mgr    *renderer.Manager
	skills []resources.SkillState
}

func NewSkills(mgr *renderer.Manager) *Skills {
	return &Skills{mgr: mgr}
}

func (s *Skills) SetState(skills []resources.SkillState) {
	s.skills = skills
}

func (s *Skills) Update() error {
	imgui.BeginV("Skills", nil, imgui.WindowFlagsAlwaysAutoResize)
	total := 0
	for _, st := range s.skills {
		total += st.Level
		imgui.Text(fmt.Sprintf("%-12s %2d", resources.SkillNames[st.Skill], st.Level))
		imgui.SameLine()
		start := resources.XPForLevel(st.Level)
		frac, overlay := float32(1), fmt.Sprintf("%d XP", st.XP)
		if st.Level < resources.MaxLevel {
			frac = float32(st.XP-start) / float32(st.NextXP-start)
			overlay = fmt.Sprintf("%d / %d XP", st.XP, st.NextXP)
		}
		// The bar's own overlay text follows the fill; draw it centred instead.
		imgui.ProgressBarV(frac, imgui.Vec2{X: 360, Y: 0}, "")
		lo, hi := imgui.ItemRectMin(), imgui.ItemRectMax()
		size := imgui.CalcTextSize(overlay, false, 0)
		imgui.WindowDrawList().AddText(
			imgui.Vec2{X: (lo.X + hi.X - size.X) / 2, Y: (lo.Y + hi.Y - size.Y) / 2},
			imgui.Packed(color.White), overlay)
		if imgui.IsItemHovered() {
			imgui.SetTooltip(fmt.Sprintf("%d XP to level %d", max(0, st.NextXP-st.XP), min(st.Level+1, resources.MaxLevel)))
		}
	}
	dimmed(fmt.Sprintf("Total level %d", total))
	imgui.End()
	return nil
}

func (s *Skills) Draw(screen *ebiten.Image) {
	s.mgr.Draw(screen)
}
