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
	smeltTime       = 3 * time.Second // per bar
	rockRespawn     = 30 * time.Second
	rockOreMin      = 5
	rockOreMax      = 10
	backpackSlots   = 20
	bankSlots       = 100
	rockWaitRetry   = 1 * time.Second
	noPathShown     = 3 * time.Second // how long "Can't get there" stays up
	offlineSlowdown = 10              // offline characters run at 1/10 speed
	discoveryCheck  = 500 * time.Millisecond
	discoveryMargin = 2 // tiles beyond a place's edge that count as finding it
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
	activityToSmith  = "Walking to the blacksmith"
	activitySmelting = "Smelting"
	activityNoOre    = "No ore to smelt"
	activityBankFull = "Bank full"
	activityNoPath   = "Can't get there"
)

type rock struct {
	obj       resources.WorldObject
	place     string // the mine it belongs to
	ore       int
	claimedBy *player
	respawnAt int // world tick; only meaningful while depleted
}

func (r *rock) tile() tile { return tile{r.obj.X, r.obj.Y} }

// freeFor reports whether p may mine r now.
func (r *rock) freeFor(p *player) bool {
	return r.ore > 0 && (r.claimedBy == nil || r.claimedBy == p)
}

// queuedAction is one entry in a player's action list.
type queuedAction struct {
	spec resources.ActionSpec
	ran  int // ticks spent running it; stopping the list pauses this
}

// ticksLeft is how much of a time goal remains.
func (a *queuedAction) ticksLeft() int {
	return max(0, ticks(time.Duration(a.spec.Minutes)*time.Minute)-a.ran)
}

type jobState int

const (
	stateWalkToRock jobState = iota + 1
	stateMining
	stateWaitForRock
	stateWalkToBank
	stateWalkToSmith
	stateSmelting
)

// job is the running state of the action at the top of the list.
type job struct {
	action    *queuedAction
	state     jobState
	rock      *rock // mining: the rock being walked to or mined, and returned to after banking
	smith     tile  // smelting: the blacksmith to work at
	countdown int   // ticks left in a swing, a smelt or a wait
	finishing bool  // goal met or time up: bank what's carried, then move on
}

func (w *World) initRocks() {
	w.rocks = map[tile]*rock{}
	w.smiths = map[string]tile{}
	for _, o := range w.wm.Objects {
		t := tile{o.X, o.Y}
		switch o.Kind {
		case resources.ObjectIronRock:
			r := &rock{obj: o, ore: rockOreMin + rand.Intn(rockOreMax-rockOreMin+1)}
			if pl := w.placeAt(t, resources.PlaceMine); pl != nil {
				r.place = pl.Name
			}
			w.rocks[t] = r
		case resources.ObjectBank:
			w.banks = append(w.banks, t)
		case resources.ObjectBlacksmith:
			if pl := w.placeAt(t, resources.PlaceVillage); pl != nil {
				w.smiths[pl.Name] = t
			}
		}
	}
}

// placeAt returns the place of the given kind covering t, if any.
func (w *World) placeAt(t tile, kind resources.PlaceKind) *resources.Place {
	for i := range w.wm.Places {
		p := &w.wm.Places[i]
		if p.Kind == kind && max(abs(p.X-t.x), abs(p.Y-t.y)) <= p.Radius {
			return p
		}
	}
	return nil
}

