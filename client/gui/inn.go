package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Inn is the quests and rumours window. It opens when the character stops
// at an inn. Rumours give directions to places the player hasn't found.
type Inn struct {
	mgr      *renderer.Manager
	contents resources.InnContents

	open    bool
	shownAt string // the village whose inn the window was last opened for
}

func NewInn(mgr *renderer.Manager) *Inn { return &Inn{mgr: mgr} }

// SetState opens the window on arriving at an inn and closes it on leaving.
func (n *Inn) SetState(c resources.InventoryContents) {
	at := ""
	if c.At == resources.BuildingInn {
		at = c.AtVillage
	}
	if at != n.shownAt {
		n.open = at != ""
		n.shownAt = at
	}
}

// SetInn takes the inn's quests and rumours from the server.
func (n *Inn) SetInn(c resources.InnContents) { n.contents = c }

func (n *Inn) Update() error {
	if !n.open || n.shownAt == "" {
		return nil
	}
	w, h := screenSize()
	placeWindow("Inn", imgui.Vec2{X: w / 2, Y: h / 2}, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.BeginV("Inn", &n.open, imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text(fmt.Sprintf("The %s inn", n.shownAt))
	imgui.Separator()

	imgui.Text("Quests")
	dimmed("No one has any work for you yet.")
	imgui.Separator()

	imgui.Text("Rumours")
	imgui.PushTextWrapPosV(520)
	switch {
	case n.contents.Village != n.shownAt:
		dimmed("Listening...")
	case len(n.contents.Rumours) == 0:
		dimmed("Nobody here knows of anywhere you haven't been.")
	default:
		for _, r := range n.contents.Rumours {
			imgui.Bullet()
			imgui.SameLine()
			imgui.Text(r)
		}
	}
	imgui.PopTextWrapPos()
	imgui.End()
	return nil
}

func (n *Inn) Draw(screen *ebiten.Image) {}
