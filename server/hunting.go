package main

import (
	"math/rand"
	"slices"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

// Hunting: deer wander the named forests. A Hunt action stalks the nearest
// one, then tries to bring it down; a miss sends it running. A catch gives
// leather and venison and trains Hunting. Deer never fight back.
const (
	deerPerForest   = 4
	huntTime        = 6 * time.Second // per attempt, before skill
	huntChanceBase  = 0.5             // at level 1
	huntChancePerLv = 0.008           // per level, up to huntChanceMax
	huntChanceMax   = 0.9
	deerFleeRange   = 6 // tiles a startled deer runs
	leatherPerDeer  = 2 // up to; at least 1
)

const (
	activityStalking = "Stalking a deer"
	activityHunting  = "Hunting"
	activityNoDeer   = "Waiting for deer"
	activityDeerFled = "The deer got away"
)

// initDeer puts deer in every named forest, at random open tiles near its
// middle that can be walked to, and adds them to what the forest offers.
func (w *World) initDeer() {
	// The world's own copy of the places: the map it was made from may be
	// shared (tests build many worlds from one island).
	w.wm.Places = slices.Clone(w.wm.Places)
	spawn := tile{w.wm.SpawnX, w.wm.SpawnY}
	for i := range w.wm.Places {
		pl := &w.wm.Places[i]
		if pl.Kind != resources.PlaceForest {
			continue
		}
		var open []tile
		for y := pl.Y - pl.Radius; y <= pl.Y+pl.Radius; y++ {
			for x := pl.X - pl.Radius; x <= pl.X+pl.Radius; x++ {
				if w.wm.TilePassable(x, y) {
					open = append(open, tile{x, y})
				}
			}
		}
		// Keep the ones that can be reached; trees can wall a clearing in.
		w.rng.Shuffle(len(open), func(a, b int) { open[a], open[b] = open[b], open[a] })
		var homes []tile
		for _, t := range open {
			if len(homes) == deerPerForest {
				break
			}
			if findPath(&w.wm, spawn, t) != nil {
				homes = append(homes, t)
			}
		}
		if len(homes) == 0 {
			continue
		}
		for n := 0; n < deerPerForest; n++ {
			home := homes[n%len(homes)]
			x, y := home.pixel()
			w.mobs = append(w.mobs, &mob{
				id: uint16(mobIDBase + len(w.mobs)), kind: resources.Deer, camp: pl.Name, home: home,
				pos: f64.Vec2{x, y}, health: resources.Deer.Health, wanderCountdown: rand.Intn(ticks(enemyWanderMax)),
			})
		}
		if !slices.Contains(pl.Materials, resources.Deer.ID) {
			pl.Materials = append(slices.Clone(pl.Materials), resources.Deer.ID)
		}
	}
}

// seekDeer claims the nearest free deer in the action's forest and walks to
// it, or waits if there are none.
func (w *World) seekDeer(p *player) {
	j := p.job
	w.releaseNode(p)
	from := p.currentTile()
	var best *mob
	for _, m := range w.mobs {
		if m.dead || m.camp != j.action.spec.Place || m.kind.ID != resources.Deer.ID || (m.fighting != nil && m.fighting != p) {
			continue
		}
		if best == nil || octile(from, m.tile()) < octile(from, best.tile()) {
			best = m
		}
	}
	if best == nil {
		j.state = stateWaitForNode
		j.countdown = ticks(rockWaitRetry)
		w.setActivity(p, activityNoDeer)
		return
	}
	best.fighting = p
	best.path = nil // it stops to graze
	j.mob = best
	j.state = stateStalking
	if !w.moveTo(p, best.tile()) {
		w.failAction(p, activityNoPath)
		return
	}
	w.startAction(p, activityStalking, walkTicks(p))
}

// tickHunt runs the stalking and hunting states.
func (w *World) tickHunt(p *player, arrived bool) {
	j := p.job
	m := j.mob
	if m == nil || m.dead || m.fighting != p {
		w.seekDeer(p)
		return
	}
	switch j.state {
	case stateStalking:
		if !arrived {
			return
		}
		if !adjacent(m.tile(), p.currentTile()) {
			w.seekDeer(p) // it moved; go again
			return
		}
		if w.exhausted(p) {
			return
		}
		j.state = stateHunting
		j.countdown = p.skillTicks(huntTime, resources.SkillHunting)
		w.startAction(p, activityHunting, j.countdown)

	case stateHunting:
		if j.countdown--; j.countdown > 0 {
			return
		}
		w.spendStamina(p, staminaPerStep)
		level := p.level(resources.SkillHunting)
		if rand.Float64() >= min(huntChanceMax, huntChanceBase+huntChancePerLv*float64(level)) {
			w.deerFlees(p, m)
			return
		}
		w.catchDeer(p, m, level)
	}
}

// deerFlees: the attempt missed and the deer runs. Nothing held from the
// stalk is paid.
func (w *World) deerFlees(p *player, m *mob) {
	w.releaseNode(p)
	from := m.tile()
	to := from.add(tile{rand.Intn(2*deerFleeRange+1) - deerFleeRange, rand.Intn(2*deerFleeRange+1) - deerFleeRange})
	if path := findPath(&w.wm, from, to); path != nil {
		m.path = path
	}
	p.job.state = stateWaitForNode
	p.job.countdown = ticks(time.Second)
	w.setActivity(p, activityDeerFled)
}

// catchDeer brings the deer down: leather and venison into the backpack,
// the XP held from the stalk plus Hunting XP, and the deer respawns later.
func (w *World) catchDeer(p *player, m *mob, level int) {
	j := p.job
	m.dead, m.fighting, m.dirty = true, nil, true
	m.respawnAt = w.ticks + ticks(enemyRespawn)
	q := min(max(10+int(float64(level)*0.6)-5+w.rng.Intn(11), 1), resources.MaxQuality)
	w.give(&p.backpack, resources.ItemLeather, 1+w.rng.Intn(leatherPerDeer), q)
	w.give(&p.backpack, resources.ItemVenison, 1, resources.DefaultQuality)
	for _, s := range resources.Skills {
		if xp := j.fightXP[s]; xp > 0 {
			w.addXP(p, s, xp)
		}
	}
	j.fightXP, j.mob = nil, nil
	w.addXP(p, resources.SkillHunting, j.material.XP)
	j.action.count++
	p.invDirty = true
	switch {
	case amountMet(p, j.action):
		w.wrapUp(p)
	case p.backpack.Full():
		w.goBank(p)
	default:
		w.seekDeer(p)
	}
}
