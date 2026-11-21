package resources

// ClassID names a character class.
type ClassID string

const (
	ClassWarrior     ClassID = "warrior"
	ClassArcher      ClassID = "archer"
	ClassMage        ClassID = "mage"
	ClassNecromancer ClassID = "necromancer"
)

// Class is what a new character starts as: their look, gear and a head
// start in the skills they fight with. Any character can train anything
// later; the class is where they begin.
type Class struct {
	ID       ClassID
	Name     string
	Desc     string
	Sprite   [2]int // pixel position in characters.png
	Gear     []Stack
	Levels   map[Skill]int // starting levels above 1
	Equipped []ItemKind    // gear that starts equipped rather than banked
	Armour   ArmourType    // the only armour this class can wear
}

// CanWear reports whether a class can wear a piece: anything that isn't
// armour, or armour of the class's type.
func (c Class) CanWear(k ItemKind) bool {
	a := Equipment[k].Armour
	return a == "" || a == c.Armour
}

// Classes lists the classes in the order the creation screen shows them.
var Classes = []Class{
	{
		ID: ClassWarrior, Name: "Warrior", Sprite: [2]int{8, 0},
		Desc:     "Fights up close with sword and shield. Starts stronger and tougher.",
		Gear:     []Stack{{Kind: ItemCopperSword, Count: 1}, {Kind: ItemCopperShield, Count: 1}},
		Equipped: []ItemKind{ItemCopperSword, ItemCopperShield},
		Levels:   map[Skill]int{SkillMelee: 5, SkillStrength: 5, SkillDefence: 3},
		Armour:   ArmourMetal,
	},
	{
		ID: ClassArcher, Name: "Archer", Sprite: [2]int{0, 24},
		Desc:     "Shoots from range with a bow. Every shot uses an arrow.",
		Gear:     []Stack{{Kind: ItemOakShortbow, Count: 1}, {Kind: ItemOakArrows, Count: 150}, {Kind: ItemLeatherBody, Count: 1}},
		Equipped: []ItemKind{ItemOakShortbow, ItemOakArrows, ItemLeatherBody},
		Levels:   map[Skill]int{SkillRanged: 5, SkillDefence: 3},
		Armour:   ArmourLeather,
	},
	{
		ID: ClassMage, Name: "Mage", Sprite: [2]int{16, 0},
		Desc:     "Casts magic bolts from range with a staff. No ammo needed.",
		Gear:     []Stack{{Kind: ItemApprenticeStaff, Count: 1}, {Kind: ItemClothRobe, Count: 1}},
		Equipped: []ItemKind{ItemApprenticeStaff, ItemClothRobe},
		Levels:   map[Skill]int{SkillMagic: 5, SkillDefence: 3},
		Armour:   ArmourCloth,
	},
	{
		ID: ClassNecromancer, Name: "Necromancer", Sprite: [2]int{0, 16},
		Desc:     "Casts draining magic from range. Hits softer than a mage but heals from the damage dealt.",
		Gear:     []Stack{{Kind: ItemBoneStaff, Count: 1}, {Kind: ItemClothRobe, Count: 1}},
		Equipped: []ItemKind{ItemBoneStaff, ItemClothRobe},
		Levels:   map[Skill]int{SkillMagic: 5, SkillDefence: 3},
		Armour:   ArmourCloth,
	},
}

// FindClass looks up a class.
func FindClass(id ClassID) (Class, bool) {
	for _, c := range Classes {
		if c.ID == id {
			return c, true
		}
	}
	return Class{}, false
}

// CreateCharacterContents is the class a new player picked.
type CreateCharacterContents struct {
	Class ClassID `json:"class"`
}

func NewChooseClassMessage() *Message {
	return &Message{MessageType: ChooseClassMessage}
}

func NewCreateCharacterMessage(c CreateCharacterContents) *Message {
	return &Message{MessageType: CreateCharacterMessage, Contents: c}
}
