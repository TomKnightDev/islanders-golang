package resources

type ItemKind string

const (
	ItemCopperOre  ItemKind = "copperOre"
	ItemIronOre    ItemKind = "ironOre"
	ItemGoldOre    ItemKind = "goldOre"
	ItemMithrilOre ItemKind = "mithrilOre"

	ItemCopperBar  ItemKind = "copperBar"
	ItemIronIngot  ItemKind = "ironIngot" // shown as "Iron bar"
	ItemGoldBar    ItemKind = "goldBar"
	ItemMithrilBar ItemKind = "mithrilBar"

	ItemOakLogs     ItemKind = "logs"
	ItemPineLogs    ItemKind = "pineLogs"
	ItemMapleLogs   ItemKind = "mapleLogs"
	ItemDarkOakLogs ItemKind = "darkOakLogs"

	ItemOakArrows     ItemKind = "arrows"
	ItemPineArrows    ItemKind = "pineArrows"
	ItemMapleArrows   ItemKind = "mapleArrows"
	ItemDarkOakArrows ItemKind = "darkOakArrows"

	ItemCabbage ItemKind = "cabbage"
	ItemCotton  ItemKind = "cotton"
)

// ItemDef is everything about an item kind.
type ItemDef struct {
	Name      string
	Unit      string // how a quantity reads: "100 ore", "20 arrows"
	Desc      string // shown when hovering over the item
	Stack     int    // per bank slot
	PackStack int    // per backpack slot; 0 means one item per slot
	Heal      int    // health restored when eaten; 0 for things you can't eat
}

var Items = map[ItemKind]ItemDef{
	ItemCopperOre:  {"Copper ore", "copper ore", "Mined from copper rocks. Smelt it into copper bars at a blacksmith.", 10, 0, 0},
	ItemIronOre:    {"Iron ore", "iron ore", "Mined from iron rocks. Smelt it into iron bars at a blacksmith.", 10, 0, 0},
	ItemGoldOre:    {"Gold ore", "gold ore", "Mined from gold rocks. Smelt it into gold bars at a blacksmith.", 10, 0, 0},
	ItemMithrilOre: {"Mithril ore", "mithril ore", "Mined from rare blue mithril rocks. Smelt it into mithril bars.", 10, 0, 0},

	ItemCopperBar:  {"Copper bar", "copper bars", "Smelted from copper ore. For crafting equipment (coming soon).", 20, 0, 0},
	ItemIronIngot:  {"Iron bar", "iron bars", "Smelted from iron ore. For crafting equipment (coming soon).", 20, 0, 0},
	ItemGoldBar:    {"Gold bar", "gold bars", "Smelted from gold ore. For crafting equipment (coming soon).", 20, 0, 0},
	ItemMithrilBar: {"Mithril bar", "mithril bars", "Smelted from mithril ore. For crafting equipment (coming soon).", 20, 0, 0},

	ItemOakLogs:     {"Oak logs", "oak logs", "Chopped from oaks. A fletcher turns each log into 10 oak arrows.", 20, 0, 0},
	ItemPineLogs:    {"Pine logs", "pine logs", "Chopped from pines. A fletcher turns each log into 10 pine arrows.", 20, 0, 0},
	ItemMapleLogs:   {"Maple logs", "maple logs", "Chopped from autumn maples. A fletcher turns each log into 10 maple arrows.", 20, 0, 0},
	ItemDarkOakLogs: {"Dark oak logs", "dark oak logs", "Chopped from dark oaks. A fletcher turns each log into 10 dark oak arrows.", 20, 0, 0},

	ItemOakArrows:     {"Oak arrows", "oak arrows", "Made from oak logs. Stacks 50 to a slot, even in the backpack.", 50, 50, 0},
	ItemPineArrows:    {"Pine arrows", "pine arrows", "Made from pine logs. Stacks 50 to a slot, even in the backpack.", 50, 50, 0},
	ItemMapleArrows:   {"Maple arrows", "maple arrows", "Made from maple logs. Stacks 50 to a slot, even in the backpack.", 50, 50, 0},
	ItemDarkOakArrows: {"Dark oak arrows", "dark oak arrows", "Made from dark oak logs. Stacks 50 to a slot, even in the backpack.", 50, 50, 0},

	ItemCabbage: {"Cabbage", "cabbages", "Grown on farms. Eaten in fights to heal 30 health.", 20, 0, 30},
	ItemCotton:  {"Cotton", "cotton", "Grown on farms. The raw material for cloth.", 20, 0, 0},
}

// Lookups built from Items, kept for brevity at call sites.
var (
	StackLimits         = map[ItemKind]int{}
	BackpackStackLimits = map[ItemKind]int{}
	ItemNames           = map[ItemKind]string{}
	ItemUnits           = map[ItemKind]string{}
	ItemDescriptions    = map[ItemKind]string{}
)

