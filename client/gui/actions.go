package gui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/client/entities"
	"github.com/tomknightdev/islanders-golang/resources"
)

var dimText = imgui.Vec4{X: 0.7, Y: 0.75, Z: 0.7, W: 1}

// Actions is the action window: places the player has found and what can be
// done there, and the action list the character works through.
type Actions struct {
	mgr    *renderer.Manager
	player *entities.Player
	places map[string]resources.Place
	// Send queues a command for the server.
	Send func(resources.ActionCommandContents)

	state   resources.InventoryContents
	picking bool // the "Add an action" window is open
}

func NewActions(mgr *renderer.Manager, player *entities.Player, places []resources.Place) *Actions {
	a := &Actions{mgr: mgr, player: player, places: map[string]resources.Place{}}
	for _, p := range places {
		a.places[p.Name] = p
	}
	return a
}

// SetState takes the latest player state from the server.
func (a *Actions) SetState(c resources.InventoryContents) { a.state = c }

func (a *Actions) send(c resources.ActionCommandContents) {
	if a.Send != nil {
		a.Send(c)
	}
}

func (a *Actions) Update() error {
	w, _ := screenSize()
	placeWindow("Actions", imgui.Vec2{X: w - windowMargin, Y: windowMargin}, imgui.Vec2{X: 1, Y: 0})
	imgui.BeginV("Actions", nil, imgui.WindowFlagsAlwaysAutoResize)
	a.controls()
	imgui.Separator()
	a.list()
	actionsLeft = imgui.WindowPos().X
	imgui.End()

	if a.picking {
		// Opens beside the actions window each time.
		imgui.SetNextWindowPosV(imgui.Vec2{X: actionsLeft - windowMargin, Y: windowMargin}, imgui.ConditionAppearing, imgui.Vec2{X: 1, Y: 0})
		imgui.BeginV("Add an action", &a.picking, imgui.WindowFlagsAlwaysAutoResize)
		a.foundPlaces()
		imgui.End()
	}
	return nil
}

func (a *Actions) controls() {
	n := len(a.state.Actions)
	label := "Add action"
	if a.picking {
		label = "Close action picker"
	}
	if button(label, n < resources.MaxActions || a.picking) {
		a.picking = !a.picking
	}
	imgui.SameLine()
	if a.state.Running {
		if imgui.Button("Stop") {
			a.send(resources.ActionCommandContents{Op: resources.ActionOpStop})
		}
		imgui.SameLine()
		dimmed("Working through the list")
		return
	}
	if button("Start", n > 0) {
		a.send(resources.ActionCommandContents{Op: resources.ActionOpStart})
	}
	imgui.SameLine()
	if n == 0 {
		dimmed("Add an action to get started")
	} else {
		dimmed("Stopped. Click the map to walk around.")
	}
}

// level is the player's level in a skill, from the latest server state.
func (a *Actions) level(s resources.Skill) int {
	for _, st := range a.state.Skills {
		if st.Skill == s {
			return st.Level
		}
	}
	return 1
}

// nearestFound is the closest found place of a kind to the player.
func (a *Actions) nearestFound(kind resources.PlaceKind) (resources.Place, int, bool) {
	var best resources.Place
	bestDist, ok := 0, false
	px, py := 0, 0
	if a.player != nil {
		px, py = resources.TileAt(a.player.Position[0], a.player.Position[1])
	}
	for _, name := range a.state.Found {
		pl, known := a.places[name]
		if !known || pl.Kind != kind {
			continue
		}
		if d := max(abs(pl.X-px), abs(pl.Y-py)); !ok || d < bestDist {
			best, bestDist, ok = pl, d, true
		}
	}
	return best, bestDist, ok
}

// foundPlaces offers each action once, at the nearest place found for it,
// with a button per material. Actions with no place found, and materials
// the player hasn't the level for, aren't shown.
func (a *Actions) foundPlaces() {
	full := len(a.state.Actions) >= resources.MaxActions
	if full {
		dimmed(fmt.Sprintf("The list is full (%d actions). Remove one to add more.", resources.MaxActions))
	} else {
		dimmed("Click to add to the end of the list.")
	}
	// Only what the player can do now: places they've found and materials
	// they have the level for.
	for _, kind := range resources.ActionKinds {
		place, dist, ok := a.nearestFound(resources.ActionPlaceKind[kind])
		if !ok {
			continue
		}
		mats := a.doable(kind, place)
		if !resources.IsVisit(kind) && len(mats) == 0 {
			continue
		}
		imgui.Text(fmt.Sprintf("%s at %s", resources.ActionNames[kind], place.Name))
		imgui.SameLine()
		dimmed(fmt.Sprintf("%d tiles away", dist))
		if resources.IsVisit(kind) {
			a.visitButtons(place, full)
		} else {
			a.materialButtons(kind, place, mats, full)
		}
	}
	dimmed("Walk into a village, mine, forest, farm or camp to find more.")
}

// doable is the materials of an action the player can use at a place, with
// recipes grouped into rows by tier, in the order tiers first appear.
func (a *Actions) doable(kind resources.ActionKind, place resources.Place) []resources.Material {
	lvl := a.level(resources.ActionSkill[kind])
	var tiers []string
	for _, m := range resources.ActionMaterials[kind] {
		if !slices.Contains(tiers, m.Tier) {
			tiers = append(tiers, m.Tier)
		}
	}
	var mats []resources.Material
	for _, tr := range tiers {
		for _, m := range resources.ActionMaterials[kind] {
			if m.Tier != tr || lvl < m.Level {
				continue
			}
			if (resources.IsGathering(kind) || resources.IsCombat(kind)) && !slices.Contains(place.Materials, m.ID) {
				continue
			}
			mats = append(mats, m)
		}
	}
	return mats
}

