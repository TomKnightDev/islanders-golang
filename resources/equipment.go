package resources

type EquipSlot string

const (
	SlotWeapon EquipSlot = "weapon"
	SlotAmmo   EquipSlot = "ammo"
	SlotHead   EquipSlot = "head"
	SlotBody   EquipSlot = "body"
	SlotShield EquipSlot = "shield"
)

// EquipSlots lists the slots in the order the equipment window shows them.
var EquipSlots = []EquipSlot{SlotWeapon, SlotAmmo, SlotHead, SlotBody, SlotShield}

var SlotNames = map[EquipSlot]string{
	SlotWeapon: "Weapon",
	SlotAmmo:   "Ammo",
	SlotHead:   "Head",
	SlotBody:   "Body",
	SlotShield: "Shield",
}

// EquipDef is what an item does when worn or wielded.
type EquipDef struct {
	Slot      EquipSlot
	Skill     Skill // skill whose level it needs
	Level     int
	Accuracy  int  // attack bonus (melee weapons, bows)
	Strength  int  // melee strength, or ranged strength for arrows
	Defence   int  // armour bonus
	Ranged    bool // a bow: attacks from range and uses arrows
	TwoHanded bool // can't be used with a shield
}

// Equipment items.
const (
	ItemTrainingSword   ItemKind = "trainingSword"
	ItemCopperSword     ItemKind = "copperSword"
	ItemIronSword       ItemKind = "ironSword"
	ItemGoldSword       ItemKind = "goldSword"
	ItemMithrilSword    ItemKind = "mithrilSword"
	ItemOakShortbow     ItemKind = "oakShortbow"
	ItemPineShortbow    ItemKind = "pineShortbow"
	ItemMapleShortbow   ItemKind = "mapleShortbow"
	ItemDarkOakShortbow ItemKind = "darkOakShortbow"
	ItemCopperHelmet    ItemKind = "copperHelmet"
	ItemIronHelmet      ItemKind = "ironHelmet"
	ItemGoldHelmet      ItemKind = "goldHelmet"
	ItemMithrilHelmet   ItemKind = "mithrilHelmet"
	ItemCopperBody      ItemKind = "copperPlatebody"
	ItemIronBody        ItemKind = "ironPlatebody"
	ItemGoldBody        ItemKind = "goldPlatebody"
	ItemMithrilBody     ItemKind = "mithrilPlatebody"
	ItemCopperShield    ItemKind = "copperShield"
	ItemIronShield      ItemKind = "ironShield"
	ItemGoldShield      ItemKind = "goldShield"
	ItemMithrilShield   ItemKind = "mithrilShield"
)

var Equipment = map[ItemKind]EquipDef{
	ItemTrainingSword: {Slot: SlotWeapon, Skill: SkillAttack, Level: 1, Accuracy: 0, Strength: 2},
	ItemCopperSword:   {Slot: SlotWeapon, Skill: SkillAttack, Level: 1, Accuracy: 6, Strength: 6},
	ItemIronSword:     {Slot: SlotWeapon, Skill: SkillAttack, Level: 15, Accuracy: 14, Strength: 13},
	ItemGoldSword:     {Slot: SlotWeapon, Skill: SkillAttack, Level: 30, Accuracy: 24, Strength: 22},
	ItemMithrilSword:  {Slot: SlotWeapon, Skill: SkillAttack, Level: 45, Accuracy: 36, Strength: 33},

	ItemOakShortbow:     {Slot: SlotWeapon, Skill: SkillRanged, Level: 1, Accuracy: 8, Ranged: true, TwoHanded: true},
	ItemPineShortbow:    {Slot: SlotWeapon, Skill: SkillRanged, Level: 15, Accuracy: 16, Ranged: true, TwoHanded: true},
	ItemMapleShortbow:   {Slot: SlotWeapon, Skill: SkillRanged, Level: 30, Accuracy: 27, Ranged: true, TwoHanded: true},
	ItemDarkOakShortbow: {Slot: SlotWeapon, Skill: SkillRanged, Level: 45, Accuracy: 40, Ranged: true, TwoHanded: true},

	ItemOakArrows:     {Slot: SlotAmmo, Skill: SkillRanged, Level: 1, Strength: 7},
	ItemPineArrows:    {Slot: SlotAmmo, Skill: SkillRanged, Level: 15, Strength: 14},
	ItemMapleArrows:   {Slot: SlotAmmo, Skill: SkillRanged, Level: 30, Strength: 23},
	ItemDarkOakArrows: {Slot: SlotAmmo, Skill: SkillRanged, Level: 45, Strength: 34},

	ItemCopperHelmet:  {Slot: SlotHead, Skill: SkillDefence, Level: 1, Defence: 4},
	ItemIronHelmet:    {Slot: SlotHead, Skill: SkillDefence, Level: 15, Defence: 9},
	ItemGoldHelmet:    {Slot: SlotHead, Skill: SkillDefence, Level: 30, Defence: 15},
	ItemMithrilHelmet: {Slot: SlotHead, Skill: SkillDefence, Level: 45, Defence: 22},
	ItemCopperBody:    {Slot: SlotBody, Skill: SkillDefence, Level: 1, Defence: 10},
	ItemIronBody:      {Slot: SlotBody, Skill: SkillDefence, Level: 15, Defence: 22},
	ItemGoldBody:      {Slot: SlotBody, Skill: SkillDefence, Level: 30, Defence: 36},
	ItemMithrilBody:   {Slot: SlotBody, Skill: SkillDefence, Level: 45, Defence: 52},
	ItemCopperShield:  {Slot: SlotShield, Skill: SkillDefence, Level: 1, Defence: 6},
	ItemIronShield:    {Slot: SlotShield, Skill: SkillDefence, Level: 15, Defence: 13},
	ItemGoldShield:    {Slot: SlotShield, Skill: SkillDefence, Level: 30, Defence: 21},
	ItemMithrilShield: {Slot: SlotShield, Skill: SkillDefence, Level: 45, Defence: 31},
}

