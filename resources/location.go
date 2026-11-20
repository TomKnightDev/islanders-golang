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
	ObjectIronRock:   "the iron rocks",
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

// Landmark returns the nearest named place within landmarkRadius tiles of
// (tx, ty), such as "the bank", or "" if there is none.
func (wm *WorldMap) Landmark(tx, ty int) string {
	best, bestDist := "", landmarkRadius+1
	for _, o := range wm.Objects {
		d := max(abs(o.X-tx), abs(o.Y-ty))
		if d < bestDist {
			best, bestDist = landmarkNames[o.Kind], d
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
