package resources

type ActionKind string

const (
	ActionMine   ActionKind = "mine"   // gather ore at a mine
	ActionChop   ActionKind = "chop"   // gather logs in a forest
	ActionFarm   ActionKind = "farm"   // gather crops at a farm
	ActionSmelt  ActionKind = "smelt"  // turn banked ore into bars at a village blacksmith
	ActionFletch ActionKind = "fletch" // turn banked logs into arrows at a village fletcher
	ActionFight  ActionKind = "fight"  // fight enemies at a camp for loot
	ActionSell   ActionKind = "sell"   // carry an item from the bank to a village store and sell it
)

// ActionKinds lists the actions in the order the action window shows them.
var ActionKinds = []ActionKind{ActionFight, ActionMine, ActionChop, ActionFarm, ActionSmelt, ActionFletch, ActionSell}

// ActionNames is the verb for each action; with a material it reads
// "Mine copper", "Fletch oak arrows".
var ActionNames = map[ActionKind]string{
	ActionMine:   "Mine",
	ActionChop:   "Chop",
	ActionFarm:   "Farm",
	ActionSmelt:  "Smelt",
	ActionFletch: "Fletch",
	ActionFight:  "Fight",
	ActionSell:   "Sell",
}

// ActionPlaceKind is the kind of place each action happens at.
var ActionPlaceKind = map[ActionKind]PlaceKind{
	ActionMine:   PlaceMine,
	ActionChop:   PlaceForest,
	ActionFarm:   PlaceFarm,
	ActionSmelt:  PlaceVillage,
	ActionFletch: PlaceVillage,
	ActionFight:  PlaceCamp,
	ActionSell:   PlaceVillage,
}

// ActionSkill is the skill each action trains, and which sets its speed
// and the materials it can use.
var ActionSkill = map[ActionKind]Skill{
	ActionMine:   SkillMining,
	ActionChop:   SkillWoodcutting,
	ActionFarm:   SkillFarming,
	ActionSmelt:  SkillSmithing,
	ActionFletch: SkillFletching,
	ActionFight:  SkillAttack, // enemies aren't level-gated; combat skills train from the fight itself
}

// Gathering actions take materials from nodes at a place; crafting actions
// turn one item from the bank into another.
func IsGathering(k ActionKind) bool { return k == ActionMine || k == ActionChop || k == ActionFarm }

// IsCombat reports whether an action is fighting, whose goal counts kills.
func IsCombat(k ActionKind) bool { return k == ActionFight }

// IsSelling reports whether an action is selling to a store. Its material
// is the item to sell, any sellable item, and its goal counts items sold.
func IsSelling(k ActionKind) bool { return k == ActionSell }

// CountsProgress reports whether an action's amount goal counts what the
// action itself did (kills, sales) rather than what's in the bank.
func CountsProgress(k ActionKind) bool { return IsCombat(k) || IsSelling(k) }

// ActionDefaultAmounts is the bank total a new action aims for.
var ActionDefaultAmounts = map[ActionKind]int{
	ActionMine:   100,
	ActionChop:   100,
	ActionFarm:   100,
	ActionSmelt:  50,
	ActionFletch: 500,
	ActionFight:  20, // kills
	ActionSell:   100,
}

// Material is one tier an action can work with.
type Material struct {
	ID    string   // unique within the action, e.g. "copper"
	Label string   // as it reads after the action's verb: "copper", "oak arrows"
	Level int      // skill level needed
	XP    int      // XP per item gathered, or per input crafted
	Item  ItemKind // what it gives: ore, logs, crops, bars, arrows
	Input ItemKind // crafting only: what it uses from the bank
	Node  int      // gathering only: the object sprite it's gathered from
}