// materialButtons offers a button per material. Recipes with tiers (metals,
// woods) get a row per tier with short labels.
func (a *Actions) materialButtons(kind resources.ActionKind, place resources.Place, mats []resources.Material, full bool) {
	first, tier := true, ""
	for _, m := range mats {
		label := strings.ToUpper(m.Label[:1]) + m.Label[1:]
		switch {
		case m.Tier != "" && m.Tier != tier:
			tier = m.Tier
			dimmed("  " + tier + ":")
			imgui.SameLine()
		case !first:
			imgui.SameLine()
		}
		if m.Tier != "" {
			label = m.Short
		}
		first = false
		if e, ok := resources.FindEnemy(m.ID); ok && resources.IsCombat(kind) {
			label = fmt.Sprintf("%s (level %d)", label, e.Level)
		}
		if button(label+"##"+string(kind)+m.ID, !full) {
			a.send(resources.ActionCommandContents{Op: resources.ActionOpAdd, Action: resources.ActionSpec{
				Kind: kind, Material: m.ID, Place: place.Name, Amount: resources.GoalAmounts[0],
			}})
		}
		if imgui.IsItemHovered() && len(m.Inputs) > 0 {
			imgui.SetTooltip(recipeTooltip(m))
		}
	}
}

// recipeTooltip says what a craft uses and makes.
func recipeTooltip(m resources.Material) string {
	var uses []string
	for _, in := range m.Inputs {
		uses = append(uses, fmt.Sprintf("%d %s", in.Count, resources.ItemUnits[in.Item]))
	}
	return fmt.Sprintf("%s: uses %s, makes %d", m.Label, strings.Join(uses, " and "), m.Output())
}

// visitButtons offers each building in the village to walk to; its window
// opens on arrival.
func (a *Actions) visitButtons(place resources.Place, full bool) {
	for i, b := range resources.Buildings {
		if i > 0 {
			imgui.SameLine()
		}
		name := resources.BuildingNames[b]
		if button(strings.ToUpper(name[:1])+name[1:]+"##visit", !full) {
			a.send(resources.ActionCommandContents{Op: resources.ActionOpAdd, Action: resources.ActionSpec{
				Kind: resources.ActionVisit, Material: string(b), Place: place.Name, Amount: 1,
			}})
		}
	}
}

func (a *Actions) list() {
	imgui.Text(fmt.Sprintf("Action list  %d/%d", len(a.state.Actions), resources.MaxActions))
	if len(a.state.Actions) == 0 {
		dimmed("Empty")
		return
	}
	for i, q := range a.state.Actions {
		imgui.PushIDInt(i)
		a.row(i, q)
		imgui.PopID()
		if i < len(a.state.Actions)-1 {
			imgui.Spacing()
		}
	}
}

func (a *Actions) row(i int, q resources.QueuedAction) {
	spec := q.ActionSpec
	last := len(a.state.Actions) - 1

	imgui.Text(fmt.Sprintf("%d. %s at %s", i+1, spec.Label(), spec.Place))
	imgui.SameLine()
	if button("Up", i > 0) {
		a.send(resources.ActionCommandContents{Op: resources.ActionOpUp, Index: i})
	}
	imgui.SameLine()
	if button("Down", i < last) {
		a.send(resources.ActionCommandContents{Op: resources.ActionOpDown, Index: i})
	}
	imgui.SameLine()
	if imgui.Button("Remove") {
		a.send(resources.ActionCommandContents{Op: resources.ActionOpRemove, Index: i})
	}

	if resources.IsVisit(spec.Kind) {
		dimmed("Walks there, then stops so you can use it. Start carries on.")
		return
	}
	verb, unit := "Gather", resources.ItemUnits[spec.Product()]
	switch {
	case resources.IsCombat(spec.Kind):
		verb, unit = "Defeat", "enemies"
	case resources.ActionSkill[spec.Kind] == resources.SkillSmithing || resources.ActionSkill[spec.Kind] == resources.SkillFletching:
		verb = "Make"
	}

	// Goal: how many, from a fixed set of amounts.
	imgui.Text(verb)
	for _, n := range resources.GoalAmounts {
		imgui.SameLine()
		label := fmt.Sprint(n)
		if n == spec.Amount {
			label = fmt.Sprintf("[%d]", n)
		}
		if button(fmt.Sprintf("%s##goal%d", label, n), n != spec.Amount) {
			edited := spec
			edited.Amount = n
			a.send(resources.ActionCommandContents{Op: resources.ActionOpEdit, Index: i, Action: edited})
		}
	}
	imgui.SameLine()
	imgui.Text(unit)

	progress := fmt.Sprintf("Done: %d/%d", min(q.Done, spec.Amount), spec.Amount)
	if i == 0 && a.state.Running {
		progress = "Now: " + progress
	}
	dimmed(progress)
}

// button is an imgui button that can be shown greyed out. This imgui
// version leaves its disabled-widget API out, so a disabled button is
// drawn dimmed and its clicks are ignored.
func button(label string, enabled bool) bool {
	if enabled {
		return imgui.Button(label)
	}
	imgui.PushStyleColor(imgui.StyleColorText, dimText)
	imgui.Button(label)
	imgui.PopStyleColor()
	return false
}

// dimmed draws secondary text in a softer colour.
func dimmed(text string) {
	imgui.PushStyleColor(imgui.StyleColorText, dimText)
	imgui.Text(text)
	imgui.PopStyleColor()
}

func (a *Actions) Draw(screen *ebiten.Image) {
	a.mgr.Draw(screen)
}
