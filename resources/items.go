package resources

type ItemKind string

const (
	ItemIronOre   ItemKind = "ironOre"
	ItemIronIngot ItemKind = "ironIngot"
)

// StackLimits is how many of an item fit in one inventory slot.
var StackLimits = map[ItemKind]int{
	ItemIronOre:   10,
	ItemIronIngot: 20,
}

var ItemNames = map[ItemKind]string{
	ItemIronOre:   "Iron ore",
	ItemIronIngot: "Iron ingot",
}

type Stack struct {
	Kind  ItemKind `json:"kind"`
	Count int      `json:"count"`
}

// Inventory is a fixed number of slots, each holding one stack of a single
// item kind. Used for backpacks and bank accounts.
type Inventory struct {
	Slots  int     `json:"slots"`
	Stacks []Stack `json:"stacks"`
}

func NewInventory(slots int) Inventory {
	return Inventory{Slots: slots}
}

// Add puts up to n of kind into the inventory, topping up partial stacks
// before starting new ones, and returns how many fit.
func (inv *Inventory) Add(kind ItemKind, n int) int {
	limit := StackLimits[kind]
	added := 0
	for i := range inv.Stacks {
		if n == 0 {
			break
		}
		s := &inv.Stacks[i]
		if s.Kind != kind || s.Count >= limit {
			continue
		}
		take := min(n, limit-s.Count)
		s.Count += take
		added += take
		n -= take
	}
	for n > 0 && len(inv.Stacks) < inv.Slots {
		take := min(n, limit)
		inv.Stacks = append(inv.Stacks, Stack{Kind: kind, Count: take})
		added += take
		n -= take
	}
	return added
}

// Full reports whether not even one more of kind would fit.
func (inv *Inventory) Full(kind ItemKind) bool {
	if len(inv.Stacks) < inv.Slots {
		return false
	}
	for _, s := range inv.Stacks {
		if s.Kind == kind && s.Count < StackLimits[kind] {
			return false
		}
	}
	return true
}

func (inv *Inventory) Count(kind ItemKind) int {
	total := 0
	for _, s := range inv.Stacks {
		if s.Kind == kind {
			total += s.Count
		}
	}
	return total
}

func (inv *Inventory) Empty() bool { return len(inv.Stacks) == 0 }

// Remove takes up to n of kind out, emptying the last stacks first, and
// returns how many were removed.
func (inv *Inventory) Remove(kind ItemKind, n int) int {
	removed := 0
	for i := len(inv.Stacks) - 1; i >= 0 && n > 0; i-- {
		s := &inv.Stacks[i]
		if s.Kind != kind {
			continue
		}
		take := min(n, s.Count)
		s.Count -= take
		removed += take
		n -= take
	}
	kept := inv.Stacks[:0]
	for _, s := range inv.Stacks {
		if s.Count > 0 {
			kept = append(kept, s)
		}
	}
	inv.Stacks = kept
	return removed
}

// TransferAll moves everything that fits into dst and reports whether the
// inventory ended up empty.
func (inv *Inventory) TransferAll(dst *Inventory) bool {
	for _, s := range append([]Stack(nil), inv.Stacks...) {
		moved := dst.Add(s.Kind, s.Count)
		inv.Remove(s.Kind, moved)
	}
	return inv.Empty()
}

// Clone returns a copy that shares no memory with inv, safe to hand to
// another goroutine (such as a connection writer) while inv keeps changing.
func (inv Inventory) Clone() Inventory {
	inv.Stacks = append([]Stack(nil), inv.Stacks...)
	return inv
}
