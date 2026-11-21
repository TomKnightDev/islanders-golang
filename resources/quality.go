package resources

import "math/rand"

// Item quality. See docs/CRAFTING-QUALITY.md and docs/ECONOMY.md.

const (
	DefaultQuality   = 10 // materials and gear of unknown quality
	StarterQuality   = 15 // starting kit and store-bought gear
	ToolStartQuality = 10 // the starting set of copper tools
	MaxQuality       = 100
)

// EffectiveQuality is quality reduced by damage.
func EffectiveQuality(quality, damage int) int {
	return quality * (100 - damage) / 100
}

// GearFactor scales a weapon's or armour's bonuses by its effective
// quality: 70% at quality 0, 130% at quality 100.
func GearFactor(effectiveQuality int) float64 {
	return 0.7 + 0.6*float64(effectiveQuality)/100
}

// ToolSpeed is how long a step takes with a tool, as a share of the time
// without one: better metal and higher quality are faster, up to 30% off
// for mithril at quality 100. With no tool, work takes twice as long.
func ToolSpeed(tool *Stack) float64 {
	if tool == nil {
		return 2
	}
	d := Equipment[tool.Kind]
	q := EffectiveQuality(tool.Quality, tool.Damage)
	return 1 - (float64(d.TierBonus)+15*float64(q)/100)/100
}

// GatheredQuality is the quality of something gathered at a skill level with
// a tool. Without one, it's only ever 1–10.
func GatheredQuality(level int, tool *Stack, r *rand.Rand) int {
	if tool == nil {
		return 1 + r.Intn(10)
	}
	d := Equipment[tool.Kind]
	q := EffectiveQuality(tool.Quality, tool.Damage)
	base := 10 + float64(level)*0.6 + float64(d.TierBonus) + float64(q)*0.2
	return clampQuality(int(base) - 5 + r.Intn(11))
}

// CraftedQuality is the quality of a bar or arrows made from an input of
// some quality: no better than the input, and capped by skill and tool.
func CraftedQuality(inputQuality, level int, tool *Stack, r *rand.Rand) int {
	cap := 15 + int(float64(level)*0.8)
	if tool != nil {
		cap += EffectiveQuality(tool.Quality, tool.Damage) / 10
	} else {
		cap /= 2
	}
	return clampQuality(min(inputQuality, cap) - 2 + r.Intn(5))
}

func clampQuality(q int) int { return min(max(q, 1), MaxQuality) }

// IsPiece reports whether an item is an individual piece with its own
// quality and damage (gear and tools) rather than a stackable material.
func IsPiece(k ItemKind) bool {
	_, ok := Equipment[k]
	return ok && StackLimits[k] == 1
}

// QualityValue scales an item's value by quality: pieces are worth half
// their base value at quality 0 and two and a half times at 100; materials
// keep their base value.
func QualityValue(k ItemKind, quality int) int {
	if !IsPiece(k) {
		return Values[k]
	}
	return int(float64(Values[k]) * (0.5 + float64(quality)/50))
}
