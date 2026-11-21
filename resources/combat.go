package resources

import "math"

// The combat formulas are scaled 10x so health and hits
// are in the tens and hundreds.

// AttackRoll is how accurate an attacker is.
func AttackRoll(level, accuracyBonus int) int {
	return (level + 8) * (accuracyBonus + 64)
}

// DefenceRoll is how hard a defender is to hit.
func DefenceRoll(level, defenceBonus int) int {
	return (level + 9) * (defenceBonus + 64)
}

// HitChance is the chance an attack roll beats a defence roll.
func HitChance(attack, defence int) float64 {
	a, d := float64(attack), float64(defence)
	if a > d {
		return 1 - (d+2)/(2*(a+1))
	}
	return a / (2 * (d + 1))
}

// MaxHit is the most damage one hit can do.
func MaxHit(level, strengthBonus int) int {
	return int(math.Floor(0.5 + float64((level+8)*(strengthBonus+64))/64))
}

// EnemyKind is one kind of enemy found at camps.
type EnemyKind struct {
	ID      string
	Name    string // singular, e.g. "Goblin"
	Plural  string // e.g. "goblins"
	Level   int    // shown to players as a guide
	Health  int
	Attack  int // attack and defence levels
	Defence int
	MaxHit  int
	Sprite  [2]int // pixel position in characters.png
	Loot    []LootDrop
	Coins   [2]int // coins dropped per kill, from–to
}

// LootDrop is one possible drop: a chance per kill and a quantity range.
type LootDrop struct {
	Item     ItemKind
	Chance   float64
	Min, Max int
}

var Enemies = []EnemyKind{
	{ID: "goblin", Name: "Goblin", Plural: "goblins", Level: 3, Health: 50, Attack: 3, Defence: 2, MaxHit: 8, Sprite: [2]int{0, 40}, Coins: [2]int{2, 8},
		Loot: []LootDrop{
			{ItemLeather, 0.2, 1, 2}, {ItemCloth, 0.15, 1, 2}, {ItemLeatherCap, 0.03, 1, 1}, {ItemClothHood, 0.03, 1, 1}, {ItemLeatherBody, 0.02, 1, 1}, {ItemClothRobe, 0.02, 1, 1},
			{ItemCopperOre, 0.4, 1, 3}, {ItemCabbage, 0.25, 1, 2}, {ItemOakArrows, 0.15, 5, 15},
			{ItemCopperSword, 0.04, 1, 1}, {ItemCopperHelmet, 0.04, 1, 1}, {ItemCopperShield, 0.03, 1, 1},
			{ItemCopperBody, 0.02, 1, 1}, {ItemOakShortbow, 0.03, 1, 1},
			{ItemApprenticeStaff, 0.02, 1, 1}, {ItemBoneStaff, 0.02, 1, 1},
		}},
	{ID: "bandit", Name: "Bandit", Plural: "bandits", Level: 12, Health: 120, Attack: 14, Defence: 12, MaxHit: 15, Sprite: [2]int{16, 192}, Coins: [2]int{8, 20},
		Loot: []LootDrop{
			{ItemLeather, 0.25, 1, 3}, {ItemCloth, 0.2, 1, 3}, {ItemHardLeatherCap, 0.025, 1, 1}, {ItemWovenHood, 0.025, 1, 1}, {ItemHardLeatherBody, 0.015, 1, 1}, {ItemWovenRobe, 0.015, 1, 1},
			{ItemIronOre, 0.35, 1, 3}, {ItemCopperBar, 0.2, 1, 2}, {ItemCabbage, 0.2, 1, 3}, {ItemPineArrows, 0.15, 5, 15},
			{ItemIronSword, 0.03, 1, 1}, {ItemIronHelmet, 0.03, 1, 1}, {ItemIronShield, 0.025, 1, 1},
			{ItemIronBody, 0.015, 1, 1}, {ItemPineShortbow, 0.025, 1, 1}, {ItemCopperBody, 0.03, 1, 1},
			{ItemAdeptStaff, 0.02, 1, 1}, {ItemGraveStaff, 0.02, 1, 1},
		}},
	{ID: "orc", Name: "Orc warrior", Plural: "orc warriors", Level: 28, Health: 250, Attack: 30, Defence: 28, MaxHit: 25, Sprite: [2]int{80, 64}, Coins: [2]int{20, 50},
		Loot: []LootDrop{
			{ItemLeather, 0.3, 2, 3}, {ItemCloth, 0.2, 1, 3}, {ItemStuddedLeatherCap, 0.02, 1, 1}, {ItemEmbroideredHood, 0.02, 1, 1}, {ItemStuddedLeatherBody, 0.012, 1, 1}, {ItemEmbroideredRobe, 0.012, 1, 1},
			{ItemGoldOre, 0.3, 1, 3}, {ItemIronIngot, 0.2, 1, 2}, {ItemCabbage, 0.2, 2, 3}, {ItemMapleArrows, 0.15, 5, 15},
			{ItemGoldSword, 0.025, 1, 1}, {ItemGoldHelmet, 0.025, 1, 1}, {ItemGoldShield, 0.02, 1, 1},
			{ItemGoldBody, 0.012, 1, 1}, {ItemMapleShortbow, 0.02, 1, 1},
			{ItemBattleStaff, 0.015, 1, 1}, {ItemWraithStaff, 0.015, 1, 1},
		}},
	{ID: "skeleton", Name: "Skeleton", Plural: "skeletons", Level: 45, Health: 400, Attack: 48, Defence: 45, MaxHit: 35, Sprite: [2]int{0, 48}, Coins: [2]int{50, 120},
		Loot: []LootDrop{
			{ItemCloth, 0.3, 2, 4}, {ItemLeather, 0.15, 1, 3}, {ItemRangerCap, 0.015, 1, 1}, {ItemEnchantedHood, 0.015, 1, 1}, {ItemRangerBody, 0.01, 1, 1}, {ItemEnchantedRobe, 0.01, 1, 1},
			{ItemMithrilOre, 0.25, 1, 2}, {ItemGoldBar, 0.2, 1, 2}, {ItemCabbage, 0.2, 2, 4}, {ItemDarkOakArrows, 0.15, 5, 15},
			{ItemMithrilSword, 0.02, 1, 1}, {ItemMithrilHelmet, 0.02, 1, 1}, {ItemMithrilShield, 0.015, 1, 1},
			{ItemMithrilBody, 0.01, 1, 1}, {ItemDarkOakShortbow, 0.015, 1, 1},
			{ItemArchmageStaff, 0.01, 1, 1}, {ItemLichStaff, 0.01, 1, 1},
		}},
}

// Deer wander the named forests. They're hunted, not fought: they flee
// rather than fight back, and give leather and venison.
var Deer = EnemyKind{ID: "deer", Name: "Deer", Plural: "deer", Level: 1, Health: 1, Sprite: [2]int{72, 104}}

// FindEnemy looks up an enemy kind.
func FindEnemy(id string) (EnemyKind, bool) {
	for _, e := range Enemies {
		if e.ID == id {
			return e, true
		}
	}
	return EnemyKind{}, false
}
