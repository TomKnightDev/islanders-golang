package resources

// TileInfo is what the player is told when they point at something.
type TileInfo struct {
	Name string
	Desc string
}

// ObjectInfo describes the things on the object layer, by sprite.
var ObjectInfo = map[int]TileInfo{
	TileBank:             {"Bank", "Deposits what you carry into your account. Every village's bank opens the same account."},
	TileBlacksmith:       {"Blacksmith", "Smelts iron ore into iron bars. Add Smelt iron for this village in the Actions window."},
	TileFletcher:         {"Fletcher's hut", "Turns logs into arrows, 10 per log. Add Fletch arrows for this village in the Actions window."},
	TileAnvil:            {"Anvil", "For crafting equipment from iron bars. Coming soon."},
	TileIronRock:         {"Iron rock", "Mine it for iron ore. Add Mine iron for this mine in the Actions window."},
	TileIronRockDepleted: {"Depleted rock", "Mined out. It refills with ore soon."},
	TileTree1:            {"Oak", "A tree."},
	TileTree2:            {"Dark oak", "A tree."},
	TilePine:             {"Pine", "A tree."},
	TileAutumnTree:       {"Autumn tree", "A tree."},
	TileFelledTree:       {"Felled tree", "Chopped down. It grows back soon."},
	TileBush:             {"Bush", "Too small to chop."},
	TileCottage:          {"Cottage", "Someone's home."},
	TileTower:            {"Tower", "Oakhaven's old watchtower."},
}
