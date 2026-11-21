package resources

type ItemKind string

const (
	ItemIronOre   ItemKind = "ironOre"
	ItemIronIngot ItemKind = "ironIngot" // shown as "Iron bar"
	ItemLogs      ItemKind = "logs"
	ItemArrows    ItemKind = "arrows"
)

// StackLimits is how many of an item fit in one slot of a stacking
// inventory (the bank).
var StackLimits = map[ItemKind]int{
	ItemIronOre:   10,
	ItemIronIngot: 20,
	ItemLogs:      20,
	ItemArrows:    50,
}

// BackpackStackLimits is how many of an item share one backpack slot. Most
// items take a slot each; arrows stack.
var BackpackStackLimits = map[ItemKind]int{
	ItemArrows: 50,
}

var ItemNames = map[ItemKind]string{
	ItemIronOre:   "Iron ore",
	ItemIronIngot: "Iron bar",
	ItemLogs:      "Logs",
	ItemArrows:    "Arrows",
}

// ItemUnits is how a quantity of the item reads: "100 ore", "20 arrows".
var ItemUnits = map[ItemKind]string{
	ItemIronOre:   "ore",
	ItemIronIngot: "bars",
	ItemLogs:      "logs",
	ItemArrows:    "arrows",
}

// ItemDescriptions are shown when hovering over an item.
var ItemDescriptions = map[ItemKind]string{
	ItemIronOre:   "Mined from iron rocks. Smelt it into bars at a blacksmith.",
	ItemIronIngot: "Smelted from iron ore. Used for crafting at the anvil (coming soon).",
	ItemLogs:      "Chopped from trees in the forests. A fletcher turns each log into 10 arrows.",
	ItemArrows:    "Made from logs at a fletcher's hut. Stacks 50 to a slot, even in the backpack.",
}

type Stack struct {
	Kind  ItemKind `json:"kind"`
	Count int      `json:"count"`
}

// Inventory is a fixed number of slots, each holding one stack of a single
// item kind. Used for backpacks and bank accounts. When Stacking is false
// (the backpack) stacks follow BackpackStackLimits: one item per slot unless
// the item says otherwise.
type Inventory struct {
	Slots    int     `json:"slots"`
	Stacking bool    `json:"stacking"`
	Stacks   []Stack `json:"stacks"`
}

func NewInventory(slots int, stacking bool) Inventory {
	return Inventory{Slots: slots, Stacking: stacking}
}

// limit is how many of kind one slot holds.
func (inv *Inventory) limit(kind ItemKind) int {
	if !inv.Stacking {
		if n, ok := BackpackStackLimits[kind]; ok {
			return n
		}
		return 1
	}
	return StackLimits[kind]
}

// Add puts up to n of kind into the inventory, topping up partial stacks
// before starting new ones, and returns how many fit.
func (inv *Inventory) Add(kind ItemKind, n int) int {
	limit := inv.limit(kind)
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
		if s.Kind == kind && s.Count < inv.limit(kind) {
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

// FreeSlots is how many slots hold nothing.
func (inv *Inventory) FreeSlots() int { return inv.Slots - len(inv.Stacks) }
