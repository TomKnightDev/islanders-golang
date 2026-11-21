package main

import "github.com/tomknightdev/islanders-golang/resources"

// Improving tuning. See docs/ECONOMY.md.
const (
	improveChanceMax  = 0.85 // success chance on a piece at or below the player's level
	improveChanceMin  = 0.10
	improveChanceDrop = 0.02 // per quality point the piece is above the player's level
	improveFailLoss   = 1    // quality lost on a failed attempt, once the damage is mended
)

// collectForImprove takes the best piece of the kind being improved from
// the bank, and as much of its material as the goal still needs, then
// heads to the station. The material must be better than the piece.
func (w *World) collectForImprove(p *player) {
	j := p.job
	piece, ok := takeBest(&p.bank, j.improving)
	if !ok {
		w.finishAction(p, activityNoPiece)
		return
	}
	in := j.material.Inputs[0].Item
	if p.bank.Count(in) == 0 || p.bank.AverageQuality(in) <= piece.Quality || piece.Quality >= resources.MaxQuality {
		p.bank.AddStack(piece)
		w.finishAction(p, activityNoBetter)
		return
	}
	n := min(j.action.spec.Amount-j.action.count, p.bank.Count(in), p.backpack.FreeSlots()-1)
	if n <= 0 {
		p.bank.AddStack(piece)
		w.finishAction(p, activityNoPiece)
		return
	}
	p.backpack.AddStack(piece)
	j.piece = piece.ID
	moveItems(&p.bank, &p.backpack, in, n)
	p.invDirty = true
	w.walkToStation(p)
}

// improveAttempt uses one unit of material on the carried piece. Success
// raises its quality, by less as it nears the material's quality; failure
// damages it, and mending that costs a little quality. Attempts on pieces
// above the player's level succeed less often but give more XP.
func (w *World) improveAttempt(p *player) {
	j := p.job
	in := j.material.Inputs[0].Item
	piece, ok := p.backpack.TakeID(j.piece)
	if !ok {
		w.goBank(p)
		return
	}
	mat := p.backpack.Take(in, 1)
	if len(mat) == 0 || mat[0].Quality <= piece.Quality {
		p.backpack.AddStack(piece)
		p.backpack.AddAll(mat)
		w.goBank(p)
		return
	}
	level, mq := p.level(j.skill), mat[0].Quality
	over := max(0, piece.Quality-level)
	chance := max(improveChanceMin, improveChanceMax-improveChanceDrop*float64(piece.Quality-level))
	chance = min(chance, improveChanceMax)
	if w.rng.Float64() < chance {
		gain := 1 + w.rng.Intn(max(1, (mq-piece.Quality)/8))
		piece.Quality = min(mq, piece.Quality+gain, resources.MaxQuality)
	} else {
		piece.Quality = max(1, piece.Quality-improveFailLoss)
	}
	p.backpack.AddStack(piece)
	j.action.count++
	w.addXP(p, j.skill, improveXP(j.material)*(20+over)/20)
	w.spendStamina(p, staminaPerStep)
	p.invDirty = true

	if p.backpack.Count(in) == 0 || amountMet(p, j.action) || piece.Quality >= resources.MaxQuality {
		w.goBank(p)
		return
	}
	if w.exhausted(p) {
		return
	}
	w.startStep(p)
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
