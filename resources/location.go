package resources

import (
	"fmt"
	"math"
)

// Locations are given as a grid square players can say out loud ("F6") plus
// the exact tile. The map is split into GridSize×GridSize-tile squares,
// lettered west to east and numbered north to south.
const (
	GridSize       = 10
	landmarkRadius = 4 // tiles; how close counts as "near" something
)

var landmarkNames = map[ObjectKind]string{
	ObjectBank:       "the bank",
	ObjectBlacksmith: "the blacksmith",
	ObjectAnvil:      "the anvil",
	ObjectInn:        "the inn",
	ObjectRock:       "the rocks",
}

// TileAt is the tile under the centre of a sprite whose top-left is at pos.
func TileAt(px, py float64) (int, int) {
	return int(math.Floor((px + TileSize/2) / TileSize)), int(math.Floor((py + TileSize/2) / TileSize))
}

// GridRef names the grid square containing tile (tx, ty), such as "F6".
func GridRef(tx, ty int) string {
	col := tx / GridSize
	if col < 0 || col >= 26 {
		return "??"
	}
	return fmt.Sprintf("%c%d", 'A'+col, ty/GridSize+1)
}

// nearRadius is how far beyond a place's edge still counts as "near" it.
const nearRadius = 8

// Describe says where tile (tx, ty) is in words players use: "in Oakhaven",
// "at Ironridge Mine", "near Saltmere", or near a landmark such as "the
// bank". It returns "" out in the wilds.
func (wm *WorldMap) Describe(tx, ty int) string {
	var near *Place
	nearDist := 0
	for i := range wm.Places {
		p := &wm.Places[i]
		d := max(abs(p.X-tx), abs(p.Y-ty))
		if d <= p.Radius {
			if p.Kind == PlaceVillage || p.Kind == PlaceForest || p.Kind == PlaceCamp {
				return "in " + p.Name
			}
			return "at " + p.Name
		}
		if d <= p.Radius+nearRadius && (near == nil || d < nearDist) {
			near, nearDist = p, d
		}
	}
	if near != nil {
		return "near " + near.Name
	}
	if l := wm.landmark(tx, ty); l != "" {
		return "near " + l
	}
	return ""
}

// landmark returns the nearest interactive object within landmarkRadius
// tiles of (tx, ty), such as "the bank", or "" if there is none.
func (wm *WorldMap) landmark(tx, ty int) string {
	best, bestDist := "", landmarkRadius+1
	for _, o := range wm.Objects {
		name, named := landmarkNames[o.Kind]
		if d := max(abs(o.X-tx), abs(o.Y-ty)); named && d < bestDist {
			best, bestDist = name, d
		}
	}
	return best
}

// Location is a shareable description of tile (tx, ty): "F6 (52, 48)".
func Location(tx, ty int) string {
	return fmt.Sprintf("%s (%d, %d)", GridRef(tx, ty), tx, ty)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
