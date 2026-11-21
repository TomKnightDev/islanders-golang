package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// Making and improving gear at an anvil or the fletcher's hut, by hand
// from the backpack, through the station window. See docs/DESIGN.md.
const (
	improveChanceMax  = 0.85 // success chance on a piece at or below the player's level
	improveChanceMin  = 0.10
	improveChanceDrop = 0.02 // per quality point the piece is above the player's level
	improveFailLoss   = 1    // quality lost on a failed attempt, once the damage is mended
)

// stationStep is how long making takes per ingredient at a station, before
// skill and tool.
var stationStep = map[resources.Building]time.Duration{
	resources.BuildingAnvil:    smithTime,
	resources.BuildingFletcher: fletchTime,
}

const (
	activityMaking  = "Making"
	activityImprove = "Improving"
)

// makeSession is a player making Count of a recipe from their backpack.
type makeSession struct {
	recipe    resources.GearRecipe
	at        visit
	left      int
	countdown int
}

// improveSession is a player improving one piece from their backpack.
type improveSession struct {
	id        uint32 // the piece
	at        visit
	recipe    resources.GearRecipe // the recipe that makes the piece: its skill, tool and material
	unitQ     int                  // quality of the unit of material in use
	unitLeft  int                  // attempts left on it; 0 means take a new one
	countdown int
}

func (s *improveSession) piece() uint32 {
	if s == nil {
		return 0
	}
	return s.id
}

func (s *makeSession) item() (resources.ItemKind, int) {
	if s == nil {
		return "", 0
	}
	return s.recipe.Item, s.left
}

// toolIn is the tool in a belt slot, or nil.
func (p *player) toolIn(slot resources.EquipSlot) *resources.Stack {
	s, ok := p.equipment[slot]
	if !ok {
		return nil
	}
	return &s
}

// stationTicks is a station step's length, scaled by skill and tool.
func (p *player) stationTicks(r resources.GearRecipe, base time.Duration) int {
	t := p.skillTicks(base, r.Skill)
	return max(1, int(float64(t)*resources.ToolSpeed(p.toolIn(r.Tool))+0.5))
}

// stationCommand starts making or improving, or stops.
func (w *World) stationCommand(p *player, c resources.StationCommandContents) {
	defer func() { p.invDirty = true }()
	switch c.Op {
	case resources.StationOpStop:
		w.stopStation(p, "Stopped.")
	case resources.StationOpMake:
		w.startMaking(p, c.Item, c.Count)
	case resources.StationOpImprove:
		w.startImproving(p, c.ID)
	}
}

func (w *World) stopStation(p *player, why string) {
	if p.making == nil && p.improving == nil {
		return
	}
	p.making, p.improving = nil, nil
	p.stationNote = why
	p.invDirty = true
	w.setActivity(p, activityIdle)
}

func (w *World) tickStation(p *player) {
	switch {
	case p.making != nil:
		w.tickMaking(p)
	case p.improving != nil:
		w.tickImprove(p)
	}
}

// --- making ---

func (w *World) startMaking(p *player, item resources.ItemKind, count int) {
	v := w.visiting(p)
	r, ok := resources.FindGearRecipe(item)
	if !ok || r.At != v.at || count < 1 {
		return
	}
	if p.level(r.Skill) < r.Level {
		p.stationNote = fmt.Sprintf("Needs %s level %d.", resources.SkillNames[r.Skill], r.Level)
		return
	}
	if !canCraft(&p.backpack, r.Material) {
		p.stationNote = "Needs " + ingredientList(r) + " in your backpack."
		return
	}
	w.stopStation(p, "")
	p.making = &makeSession{recipe: r, at: v, left: min(count, resources.MaxGoalAmount)}
	p.stationNote = fmt.Sprintf("Making %s.", r.Label)
	w.startMake(p)
}

func (w *World) startMake(p *player) {
	s := p.making
	units := 0
	for _, in := range s.recipe.Inputs {
		units += in.Count
	}
	s.countdown = p.stationTicks(s.recipe, stationStep[s.at.at]*time.Duration(units))
	w.startAction(p, activityMaking, s.countdown)
}

