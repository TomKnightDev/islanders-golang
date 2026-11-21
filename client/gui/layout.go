package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

// windowMargin is the gap between a window and the screen edge, or the
// window next to it.
const windowMargin = 10

// placeFrames is how many frames a window is held in its spot when it first
// opens. Auto-sizing windows only know their size after a frame or two, and
// windows anchored to the right or bottom need it to sit flush.
const placeFrames = 5

var placed = map[string]int{}

// placeWindow puts the next window at pos, with pivot choosing which of its
// corners sits there: (0,0) top-left, (1,1) bottom-right. It applies for the
// first few frames each time the game loads, overriding positions saved in
// imgui.ini, and then leaves the player free to move the window.
func placeWindow(id string, pos, pivot imgui.Vec2) {
	if placed[id] >= placeFrames {
		return
	}
	placed[id]++
	imgui.SetNextWindowPosV(pos, imgui.ConditionAlways, pivot)
}

// screenSize is the window size in UI pixels.
func screenSize() (float32, float32) {
	w, h := ebiten.WindowSize()
	return float32(w), float32(h)
}

// infoBottom is where the info window ends, so the skills window can sit
// below it.
var infoBottom float32 = windowMargin

// actionsLeft is the left edge of the actions window, so the action picker
// can open beside it.
var actionsLeft float32

// inventoryTop is where the inventory window starts, so the equipment window
// can sit above it.
var inventoryTop float32
