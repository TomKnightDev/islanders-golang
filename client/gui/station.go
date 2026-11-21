package gui

import (
	"fmt"
	"slices"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Station is the window at an anvil or the fletcher's hut, with two tabs:
// Make gear from what's in the backpack, or Improve a piece in it.
type Station struct {
	mgr   *renderer.Manager
	send  func(resources.StationCommandContents)
	state resources.InventoryContents

	open     bool
	shownAt  resources.Building
	recipe   resources.ItemKind // the recipe picked on the Make tab
	selected uint32             // the piece picked on the Improve tab
}

func NewStation(mgr *renderer.Manager, send func(resources.StationCommandContents)) *Station {
	return &Station{mgr: mgr, send: send}
}

// SetState opens the window on arriving at an anvil or the fletcher's hut
// and closes it on leaving.
func (m *Station) SetState(c resources.InventoryContents) {
	m.state = c
	at := resources.Building("")
	if c.At == resources.BuildingAnvil || c.At == resources.BuildingFletcher {
		at = c.At
	}
	if at != m.shownAt {
		m.open = at != ""
		m.shownAt = at
		m.recipe, m.selected = "", 0
	}
	if c.Improving != 0 {
		m.selected = c.Improving
	}
	if c.Making != "" {
		m.recipe = c.Making
	}
}

func (m *Station) level(s resources.Skill) int {
	for _, st := range m.state.Skills {
		if st.Skill == s {
			return st.Level
		}
	}
	return 1
}

func (m *Station) busy() bool { return m.state.Making != "" || m.state.Improving != 0 }

func (m *Station) Update() error {
	if !m.open || m.shownAt == "" {
		return nil
	}
	w, h := screenSize()
	title := "Anvil"
	if m.shownAt == resources.BuildingFletcher {
		title = "Fletcher"
	}
	placeWindow(title, imgui.Vec2{X: w / 2, Y: h / 2}, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.BeginV(title, &m.open, imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text(fmt.Sprintf("The %s %s", m.state.AtVillage, resources.BuildingNames[m.shownAt]))
	if imgui.BeginTabBar("station") {
		if imgui.BeginTabItem("Make") {
			m.makeTab()
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Improve") {
			m.improveTab()
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}
	if m.state.StationNote != "" {
		imgui.Separator()
		imgui.PushTextWrapPosV(600)
		dimmed(m.state.StationNote)
		imgui.PopTextWrapPos()
	}
	imgui.End()
	return nil
}

// --- Make ---

// makeTab lists the recipes made here that the player has the level for,
// a row per metal or wood; picking one shows what it needs.
func (m *Station) makeTab() {
	imgui.BeginGroup()
	tier, any := "", false
	for _, r := range resources.GearRecipes {
		if r.At != m.shownAt || m.level(r.Skill) < r.Level {
			continue
		}
		any = true
		if r.Tier != tier {
			tier = r.Tier
			dimmed(tier + ":")
		} else {
			imgui.SameLine()
		}
		// Sized to its text: a selectable fills the row by default, so the
		// ones after it on the same line would sit on top of it.
		size := imgui.CalcTextSize(r.Short, false, 0)
		if imgui.SelectableV(r.Short+"##make"+r.ID, r.Item == m.recipe, 0, imgui.Vec2{X: size.X}) && !m.busy() {
			m.recipe = r.Item
		}
		if imgui.IsItemHovered() {
			imgui.SetTooltip(r.Label)
		}
	}
	if !any {
		dimmed("Nothing you can make here yet.")
	}
	if next, ok := m.nextUnlock(); ok {
		dimmed(fmt.Sprintf("Next: %s at %s level %d. The Guide lists every recipe.", next.Label, resources.SkillNames[next.Skill], next.Level))
	}
	imgui.EndGroup()

	r, ok := resources.FindGearRecipe(m.recipe)
	if !ok || r.At != m.shownAt {
		imgui.Separator()
		dimmed("Pick something to make.")
		return
	}
	imgui.Separator()
	imgui.Text(fmt.Sprintf("%s (%s level %d)", resources.ItemNames[r.Item], resources.SkillNames[r.Skill], r.Level))
	crafts := resources.MaxGoalAmount
	for _, in := range r.Inputs {
		have := m.state.Backpack.Count(in.Item)
		crafts = min(crafts, have/in.Count)
		text := fmt.Sprintf("  %d %s: you carry %d", in.Count, resources.ItemUnits[in.Item], have)
		if have >= in.Count {
			imgui.Text(text)
		} else {
			dimmed(text + " (missing)")
		}
	}
	m.toolLine(r, false)
	dimmed("New gear comes out rough: about half your level, never better than its materials.")

	if m.state.Making == r.Item {
		imgui.Text(fmt.Sprintf("Making: %d left", m.state.MakeLeft))
		imgui.SameLine()
		if imgui.Button("Stop##make") {
			m.send(resources.StationCommandContents{Op: resources.StationOpStop})
		}
		return
	}
	for _, n := range []int{1, 5} {
		if button(fmt.Sprintf("Make %d", n), crafts >= n && !m.busy()) {
			m.send(resources.StationCommandContents{Op: resources.StationOpMake, Item: r.Item, Count: n})
		}
		imgui.SameLine()
	}
	if button(fmt.Sprintf("Make all (%d)", crafts), crafts > 0 && !m.busy()) {
		m.send(resources.StationCommandContents{Op: resources.StationOpMake, Item: r.Item, Count: crafts})
	}
}

// nextUnlock is the lowest-level recipe here the player can't make yet.
func (m *Station) nextUnlock() (resources.GearRecipe, bool) {
	var next resources.GearRecipe
	found := false
	for _, r := range resources.GearRecipes {
		if r.At == m.shownAt && m.level(r.Skill) < r.Level && (!found || r.Level < next.Level) {
			next, found = r, true
		}
	}
	return next, found
}

// toolLine shows the recipe's tool on the belt, or that it's missing.
func (m *Station) toolLine(r resources.GearRecipe, required bool) bool {
	for _, e := range m.state.Equipment {
		if e.Slot == r.Tool {
			imgui.Text(fmt.Sprintf("  %s q%d, %d%% worn", resources.ItemNames[e.Item], e.Quality, e.Damage))
			return true
		}
	}
	if required {
		dimmed("  A " + resources.SlotNames[r.Tool] + " on your tool belt (missing)")
	} else {
		dimmed("  No " + resources.SlotNames[r.Tool] + " on your belt: half speed and rougher work")
	}
	return false
}

// --- Improve ---

// improveTab lists the pieces in the backpack that can be improved here;
// picking one shows what it needs.
func (m *Station) improveTab() {
	imgui.BeginGroup()
	var picked *resources.Stack
	any := false
	for i, s := range m.state.Backpack.Stacks {
		if r, ok := resources.FindGearRecipe(s.Kind); !ok || r.At != m.shownAt {
			continue
		}
		any = true
		label := fmt.Sprintf("%s q%d##piece%d", resources.ItemNames[s.Kind], s.Quality, i)
		if imgui.SelectableV(label, s.ID == m.selected, 0, imgui.Vec2{X: 260}) && !m.busy() {
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
		m.improveDetails(*picked)
	}
	imgui.EndGroup()
}

func (m *Station) improveDetails(piece resources.Stack) {
	r, _ := resources.FindGearRecipe(piece.Kind)
	material := r.Inputs[0].Item
	imgui.Text(fmt.Sprintf("%s, quality %d", resources.ItemNames[piece.Kind], piece.Quality))

	ready := true
	lvl := m.level(r.Skill)
	if text := fmt.Sprintf("  %s level %d (you have %d)", resources.SkillNames[r.Skill], r.Level, lvl); lvl >= r.Level {
		imgui.Text(text)
	} else {
		dimmed(text + " (missing)")
		ready = false
	}

	var better []int
	for _, s := range m.state.Backpack.Stacks {
		if s.Kind == material && s.Quality > piece.Quality {
			for i := 0; i < s.Count; i++ {
				better = append(better, s.Quality)
			}
		}
	}
	matText := fmt.Sprintf("  %s better than quality %d: you carry %d", resources.ItemUnits[material], piece.Quality, len(better))
	if len(better) > 0 {
		imgui.Text(matText + fmt.Sprintf(", best q%d", slices.Max(better)))
	} else {
		dimmed(matText + " (missing)")
		ready = false
	}
	if !m.toolLine(r, true) {
		ready = false
	}

	chance := min(85, max(10, 85-2*(piece.Quality-lvl)))
	dimmed(fmt.Sprintf("Success chance about %d%%. Each %s lasts %d attempts;", chance,
		resources.ItemUnits[material], resources.ImproveUnitAttempts))
	dimmed(fmt.Sprintf("each attempt wears the tool %d%%. A failure costs 1 quality.", resources.ImproveToolWear))

	if m.state.Improving == piece.ID {
		if imgui.Button("Stop##improve") {
			m.send(resources.StationCommandContents{Op: resources.StationOpStop})
		}
	} else if button("Start improving", ready && !m.busy()) {
		m.send(resources.StationCommandContents{Op: resources.StationOpImprove, ID: piece.ID})
	}
}

func (m *Station) Draw(screen *ebiten.Image) {}
