package main

import (
	"container/heap"

	"github.com/tomknightdev/islanders-golang/resources"
)

type tile struct{ x, y int }

var neighbours = [4]tile{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

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
// to goal. If goal is blocked, the path ends on the nearest walkable tile next
// to it, so clicking a rock or a building walks up to it. It returns nil when
// there is no way there, and an empty path when already there.
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

	// Manhattan distance, less one when the goal itself is blocked, so the
	// estimate never overshoots the real distance to a tile beside it.
	h := func(t tile) int { return max(0, abs(t.x-goal.x)+abs(t.y-goal.y)-slack) }

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
		for _, d := range neighbours {
			n := cur.t.add(d)
			if !wm.TilePassable(n.x, n.y) {
				continue
			}
			g := cur.g + 1
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