// tickMaking makes one piece when its time is up: the ingredients come out
// of the backpack and the piece goes in, rough at first (see MadeQuality).
func (w *World) tickMaking(p *player) {
	s := p.making
	if w.visiting(p) != s.at {
		w.stopStation(p, "You walked away.")
		return
	}
	if s.countdown--; s.countdown > 0 {
		return
	}
	r := s.recipe
	if !canCraft(&p.backpack, r.Material) {
		w.stopStation(p, "Out of materials: needs "+ingredientList(r)+".")
		return
	}
	qsum, n := 0, 0
	for _, in := range r.Inputs {
		for _, part := range p.backpack.Take(in.Item, in.Count) {
			qsum += part.Quality * part.Count
			n += part.Count
		}
	}
	q := resources.MadeQuality(qsum/max(n, 1), p.level(r.Skill), p.toolIn(r.Tool), w.rng)
	w.give(&p.backpack, r.Item, 1, q) // the ingredients' slots are free, so this fits
	w.addXP(p, r.Skill, r.XP)
	w.spendStamina(p, staminaPerStep)
	s.left--
	p.stationNote = fmt.Sprintf("Made a %s, quality %d.", r.Label, q)
	p.invDirty = true
	switch {
	case s.left == 0:
		w.stopStation(p, p.stationNote)
	case !canCraft(&p.backpack, r.Material):
		w.stopStation(p, p.stationNote+" Out of materials.")
	case w.exhausted(p):
		w.stopStation(p, p.stationNote+" Too tired to go on.")
	default:
		w.startMake(p)
	}
}

// ingredientList reads "2 copper bars" or "2 oak logs and 1 copper bars".
func ingredientList(r resources.GearRecipe) string {
	var parts []string
	for _, in := range r.Inputs {
		parts = append(parts, fmt.Sprintf("%d %s", in.Count, resources.ItemUnits[in.Item]))
	}
	return list(parts)
}

// --- improving ---

func (w *World) startImproving(p *player, id uint32) {
	v := w.visiting(p)
	i := slices.IndexFunc(p.backpack.Stacks, func(s resources.Stack) bool { return s.ID == id && id != 0 })
	if i < 0 {
		return
	}
	piece := p.backpack.Stacks[i]
	r, ok := resources.FindGearRecipe(piece.Kind)
	if !ok || r.At != v.at {
		return
	}
	if why := w.cantImprove(p, r, piece, true); why != "" {
		p.stationNote = why
		return
	}
	w.stopStation(p, "")
	p.improving = &improveSession{id: id, at: v, recipe: r}
	p.stationNote = fmt.Sprintf("Improving %s.", resources.ItemNames[piece.Kind])
	w.startAttempt(p)
}

// cantImprove says why p can't work on a piece now, or "" if they can.
// needUnit asks for a fresh unit of material better than the piece.
func (w *World) cantImprove(p *player, r resources.GearRecipe, piece resources.Stack, needUnit bool) string {
	switch {
	case p.level(r.Skill) < r.Level:
		return fmt.Sprintf("Needs %s level %d.", resources.SkillNames[r.Skill], r.Level)
	case p.toolIn(r.Tool) == nil:
		return fmt.Sprintf("Needs a %s on your tool belt.", resources.SlotNames[r.Tool])
	case piece.Quality >= resources.MaxQuality:
		return "It can't get any better."
	case needUnit && bestUnit(&p.backpack, r.Inputs[0].Item, piece.Quality) < 0:
		return fmt.Sprintf("Needs %s better than quality %d in your backpack.", resources.ItemUnits[r.Inputs[0].Item], piece.Quality)
	}
	return ""
}

// bestUnit is the backpack slot of the cheapest unit of material that's
// still better than quality q, or -1.
func bestUnit(inv *resources.Inventory, k resources.ItemKind, q int) int {
	best := -1
	for i, s := range inv.Stacks {
		if s.Kind == k && s.Quality > q && (best < 0 || s.Quality < inv.Stacks[best].Quality) {
			best = i
		}
	}
	return best
}

