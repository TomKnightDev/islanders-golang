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
	Magic     bool // a staff: casts from range using Magic, no ammo
	Drain     int  // percent of damage dealt that heals the wielder
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

	// Mage staves hit hardest; necromancer staves heal you for part of
	// the damage they deal.
	ItemApprenticeStaff ItemKind = "apprenticeStaff"
	ItemAdeptStaff      ItemKind = "adeptStaff"
	ItemBattleStaff     ItemKind = "battleStaff"
	ItemArchmageStaff   ItemKind = "archmageStaff"
	ItemBoneStaff       ItemKind = "boneStaff"
	ItemGraveStaff      ItemKind = "graveStaff"
	ItemWraithStaff     ItemKind = "wraithStaff"
	ItemLichStaff       ItemKind = "lichStaff"
)

// necromancerDrain is how much of the damage a necromancer staff deals comes
// back as health.
const necromancerDrain = 25

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

	ItemApprenticeStaff: {Slot: SlotWeapon, Skill: SkillMagic, Level: 1, Accuracy: 6, Strength: 6, Magic: true, TwoHanded: true},
	ItemAdeptStaff:      {Slot: SlotWeapon, Skill: SkillMagic, Level: 15, Accuracy: 14, Strength: 13, Magic: true, TwoHanded: true},
	ItemBattleStaff:     {Slot: SlotWeapon, Skill: SkillMagic, Level: 30, Accuracy: 24, Strength: 22, Magic: true, TwoHanded: true},
	ItemArchmageStaff:   {Slot: SlotWeapon, Skill: SkillMagic, Level: 45, Accuracy: 36, Strength: 33, Magic: true, TwoHanded: true},
	ItemBoneStaff:       {Slot: SlotWeapon, Skill: SkillMagic, Level: 1, Accuracy: 6, Strength: 4, Magic: true, Drain: necromancerDrain, TwoHanded: true},
	ItemGraveStaff:      {Slot: SlotWeapon, Skill: SkillMagic, Level: 15, Accuracy: 14, Strength: 10, Magic: true, Drain: necromancerDrain, TwoHanded: true},
	ItemWraithStaff:     {Slot: SlotWeapon, Skill: SkillMagic, Level: 30, Accuracy: 24, Strength: 18, Magic: true, Drain: necromancerDrain, TwoHanded: true},
	ItemLichStaff:       {Slot: SlotWeapon, Skill: SkillMagic, Level: 45, Accuracy: 36, Strength: 28, Magic: true, Drain: necromancerDrain, TwoHanded: true},

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
		ItemApprenticeStaff: "Apprentice staff", ItemAdeptStaff: "Adept staff", ItemBattleStaff: "Battle staff", ItemArchmageStaff: "Archmage staff",
		ItemBoneStaff: "Bone staff", ItemGraveStaff: "Grave staff", ItemWraithStaff: "Wraith staff", ItemLichStaff: "Lich staff",
	}
	for k, name := range gear {
		d := Equipment[k]
		desc := "Equip it from the bank in the Equipment window."
		switch {
		case d.Drain > 0:
			desc = "A necromancer's staff, two-handed. Casts from range with Magic and heals you for part of the damage. " + desc
		case d.Magic:
			desc = "A mage's staff, two-handed. Casts from range with Magic; no ammo needed. " + desc
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
	Ranged                      bool // fighting with a bow
	Magic                       bool // fighting with a staff
	Drain                       int  // percent of damage healed
}

// Style names how a character fights with their weapon.
func (b Bonuses) Style() string {
	switch {
	case b.Drain > 0:
		return "Necromancy"
	case b.Magic:
		return "Magic"
	case b.Ranged:
		return "Ranged"
	}
	return "Melee"
}

// Skills are the skills a fighting style uses for accuracy and max hit.
func (b Bonuses) Skills() (accuracy, strength Skill) {
	switch {
	case b.Magic:
		return SkillMagic, SkillMagic
	case b.Ranged:
		return SkillRanged, SkillRanged
	}
	return SkillAttack, SkillStrength
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
			b.Ranged, b.Magic, b.Drain = d.Ranged, d.Magic, d.Drain
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
