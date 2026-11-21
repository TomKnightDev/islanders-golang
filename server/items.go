package main

import "github.com/tomknightdev/islanders-golang/resources"

// newPiece makes a new piece of gear or a tool, with its own ID.
func (w *World) newPiece(k resources.ItemKind, quality int) resources.Stack {
	w.nextItemID++
	return resources.Stack{Kind: k, Count: 1, Quality: quality, ID: w.nextItemID}
}

// give adds n of an item at a quality to an inventory and returns how many
// fit. Gear and tools are made as separate pieces, each with its own ID.
func (w *World) give(inv *resources.Inventory, k resources.ItemKind, n, quality int) int {
	if !resources.IsPiece(k) {
		return inv.AddStack(resources.Stack{Kind: k, Count: n, Quality: quality})
	}
	added := 0
	for ; added < n; added++ {
		if inv.AddStack(w.newPiece(k, quality)) == 0 {
			break
		}
	}
	return added
}

// moveItems moves up to n of an item from one inventory to another, keeping
// each item's quality, damage and ID, and returns how many moved. Anything
// that doesn't fit stays where it was.
func moveItems(from, to *resources.Inventory, k resources.ItemKind, n int) int {
	moved := 0
	for _, part := range from.Take(k, n) {
		got := to.AddStack(part)
		moved += got
		if got < part.Count {
			rest := part
			rest.Count -= got
			from.AddStack(rest)
		}
	}
	return moved
}

// takeBest removes the highest-quality piece of a kind from an inventory.
func takeBest(inv *resources.Inventory, k resources.ItemKind) (resources.Stack, bool) {
	best := -1
	for i, s := range inv.Stacks {
		if s.Kind == k && (best < 0 || s.Quality > inv.Stacks[best].Quality) {
			best = i
		}
	}
	if best < 0 {
		return resources.Stack{}, false
	}
	piece := inv.Stacks[best]
	piece.Count = 1
	inv.Stacks[best].Count--
	if inv.Stacks[best].Count == 0 {
		inv.Stacks = append(inv.Stacks[:best], inv.Stacks[best+1:]...)
	}
	return piece, true
}

// tool is the tool p has on their belt for an action, or nil.
func (p *player) tool(kind resources.ActionKind) *resources.Stack {
	slot, ok := resources.ActionTool[kind]
	if !ok {
		return nil
	}
	s, ok := p.equipment[slot]
	if !ok {
		return nil
	}
	return &s
}

// stepTicks is how long a step of an action takes: the base time, scaled
// by skill level and by the tool for it (twice as long without one).
func (p *player) stepTicks(kind resources.ActionKind, base int) int {
	return max(1, int(float64(base)*resources.ToolSpeed(p.tool(kind))+0.5))
}
