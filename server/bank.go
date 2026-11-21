package main

import "github.com/tomknightdev/islanders-golang/resources"

// bankReach is how close (in tiles, either way) a player must stand to a
// bank to use it.
const bankReach = 1

// bankOpen is the village whose bank p is standing at, if any. The bank is
// only open to a player who's stopped there with the action list stopped;
// jobs bank on their own.
func (w *World) bankOpen(p *player) (string, bool) {
	if p.running || len(p.path) > 0 {
		return "", false
	}
	at := p.currentTile()
	for _, b := range w.banks {
		if max(abs(b.x-at.x), abs(b.y-at.y)) <= bankReach {
			return w.villageAt(b), true
		}
	}
	return "", false
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

// tickBankWindow tells the client when p arrives at or leaves a bank, so
// the bank window opens and closes.
func (w *World) tickBankWindow(p *player) {
	at, _ := w.bankOpen(p)
	if at != p.bankShown {
		p.bankShown = at
		p.invDirty = true
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

// moveOne moves a piece by ID, or Count of an item kind.
func moveOne(from, to *resources.Inventory, c resources.BankCommandContents) {
	if c.ID != 0 {
		if piece, ok := from.TakeID(c.ID); ok && to.AddStack(piece) == 0 {
			from.AddStack(piece) // no room
		}
		return
	}
	if c.Count > 0 {
		moveItems(from, to, c.Item, c.Count)
	}
}
