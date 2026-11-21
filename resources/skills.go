package resources

import "math"

type Skill string

const (
	SkillAttack      Skill = "attack"    // melee accuracy
	SkillStrength    Skill = "strength"  // melee max hit
	SkillDefence     Skill = "defence"   // how often enemies hit you
	SkillRanged      Skill = "ranged"    // bow accuracy and max hit
	SkillMagic       Skill = "magic"     // staff accuracy and max hit
	SkillHitpoints   Skill = "hitpoints" // max health
	SkillMining      Skill = "mining"
	SkillWoodcutting Skill = "woodcutting"
	SkillFarming     Skill = "farming"
	SkillSmithing    Skill = "smithing"
	SkillFletching   Skill = "fletching"
)

// Skills lists every skill in the order the skills window shows them.
var Skills = []Skill{
	SkillAttack, SkillStrength, SkillDefence, SkillRanged, SkillMagic, SkillHitpoints,
	SkillMining, SkillWoodcutting, SkillFarming, SkillSmithing, SkillFletching,
}

var SkillNames = map[Skill]string{
	SkillAttack:      "Attack",
	SkillStrength:    "Strength",
	SkillDefence:     "Defence",
	SkillRanged:      "Ranged",
	SkillMagic:       "Magic",
	SkillHitpoints:   "Hitpoints",
	SkillMining:      "Mining",
	SkillWoodcutting: "Woodcutting",
	SkillFarming:     "Farming",
	SkillSmithing:    "Smithing",
	SkillFletching:   "Fletching",
}

const MaxLevel = 99

// StartingHitpoints is the Hitpoints level every character starts at.
const StartingHitpoints = 10

// MaxHealth is a character's full health at a Hitpoints level.
func MaxHealth(hitpointsLevel int) int { return 10 * hitpointsLevel }

// levelXP[l] is the total XP needed to reach level l. The curve roughly
// doubles every seven levels, so level 99 takes about 13 million XP: the
// game is meant to be played for hundreds of hours.
var levelXP = func() [MaxLevel + 1]int {
	var t [MaxLevel + 1]int
	points := 0.0
	for l := 2; l <= MaxLevel; l++ {
		points += math.Floor(float64(l-1) + 300*math.Pow(2, float64(l-1)/7))
		t[l] = int(points / 4)
	}
	return t
}()

// XPForLevel is the total XP needed to reach a level.
func XPForLevel(level int) int {
	return levelXP[min(max(level, 1), MaxLevel)]
}

// LevelForXP is the level a total amount of XP gives.
func LevelForXP(xp int) int {
	l := 1
	for l < MaxLevel && xp >= levelXP[l+1] {
		l++
	}
	return l
}

// SpeedFactor scales how long an action takes at a skill level: every level
// takes 0.4% off, so level 99 works in about 61% of the base time.
func SpeedFactor(level int) float64 {
	return 1 - 0.004*float64(min(max(level, 1), MaxLevel)-1)
}

// SkillState is one skill as the server reports it.
type SkillState struct {
	Skill  Skill `json:"skill"`
	Level  int   `json:"level"`
	XP     int   `json:"xp"`
	NextXP int   `json:"nextXP"` // total XP for the next level; equals XP at the max level
}