// ActionMaterials lists each action's tiers, lowest first.
var ActionMaterials = map[ActionKind][]Material{
	ActionMine: {
		{ID: "copper", Label: "copper", Level: 1, XP: 10, Item: ItemCopperOre, Node: TileCopperRock},
		{ID: "iron", Label: "iron", Level: 15, XP: 25, Item: ItemIronOre, Node: TileIronRock},
		{ID: "gold", Label: "gold", Level: 30, XP: 45, Item: ItemGoldOre, Node: TileGoldRock},
		{ID: "mithril", Label: "mithril", Level: 45, XP: 80, Item: ItemMithrilOre, Node: TileMithrilRock},
	},
	ActionChop: {
		{ID: "oak", Label: "oak", Level: 1, XP: 10, Item: ItemOakLogs, Node: TileTree1},
		{ID: "pine", Label: "pine", Level: 15, XP: 25, Item: ItemPineLogs, Node: TilePine},
		{ID: "maple", Label: "maple", Level: 30, XP: 45, Item: ItemMapleLogs, Node: TileAutumnTree},
		{ID: "darkoak", Label: "dark oak", Level: 45, XP: 80, Item: ItemDarkOakLogs, Node: TileTree2},
	},
	ActionFarm: {
		{ID: "cabbage", Label: "cabbages", Level: 1, XP: 8, Item: ItemCabbage, Node: TileCabbage},
		{ID: "cotton", Label: "cotton", Level: 10, XP: 18, Item: ItemCotton, Node: TileCotton},
	},
	ActionSmelt: {
		{ID: "copper", Label: "copper bars", Level: 1, XP: 12, Item: ItemCopperBar, Input: ItemCopperOre},
		{ID: "iron", Label: "iron bars", Level: 15, XP: 28, Item: ItemIronIngot, Input: ItemIronOre},
		{ID: "gold", Label: "gold bars", Level: 30, XP: 50, Item: ItemGoldBar, Input: ItemGoldOre},
		{ID: "mithril", Label: "mithril bars", Level: 45, XP: 90, Item: ItemMithrilBar, Input: ItemMithrilOre},
	},
	ActionFletch: {
		{ID: "oak", Label: "oak arrows", Level: 1, XP: 12, Item: ItemOakArrows, Input: ItemOakLogs},
		{ID: "pine", Label: "pine arrows", Level: 15, XP: 28, Item: ItemPineArrows, Input: ItemPineLogs},
		{ID: "maple", Label: "maple arrows", Level: 30, XP: 50, Item: ItemMapleArrows, Input: ItemMapleLogs},
		{ID: "darkoak", Label: "dark oak arrows", Level: 45, XP: 90, Item: ItemDarkOakArrows, Input: ItemDarkOakLogs},
	},
}

func init() {
	// Fighting's "materials" are the enemy kinds.
	for _, e := range Enemies {
		ActionMaterials[ActionFight] = append(ActionMaterials[ActionFight], Material{
			ID: e.ID, Label: e.Plural, Level: 1,
		})
	}
}

// FindMaterial looks up one of an action's materials.
func FindMaterial(kind ActionKind, id string) (Material, bool) {
	for _, m := range ActionMaterials[kind] {
		if m.ID == id {
			return m, true
		}
	}
	return Material{}, false
}

// NodeMaterial is the gathering action and material for an object sprite,
// if it's something that can be gathered.
func NodeMaterial(gid int) (ActionKind, Material, bool) {
	for _, k := range ActionKinds {
		for _, m := range ActionMaterials[k] {
			if m.Node != 0 && m.Node == gid {
				return k, m, true
			}
		}
	}
	return "", Material{}, false
}

type GoalKind string

const (
	GoalAmount GoalKind = "amount" // until the bank holds Amount of the action's product
	GoalTime   GoalKind = "time"   // for Minutes of the character's time
)

const (
	MaxActions    = 5
	MaxGoalAmount = 10000
	MaxGoalMins   = 24 * 60
)

// ActionSpec is one entry in a player's action list.
type ActionSpec struct {
	Kind     ActionKind `json:"kind"`
	Material string     `json:"material"` // a Material ID for this kind
	Place    string     `json:"place"`
	Goal     GoalKind   `json:"goal"`
	Amount   int        `json:"amount"`
	Minutes  int        `json:"minutes"`
}

// Product is the item an action's amount goal counts in the bank.
func (a ActionSpec) Product() ItemKind {
	m, _ := FindMaterial(a.Kind, a.Material)
	return m.Item
}

// Label is how the action reads: "Mine copper", "Smelt iron bars",
// "Sell iron ore".
func (a ActionSpec) Label() string {
	if IsSelling(a.Kind) {
		return ActionNames[a.Kind] + " " + Items[ItemKind(a.Material)].Unit
	}
	m, _ := FindMaterial(a.Kind, a.Material)
	return ActionNames[a.Kind] + " " + m.Label
}

// QueuedAction is an entry in the action list as the server reports it.
type QueuedAction struct {
	ActionSpec
	// For the action that's running: how many of the product are banked,
	// and for a time goal, the seconds left.
	Banked      int     `json:"banked"`
	SecondsLeft float64 `json:"secondsLeft"`
}

// Action list commands, sent by the client.
const (
	ActionOpAdd    = "add"
	ActionOpEdit   = "edit"
	ActionOpRemove = "remove"
	ActionOpUp     = "up"
	ActionOpDown   = "down"
	ActionOpStart  = "start"
	ActionOpStop   = "stop"
)

type ActionCommandContents struct {
	Op     string     `json:"op"`
	Index  int        `json:"index"`
	Action ActionSpec `json:"action"`
}

func NewActionCommandMessage(contents ActionCommandContents) *Message {
	return &Message{
		MessageType: ActionCommandMessage,
		Contents:    contents,
	}
}
