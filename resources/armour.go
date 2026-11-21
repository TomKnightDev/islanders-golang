package resources

// Leather armour for archers and cloth armour for mages and necromancers,
// in four tiers like metal. Leather comes from deer and enemies, cloth from
// cotton woven at a tailor or from enemies; the tailor makes both.

// softArmour is one tier of leather or cloth armour.
type softArmour struct {
	head, body           ItemKind
	headName, bodyName   string
	level                int
	headDef, bodyDef     int
	headValue, bodyValue int
	headUses, bodyUses   int // leather or cloth to make it
	xp                   int // per unit used
}

const (
	ItemLeatherCap         ItemKind = "leatherCap"
	ItemLeatherBody        ItemKind = "leatherBody"
	ItemHardLeatherCap     ItemKind = "hardLeatherCap"
	ItemHardLeatherBody    ItemKind = "hardLeatherBody"
	ItemStuddedLeatherCap  ItemKind = "studdedLeatherCap"
	ItemStuddedLeatherBody ItemKind = "studdedLeatherBody"
	ItemRangerCap          ItemKind = "rangerCap"
	ItemRangerBody         ItemKind = "rangerBody"

	ItemClothHood       ItemKind = "clothHood"
	ItemClothRobe       ItemKind = "clothRobe"
	ItemWovenHood       ItemKind = "wovenHood"
	ItemWovenRobe       ItemKind = "wovenRobe"
	ItemEmbroideredHood ItemKind = "embroideredHood"
	ItemEmbroideredRobe ItemKind = "embroideredRobe"
	ItemEnchantedHood   ItemKind = "enchantedHood"
	ItemEnchantedRobe   ItemKind = "enchantedRobe"
)

// Leather guards a little less than metal, and cloth less again: archers
// and casters fight from range, so they're hit less often.
var leatherArmour = []softArmour{
	{ItemLeatherCap, ItemLeatherBody, "Leather cap", "Leather body", 1, 3, 8, 25, 50, 2, 4, 12},
	{ItemHardLeatherCap, ItemHardLeatherBody, "Hardened leather cap", "Hardened leather body", 15, 7, 18, 75, 150, 3, 6, 28},
	{ItemStuddedLeatherCap, ItemStuddedLeatherBody, "Studded leather cap", "Studded leather body", 30, 12, 29, 210, 420, 4, 8, 50},
	{ItemRangerCap, ItemRangerBody, "Ranger's cap", "Ranger's body", 45, 18, 42, 600, 1200, 5, 10, 90},
}

var clothArmour = []softArmour{
	{ItemClothHood, ItemClothRobe, "Cloth hood", "Cloth robe", 1, 2, 6, 20, 40, 2, 4, 12},
	{ItemWovenHood, ItemWovenRobe, "Woven hood", "Woven robe", 15, 5, 13, 60, 120, 3, 6, 28},
	{ItemEmbroideredHood, ItemEmbroideredRobe, "Embroidered hood", "Embroidered robe", 30, 9, 22, 180, 350, 4, 8, 50},
	{ItemEnchantedHood, ItemEnchantedRobe, "Enchanted hood", "Enchanted robe", 45, 13, 32, 500, 1000, 5, 10, 90},
}

func init() {
	add := func(tiers []softArmour, armour ArmourType, material ItemKind, who string) {
		for _, a := range tiers {
			for _, piece := range []struct {
				item       ItemKind
				name       string
				slot       EquipSlot
				def, value int
				uses       int
				extra      int // level above the tier's
			}{
				// Caps and hoods open at the tier's level, so Tailoring 1 can
				// make something; bodies and robes come a few levels later.
				{a.head, a.headName, SlotHead, a.headDef, a.headValue, a.headUses, 0},
				{a.body, a.bodyName, SlotBody, a.bodyDef, a.bodyValue, a.bodyUses, 4},
			} {
				Equipment[piece.item] = EquipDef{Slot: piece.slot, Skill: SkillDefence, Level: a.level, Defence: piece.def, Armour: armour}
				desc := ArmourNames[armour] + " armour, for " + who + " only. Enemies hit you less often."
				Items[piece.item] = ItemDef{Name: piece.name, Unit: piece.name, Desc: desc, Stack: 1}
				StackLimits[piece.item] = 1
				ItemNames[piece.item] = piece.name
				ItemUnits[piece.item] = piece.name
				ItemDescriptions[piece.item] = desc
				Values[piece.item] = piece.value
				short := "hood"
				if armour == ArmourLeather {
					short = "cap"
				}
				if piece.slot == SlotBody {
					short = "body"
					if armour == ArmourCloth {
						short = "robe"
					}
				}
				GearRecipes = append(GearRecipes, GearRecipe{
					Material: Material{
						ID: string(piece.item), Label: lowerFirst(piece.name), Tier: tierName(piece.name), Short: short,
						Level: a.level + piece.extra, XP: a.xp * piece.uses, Item: piece.item,
						Inputs: []Ingredient{{material, piece.uses}},
					},
					At: BuildingTailor, Skill: SkillTailoring, Tool: SlotNeedle,
				})
			}
		}
	}
	add(leatherArmour, ArmourLeather, ItemLeather, "archers")
	add(clothArmour, ArmourCloth, ItemCloth, "mages and necromancers")
}

// tierName is a piece's name without its last word: "Hardened leather cap"
// is in the "Hardened leather" row.
func tierName(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == ' ' {
			return name[:i]
		}
	}
	return name
}