func init() {
	for k, d := range Items {
		StackLimits[k] = d.Stack
		if d.PackStack > 0 {
			BackpackStackLimits[k] = d.PackStack
		}
		ItemNames[k] = d.Name
		ItemUnits[k] = d.Unit
		ItemDescriptions[k] = d.Desc
	}
}

// Stack is some of one item kind in one slot. Materials stack, keeping the
// average quality of what's in them; gear and tools are always a stack of
// one, carrying that piece's own quality, damage and ID.
type Stack struct {
	Kind    ItemKind `json:"kind"`
	Count   int      `json:"count"`
	Quality int      `json:"quality,omitempty"` // 1–100
	Damage  int      `json:"damage,omitempty"`  // 0–100; gear and tools only
	ID      uint32   `json:"id,omitempty"`      // gear and tools only; 0 if unassigned
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

// Add puts up to n of kind into the inventory at the default quality and
// returns how many fit.
func (inv *Inventory) Add(kind ItemKind, n int) int {
	return inv.AddStack(Stack{Kind: kind, Count: n, Quality: DefaultQuality})
}

// AddStack puts up to s.Count items into the inventory, topping up partial
// stacks of the same kind before starting new ones, and returns how many
// fit. A topped-up stack's quality becomes the weighted average. Items that
// never stack (gear, tools) each take their own slot, keeping their quality,
// damage and ID.
func (inv *Inventory) AddStack(s Stack) int {
	limit := inv.limit(s.Kind)
	n, added := s.Count, 0
	if limit > 1 {
		for i := range inv.Stacks {
			if n == 0 {
				break
			}
			cur := &inv.Stacks[i]
			if cur.Kind != s.Kind || cur.Count >= limit {
				continue
			}
			take := min(n, limit-cur.Count)
			cur.Quality = (cur.Quality*cur.Count + s.Quality*take) / (cur.Count + take)
			cur.Count += take
			added += take
			n -= take
		}
	}
	for n > 0 && len(inv.Stacks) < inv.Slots {
		take := min(n, limit)
		piece := s
		piece.Count = take
		if added > 0 && limit == 1 {
			piece.ID = 0 // only the first piece keeps an ID; callers add gear one at a time
		}
		inv.Stacks = append(inv.Stacks, piece)
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

// AverageQuality is the average quality of everything of kind held, or 0.
func (inv *Inventory) AverageQuality(kind ItemKind) int {
	sum, n := 0, 0
	for _, s := range inv.Stacks {
		if s.Kind == kind {
			sum += s.Quality * s.Count
			n += s.Count
		}
	}
	if n == 0 {
		return 0
	}
	return sum / n
}

func (inv *Inventory) Empty() bool { return len(inv.Stacks) == 0 }

// Remove takes up to n of kind out, emptying the last stacks first, and
// returns how many were removed.
func (inv *Inventory) Remove(kind ItemKind, n int) int {
	removed := 0
	for _, s := range inv.Take(kind, n) {
		removed += s.Count
	}
	return removed
}

// Take removes up to n of kind, emptying the last stacks first, and returns
// what it took with each part's quality, damage and ID.
func (inv *Inventory) Take(kind ItemKind, n int) []Stack {
	var taken []Stack
	for i := len(inv.Stacks) - 1; i >= 0 && n > 0; i-- {
		s := &inv.Stacks[i]
		if s.Kind != kind {
			continue
		}
		part := *s
		part.Count = min(n, s.Count)
		s.Count -= part.Count
		n -= part.Count
		taken = append(taken, part)
	}
	inv.compact()
	return taken
}

// TakeID removes the piece with this ID.
func (inv *Inventory) TakeID(id uint32) (Stack, bool) {
	for i, s := range inv.Stacks {
		if s.ID == id && id != 0 {
			inv.Stacks = append(inv.Stacks[:i], inv.Stacks[i+1:]...)
			return s, true
		}
	}
	return Stack{}, false
}

func (inv *Inventory) compact() {
	kept := inv.Stacks[:0]
	for _, s := range inv.Stacks {
		if s.Count > 0 {
			kept = append(kept, s)
		}
	}
	inv.Stacks = kept
}

// AddAll adds stacks, as taken from another inventory, and returns how many
// items fit.
func (inv *Inventory) AddAll(stacks []Stack) int {
	added := 0
	for _, s := range stacks {
		added += inv.AddStack(s)
	}
	return added
}

// TransferAll moves everything that fits into dst, keeping each item's
// quality, and reports whether the inventory ended up empty.
func (inv *Inventory) TransferAll(dst *Inventory) bool {
	for i := range inv.Stacks {
		s := &inv.Stacks[i]
		s.Count -= dst.AddStack(*s)
	}
	inv.compact()
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
