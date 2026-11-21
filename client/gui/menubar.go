package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

// MenuBar is the bar along the top of the screen: a link per window, which
// opens or closes it, and Log out.
type MenuBar struct {
	items []menuItem
	// Logout disconnects and goes back to the login screen.
	Logout func()
}

type menuItem struct {
	label string
	open  *bool
}

func NewMenuBar() *MenuBar { return &MenuBar{} }

// Add puts a window's link on the bar; open is the window's Open flag.
func (m *MenuBar) Add(label string, open *bool) {
	m.items = append(m.items, menuItem{label, open})
}

// menuBarHeight is the bar's height, so windows pinned to the top of the
// screen sit below it.
var menuBarHeight float32

// topMargin is where windows pinned to the top of the screen start.
func topMargin() float32 { return menuBarHeight + windowMargin }

func (m *MenuBar) Update() error {
	if !imgui.BeginMainMenuBar() {
		return nil
	}
	menuBarHeight = imgui.WindowSize().Y
	for _, it := range m.items {
		if imgui.MenuItemV(it.label, "", *it.open, true) {
			*it.open = !*it.open
		}
	}
	if m.Logout != nil {
		// Off to the right, away from the window links.
		w := imgui.CalcTextSize("Log out", false, 0).X + 3*windowMargin
		imgui.SetCursorPos(imgui.Vec2{X: imgui.WindowSize().X - w, Y: imgui.CursorPosY()})
		if imgui.MenuItem("Log out") {
			m.Logout()
		}
	}
	imgui.EndMainMenuBar()
	return nil
}

func (m *MenuBar) Draw(screen *ebiten.Image) {}

// ResetLayout puts every window back in its starting place the next time
// it's shown, as on a fresh start (after logging out and in again).
func ResetLayout() {
	placed = map[string]int{}
	menuBarHeight = 0
}
