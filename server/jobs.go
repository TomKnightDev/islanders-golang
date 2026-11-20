package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// Tuning values. See docs/DESIGN.md.
const (
	swingTime       = 2 * time.Second
	rockRespawn     = 30 * time.Second
	rockOreMin      = 5
	rockOreMax      = 10
	backpackSlots   = 20
	bankSlots       = 100
	rockWaitRetry   = 1 * time.Second
	noPathShown     = 3 * time.Second // how long "Can't get there" stays up
	offlineSlowdown = 10              // offline characters run at 1/10 speed
)

func ticks(d time.Duration) int { return int(d / tickInterval) }

// Activities shown to the player.
const (
	activityIdle     = "Idle"
	activityWalking  = "Walking"
	activityToRock   = "Walking to a rock"
	activityMining   = "Mining"
	activityToBank   = "Walking to the bank"
	activityWaiting  = "Waiting for a free rock"
	activityBankFull = "Bank full"
	activityNoPath   = "Can't get there"
)

type rock struct {
	obj       resources.WorldObject
	ore       int
	claimedBy *player
	respawnAt int // world tick; only meaningful while depleted
}

func (r *rock) tile() tile { return tile{r.obj.X, r.obj.Y} }

// freeFor reports whether p may mine r now.
func (r *rock) freeFor(p *player) bool {
	return r.ore > 0 && (r.claimedBy == nil || r.claimedBy == p)
}

type jobKind int

const (
	jobMine jobKind = iota + 1
	jobBankTrip
)

type jobState int

const (
	stateWalkToRock jobState = iota + 1
	stateMining
	stateWaitForRock
	stateWalkToBank
)

type job struct {
	kind      jobKind
	state     jobState
	rock      *rock // the rock being walked to or mined; the one to return to after banking
	countdown int   // ticks left in a swing or a wait
}

func (w *World) initRocks() {
	w.rocks = map[tile]*rock{}
	for _, o := range w.wm.Objects {
		switch o.Kind {
		case resources.ObjectIronRock:
			w.rocks[tile{o.X, o.Y}] = &rock{obj: o, ore: rockOreMin + rand.Intn(rockOreMax-rockOreMin+1)}
		case resources.ObjectBank:
			w.banks = append(w.banks, tile{o.X, o.Y})
		}
	}
}

// setActivity changes what p is shown doing, for actions with no set length.
func (w *World) setActivity(p *player, a string) {
	if p.activity != a || p.actionTicks != 0 {
		w.startAction(p, a, 0)
	}
}

// startAction begins an action lasting d (0 for open-ended) and tells the
// client, which animates the progress bar from it.
func (w *World) startAction(p *player, a string, d int) {
	p.activity = a
	p.actionStart = p.ticks
	p.actionTicks = d
	p.invDirty = true
}

// walkTicks estimates how long p's current path takes to walk.
func walkTicks(p *player) int {
	dist, x, y := 0.0, p.pos[0], p.pos[1]
	for _, t := range p.path {
		tx, ty := t.pixel()
		dist += math.Hypot(tx-x, ty-y)
		x, y = tx, ty
	}
	return int(math.Ceil(dist / walkStep))
}

// click is what a player's click means: mine a rock, visit the bank, or walk.
func (w *World) click(p *player, t tile) {
	w.stopJob(p)
	if r, ok := w.rocks[t]; ok {
		w.startMining(p, r)
		return
	}
	for _, b := range w.banks {
		if t == b {
			p.job = &job{kind: jobBankTrip}
			w.walkToBank(p, b)
			return
		}
	}
	if w.moveTo(p, t) {
		w.startAction(p, activityWalking, walkTicks(p))
	} else {
		w.setActivity(p, activityNoPath)
	}
}

// stopJob ends p's current job and gives up any rock it was holding.
func (w *World) stopJob(p *player) {
	if p.job != nil && p.job.rock != nil && p.job.rock.claimedBy == p {
		p.job.rock.claimedBy = nil
	}
	p.job = nil
	w.setActivity(p, activityIdle)
}

func (w *World) startMining(p *player, clicked *rock) {
	p.job = &job{kind: jobMine, rock: clicked}
	if r := w.nearestFreeRock(p, clicked); r != nil {
		w.walkToRock(p, r)
		return
	}
	// Every rock is taken or empty: wait beside the one that was clicked.
	w.moveTo(p, clicked.tile())
	w.waitForRock(p)
}

// nearestFreeRock prefers the given rock if p can mine it, otherwise the
// closest rock nobody else is mining that still has ore.
func (w *World) nearestFreeRock(p *player, prefer *rock) *rock {
	if prefer != nil && prefer.freeFor(p) {
		return prefer
	}
	from := p.currentTile()
	var best *rock
	bestDist := 0
	for _, r := range w.rocks {
		if !r.freeFor(p) {
			continue
		}
		if d := octile(from, r.tile()); best == nil || d < bestDist || (d == bestDist && r.obj.ID < best.obj.ID) {
			best, bestDist = r, d
		}
	}
	return best
}

