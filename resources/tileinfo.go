package resources

import "fmt"

// TileInfo is what the player is told when they point at something.
type TileInfo struct {
	Name string
	Desc string
}

// ObjectInfo describes the things on the object layer, by sprite.
// Gatherable nodes (rocks, trees, crops) are described from their material.
var ObjectInfo = map[int]TileInfo{
	TileBank:             {"Bank", "Deposits what you carry into your account. Every village's bank opens the same account."},
	TileBlacksmith:       {"Blacksmith", "Smelts ore into bars, training Smithing. Add Smelt for this village in the Actions window."},
	TileFletcher:         {"Fletcher's hut", "Turns logs into arrows, 10 per log, training Fletching. Add Fletch for this village in the Actions window."},
	TileAnvil:            {"Anvil", "For crafting equipment from bars. Coming soon."},
	TileStore:            {"General store", "Sells food, starter gear and basic materials (open the Shop), and buys anything of value. Add a Sell action to carry goods here; prices drop as a store buys more."},
	TileInn:              {"Inn", "Tired characters relax here to recover stamina. They come here on their own when they're worn out, then go back to work."},
	TileIronRockDepleted: {"Depleted rock", "Mined out. It refills with ore soon."},
	TileFelledTree:       {"Felled tree", "Chopped down. It grows back soon."},
	TileHarvested:        {"Harvested plot", "Picked clean. It grows back soon."},
	TileBush:             {"Bush", "Too small to chop."},
	TileCottage:          {"Cottage", "Someone's home."},
	TileTower:            {"Tower", "Oakhaven's old watchtower."},
	209:                  {"Hut", "Rough shelter. Whoever lives here won't welcome you."},
	122:                  {"Ruins", "Crumbling stone. Something still walks here."},
}

// NodeInfo describes a gatherable sprite: what it is, what it gives, and
// the level needed.
func NodeInfo(gid int) (TileInfo, bool) {
	kind, m, ok := NodeMaterial(gid)
	if !ok {
		return TileInfo{}, false
	}
	name := ItemNames[m.Item]
	switch kind {
	case ActionMine:
		name = Items[m.Item].Name[:len(Items[m.Item].Name)-len(" ore")] + " rock"
	case ActionChop:
		name = Items[m.Item].Name[:len(Items[m.Item].Name)-len(" logs")] + " tree"
	}
	desc := fmt.Sprintf("%s it for %s. Needs %s level %d.", ActionNames[kind], ItemUnits[m.Item], SkillNames[ActionSkill[kind]], m.Level)
	return TileInfo{Name: name, Desc: desc}, true
}
