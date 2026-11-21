package main

import (
	"fmt"
	"slices"

	"github.com/tomknightdev/islanders-golang/resources"
)

// Improving tuning. See docs/DESIGN.md.
const (
	improveChanceMax  = 0.85 // success chance on a piece at or below the player's level
	improveChanceMin  = 0.10
	improveChanceDrop = 0.02 // per quality point the piece is above the player's level
	improveFailLoss   = 1    // quality lost on a failed attempt, once the damage is mended
)

// improveSession is a player improving one piece from their backpack at an
// anvil or the fletcher's hut. It runs while they stay there.
type improveSession struct {
	id        uint32               // the piece
	at        visit                // where it's happening
	kind      resources.ActionKind // the recipe that makes the piece: its skill and tool
	material  resources.Material
	unitQ     int // quality of the unit of material in use
	unitLeft  int // attempts left on it; 0 means take a new one
	countdown int
}

func (s *improveSession) piece() uint32 {
	if s == nil {
		return 0
	}
	return s.id
}

// improveCommand starts or stops improving a piece in the backpack.
func (w *World) improveCommand(p *player, c resources.ImproveCommandContents) {
	defer func() { p.invDirty = true }()
	if c.Op == resources.ImproveOpStop {
		w.stopImproving(p, "Stopped.")
		return
	}
	v := w.visiting(p)
	var piece *resources.Stack
	for i := range p.backpack.Stacks {
		if p.backpack.Stacks[i].ID == c.ID && c.ID != 0 {
			piece = &p.backpack.Stacks[i]
		}
	}
	if piece == nil {
		return
	}
	if at, ok := resources.ImproveAt(piece.Kind); !ok || at != v.at {
		return
	}
	kind, m, _ := resources.ImproveRecipe(piece.Kind)
	if why := w.cantImprove(p, kind, m, *piece, true); why != "" {
		p.improveNote = why
		return
	}
	p.improving = &improveSession{id: c.ID, at: v, kind: kind, material: m}
	p.improveNote = fmt.Sprintf("Improving %s.", resources.ItemNames[piece.Kind])
	w.startAttempt(p)
}

// cantImprove says why p can't work on a piece now, or "" if they can.
// needUnit asks for a fresh unit of material better than the piece.
func (w *World) cantImprove(p *player, kind resources.ActionKind, m resources.Material, piece resources.Stack, needUnit bool) string {
	skill := resources.ActionSkill[kind]
	switch {
	case p.level(skill) < m.Level:
		return fmt.Sprintf("Needs %s level %d.", resources.SkillNames[skill], m.Level)
	case p.tool(kind) == nil:
		return fmt.Sprintf("Needs a %s on your tool belt.", resources.SlotNames[resources.ActionTool[kind]])
	case piece.Quality >= resources.MaxQuality:
		return "It can't get any better."
	case needUnit && bestUnit(&p.backpack, m.Inputs[0].Item, piece.Quality) < 0:
		return fmt.Sprintf("Needs %s better than quality %d in your backpack.", resources.ItemUnits[m.Inputs[0].Item], piece.Quality)
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
	s.countdown = p.stepTicks(s.kind, p.skillTicks(improveTime, resources.ActionSkill[s.kind]))
	w.startAction(p, activityImprove, s.countdown)
}

func (w *World) stopImproving(p *player, why string) {
	if p.improving == nil {
		return
	}
	p.improving = nil
	p.improveNote = why
	p.invDirty = true
	w.setActivity(p, activityIdle)
}

// tickImprove runs the attempts. Each uses part of a unit of material
// better than the piece, and wears the tool; success raises the piece's
// quality (never past the material's), failure costs a little. Attempts on
// pieces above the player's level succeed less often but give more XP.
func (w *World) tickImprove(p *player) {
	s := p.improving
	if s == nil {
		return
	}
	if w.visiting(p) != s.at {
		w.stopImproving(p, "You walked away.")
		return
	}
	if s.countdown--; s.countdown > 0 {
		return
	}
	i := slices.IndexFunc(p.backpack.Stacks, func(st resources.Stack) bool { return st.ID == s.id })
	if i < 0 {
		w.stopImproving(p, "The piece isn't in your backpack.")
		return
	}
	piece := &p.backpack.Stacks[i]
	if s.unitLeft > 0 && s.unitQ <= piece.Quality {
		s.unitLeft = 0 // what's left of this unit is no better than the piece now
	}
	if why := w.cantImprove(p, s.kind, s.material, *piece, s.unitLeft == 0); why != "" {
		w.stopImproving(p, why)
		return
	}
	if s.unitLeft == 0 {
		u := bestUnit(&p.backpack, s.material.Inputs[0].Item, piece.Quality)
		s.unitQ, s.unitLeft = p.backpack.Stacks[u].Quality, resources.ImproveUnitAttempts
		p.backpack.Stacks = append(p.backpack.Stacks[:u], p.backpack.Stacks[u+1:]...)
		i = slices.IndexFunc(p.backpack.Stacks, func(st resources.Stack) bool { return st.ID == s.id })
		piece = &p.backpack.Stacks[i]
	}
	s.unitLeft--

	skill := resources.ActionSkill[s.kind]
	level, before := p.level(skill), piece.Quality
	chance := min(improveChanceMax, max(improveChanceMin, improveChanceMax-improveChanceDrop*float64(piece.Quality-level)))
	if w.rng.Float64() < chance {
		gain := 1 + w.rng.Intn(max(1, (s.unitQ-piece.Quality)/8))
		piece.Quality = min(s.unitQ, piece.Quality+gain, resources.MaxQuality)
		p.improveNote = fmt.Sprintf("Success: quality %d → %d.", before, piece.Quality)
	} else {
		piece.Quality = max(1, piece.Quality-improveFailLoss)
		p.improveNote = fmt.Sprintf("Failed: quality %d → %d.", before, piece.Quality)
	}
	w.addXP(p, skill, improveXP(s.material)*(20+max(0, before-level))/20)
	w.spendStamina(p, staminaPerStep)
	p.invDirty = true

	if w.wearTool(p, s.kind) {
		w.stopImproving(p, p.improveNote+" Your tool broke.")
		return
	}
	if w.exhausted(p) {
		w.stopImproving(p, p.improveNote+" Too tired to go on.")
		return
	}
	w.startAttempt(p)
}

// wearTool damages the tool used for an action and reports whether it
// broke (and is gone).
func (w *World) wearTool(p *player, kind resources.ActionKind) bool {
	slot := resources.ActionTool[kind]
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
func improveXP(m resources.Material) int {
	units := 0
	for _, in := range m.Inputs {
		units += in.Count
	}
	return max(1, m.XP/max(units, 1))
}
