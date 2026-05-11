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
)

func GenerateWorldMap() resources.WorldMap {
	grid := [mapWidth][mapHeight]int{}

	// Fill base terrain by distance from edge
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			grid[x][y] = baseTile(x, y)
		}
	}

	// Scatter trees in the grass interior, keeping a clear spawn area at (1,1)
	placed := 0
	attempts := 0
	for placed < treeCount && attempts < treeCount*20 {
		attempts++
		tx := sandEdge + rand.Intn(mapWidth-sandEdge*2)
		ty := sandEdge + rand.Intn(mapHeight-sandEdge*2)

		// Keep a small clear zone around the spawn point
		if tx >= 47 && tx <= 53 && ty >= 47 && ty <= 53 {
			continue
		}

		if grid[tx][ty] == resources.TileGrass {
			if rand.Intn(2) == 0 {
				grid[tx][ty] = resources.TileTree1
			} else {
				grid[tx][ty] = resources.TileTree2
			}
			placed++
		}
	}

	// Flatten to row-major data slice: Data[y*width+x] = tile at (x,y)
	data := make([]int, mapWidth*mapHeight)
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			data[y*mapWidth+x] = grid[x][y]
		}
	}

	return resources.WorldMap{
		Height: mapHeight,
		Width:  mapWidth,
		Layers: []resources.Layer{
			{
				Data:    data,
				Height:  mapHeight,
				ID:      1,
				Name:    "Tile Layer 1",
				Opacity: 1,
				Type:    "tilelayer",
				Visible: true,
				Width:   mapWidth,
			},
		},
	}
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
