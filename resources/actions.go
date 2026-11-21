package resources

type ActionKind string

const (
	ActionMine    ActionKind = "mine"    // gather ore at a mine
	ActionChop    ActionKind = "chop"    // gather logs in a forest
	ActionFarm    ActionKind = "farm"    // gather crops at a farm
	ActionSmelt   ActionKind = "smelt"   // turn banked ore into bars at a village blacksmith
	ActionFletch  ActionKind = "fletch"  // turn banked logs (and bars) into arrows, bows and staves at a village fletcher
	ActionSmith   ActionKind = "smith"   // turn banked bars into weapons, armour and tools at a village anvil
	ActionImprove ActionKind = "improve" // raise the quality of a piece of gear in the bank, at its station
	ActionFight   ActionKind = "fight"   // fight enemies at a camp for loot
	ActionVisit   ActionKind = "visit"   // walk to a village building and open its window
)

// ActionKinds lists the actions in the order the action window shows them.
var ActionKinds = []ActionKind{ActionFight, ActionMine, ActionChop, ActionFarm, ActionSmelt, ActionSmith, ActionFletch, ActionImprove, ActionVisit}

// ActionNames is the verb for each action; with a material it reads
// "Mine copper", "Fletch oak arrows".
var ActionNames = map[ActionKind]string{
	ActionMine:    "Mine",
	ActionChop:    "Chop",
	ActionFarm:    "Farm",
	ActionSmelt:   "Smelt",
	ActionFletch:  "Fletch",
	ActionSmith:   "Smith",
	ActionImprove: "Improve",
	ActionFight:   "Fight",
	ActionVisit:   "Visit",
}

// ActionPlaceKind is the kind of place each action happens at.
var ActionPlaceKind = map[ActionKind]PlaceKind{
	ActionMine:    PlaceMine,
	ActionChop:    PlaceForest,
	ActionFarm:    PlaceFarm,
	ActionSmelt:   PlaceVillage,
	ActionFletch:  PlaceVillage,
	ActionSmith:   PlaceVillage,
	ActionImprove: PlaceVillage,
	ActionFight:   PlaceCamp,
	ActionVisit:   PlaceVillage,
}

// ActionSkill is the skill each action trains, and which sets its speed
// and the materials it can use.
var ActionSkill = map[ActionKind]Skill{
	ActionMine:   SkillMining,
	ActionChop:   SkillWoodcutting,
	ActionFarm:   SkillFarming,
	ActionSmelt:  SkillSmithing,
	ActionFletch: SkillFletching,
	ActionSmith:  SkillSmithing,
	ActionFight:  SkillMelee, // enemies aren't level-gated; combat skills train from the fight itself
}

// Gathering actions take materials from nodes at a place; crafting actions
// turn one item from the bank into another.
func IsGathering(k ActionKind) bool { return k == ActionMine || k == ActionChop || k == ActionFarm }

// IsCombat reports whether an action is fighting, whose goal counts kills.
func IsCombat(k ActionKind) bool { return k == ActionFight }

// IsImprove reports whether an action improves gear. Its material is the
// item kind; its skill, station and material come from the recipe that
// makes that item (see ImproveRecipe), and its goal counts attempts.
func IsImprove(k ActionKind) bool { return k == ActionImprove }

// IsVisit reports whether an action is a visit to a village building. Its
// material is the Building, and it has no amount: it's done on arrival.
func IsVisit(k ActionKind) bool { return k == ActionVisit }

// Building is a village building a player can visit; standing at one with
// the action list stopped opens its window.
type Building string

const (
	BuildingBank  Building = "bank"
	BuildingStore Building = "store"
	BuildingInn   Building = "inn"
)

// Buildings lists them in the order the action window offers them.
var Buildings = []Building{BuildingBank, BuildingStore, BuildingInn}

var BuildingNames = map[Building]string{
	BuildingBank:  "bank",
	BuildingStore: "store",
	BuildingInn:   "inn",
}

// GoalAmounts are the amounts the action window offers: how many to gather,
// make or defeat. New actions start at the first.
var GoalAmounts = []int{10, 20, 50, 100}

