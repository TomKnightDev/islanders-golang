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
	Open   bool // shown; the menu bar toggles it
	mgr    *renderer.Manager
	skills []resources.SkillState
}

func NewSkills(mgr *renderer.Manager) *Skills {
	return &Skills{Open: true, mgr: mgr}
}

func (s *Skills) SetState(skills []resources.SkillState) {
	s.skills = skills
}

func (s *Skills) Update() error {
	if !s.Open {
		return nil
	}
	placeWindow("Skills", imgui.Vec2{X: windowMargin, Y: infoBottom + windowMargin}, imgui.Vec2{})
	imgui.BeginV("Skills", &s.Open, imgui.WindowFlagsAlwaysAutoResize)
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
		centredBar(frac, 360, overlay)
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

// centredBar is a progress bar with its text centred on it; imgui's own
// overlay text follows the fill instead.
func centredBar(frac, width float32, text string) {
	imgui.ProgressBarV(frac, imgui.Vec2{X: width, Y: 0}, "")
	lo, hi := imgui.ItemRectMin(), imgui.ItemRectMax()
	size := imgui.CalcTextSize(text, false, 0)
	imgui.WindowDrawList().AddText(
		imgui.Vec2{X: (lo.X + hi.X - size.X) / 2, Y: (lo.Y + hi.Y - size.Y) / 2},
		imgui.Packed(color.White), text)
}
