package gui

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// tooltipWidth is how wide a tooltip's text runs before wrapping, in pixels.
const tooltipWidth = 520

// Hover shows a tooltip naming whatever is under the mouse on the map and
// saying what it's for.
type Hover struct {
	world *resources.WorldMap
	// cursorTile returns the map tile under the mouse.
	cursorTile func() (int, int)
}

func NewHover(world *resources.WorldMap, cursorTile func() (int, int)) *Hover {
	return &Hover{world: world, cursorTile: cursorTile}
}

func (h *Hover) Update() error {
	if h.world == nil || imgui.CurrentIO().WantCaptureMouse() {
		return nil
	}
	tx, ty := h.cursorTile()
	wm := h.world
	if tx < 0 || ty < 0 || tx >= wm.Width || ty >= wm.Height {
		return nil
	}
	gid := wm.Layers[resources.LayerObject].Data[ty*wm.Width+tx]
	info, ok := resources.ObjectInfo[gid]
	if !ok {
		return nil
	}
	name, desc := info.Name, info.Desc

	switch {
	case resources.IsTree(gid):
		if forest := h.placeAt(tx, ty, resources.PlaceForest); forest != nil {
			desc = fmt.Sprintf("Chop it for logs: add Chop wood at %s in the Actions window.", forest.Name)
		} else {
			desc = "A tree. Only trees in a named forest can be chopped."
		}
	case gid == resources.TileIronRock || gid == resources.TileIronRockDepleted:
		if mine := h.placeAt(tx, ty, resources.PlaceMine); mine != nil {
			name += " at " + mine.Name
		}
	default:
		if village := h.placeAt(tx, ty, resources.PlaceVillage); village != nil {
			name = village.Name + "'s " + strings.ToLower(name[:1]) + name[1:]
		}
	}

	imgui.BeginTooltip()
	imgui.PushTextWrapPosV(tooltipWidth)
	imgui.Text(name)
	dimmed(desc)
	imgui.PopTextWrapPos()
	imgui.EndTooltip()
	return nil
}

func (h *Hover) placeAt(tx, ty int, kind resources.PlaceKind) *resources.Place {
	for i := range h.world.Places {
		p := &h.world.Places[i]
		if p.Kind == kind && max(abs(p.X-tx), abs(p.Y-ty)) <= p.Radius {
			return p
		}
	}
	return nil
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Draw does nothing: the tooltip is drawn with the other imgui windows.
func (h *Hover) Draw(screen *ebiten.Image) {}
