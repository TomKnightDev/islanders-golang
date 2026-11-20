package main

import (
	"math/rand"

	"github.com/tomknightdev/islanders-golang/resources"
)

const (
	mapWidth  = 100
	mapHeight = 100

	// Border zone thresholds (distance from edge in tiles)
	deepWaterEdge    = 3
	shallowWaterEdge = 5
	sandEdge         = 7

	treeCount = 180

	// Spawn clearing, in tiles. Players spawn at its centre (50,50).
	spawnMin = 47
	spawnMax = 53
)

// Fixed town layout, in tiles. The town sits just north of the spawn clearing
// and the iron rocks are a cluster to the south-east.
var townObjects = []struct {
	kind resources.ObjectKind
	x, y int
}{
	{resources.ObjectBank, 47, 44},
	{resources.ObjectBlacksmith, 52, 44},
	{resources.ObjectAnvil, 53, 44},
	{resources.ObjectIronRock, 59, 57},
	{resources.ObjectIronRock, 61, 57},
	{resources.ObjectIronRock, 60, 59},
	{resources.ObjectIronRock, 62, 59},
	{resources.ObjectIronRock, 61, 61},
}

func GenerateWorldMap() resources.WorldMap {
	ground := make([]int, mapWidth*mapHeight)
	objects := make([]int, mapWidth*mapHeight)
	at := func(x, y int) int { return y*mapWidth + x }

	// Fill base terrain by distance from edge
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			ground[at(x, y)] = baseTile(x, y)
		}
	}

	// Place interactive objects
	var worldObjects []resources.WorldObject
	for i, o := range townObjects {
		objects[at(o.x, o.y)] = resources.ObjectTiles[o.kind]
		worldObjects = append(worldObjects, resources.WorldObject{
			ID:   uint16(i + 1),
			Kind: o.kind,
			X:    o.x,
			Y:    o.y,
		})
	}

	// Scatter trees in the grass interior, keeping the spawn clearing and the
	// ring of tiles around every object free so they stay reachable.
	placed := 0
	attempts := 0
	for placed < treeCount && attempts < treeCount*20 {
		attempts++
		tx := sandEdge + rand.Intn(mapWidth-sandEdge*2)
		ty := sandEdge + rand.Intn(mapHeight-sandEdge*2)

		if tx >= spawnMin && tx <= spawnMax && ty >= spawnMin && ty <= spawnMax {
			continue
		}
		if nearObject(worldObjects, tx, ty) {
			continue
		}

		if ground[at(tx, ty)] == resources.TileGrass && objects[at(tx, ty)] == resources.TileNone {
			if rand.Intn(2) == 0 {
				objects[at(tx, ty)] = resources.TileTree1
			} else {
				objects[at(tx, ty)] = resources.TileTree2
			}
			placed++
		}
	}

	return resources.WorldMap{
		Height: mapHeight,
		Width:  mapWidth,
		Layers: []resources.Layer{
			newLayer(1, "Ground", ground),
			newLayer(2, "Objects", objects),
		},
		Objects: worldObjects,
	}
}

func newLayer(id int, name string, data []int) resources.Layer {
	return resources.Layer{
		Data:    data,
		Height:  mapHeight,
		ID:      id,
		Name:    name,
		Opacity: 1,
		Type:    "tilelayer",
		Visible: true,
		Width:   mapWidth,
	}
}

func nearObject(objects []resources.WorldObject, x, y int) bool {
	for _, o := range objects {
		if x >= o.X-2 && x <= o.X+2 && y >= o.Y-2 && y <= o.Y+2 {
			return true
		}
	}
	return false
}

func baseTile(x, y int) int {
	dist := min(x, y, mapWidth-1-x, mapHeight-1-y)
	switch {
	case dist < deepWaterEdge:
		return resources.TileWaterDeep
	case dist < shallowWaterEdge:
		return resources.TileWaterShallow
	case dist < sandEdge:
		return resources.TileSand
	default:
		return resources.TileGrass
	}
}
