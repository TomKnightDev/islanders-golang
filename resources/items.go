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
	Name  string
	Unit  string // how a quantity reads: "100 ore", "20 arrows"
	Desc  string // shown when hovering over the item
	Stack int    // per slot, in the backpack or the bank
	Heal  int    // health restored when eaten; 0 for things you can't eat
}

var Items = map[ItemKind]ItemDef{
	ItemCopperOre:  {"Copper ore", "copper ore", "Mined from copper rocks. Smelt it into copper bars at a blacksmith.", 10, 0},
	ItemIronOre:    {"Iron ore", "iron ore", "Mined from iron rocks. Smelt it into iron bars at a blacksmith.", 10, 0},
	ItemGoldOre:    {"Gold ore", "gold ore", "Mined from gold rocks. Smelt it into gold bars at a blacksmith.", 10, 0},
	ItemMithrilOre: {"Mithril ore", "mithril ore", "Mined from rare blue mithril rocks. Smelt it into mithril bars.", 10, 0},

	ItemCopperBar:  {"Copper bar", "copper bars", "Smelted from copper ore. For crafting equipment (coming soon).", 20, 0},
	ItemIronIngot:  {"Iron bar", "iron bars", "Smelted from iron ore. For crafting equipment (coming soon).", 20, 0},
	ItemGoldBar:    {"Gold bar", "gold bars", "Smelted from gold ore. For crafting equipment (coming soon).", 20, 0},
	ItemMithrilBar: {"Mithril bar", "mithril bars", "Smelted from mithril ore. For crafting equipment (coming soon).", 20, 0},

	ItemOakLogs:     {"Oak logs", "oak logs", "Chopped from oaks. A fletcher turns each log into 10 oak arrows.", 20, 0},
	ItemPineLogs:    {"Pine logs", "pine logs", "Chopped from pines. A fletcher turns each log into 10 pine arrows.", 20, 0},
	ItemMapleLogs:   {"Maple logs", "maple logs", "Chopped from autumn maples. A fletcher turns each log into 10 maple arrows.", 20, 0},
	ItemDarkOakLogs: {"Dark oak logs", "dark oak logs", "Chopped from dark oaks. A fletcher turns each log into 10 dark oak arrows.", 20, 0},

	ItemOakArrows:     {"Oak arrows", "oak arrows", "Made from oak logs. Stacks 50 to a slot.", 50, 0},
	ItemPineArrows:    {"Pine arrows", "pine arrows", "Made from pine logs. Stacks 50 to a slot.", 50, 0},
	ItemMapleArrows:   {"Maple arrows", "maple arrows", "Made from maple logs. Stacks 50 to a slot.", 50, 0},
	ItemDarkOakArrows: {"Dark oak arrows", "dark oak arrows", "Made from dark oak logs. Stacks 50 to a slot.", 50, 0},

	ItemCabbage: {"Cabbage", "cabbages", "Grown on farms. Eaten in fights to heal 30 health.", 20, 30},
	ItemCotton:  {"Cotton", "cotton", "Grown on farms. The raw material for cloth.", 20, 0},
}

// Lookups built from Items, kept for brevity at call sites.
var (
	StackLimits      = map[ItemKind]int{}
	ItemNames        = map[ItemKind]string{}
	ItemUnits        = map[ItemKind]string{}
	ItemDescriptions = map[ItemKind]string{}
)

func init() {
	for k, d := range Items {
		StackLimits[k] = d.Stack
		ItemNames[k] = d.Name
		ItemUnits[k] = d.Unit
		ItemDescriptions[k] = d.Desc
	}
}

// Stack is what's in one slot: some of one material at one quality, or a
// single piece of gear or a tool with its own quality, damage and ID.
type Stack struct {
	Kind    ItemKind `json:"kind"`
	Count   int      `json:"count"`
	Quality int      `json:"quality,omitempty"` // 1–100
	Damage  int      `json:"damage,omitempty"`  // 0–100; gear and tools only
	ID      uint32   `json:"id,omitempty"`      // gear and tools only; 0 if unassigned
}

