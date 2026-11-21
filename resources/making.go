package resources

// Making gear: at a village anvil (Smithing) or fletcher's hut
// (Fletching), by hand from the backpack, in the station window. The same
// recipes set how a piece is improved. See docs/DESIGN.md.

// ArrowsPerLog is how many arrows one log makes.
const ArrowsPerLog = 10

// GearRecipe is a piece of gear or a tool made at a station.
type GearRecipe struct {
	Material           // Item, Inputs, Level, XP, and Tier/Short for laying out the window
	At       Building  // BuildingAnvil, BuildingFletcher or BuildingTailor
	Skill    Skill     // the skill it needs and trains
	Tool     EquipSlot // the tool it's made and improved with
}

// GearRecipes lists every recipe, a tier at a time.
var GearRecipes []GearRecipe

// smithMetals are the four metals, with their bar, base level and XP per
// bar.
var smithMetals = []struct {
	name  string
	bar   ItemKind
	level int
	xp    int
	gear  map[EquipSlot]ItemKind
}{
	{"copper", ItemCopperBar, 1, 12, map[EquipSlot]ItemKind{SlotWeapon: ItemCopperSword, SlotHead: ItemCopperHelmet, SlotBody: ItemCopperBody, SlotShield: ItemCopperShield}},
	{"iron", ItemIronIngot, 15, 28, map[EquipSlot]ItemKind{SlotWeapon: ItemIronSword, SlotHead: ItemIronHelmet, SlotBody: ItemIronBody, SlotShield: ItemIronShield}},
	{"gold", ItemGoldBar, 30, 50, map[EquipSlot]ItemKind{SlotWeapon: ItemGoldSword, SlotHead: ItemGoldHelmet, SlotBody: ItemGoldBody, SlotShield: ItemGoldShield}},
	{"mithril", ItemMithrilBar, 45, 90, map[EquipSlot]ItemKind{SlotWeapon: ItemMithrilSword, SlotHead: ItemMithrilHelmet, SlotBody: ItemMithrilBody, SlotShield: ItemMithrilShield}},
}

// smithPieces are what the anvil makes from each metal: bars used and the
// level above the metal's base level.
var smithPieces = []struct {
	slot  EquipSlot
	name  string
	bars  int
	extra int
}{
	{SlotPickaxe, "pickaxe", 1, 0}, {SlotHatchet, "hatchet", 1, 0}, {SlotSickle, "sickle", 1, 0},
	{SlotHammer, "hammer", 1, 0}, {SlotKnife, "knife", 1, 0}, {SlotNeedle, "needle", 1, 0},
	{SlotHead, "helmet", 2, 2}, {SlotWeapon, "sword", 2, 4}, {SlotShield, "shield", 3, 6}, {SlotBody, "body", 4, 8},
}

// fletchWoods are the four woods, with the bar a staff of that tier needs.
var fletchWoods = []struct {
	name      string
	log, bar  ItemKind
	level, xp int
	bow       ItemKind
	staves    []ItemKind // mage, then necromancer
}{
	{"oak", ItemOakLogs, ItemCopperBar, 1, 12, ItemOakShortbow, []ItemKind{ItemApprenticeStaff, ItemBoneStaff}},
	{"pine", ItemPineLogs, ItemIronIngot, 15, 28, ItemPineShortbow, []ItemKind{ItemAdeptStaff, ItemGraveStaff}},
	{"maple", ItemMapleLogs, ItemGoldBar, 30, 50, ItemMapleShortbow, []ItemKind{ItemBattleStaff, ItemWraithStaff}},
	{"dark oak", ItemDarkOakLogs, ItemMithrilBar, 45, 90, ItemDarkOakShortbow, []ItemKind{ItemArchmageStaff, ItemLichStaff}},
}

func init() {
	for _, m := range smithMetals {
		for _, piece := range smithPieces {
			item, ok := m.gear[piece.slot]
			if !ok {
				item = ToolItem(upperFirst(m.name), piece.slot)
			}
			GearRecipes = append(GearRecipes, GearRecipe{
				Material: Material{
					ID: string(item), Label: m.name + " " + piece.name, Tier: upperFirst(m.name), Short: piece.name,
					Level: m.level + piece.extra, XP: m.xp * piece.bars, Item: item,
					Inputs: []Ingredient{{m.bar, piece.bars}},
				},
				At: BuildingAnvil, Skill: SkillSmithing, Tool: SlotHammer,
			})
		}
	}
	for _, wd := range fletchWoods {
		GearRecipes = append(GearRecipes, GearRecipe{
			Material: Material{
				ID: string(wd.bow), Label: wd.name + " shortbow", Tier: upperFirst(wd.name), Short: "shortbow",
				Level: wd.level + 4, XP: wd.xp * 3, Item: wd.bow,
				Inputs: []Ingredient{{wd.log, 3}},
			},
			At: BuildingFletcher, Skill: SkillFletching, Tool: SlotKnife,
		})
		for _, staff := range wd.staves {
			GearRecipes = append(GearRecipes, GearRecipe{
				Material: Material{
					ID: string(staff), Label: lowerFirst(Items[staff].Name), Tier: upperFirst(wd.name), Short: lowerFirst(Items[staff].Name),
					Level: wd.level + 8, XP: wd.xp * 3, Item: staff,
					Inputs: []Ingredient{{wd.log, 2}, {wd.bar, 1}},
				},
				At: BuildingFletcher, Skill: SkillFletching, Tool: SlotKnife,
			})
		}
	}
}

// FindGearRecipe is the recipe that makes an item, which also sets how
// it's improved: at the same station, with the same skill and tool, using
// its first ingredient.
func FindGearRecipe(k ItemKind) (GearRecipe, bool) {
	for _, r := range GearRecipes {
		if r.Item == k {
			return r, true
		}
	}
	return GearRecipe{}, false
}

// ImproveUnitAttempts is how many improving attempts one unit of material
// (a bar or a log) lasts.
const ImproveUnitAttempts = 4

// ImproveToolWear is the damage a tool takes from each improving attempt;
// at 100 it breaks.
const ImproveToolWear = 1

// Station window commands, sent by the client while at an anvil or the
// fletcher's hut: make Count of Item, or improve the piece with ID.
const (
	StationOpMake    = "make"
	StationOpImprove = "improve"
	StationOpStop    = "stop"
)

type StationCommandContents struct {
	Op    string   `json:"op"`
	Item  ItemKind `json:"item"`  // make: the recipe's item
	Count int      `json:"count"` // make: how many
	ID    uint32   `json:"id"`    // improve: the piece in the backpack
}

func NewStationCommandMessage(c StationCommandContents) *Message {
	return &Message{MessageType: StationCommandMessage, Contents: c}
}