// Material is one tier an action can work with.
type Material struct {
	ID     string       // unique within the action, e.g. "copper"
	Label  string       // as it reads after the action's verb: "copper", "oak arrows"
	Tier   string       // optional row in the action picker, e.g. "Copper"; Short is the label within it
	Short  string       // optional shorter label for the action picker, within its tier's row
	Level  int          // skill level needed
	XP     int          // XP per item gathered, or per craft
	Item   ItemKind     // what it gives: ore, logs, crops, bars, arrows, gear
	Inputs []Ingredient // crafting only: what one craft uses from the bank
	Makes  int          // crafting only: how many one craft makes (0 means 1)
	Node   int          // gathering only: the object sprite it's gathered from
}

// Ingredient is Count of an item used by one craft.
type Ingredient struct {
	Item  ItemKind
	Count int
}

// Output is how many one craft makes.
func (m Material) Output() int { return max(m.Makes, 1) }

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
		{ID: "copper", Label: "copper bars", Level: 1, XP: 12, Item: ItemCopperBar, Inputs: []Ingredient{{ItemCopperOre, 1}}},
		{ID: "iron", Label: "iron bars", Level: 15, XP: 28, Item: ItemIronIngot, Inputs: []Ingredient{{ItemIronOre, 1}}},
		{ID: "gold", Label: "gold bars", Level: 30, XP: 50, Item: ItemGoldBar, Inputs: []Ingredient{{ItemGoldOre, 1}}},
		{ID: "mithril", Label: "mithril bars", Level: 45, XP: 90, Item: ItemMithrilBar, Inputs: []Ingredient{{ItemMithrilOre, 1}}},
	},
	ActionFletch: {
		{ID: "oak", Label: "oak arrows", Tier: "Oak", Short: "arrows", Level: 1, XP: 12, Item: ItemOakArrows, Inputs: []Ingredient{{ItemOakLogs, 1}}, Makes: ArrowsPerLog},
		{ID: "pine", Label: "pine arrows", Tier: "Pine", Short: "arrows", Level: 15, XP: 28, Item: ItemPineArrows, Inputs: []Ingredient{{ItemPineLogs, 1}}, Makes: ArrowsPerLog},
		{ID: "maple", Label: "maple arrows", Tier: "Maple", Short: "arrows", Level: 30, XP: 50, Item: ItemMapleArrows, Inputs: []Ingredient{{ItemMapleLogs, 1}}, Makes: ArrowsPerLog},
		{ID: "darkoak", Label: "dark oak arrows", Tier: "Dark oak", Short: "arrows", Level: 45, XP: 90, Item: ItemDarkOakArrows, Inputs: []Ingredient{{ItemDarkOakLogs, 1}}, Makes: ArrowsPerLog},
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

const (
	MaxActions    = 5
	MaxGoalAmount = 10000
)

// ActionSpec is one entry in a player's action list. It's done once it has
// gathered, made, defeated or sold Amount.
type ActionSpec struct {
	Kind     ActionKind `json:"kind"`
	Material string     `json:"material"` // a Material ID for this kind
	Place    string     `json:"place"`
	Amount   int        `json:"amount"`
}

// Product is the item an action gathers, makes or improves.
func (a ActionSpec) Product() ItemKind {
	if IsImprove(a.Kind) {
		return ItemKind(a.Material)
	}
	m, _ := FindMaterial(a.Kind, a.Material)
	return m.Item
}

// Label is how the action reads: "Mine copper", "Smelt iron bars",
// "Visit the store".
func (a ActionSpec) Label() string {
	if IsVisit(a.Kind) {
		return ActionNames[a.Kind] + " the " + BuildingNames[Building(a.Material)]
	}
	if IsImprove(a.Kind) {
		return ActionNames[a.Kind] + " " + lowerFirst(Items[ItemKind(a.Material)].Name)
	}
	m, _ := FindMaterial(a.Kind, a.Material)
	return ActionNames[a.Kind] + " " + m.Label
}

// QueuedAction is an entry in the action list as the server reports it.
type QueuedAction struct {
	ActionSpec
	// Done is how many it has gathered, made or defeated so far.
	Done int `json:"done"`
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
