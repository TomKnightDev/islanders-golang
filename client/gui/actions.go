package gui

import (
	"fmt"
	"time"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// defaultMinutes is the time goal a new action starts with; amounts come
// from resources.ActionDefaultAmounts. The player can change both.
const defaultMinutes = 10

var dimText = imgui.Vec4{X: 0.7, Y: 0.75, Z: 0.7, W: 1}

// Actions is the action window: places the player has found and what can be
// done there, and the action list the character works through.
type Actions struct {
	mgr    *renderer.Manager
	places map[string]resources.Place
	// Send queues a command for the server.
	Send func(resources.ActionCommandContents)

	state    resources.InventoryContents
	received time.Time

	// Goal inputs being typed into keep their own value until the player
	// moves on, so server updates don't overwrite them mid-edit.
	amounts [resources.MaxActions]int32
	minutes [resources.MaxActions]int32
	editing [resources.MaxActions][2]bool
}

func NewActions(mgr *renderer.Manager, places []resources.Place) *Actions {
	a := &Actions{mgr: mgr, places: map[string]resources.Place{}}
	for _, p := range places {
		a.places[p.Name] = p
	}
	return a
}

// SetState takes the latest player state from the server.
func (a *Actions) SetState(c resources.InventoryContents) {
	a.state = c
	a.received = time.Now()
}

func (a *Actions) send(c resources.ActionCommandContents) {
	if a.Send != nil {
		a.Send(c)
	}
}

func (a *Actions) Update() error {
	winW, _ := ebiten.WindowSize()
	imgui.SetNextWindowPosV(imgui.Vec2{X: float32(winW) - 640, Y: 40}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.BeginV("Actions", nil, imgui.WindowFlagsAlwaysAutoResize)

	a.controls()
	imgui.Separator()
	a.foundPlaces()
	imgui.Separator()
	a.list()

	imgui.End()
	return nil
}

func (a *Actions) controls() {
	n := len(a.state.Actions)
	if a.state.Running {
		if imgui.Button("Stop") {
			a.send(resources.ActionCommandContents{Op: resources.ActionOpStop})
		}
		imgui.SameLine()
		imgui.Text("Working through the list")
		return
	}
	if button("Start", n > 0) {
		a.send(resources.ActionCommandContents{Op: resources.ActionOpStart})
	}
	imgui.SameLine()
	if n == 0 {
		dimmed("Add an action below to get started")
	} else {
		dimmed("Stopped. Click the map to walk around.")
	}
}

func (a *Actions) foundPlaces() {
	imgui.Text("Places you've found")
	full := len(a.state.Actions) >= resources.MaxActions
	for _, name := range a.state.Found {
		place, ok := a.places[name]
		if !ok {
			continue
		}
		for _, kind := range resources.ActionKinds {
			if resources.ActionPlaceKind[kind] != place.Kind {
				continue
			}
			if button(resources.ActionNames[kind]+"##"+name, !full) {
				a.send(resources.ActionCommandContents{Op: resources.ActionOpAdd, Action: resources.ActionSpec{
					Kind: kind, Place: name, Goal: resources.GoalAmount,
					Amount: resources.ActionDefaultAmounts[kind], Minutes: defaultMinutes,
				}})
			}
			imgui.SameLine()
			imgui.Text(name)
		}
	}
	dimmed("Walk into a village, mine or forest to find it.")
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

	imgui.Text(fmt.Sprintf("%d. %s at %s", i+1, resources.ActionNames[spec.Kind], spec.Place))
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

	unit := resources.ItemUnits[spec.Product()]
	edited := spec

	// Goal: a bank total, or a length of time.
	if imgui.RadioButton("Until the bank has", spec.Goal == resources.GoalAmount) {
		edited.Goal = resources.GoalAmount
	}
	imgui.SameLine()
	if a.goalInput("##amount", i, 0, &a.amounts[i], spec.Amount, resources.MaxGoalAmount) {
		edited.Amount = int(a.amounts[i])
	}
	imgui.SameLine()
	imgui.Text(unit)

	if imgui.RadioButton("For", spec.Goal == resources.GoalTime) {
		edited.Goal = resources.GoalTime
	}
	imgui.SameLine()
	if a.goalInput("##minutes", i, 1, &a.minutes[i], spec.Minutes, resources.MaxGoalMins) {
		edited.Minutes = int(a.minutes[i])
	}
	imgui.SameLine()
	imgui.Text("minutes")

	if edited != spec {
		a.send(resources.ActionCommandContents{Op: resources.ActionOpEdit, Index: i, Action: edited})
	}

	// Progress, for the goal that's set.
	var progress string
	if spec.Goal == resources.GoalAmount {
		progress = fmt.Sprintf("Bank: %d/%d %s", min(q.Banked, spec.Amount), spec.Amount, unit)
	} else {
		left := q.SecondsLeft
		if i == 0 && a.state.Running {
			left = max(0, left-time.Since(a.received).Seconds())
		}
		progress = fmt.Sprintf("%d:%02d left", int(left)/60, int(left)%60)
	}
	if i == 0 && a.state.Running {
		progress = "Now: " + progress
	}
	dimmed(progress)
}

// goalInput draws a number box for a goal value. While the player is typing
// in it, it keeps their value; otherwise it shows the server's. It reports
// whether the player changed the value.
func (a *Actions) goalInput(id string, i, which int, value *int32, server, limit int) bool {
	if !a.editing[i][which] {
		*value = int32(server)
	}
	imgui.PushItemWidth(170)
	changed := imgui.InputInt(id, value)
	imgui.PopItemWidth()
	a.editing[i][which] = imgui.IsItemActive()
	*value = min(max(*value, 1), int32(limit))
	return changed && int(*value) != server
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