// Inventory is a fixed number of slots, used for backpacks and bank
// accounts. Each slot holds one piece of gear, or a stack of one material
// at one quality, up to the item's StackLimits. New items only join a stack
// of the same quality; Combine is how different qualities mix.
type Inventory struct {
	Slots  int     `json:"slots"`
	Stacks []Stack `json:"stacks"`
}

func NewInventory(slots int) Inventory {
	return Inventory{Slots: slots}
}

// limit is how many of kind one slot holds.
func limit(kind ItemKind) int { return max(StackLimits[kind], 1) }

// Add puts up to n of kind into the inventory at the default quality and
// returns how many fit.
func (inv *Inventory) Add(kind ItemKind, n int) int {
	return inv.AddStack(Stack{Kind: kind, Count: n, Quality: DefaultQuality})
}

// AddStack puts up to s.Count items into the inventory and returns how many
// fit. They top up stacks of the same kind and quality first, then start
// new ones. Gear and tools each take their own slot, keeping their quality,
// damage and ID.
func (inv *Inventory) AddStack(s Stack) int {
	lim := limit(s.Kind)
	n, added := s.Count, 0
	if lim > 1 {
		for i := range inv.Stacks {
			if n == 0 {
				break
			}
			cur := &inv.Stacks[i]
			if cur.Kind != s.Kind || cur.Quality != s.Quality || cur.Count >= lim {
				continue
			}
			take := min(n, lim-cur.Count)
			cur.Count += take
			added += take
			n -= take
		}
	}
	for n > 0 && len(inv.Stacks) < inv.Slots {
		take := min(n, lim)
		piece := s
		piece.Count = take
		if added > 0 && lim == 1 {
			piece.ID = 0 // only the first piece keeps an ID; callers add gear one at a time
		}
		inv.Stacks = append(inv.Stacks, piece)
		added += take
		n -= take
	}
	return added
}

// Full reports whether there's no slot free for a new stack.
func (inv *Inventory) Full() bool { return len(inv.Stacks) >= inv.Slots }

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
	return inv.take(n, func(s Stack) bool { return s.Kind == kind })
}

// TakeQuality is Take, but only from stacks of one quality.
func (inv *Inventory) TakeQuality(kind ItemKind, quality, n int) []Stack {
	return inv.take(n, func(s Stack) bool { return s.Kind == kind && s.Quality == quality })
}

// TakeSlot removes up to n from one slot.
func (inv *Inventory) TakeSlot(i, n int) (Stack, bool) {
	if i < 0 || i >= len(inv.Stacks) || n < 1 {
		return Stack{}, false
	}
	part := inv.Stacks[i]
	part.Count = min(n, part.Count)
	inv.Stacks[i].Count -= part.Count
	inv.compact()
	return part, true
}

func (inv *Inventory) take(n int, match func(Stack) bool) []Stack {
	var taken []Stack
	for i := len(inv.Stacks) - 1; i >= 0 && n > 0; i-- {
		s := &inv.Stacks[i]
		if !match(*s) {
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

// Combine merges every stack of a material into as few slots as it fits
// in, all at their average quality. It reports whether anything changed.
func (inv *Inventory) Combine(kind ItemKind) bool {
	if IsPiece(kind) {
		return false
	}
	stacks, total, sum := 0, 0, 0
	for _, s := range inv.Stacks {
		if s.Kind == kind {
			stacks++
			total += s.Count
			sum += s.Quality * s.Count
		}
	}
	if stacks < 2 || total == 0 {
		return false
	}
	inv.Take(kind, total)
	inv.AddStack(Stack{Kind: kind, Count: total, Quality: sum / total}) // fits: it needs no more slots than it had
	return true
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