func init() {
	gear := map[ItemKind]string{
		ItemTrainingSword: "Training sword", ItemCopperSword: "Copper sword", ItemIronSword: "Iron sword",
		ItemGoldSword: "Gold sword", ItemMithrilSword: "Mithril sword",
		ItemOakShortbow: "Oak shortbow", ItemPineShortbow: "Pine shortbow", ItemMapleShortbow: "Maple shortbow",
		ItemDarkOakShortbow: "Dark oak shortbow",
		ItemCopperHelmet:    "Copper helmet", ItemIronHelmet: "Iron helmet", ItemGoldHelmet: "Gold helmet", ItemMithrilHelmet: "Mithril helmet",
		ItemCopperBody: "Copper platebody", ItemIronBody: "Iron platebody", ItemGoldBody: "Gold platebody", ItemMithrilBody: "Mithril platebody",
		ItemCopperShield: "Copper shield", ItemIronShield: "Iron shield", ItemGoldShield: "Gold shield", ItemMithrilShield: "Mithril shield",
	}
	for k, name := range gear {
		d := Equipment[k]
		desc := "Equip it from the bank in the Equipment window."
		switch {
		case d.Ranged:
			desc = "A two-handed bow. Fires arrows from your ammo slot. " + desc
		case d.Slot == SlotWeapon:
			desc = "A sword for melee fighting. " + desc
		default:
			desc = "Armour. Enemies hit you less often. " + desc
		}
		Items[k] = ItemDef{Name: name, Unit: name, Desc: desc, Stack: 1}
		StackLimits[k] = 1
		ItemNames[k] = name
		ItemUnits[k] = name
		ItemDescriptions[k] = desc
	}
}

// Bonuses are the summed bonuses of what a character has equipped.
type Bonuses struct {
	Accuracy, Strength, Defence int
	Ranged                      bool
}

// EquipmentBonuses adds up the bonuses of equipped items.
func EquipmentBonuses(equipped map[EquipSlot]Stack) Bonuses {
	var b Bonuses
	for slot, s := range equipped {
		d, ok := Equipment[s.Kind]
		if !ok || s.Count == 0 {
			continue
		}
		b.Defence += d.Defence
		switch slot {
		case SlotWeapon:
			b.Accuracy += d.Accuracy
			b.Ranged = d.Ranged
			if !d.Ranged {
				b.Strength += d.Strength
			}
		case SlotAmmo:
			// Arrows only count when shot from a bow; added below.
		}
	}
	if b.Ranged {
		if ammo, ok := equipped[SlotAmmo]; ok && ammo.Count > 0 {
			b.Strength += Equipment[ammo.Kind].Strength
		}
	}
	return b
}

// Equipment window commands, sent by the client.
const (
	EquipOpEquip   = "equip"   // from the bank into its slot
	EquipOpUnequip = "unequip" // from a slot back to the bank
)

type EquipCommandContents struct {
	Op   string    `json:"op"`
	Item ItemKind  `json:"item"` // for equip
	Slot EquipSlot `json:"slot"` // for unequip
}

func NewEquipCommandMessage(c EquipCommandContents) *Message {
	return &Message{MessageType: EquipCommandMessage, Contents: c}
}

// EquippedItem is one filled slot as the server reports it.
type EquippedItem struct {
	Slot  EquipSlot `json:"slot"`
	Item  ItemKind  `json:"item"`
	Count int       `json:"count"`
}
