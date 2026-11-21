package resources

// Making gear: the anvil's recipes (Smithing) and the fletcher's bows and
// staves (Fletching), added to ActionMaterials. See docs/ECONOMY.md.

// ArrowsPerLog is how many arrows one log makes.
const ArrowsPerLog = 10

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
	{SlotHammer, "hammer", 1, 0}, {SlotKnife, "knife", 1, 0},
	{SlotHead, "helmet", 2, 2}, {SlotWeapon, "sword", 2, 4}, {SlotShield, "shield", 3, 6}, {SlotBody, "body", 4, 8},
}

// fletchWoods are the four woods, with the bar a staff of that tier needs.
var fletchWoods = []struct {
	id, name  string
	log, bar  ItemKind
	level, xp int
	bow       ItemKind
	staves    []ItemKind // mage, then necromancer
}{
	{"oak", "oak", ItemOakLogs, ItemCopperBar, 1, 12, ItemOakShortbow, []ItemKind{ItemApprenticeStaff, ItemBoneStaff}},
	{"pine", "pine", ItemPineLogs, ItemIronIngot, 15, 28, ItemPineShortbow, []ItemKind{ItemAdeptStaff, ItemGraveStaff}},
	{"maple", "maple", ItemMapleLogs, ItemGoldBar, 30, 50, ItemMapleShortbow, []ItemKind{ItemBattleStaff, ItemWraithStaff}},
	{"darkoak", "dark oak", ItemDarkOakLogs, ItemMithrilBar, 45, 90, ItemDarkOakShortbow, []ItemKind{ItemArchmageStaff, ItemLichStaff}},
}

func init() {
	for _, m := range smithMetals {
		for _, piece := range smithPieces {
			item, ok := m.gear[piece.slot]
			if !ok {
				item = ToolItem(upperFirst(m.name), piece.slot)
			}
			ActionMaterials[ActionSmith] = append(ActionMaterials[ActionSmith], Material{
				ID: string(item), Label: m.name + " " + piece.name, Tier: upperFirst(m.name), Short: piece.name,
				Level: m.level + piece.extra, XP: m.xp * piece.bars, Item: item,
				Inputs: []Ingredient{{m.bar, piece.bars}},
			})
		}
	}
	for _, wd := range fletchWoods {
		ActionMaterials[ActionFletch] = append(ActionMaterials[ActionFletch], Material{
			ID: string(wd.bow), Label: wd.name + " shortbow", Tier: upperFirst(wd.name), Short: "shortbow",
			Level: wd.level + 4, XP: wd.xp * 3, Item: wd.bow,
			Inputs: []Ingredient{{wd.log, 3}},
		})
		for _, staff := range wd.staves {
			ActionMaterials[ActionFletch] = append(ActionMaterials[ActionFletch], Material{
				ID: string(staff), Label: lowerFirst(Items[staff].Name), Tier: upperFirst(wd.name), Short: lowerFirst(Items[staff].Name),
				Level: wd.level + 8, XP: wd.xp*2 + wd.xp, Item: staff,
				Inputs: []Ingredient{{wd.log, 2}, {wd.bar, 1}},
			})
		}
	}
}

// ImproveRecipe is the recipe that makes an item, which sets how it's
// improved: at that recipe's station, with its skill, using its first
// ingredient.
func ImproveRecipe(k ItemKind) (ActionKind, Material, bool) {
	if !IsPiece(k) {
		return "", Material{}, false
	}
	for _, kind := range []ActionKind{ActionSmith, ActionFletch} {
		for _, m := range ActionMaterials[kind] {
			if m.Item == k {
				return kind, m, true
			}
		}
	}
	return "", Material{}, false
}

// Improvable lists every item that can be improved, in recipe order.
func Improvable() []ItemKind {
	var out []ItemKind
	for _, kind := range []ActionKind{ActionSmith, ActionFletch} {
		for _, m := range ActionMaterials[kind] {
			if IsPiece(m.Item) {
				out = append(out, m.Item)
			}
		}
	}
	return out
}
