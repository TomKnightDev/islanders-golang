package main

import (
	"container/heap"

	"github.com/tomknightdev/islanders-golang/resources"
)

type tile struct{ x, y int }

// neighbours are the tiles a blocked goal can be approached from.
var neighbours = [4]tile{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

// steps are the moves a walk can make: straight or diagonal.
var steps = [8]tile{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

// Step costs: a diagonal is about √2 times a straight step.
const (
	straightCost = 10
	diagonalCost = 14
)

// octile is the cheapest cost between two tiles on an open map.
func octile(a, b tile) int {
	dx, dy := abs(a.x-b.x), abs(a.y-b.y)
	return straightCost*max(dx, dy) + (diagonalCost-straightCost)*min(dx, dy)
}

// canStep reports whether a walker on from can move by d. Diagonal steps
// need both tiles they pass between to be free, so sprites never clip the
// corner of a tree or rock.
func canStep(wm *resources.WorldMap, from, d tile) bool {
	to := from.add(d)
	if !wm.TilePassable(to.x, to.y) {
		return false
	}
	if d.x != 0 && d.y != 0 {
		return wm.TilePassable(from.x+d.x, from.y) && wm.TilePassable(from.x, from.y+d.y)
	}
	return true
}

func (t tile) add(d tile) tile { return tile{t.x + d.x, t.y + d.y} }

// pixel returns the top-left pixel of the tile, which is where a sprite
// standing on it is drawn.
func (t tile) pixel() (float64, float64) {
	return float64(t.x * resources.TileSize), float64(t.y * resources.TileSize)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// findPath returns the tiles to walk through, in order, from start (excluded)
// to goal, moving in 8 directions. If goal is blocked, the path ends on the
// nearest walkable tile beside it, so clicking a rock or a building walks up
// to it. It returns nil when there is no way there, and an empty path when
// already there.
func findPath(wm *resources.WorldMap, start, goal tile) []tile {
	goals := map[tile]bool{}
	slack := 0
	if wm.TilePassable(goal.x, goal.y) {
		goals[goal] = true
	} else {
		slack = 1
		for _, d := range neighbours {
			if n := goal.add(d); wm.TilePassable(n.x, n.y) {
				goals[n] = true
			}
		}
	}
	if len(goals) == 0 {
		return nil
	}
	if goals[start] {
		return []tile{}
	}

	// Octile distance, less one step when the goal itself is blocked, so the
	// estimate never overshoots the real cost to a tile beside it.
	h := func(t tile) int { return max(0, octile(t, goal)-slack*straightCost) }

	cameFrom := map[tile]tile{}
	cost := map[tile]int{start: 0}
	open := &tileHeap{{t: start, f: h(start)}}

	for open.Len() > 0 {
		cur := heap.Pop(open).(tileNode)
		if goals[cur.t] {
			var path []tile
			for t := cur.t; t != start; t = cameFrom[t] {
				path = append(path, t)
			}
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path
		}
		if cur.g > cost[cur.t] {
			continue // stale heap entry
		}
		for _, d := range steps {
			if !canStep(wm, cur.t, d) {
				continue
			}
			n := cur.t.add(d)
			g := cur.g + straightCost
			if d.x != 0 && d.y != 0 {
				g = cur.g + diagonalCost
			}
			if old, seen := cost[n]; seen && g >= old {
				continue
			}
			cost[n] = g
			cameFrom[n] = cur.t
			heap.Push(open, tileNode{t: n, g: g, f: g + h(n)})
		}
	}
	return nil
}

type tileNode struct {
	t    tile
	g, f int
}

type tileHeap []tileNode

func (h tileHeap) Len() int           { return len(h) }
func (h tileHeap) Less(i, j int) bool { return h[i].f < h[j].f }
func (h tileHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *tileHeap) Push(x any)        { *h = append(*h, x.(tileNode)) }
func (h *tileHeap) Pop() any {
	old := *h
	n := old[len(old)-1]
	*h = old[:len(old)-1]
	return n
}