func (w *World) walkToRock(p *player, r *rock) {
	if p.job.rock != nil && p.job.rock != r && p.job.rock.claimedBy == p {
		p.job.rock.claimedBy = nil
	}
	r.claimedBy = p // reserve it now so two players don't walk to the same rock
	p.job.rock = r
	p.job.state = stateWalkToRock
	if !w.moveTo(p, r.tile()) {
		w.stopJob(p)
		w.setActivity(p, activityNoPath)
		return
	}
	w.startAction(p, activityToRock, walkTicks(p))
}

func (w *World) waitForRock(p *player) {
	p.job.state = stateWaitForRock
	p.job.countdown = ticks(rockWaitRetry)
	w.setActivity(p, activityWaiting)
}

// nearestBank is the bank closest to p as the crow flies.
func (w *World) nearestBank(p *player) (tile, bool) {
	from := p.currentTile()
	var best tile
	found := false
	for _, b := range w.banks {
		if !found || octile(from, b) < octile(from, best) {
			best, found = b, true
		}
	}
	return best, found
}

func (w *World) walkToBank(p *player, bank tile) {
	if p.job.rock != nil && p.job.rock.claimedBy == p {
		p.job.rock.claimedBy = nil
	}
	p.job.state = stateWalkToBank
	if !w.moveTo(p, bank) {
		w.stopJob(p)
		w.setActivity(p, activityNoPath)
		return
	}
	w.startAction(p, activityToBank, walkTicks(p))
}

// tickJob advances p's job by one tick. Called after p has moved.
func (w *World) tickJob(p *player) {
	j := p.job
	arrived := len(p.path) == 0
	if j == nil {
		switch {
		case p.activity == activityWalking && arrived:
			w.setActivity(p, activityIdle)
		case p.activity == activityNoPath && p.ticks-p.actionStart >= ticks(noPathShown):
			w.setActivity(p, activityIdle)
		}
		return
	}

	switch j.state {
	case stateWalkToRock:
		if !arrived {
			return
		}
		if !j.rock.freeFor(p) { // emptied while we walked
			w.nextRock(p)
			return
		}
		j.state = stateMining
		j.countdown = ticks(swingTime)
		w.startAction(p, activityMining, j.countdown)

	case stateMining:
		j.countdown--
		if j.countdown > 0 {
			return
		}
		r := j.rock
		r.ore--
		p.backpack.Add(resources.ItemIronOre, 1)
		p.invDirty = true
		if r.ore == 0 {
			w.depleteRock(r)
		}
		switch {
		case p.backpack.Full(resources.ItemIronOre):
			if b, ok := w.nearestBank(p); ok {
				w.walkToBank(p, b)
			} else {
				w.stopJob(p)
				w.setActivity(p, activityNoPath)
			}
		case r.ore == 0:
			w.nextRock(p)
		default:
			j.countdown = ticks(swingTime)
			w.startAction(p, activityMining, j.countdown)
		}

	case stateWaitForRock:
		j.countdown--
		if j.countdown > 0 {
			return
		}
		if r := w.nearestFreeRock(p, j.rock); r != nil {
			w.walkToRock(p, r)
		} else {
			j.countdown = ticks(rockWaitRetry)
		}

	case stateWalkToBank:
		if !arrived {
			return
		}
		emptied := p.backpack.TransferAll(&p.bank)
		p.invDirty = true
		switch {
		case !emptied:
			w.stopJob(p)
			w.setActivity(p, activityBankFull)
		case j.kind == jobBankTrip:
			w.stopJob(p)
		default:
			w.nextRock(p) // back to the same rock if it's still free
		}
	}
}

func (w *World) nextRock(p *player) {
	if r := w.nearestFreeRock(p, p.job.rock); r != nil {
		w.walkToRock(p, r)
		return
	}
	w.waitForRock(p)
}

func (w *World) setObjectTile(t tile, gid int) {
	w.wm.Layers[resources.LayerObject].Data[t.y*w.wm.Width+t.x] = gid
	w.broadcast(resources.NewObjectUpdateMessage(resources.ObjectUpdateContents{X: t.x, Y: t.y, Tile: gid}))
}

func (w *World) depleteRock(r *rock) {
	r.claimedBy = nil
	r.respawnAt = w.ticks + ticks(rockRespawn)
	w.setObjectTile(r.tile(), resources.TileIronRockDepleted)
}

func (w *World) tickRocks() {
	for _, r := range w.rocks {
		if r.ore == 0 && w.ticks >= r.respawnAt {
			r.ore = rockOreMin + rand.Intn(rockOreMax-rockOreMin+1)
			w.setObjectTile(r.tile(), resources.TileIronRock)
		}
	}
}

func (w *World) inventoryMessage(p *player) *resources.Message {
	seconds := func(t int) float64 { return float64(t) * tickInterval.Seconds() }
	return resources.NewInventoryMessage(resources.InventoryContents{
		Backpack:      p.backpack.Clone(),
		Bank:          p.bank.Clone(),
		Activity:      p.activity,
		ActionSeconds: seconds(p.actionTicks),
		ActionElapsed: seconds(p.ticks - p.actionStart),
	})
}