func (w *World) place(name string) *resources.Place {
	for i := range w.wm.Places {
		if w.wm.Places[i].Name == name {
			return &w.wm.Places[i]
		}
	}
	return nil
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

// click walks p to a tile. While the action list is running, clicks are
// ignored so they never cancel what the character is doing.
func (w *World) click(p *player, t tile) {
	if p.running {
		return
	}
	if w.moveTo(p, t) {
		w.startAction(p, activityWalking, walkTicks(p))
	} else {
		w.setActivity(p, activityNoPath)
	}
}

// --- the action list ---

func (w *World) validAction(p *player, a resources.ActionSpec) bool {
	kind, ok := resources.ActionPlaceKind[a.Kind]
	if !ok || !p.found[a.Place] {
		return false
	}
	if pl := w.place(a.Place); pl == nil || pl.Kind != kind {
		return false
	}
	if a.Kind == resources.ActionSmelt {
		if _, ok := w.smiths[a.Place]; !ok {
			return false
		}
	}
	switch a.Goal {
	case resources.GoalAmount:
		return a.Amount >= 1 && a.Amount <= resources.MaxGoalAmount
	case resources.GoalTime:
		return a.Minutes >= 1 && a.Minutes <= resources.MaxGoalMins
	}
	return false
}

// actionCommand applies one edit to p's action list from the action window.
func (w *World) actionCommand(p *player, c resources.ActionCommandContents) {
	inRange := c.Index >= 0 && c.Index < len(p.actions)
	var current *queuedAction
	if len(p.actions) > 0 {
		current = p.actions[0]
	}
	switch c.Op {
	case resources.ActionOpAdd:
		if len(p.actions) < resources.MaxActions && w.validAction(p, c.Action) {
			p.actions = append(p.actions, &queuedAction{spec: c.Action})
		}
	case resources.ActionOpEdit:
		if inRange && w.validAction(p, c.Action) {
			old := p.actions[c.Index].spec
			p.actions[c.Index].spec = c.Action
			if c.Index == 0 && (old.Kind != c.Action.Kind || old.Place != c.Action.Place) {
				w.cancelJob(p) // a different job now; restart it
			}
		}
	case resources.ActionOpRemove:
		if inRange {
			p.actions = append(p.actions[:c.Index], p.actions[c.Index+1:]...)
		}
	case resources.ActionOpUp:
		if inRange && c.Index > 0 {
			p.actions[c.Index], p.actions[c.Index-1] = p.actions[c.Index-1], p.actions[c.Index]
		}
	case resources.ActionOpDown:
		if inRange && c.Index < len(p.actions)-1 {
			p.actions[c.Index], p.actions[c.Index+1] = p.actions[c.Index+1], p.actions[c.Index]
		}
	case resources.ActionOpStart:
		if len(p.actions) > 0 {
			p.running = true
			p.path = p.path[:min(1, len(p.path))] // stop any manual walk after this step
		}
	case resources.ActionOpStop:
		p.running = false
		w.cancelJob(p)
	}
	// Whatever was running is no longer first in the list: stop it so the
	// new first action starts.
	if p.job != nil && (len(p.actions) == 0 || p.actions[0] != current) {
		w.cancelJob(p)
	}
	if len(p.actions) == 0 {
		p.running = false
	}
	p.invDirty = true
}

// cancelJob stops the running job where it is. Anything carried stays in
// the backpack.
func (w *World) cancelJob(p *player) {
	if p.job == nil {
		return
	}
	w.releaseRock(p)
	p.job = nil
	p.path = p.path[:min(1, len(p.path))] // finish the current step, no further
	w.setActivity(p, activityIdle)
}

func (w *World) releaseRock(p *player) {
	if p.job != nil && p.job.rock != nil && p.job.rock.claimedBy == p {
		p.job.rock.claimedBy = nil
	}
}

func amountMet(p *player, a *queuedAction) bool {
	return a.spec.Goal == resources.GoalAmount && p.bank.Count(a.spec.Product()) >= a.spec.Amount
}

func timeUp(a *queuedAction) bool {
	return a.spec.Goal == resources.GoalTime && a.ticksLeft() == 0
}

func (w *World) startNextAction(p *player) {
	if len(p.actions) == 0 {
		p.running = false
		return
	}
	a := p.actions[0]
	if amountMet(p, a) || timeUp(a) {
		w.finishAction(p, activityIdle)
		return
	}
	switch a.spec.Kind {
	case resources.ActionMine:
		p.job = &job{action: a}
		w.mineNext(p)
	case resources.ActionSmelt:
		p.job = &job{action: a, smith: w.smiths[a.spec.Place]}
		w.goBank(p) // a smelting trip starts at the bank, collecting ore
	}
}

// finishAction moves on to the next action. note is shown if the list is
// now empty.
func (w *World) finishAction(p *player, note string) {
	w.releaseRock(p)
	p.job = nil
	p.actions = p.actions[1:]
	if len(p.actions) == 0 {
		p.running = false
		w.setActivity(p, note)
	}
	p.invDirty = true
}

// failAction stops the whole list, leaving the action in place so the
// player can sort out the problem and start again.
func (w *World) failAction(p *player, why string) {
	w.releaseRock(p)
	p.job = nil
	p.running = false
	w.setActivity(p, why)
}

// wrapUp ends the action once anything carried has been banked.
func (w *World) wrapUp(p *player) {
	p.job.finishing = true
	w.releaseRock(p)
	if p.backpack.Empty() {
		w.finishAction(p, activityIdle)
		return
	}
	w.goBank(p)
}

// --- mining ---

// nearestFreeRock prefers the given rock if p can mine it, otherwise the
// closest rock in the mine that nobody else is mining and that has ore.
func (w *World) nearestFreeRock(p *player, prefer *rock, mine string) *rock {
	if prefer != nil && prefer.place == mine && prefer.freeFor(p) {
		return prefer
	}
	from := p.currentTile()
	var best *rock
	bestDist := 0
	for _, r := range w.rocks {
		if r.place != mine || !r.freeFor(p) {
			continue
		}
		if d := octile(from, r.tile()); best == nil || d < bestDist || (d == bestDist && r.obj.ID < best.obj.ID) {
			best, bestDist = r, d
		}
	}
	return best
}

// mineNext sends p to a free rock in the action's mine, or to wait there.
func (w *World) mineNext(p *player) {
	j := p.job
	mine := j.action.spec.Place
	if r := w.nearestFreeRock(p, j.rock, mine); r != nil {
		w.walkToRock(p, r)
		return
	}
	if j.rock == nil {
		// Every rock is busy and we're not there yet: walk to the mine.
		if pl := w.place(mine); pl != nil {
			w.moveTo(p, tile{pl.X, pl.Y})
		}
	}
	j.state = stateWaitForRock
	j.countdown = ticks(rockWaitRetry)
	w.setActivity(p, activityWaiting)
}

func (w *World) walkToRock(p *player, r *rock) {
	if p.job.rock != r {
		w.releaseRock(p)
	}
	r.claimedBy = p // reserve it now so two players don't walk to the same rock
	p.job.rock = r
	p.job.state = stateWalkToRock
	if !w.moveTo(p, r.tile()) {
		w.failAction(p, activityNoPath)
		return
	}
	w.startAction(p, activityToRock, walkTicks(p))
}

// --- banking ---

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

func (w *World) goBank(p *player) {
	w.releaseRock(p)
	b, ok := w.nearestBank(p)
	if !ok || !w.moveTo(p, b) {
		w.failAction(p, activityNoPath)
		return
	}
	p.job.state = stateWalkToBank
	w.startAction(p, activityToBank, walkTicks(p))
}

// atBank deposits everything carried, then either moves on or starts the
// next trip.
func (w *World) atBank(p *player) {
	j := p.job
	if !p.backpack.TransferAll(&p.bank) {
		p.invDirty = true
		w.failAction(p, activityBankFull)
		return
	}
	p.invDirty = true
	if j.finishing || amountMet(p, j.action) {
		w.finishAction(p, activityIdle)
		return
	}
	switch j.action.spec.Kind {
	case resources.ActionMine:
		w.mineNext(p) // back to the same rock if it's still free
	case resources.ActionSmelt:
		w.collectOre(p)
	}
}

// --- smelting ---

// collectOre takes ore from the bank for one smelting trip: as much as the
// backpack holds, but no more than the goal still needs.
func (w *World) collectOre(p *player) {
	j := p.job
	n := min(backpackSlots-len(p.backpack.Stacks), p.bank.Count(resources.ItemIronOre))
	if j.action.spec.Goal == resources.GoalAmount {
		n = min(n, j.action.spec.Amount-p.bank.Count(resources.ItemIronIngot))
	}
	if n <= 0 {
		w.finishAction(p, activityNoOre)
		return
	}
	p.bank.Remove(resources.ItemIronOre, n)
	p.backpack.Add(resources.ItemIronOre, n)
	p.invDirty = true
	j.state = stateWalkToSmith
	if !w.moveTo(p, j.smith) {
		w.failAction(p, activityNoPath)
		return
	}
	w.startAction(p, activityToSmith, walkTicks(p))
}

// --- the tick ---

// tickJob advances p's job by one tick. Called after p has moved.
func (w *World) tickJob(p *player) {
	j := p.job
	arrived := len(p.path) == 0
	if j == nil {
		switch {
		case p.running:
			w.startNextAction(p)
		case p.activity == activityWalking && arrived:
			w.setActivity(p, activityIdle)
		case p.activity == activityNoPath && p.ticks-p.actionStart >= ticks(noPathShown):
			w.setActivity(p, activityIdle)
		}
		return
	}

	j.action.ran++
	if !j.finishing && timeUp(j.action) {
		w.wrapUp(p)
		return
	}

	switch j.state {
	case stateWalkToRock:
		if !arrived {
			return
		}
		if !j.rock.freeFor(p) { // emptied while we walked
			w.mineNext(p)
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
		goal := j.action.spec
		enough := goal.Goal == resources.GoalAmount &&
			p.bank.Count(resources.ItemIronOre)+p.backpack.Count(resources.ItemIronOre) >= goal.Amount
		switch {
		case p.backpack.Full(resources.ItemIronOre) || enough:
			w.goBank(p)
		case r.ore == 0:
			w.mineNext(p)
		default:
			j.countdown = ticks(swingTime)
			w.startAction(p, activityMining, j.countdown)
		}

	case stateWaitForRock:
		j.countdown--
		if j.countdown > 0 {
			return
		}
		if r := w.nearestFreeRock(p, j.rock, j.action.spec.Place); r != nil {
			w.walkToRock(p, r)
		} else {
			j.countdown = ticks(rockWaitRetry)
		}

	case stateWalkToBank:
		if arrived {
			w.atBank(p)
		}

	case stateWalkToSmith:
		if !arrived {
			return
		}
		j.state = stateSmelting
		j.countdown = ticks(smeltTime)
		w.startAction(p, activitySmelting, j.countdown)

	case stateSmelting:
		j.countdown--
		if j.countdown > 0 {
			return
		}
		p.backpack.Remove(resources.ItemIronOre, 1)
		p.backpack.Add(resources.ItemIronIngot, 1)
		p.invDirty = true
		if p.backpack.Count(resources.ItemIronOre) == 0 {
			w.goBank(p)
			return
		}
		j.countdown = ticks(smeltTime)
		w.startAction(p, activitySmelting, j.countdown)
	}
}

// tickDiscovery marks places p walks into as found.
func (w *World) tickDiscovery(p *player) {
	if p.ticks%ticks(discoveryCheck) != 0 {
		return
	}
	tx, ty := resources.TileAt(p.pos[0], p.pos[1])
	for _, pl := range w.wm.Places {
		if !p.found[pl.Name] && max(abs(pl.X-tx), abs(pl.Y-ty)) <= pl.Radius+discoveryMargin {
			p.found[pl.Name] = true
			p.invDirty = true
		}
	}
}

// --- rocks ---

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
	var found []string
	for _, pl := range w.wm.Places {
		if p.found[pl.Name] {
			found = append(found, pl.Name)
		}
	}
	actions := make([]resources.QueuedAction, len(p.actions))
	for i, a := range p.actions {
		actions[i] = resources.QueuedAction{ActionSpec: a.spec, Banked: p.bank.Count(a.spec.Product())}
		if a.spec.Goal == resources.GoalTime {
			actions[i].SecondsLeft = seconds(a.ticksLeft())
		}
	}
	return resources.NewInventoryMessage(resources.InventoryContents{
		Backpack:      p.backpack.Clone(),
		Bank:          p.bank.Clone(),
		Activity:      p.activity,
		ActionSeconds: seconds(p.actionTicks),
		ActionElapsed: seconds(p.ticks - p.actionStart),
		Found:         found,
		Actions:       actions,
		Running:       p.running,
	})
}
