package main

import (
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// Stamina tuning. See docs/DESIGN.md.
const (
	staminaPerStep   = 1                      // each gather swing, smelt or fletch
	staminaPerAttack = 2                      // fighting tires you faster
	innRestTime      = 500 * time.Millisecond // per point of stamina recovered at an inn
)

const (
	activityToInn    = "Walking to the inn"
	activityRelaxing = "Relaxing at the inn"
	activityNoInn    = "Too tired, and no inn to rest at"
)

func (p *player) maxStamina() int {
	return resources.MaxStamina(p.level(resources.SkillStamina))
}

// spendStamina uses up stamina on a step of work, training Stamina.
func (w *World) spendStamina(p *player, n int) {
	p.stamina = max(0, p.stamina-n)
	w.addXP(p, resources.SkillStamina, n)
	p.invDirty = true
}

// exhausted sends p to an inn if they have no stamina left, and reports
// whether it did. It's checked at the start of each step of work, so a
// character never stops part-way through a swing.
func (w *World) exhausted(p *player) bool {
	if p.stamina > 0 {
		return false
	}
	w.goToInn(p)
	return true
}

// goToInn lets go of whatever the job holds and heads for the nearest inn.
// The job stays, so the character goes back to it once rested.
func (w *World) goToInn(p *player) {
	w.releaseNode(p)
	inn, ok := w.nearestOf(p, w.inns)
	if !ok || !w.moveTo(p, inn) {
		w.failAction(p, activityNoInn)
		return
	}
	p.job.state = stateWalkToInn
	w.startAction(p, activityToInn, walkTicks(p))
}

// nearestOf is the closest of some tiles to p as the crow flies.
func (w *World) nearestOf(p *player, tiles []tile) (tile, bool) {
	from := p.currentTile()
	var best tile
	found := false
	for _, t := range tiles {
		if !found || octile(from, t) < octile(from, best) {
			best, found = t, true
		}
	}
	return best, found
}

// startRelaxing pays the inn fee and rests. A character who can't afford
// it sleeps in the stable: free, but slower.
func (w *World) startRelaxing(p *player) {
	j := p.job
	j.state = stateRelaxing
	j.restTicks = ticks(innRestTime)
	activity := activityRelaxing
	if p.coins >= resources.InnFee {
		p.coins -= resources.InnFee
		p.invDirty = true
	} else {
		j.restTicks *= stableSlowdown
		activity = activityStable
	}
	j.countdown = j.restTicks
	w.startAction(p, activity, (p.maxStamina()-p.stamina)*j.restTicks)
}

// tickRelaxing recovers stamina at the inn. Once rested, the job is
// dropped so the action starts again from the top, picking up where the
// list left off.
func (w *World) tickRelaxing(p *player) {
	j := p.job
	j.countdown--
	if j.countdown > 0 {
		return
	}
	j.countdown = j.restTicks
	p.stamina = min(p.maxStamina(), p.stamina+1)
	p.invDirty = true
	if p.stamina >= p.maxStamina() {
		p.job = nil
		w.setActivity(p, activityIdle)
	}
}

// atInn reports whether the job is on an inn trip, which doesn't count
// towards an action's time goal.
func (j *job) atInn() bool { return j.state == stateWalkToInn || j.state == stateRelaxing }