func (w *World) startAttempt(p *player) {
	s := p.improving
	s.countdown = p.stationTicks(s.recipe, improveTime)
	w.startAction(p, activityImprove, s.countdown)
}

// tickImprove runs the attempts. Each uses part of a unit of material
// better than the piece, and wears the tool; success raises the piece's
// quality (never past the material's), failure costs a little. Attempts on
// pieces above the player's level succeed less often but give more XP.
func (w *World) tickImprove(p *player) {
	s := p.improving
	if w.visiting(p) != s.at {
		w.stopStation(p, "You walked away.")
		return
	}
	if s.countdown--; s.countdown > 0 {
		return
	}
	pieceAt := func() int {
		return slices.IndexFunc(p.backpack.Stacks, func(st resources.Stack) bool { return st.ID == s.id })
	}
	i := pieceAt()
	if i < 0 {
		w.stopStation(p, "The piece isn't in your backpack.")
		return
	}
	if s.unitLeft > 0 && s.unitQ <= p.backpack.Stacks[i].Quality {
		s.unitLeft = 0 // what's left of this unit is no better than the piece now
	}
	if why := w.cantImprove(p, s.recipe, p.backpack.Stacks[i], s.unitLeft == 0); why != "" {
		w.stopStation(p, why)
		return
	}
	if s.unitLeft == 0 {
		u := bestUnit(&p.backpack, s.recipe.Inputs[0].Item, p.backpack.Stacks[i].Quality)
		unit, _ := p.backpack.TakeSlot(u, 1)
		s.unitQ, s.unitLeft = unit.Quality, resources.ImproveUnitAttempts
		i = pieceAt()
	}
	s.unitLeft--
	piece := &p.backpack.Stacks[i]

	level, before := p.level(s.recipe.Skill), piece.Quality
	chance := min(improveChanceMax, max(improveChanceMin, improveChanceMax-improveChanceDrop*float64(piece.Quality-level)))
	if w.rng.Float64() < chance {
		gain := 1 + w.rng.Intn(max(1, (s.unitQ-piece.Quality)/8))
		piece.Quality = min(s.unitQ, piece.Quality+gain, resources.MaxQuality)
		p.stationNote = fmt.Sprintf("Success: quality %d → %d.", before, piece.Quality)
	} else {
		piece.Quality = max(1, piece.Quality-improveFailLoss)
		p.stationNote = fmt.Sprintf("Failed: quality %d → %d.", before, piece.Quality)
	}
	w.addXP(p, s.recipe.Skill, improveXP(s.recipe)*(20+max(0, before-level))/20)
	w.spendStamina(p, staminaPerStep)
	p.invDirty = true

	if w.wearTool(p, s.recipe.Tool) {
		w.stopStation(p, p.stationNote+" Your tool broke.")
		return
	}
	if w.exhausted(p) {
		w.stopStation(p, p.stationNote+" Too tired to go on.")
		return
	}
	w.startAttempt(p)
}

// wearTool damages the tool in a belt slot and reports whether it broke
// (and is gone).
func (w *World) wearTool(p *player, slot resources.EquipSlot) bool {
	tool, ok := p.equipment[slot]
	if !ok {
		return false
	}
	tool.Damage += resources.ImproveToolWear
	if tool.Damage >= 100 {
		delete(p.equipment, slot)
		if p.sess != nil {
			p.sess.push(resources.NewChatMessage(p.id, fmt.Sprintf("Your %s broke.", resources.ItemNames[tool.Kind])))
		}
		return true
	}
	p.equipment[slot] = tool
	return false
}

// improveXP is the XP for one attempt before any bonus: the recipe's XP
// for one unit of its ingredients.
func improveXP(r resources.GearRecipe) int {
	units := 0
	for _, in := range r.Inputs {
		units += in.Count
	}
	return max(1, r.XP/max(units, 1))
}
