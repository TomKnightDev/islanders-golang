package main

import "github.com/tomknightdev/islanders-golang/resources"

// buildingReach is how close (in tiles, either way) a player must stand to
// a building to use it.
const buildingReach = 1

// visit is a building a player is standing at, in a village.
type visit struct {
	at      resources.Building
	village string
}

// buildingTiles is every tile of a kind of building.
func (w *World) buildingTiles(b resources.Building) []tile {
	switch b {
	case resources.BuildingBank:
		return w.banks
	case resources.BuildingInn:
		return w.inns
	case resources.BuildingStore:
		var out []tile
		for _, s := range w.stores {
			out = append(out, s.t)
		}
		return out
	case resources.BuildingAnvil:
		return stationTiles(w.stations[resources.ObjectAnvil])
	case resources.BuildingFletcher:
		return stationTiles(w.stations[resources.ObjectFletcher])
	}
	return nil
}

func stationTiles(byVillage map[string]tile) []tile {
	var out []tile
	for _, t := range byVillage {
		out = append(out, t)
	}
	return out
}

// buildingTile is a village's building of a kind.
func (w *World) buildingTile(b resources.Building, village string) (tile, bool) {
	for _, t := range w.buildingTiles(b) {
		if w.villageAt(t) == village {
			return t, true
		}
	}
	return tile{}, false
}

// visiting is the building p is standing at, if any. Buildings only open to
// a player who's stopped there with the action list stopped; jobs bank and
// rest on their own. Where several are in reach, the one the player clicked
// or visited wins, then the nearest.
func (w *World) visiting(p *player) visit {
	if p.running || len(p.path) > 0 {
		return visit{}
	}
	at := p.currentTile()
	var best visit
	bestDist, found := 0, false
	for _, b := range resources.Buildings {
		for _, t := range w.buildingTiles(b) {
			if max(abs(t.x-at.x), abs(t.y-at.y)) > buildingReach {
				continue
			}
			d := abs(t.x-at.x) + abs(t.y-at.y)
			if t == p.visitPrefer {
				d = -1
			}
			if !found || d < bestDist {
				best, bestDist, found = visit{b, w.villageAt(t)}, d, true
			}
		}
	}
	return best
}

// bankOpen is the village whose bank p is standing at, if any.
func (w *World) bankOpen(p *player) (string, bool) {
	v := w.visiting(p)
	return v.village, v.at == resources.BuildingBank
}

// villageAt is the name of the village nearest a tile.
func (w *World) villageAt(t tile) string {
	name, best := "", -1
	for _, pl := range w.wm.Places {
		if pl.Kind != resources.PlaceVillage {
			continue
		}
		if d := max(abs(pl.X-t.x), abs(pl.Y-t.y)); best < 0 || d < best {
			name, best = pl.Name, d
		}
	}
	return name
}

// tickVisit tells the client when p arrives at or leaves a building, so its
// window opens and closes, and sends what the window shows.
func (w *World) tickVisit(p *player) {
	v := w.visiting(p)
	if v == p.visit {
		return
	}
	p.visit = v
	p.invDirty = true
	if p.sess == nil {
		return
	}
	switch v.at {
	case resources.BuildingStore:
		p.sess.push(resources.NewStoreMessage(w.storeContents(p, w.storeIn(v.village))))
	case resources.BuildingInn:
		p.sess.push(resources.NewInnMessage(resources.InnContents{Village: v.village, Rumours: w.rumours(p, v.village)}))
	}
}

// bankCommand moves items between the backpack and the bank, keeping their
// quality. Only allowed while standing at a bank.
func (w *World) bankCommand(p *player, c resources.BankCommandContents) {
	if _, ok := w.bankOpen(p); !ok {
		return
	}
	defer func() { p.invDirty = true }()
	switch c.Op {
	case resources.BankOpDepositAll:
		p.backpack.TransferAll(&p.bank)
	case resources.BankOpDeposit:
		moveOne(&p.backpack, &p.bank, c)
	case resources.BankOpWithdraw:
		moveOne(&p.bank, &p.backpack, c)
	}
}

// moveOne moves a piece by ID, or Count of an item kind (only of one
// quality, if the command names one).
func moveOne(from, to *resources.Inventory, c resources.BankCommandContents) {
	switch {
	case c.ID != 0:
		if piece, ok := from.TakeID(c.ID); ok && to.AddStack(piece) == 0 {
			from.AddStack(piece) // no room
		}
	case c.Count > 0 && c.Quality > 0:
		for _, part := range from.TakeQuality(c.Item, c.Quality, c.Count) {
			if got := to.AddStack(part); got < part.Count {
				part.Count -= got
				from.AddStack(part) // no room for the rest
			}
		}
	case c.Count > 0:
		moveItems(from, to, c.Item, c.Count)
	}
}

// combineCommand merges a material's stacks, averaging their quality: in
// the backpack anywhere, in the bank only standing at one.
func (w *World) combineCommand(p *player, c resources.CombineCommandContents) {
	inv := &p.backpack
	if c.Bank {
		if _, ok := w.bankOpen(p); !ok {
			return
		}
		inv = &p.bank
	}
	if inv.Combine(c.Item, c.Qualities) {
		p.invDirty = true
	}
}
