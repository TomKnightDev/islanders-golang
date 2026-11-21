package gui

import (
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
	var name, desc string
	if kind, _, ok := resources.NodeMaterial(gid); ok {
		info, _ := resources.NodeInfo(gid)
		pl := h.placeAt(tx, ty, resources.ActionPlaceKind[kind])
		switch {
		case pl != nil:
			name, desc = info.Name+" in "+pl.Name, info.Desc
		case kind == resources.ActionChop:
			name, desc = info.Name, "Only trees in a named forest can be chopped."
		default:
			// Bushes and flowers outside farms are just scenery.
			obj, ok := resources.ObjectInfo[gid]
			if !ok {
				return nil
			}
			name, desc = obj.Name, obj.Desc
		}
	} else {
		obj, ok := resources.ObjectInfo[gid]
		if !ok {
			return nil
		}
		name, desc = obj.Name, obj.Desc
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
