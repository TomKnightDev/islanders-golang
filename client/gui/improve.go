package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Improve is the window at an anvil or the fletcher's hut. It lists the
// pieces in the backpack that can be improved there; picking one shows
// what it needs, and Start works on it an attempt at a time.
type Improve struct {
	mgr   *renderer.Manager
	send  func(resources.ImproveCommandContents)
	state resources.InventoryContents

	open     bool
	shownAt  resources.Building
	selected uint32 // the piece picked in the list
}

func NewImprove(mgr *renderer.Manager, send func(resources.ImproveCommandContents)) *Improve {
	return &Improve{mgr: mgr, send: send}
}

// SetState opens the window on arriving at an anvil or the fletcher's hut
// and closes it on leaving.
func (m *Improve) SetState(c resources.InventoryContents) {
	m.state = c
	at := resources.Building("")
	if c.At == resources.BuildingAnvil || c.At == resources.BuildingFletcher {
		at = c.At
	}
	if at != m.shownAt {
		m.open = at != ""
		m.shownAt = at
	}
	if c.Improving != 0 {
		m.selected = c.Improving
	}
}

func (m *Improve) level(s resources.Skill) int {
	for _, st := range m.state.Skills {
		if st.Skill == s {
			return st.Level
		}
	}
	return 1
}

func (m *Improve) Update() error {
	if !m.open || m.shownAt == "" {
		return nil
	}
	w, h := screenSize()
	placeWindow("Improve", imgui.Vec2{X: w / 2, Y: h / 2}, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.BeginV("Improve", &m.open, imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text(fmt.Sprintf("Improving at the %s %s", m.state.AtVillage, resources.BuildingNames[m.shownAt]))
	imgui.Separator()

	// What can be improved here, from the backpack.
	imgui.BeginGroup()
	imgui.Text("Your backpack")
	var picked *resources.Stack
	any := false
	for i, s := range m.state.Backpack.Stacks {
		if at, ok := resources.ImproveAt(s.Kind); !ok || at != m.shownAt {
			continue
		}
		any = true
		label := fmt.Sprintf("%s q%d##piece%d", resources.ItemNames[s.Kind], s.Quality, i)
		if imgui.SelectableV(label, s.ID == m.selected, 0, imgui.Vec2{X: 260}) && m.state.Improving == 0 {
			m.selected = s.ID
		}
		if s.ID == m.selected {
			picked = &m.state.Backpack.Stacks[i]
		}
	}
	if !any {
		what := "metal gear or tools"
		if m.shownAt == resources.BuildingFletcher {
			what = "bows or staves"
		}
		dimmed("No " + what + " in your backpack.")
	}
	imgui.EndGroup()

	imgui.SameLineV(0, 30)

	imgui.BeginGroup()
	if picked == nil {
		dimmed("Pick something to improve.")
	} else {
		m.details(*picked)
	}
	imgui.EndGroup()

	if m.state.ImproveNote != "" {
		imgui.Separator()
		imgui.PushTextWrapPosV(600)
		dimmed(m.state.ImproveNote)
		imgui.PopTextWrapPos()
	}
	imgui.End()
	return nil
}

// details shows what improving a piece needs and the Start/Stop button.
func (m *Improve) details(piece resources.Stack) {
	kind, recipe, _ := resources.ImproveRecipe(piece.Kind)
	skill := resources.ActionSkill[kind]
	material := recipe.Inputs[0].Item
	imgui.Text(fmt.Sprintf("%s, quality %d", resources.ItemNames[piece.Kind], piece.Quality))

	ready := true
	need := func(ok bool, text string) {
		if ok {
			imgui.Text("  " + text)
		} else {
			dimmed("  " + text + " (missing)")
			ready = false
		}
	}
	lvl := m.level(skill)
	need(lvl >= recipe.Level, fmt.Sprintf("%s level %d (you have %d)", resources.SkillNames[skill], recipe.Level, lvl))

	better, best := 0, 0
	for _, s := range m.state.Backpack.Stacks {
		if s.Kind == material && s.Quality > piece.Quality {
			better += s.Count
			best = max(best, s.Quality)
		}
	}
	matText := fmt.Sprintf("%s better than quality %d: you carry %d", resources.ItemUnits[material], piece.Quality, better)
	if better > 0 {
		matText += fmt.Sprintf(", best q%d", best)
	}
	need(better > 0, matText)

	slot := resources.ActionTool[kind]
	toolText := "A " + resources.SlotNames[slot] + " on your tool belt"
	hasTool := false
	for _, e := range m.state.Equipment {
		if e.Slot == slot {
			hasTool = true
			toolText = fmt.Sprintf("%s q%d, %d%% worn", resources.ItemNames[e.Item], e.Quality, e.Damage)
		}
	}
	need(hasTool, toolText)

	chance := min(85, max(10, 85-2*(piece.Quality-lvl)))
	dimmed(fmt.Sprintf("Success chance about %d%%. Each %s lasts %d attempts;", chance,
		resources.ItemUnits[material], resources.ImproveUnitAttempts))
	dimmed(fmt.Sprintf("each attempt wears the tool %d%%. A failure costs 1 quality.", resources.ImproveToolWear))

	if m.state.Improving == piece.ID {
		if imgui.Button("Stop") {
			m.send(resources.ImproveCommandContents{Op: resources.ImproveOpStop})
		}
	} else if button("Start improving", ready && m.state.Improving == 0) {
		m.send(resources.ImproveCommandContents{Op: resources.ImproveOpStart, ID: piece.ID})
	}
}

func (m *Improve) Draw(screen *ebiten.Image) {}
