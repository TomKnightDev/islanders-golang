package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// Tuning values. See docs/DESIGN.md.
const (
	swingTime       = 2 * time.Second // per ore
	chopTime        = 3 * time.Second // per log
	smeltTime       = 3 * time.Second // per bar
	fletchTime      = 2 * time.Second // per log
	arrowsPerLog    = 10
	rockRespawn     = 30 * time.Second
	rockOreMin      = 5
	rockOreMax      = 10
	treeRegrow      = 60 * time.Second
	treeLogsMin     = 1
	treeLogsMax     = 4
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
	activityChopping = "Chopping"
	activityToTree   = "Walking to a tree"
	activityNoTree   = "Waiting for a free tree"
	activityToSmith  = "Walking to the blacksmith"
	activitySmelting = "Smelting"
	activityNoOre    = "No ore to smelt"
	activityToFletch = "Walking to the fletcher"
	activityFletch   = "Fletching"
	activityNoLogs   = "No logs to fletch"
	activityBankFull = "Bank full"
	activityNoPath   = "Can't get there"
)

// gatherKind describes a gathering action: what it collects, from what,
// and how fast.
type gatherKind struct {
	item               resources.ItemKind
	swing              time.Duration
	minYield, maxYield int // how much a node holds when full
	respawn            time.Duration
	spent              int // sprite while the node is empty
	working, walking   string
	waiting            string
}

var gatherKinds = map[resources.ActionKind]*gatherKind{
	resources.ActionMine: {
		item: resources.ItemIronOre, swing: swingTime, minYield: rockOreMin, maxYield: rockOreMax,
		respawn: rockRespawn, spent: resources.TileIronRockDepleted,
		working: activityMining, walking: activityToRock, waiting: activityWaiting,
	},
	resources.ActionChop: {
		item: resources.ItemLogs, swing: chopTime, minYield: treeLogsMin, maxYield: treeLogsMax,
		respawn: treeRegrow, spent: resources.TileFelledTree,
		working: activityChopping, walking: activityToTree, waiting: activityNoTree,
	},
}

// recipe describes a crafting action: turn an input from the bank into an
// output at a station in a village.
type recipe struct {
	input, output    resources.ItemKind
	perInput         int // outputs made from one input
	step             time.Duration
	station          resources.ObjectKind
	working, walking string
	noInput          string
}

var recipes = map[resources.ActionKind]*recipe{
	resources.ActionSmelt: {
		input: resources.ItemIronOre, output: resources.ItemIronIngot, perInput: 1, step: smeltTime,
		station: resources.ObjectBlacksmith, working: activitySmelting, walking: activityToSmith, noInput: activityNoOre,
	},
	resources.ActionFletch: {
		input: resources.ItemLogs, output: resources.ItemArrows, perInput: arrowsPerLog, step: fletchTime,
		station: resources.ObjectFletcher, working: activityFletch, walking: activityToFletch, noInput: activityNoLogs,
	},
}

// node is something to gather from: an iron rock or a tree.
type node struct {
	kind        *gatherKind
	t           tile
	place       string // the mine or forest it belongs to
	live        int    // sprite while it has something to gather
	amount      int
	claimedBy   *player
	respawnAt   int  // world tick; only meaningful while empty
	unreachable bool // no walkable tile beside it
}

func (n *node) tile() tile { return n.t }

// freeFor reports whether p may gather from n now.
func (n *node) freeFor(p *player) bool {
	return n.amount > 0 && !n.unreachable && (n.claimedBy == nil || n.claimedBy == p)
}

func (n *node) refill() {
	n.amount = n.kind.minYield + rand.Intn(n.kind.maxYield-n.kind.minYield+1)
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
	stateWalkToNode jobState = iota + 1
	stateGathering
	stateWaitForNode
	stateWalkToBank
	stateWalkToStation
	stateCrafting
)

// job is the running state of the action at the top of the list.
type job struct {
	action    *queuedAction
	state     jobState
	gather    *gatherKind // gathering actions
	node      *node       // the node being walked to or gathered, and returned to after banking
	recipe    *recipe     // crafting actions
	station   tile        // where to craft
	countdown int         // ticks left in a swing, a craft step or a wait
	finishing bool        // goal met or time up: bank what's carried, then move on
}

// initNodes finds everything to gather from and every crafting station.
// Rocks are listed objects; trees are any tree tile inside a named forest.
func (w *World) initNodes() {
	w.nodes = map[tile]*node{}
	w.stations = map[resources.ObjectKind]map[string]tile{}
	objects := w.wm.Layers[resources.LayerObject].Data
	for _, o := range w.wm.Objects {
		t := tile{o.X, o.Y}
		switch o.Kind {
		case resources.ObjectIronRock:
			n := &node{kind: gatherKinds[resources.ActionMine], t: t, live: objects[t.y*w.wm.Width+t.x]}
			if pl := w.placeAt(t, resources.PlaceMine); pl != nil {
				n.place = pl.Name
			}
			n.refill()
			w.nodes[t] = n
		case resources.ObjectBank:
			w.banks = append(w.banks, t)
		case resources.ObjectBlacksmith, resources.ObjectFletcher:
			if pl := w.placeAt(t, resources.PlaceVillage); pl != nil {
				if w.stations[o.Kind] == nil {
					w.stations[o.Kind] = map[string]tile{}
				}
				w.stations[o.Kind][pl.Name] = t
			}
		}
	}
	for _, pl := range w.wm.Places {
		if pl.Kind != resources.PlaceForest {
			continue
		}
		for y := pl.Y - pl.Radius; y <= pl.Y+pl.Radius; y++ {
			for x := pl.X - pl.Radius; x <= pl.X+pl.Radius; x++ {
				t := tile{x, y}
				if x < 0 || y < 0 || x >= w.wm.Width || y >= w.wm.Height || w.nodes[t] != nil {
					continue
				}
				if gid := objects[y*w.wm.Width+x]; resources.IsTree(gid) {
					n := &node{kind: gatherKinds[resources.ActionChop], t: t, place: pl.Name, live: gid}
					n.refill()
					w.nodes[t] = n
				}
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
	if r := recipes[a.Kind]; r != nil {
		if _, ok := w.stations[r.station][a.Place]; !ok {
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
	w.releaseNode(p)
	p.job = nil
	p.path = p.path[:min(1, len(p.path))] // finish the current step, no further
	w.setActivity(p, activityIdle)
}

func (w *World) releaseNode(p *player) {
	if p.job != nil && p.job.node != nil && p.job.node.claimedBy == p {
		p.job.node.claimedBy = nil
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
	if g := gatherKinds[a.spec.Kind]; g != nil {
		p.job = &job{action: a, gather: g}
		w.gatherNext(p)
		return
	}
	if r := recipes[a.spec.Kind]; r != nil {
		p.job = &job{action: a, recipe: r, station: w.stations[r.station][a.spec.Place]}
		w.goBank(p) // a crafting trip starts at the bank, collecting its input
	}
}

// finishAction moves on to the next action. note is shown if the list is
// now empty.
func (w *World) finishAction(p *player, note string) {
	w.releaseNode(p)
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
	w.releaseNode(p)
	p.job = nil
	p.running = false
	w.setActivity(p, why)
}

// wrapUp ends the action once anything carried has been banked.
func (w *World) wrapUp(p *player) {
	p.job.finishing = true
	w.releaseNode(p)
	if p.backpack.Empty() {
		w.finishAction(p, activityIdle)
		return
	}
	w.goBank(p)
}

// --- gathering ---

// nearestFreeNode prefers the given node if p can use it, otherwise the
// closest node of the same kind in the place that nobody else is using.
func (w *World) nearestFreeNode(p *player, prefer *node, place string, kind *gatherKind) *node {
	if prefer != nil && prefer.place == place && prefer.freeFor(p) {
		return prefer
	}
	from := p.currentTile()
	var best *node
	bestDist := 0
	for _, n := range w.nodes {
		if n.kind != kind || n.place != place || !n.freeFor(p) {
			continue
		}
		if d := octile(from, n.t); best == nil || d < bestDist ||
			(d == bestDist && (n.t.y < best.t.y || (n.t.y == best.t.y && n.t.x < best.t.x))) {
			best, bestDist = n, d
		}
	}
	return best
}

// gatherNext sends p to a free node in the action's place, or to wait there.
func (w *World) gatherNext(p *player) {
	j := p.job
	place := j.action.spec.Place
	for {
		n := w.nearestFreeNode(p, j.node, place, j.gather)
		if n == nil {
			break
		}
		if w.walkToNode(p, n) {
			return
		}
		if p.job == nil {
			return
		}
	}
	if j.node == nil {
		// Everything is busy and we're not there yet: walk to the place.
		if pl := w.place(place); pl != nil {
			w.moveTo(p, tile{pl.X, pl.Y})
		}
	}
	j.state = stateWaitForNode
	j.countdown = ticks(rockWaitRetry)
	w.setActivity(p, j.gather.waiting)
}

// walkToNode reserves n and walks p beside it. A node with no way to it is
// marked unreachable and skipped from then on.
func (w *World) walkToNode(p *player, n *node) bool {
	if p.job.node != n {
		w.releaseNode(p)
	}
	if !w.moveTo(p, n.t) {
		n.unreachable = true
		return false
	}
	n.claimedBy = p // reserve it now so two players don't head for the same one
	p.job.node = n
	p.job.state = stateWalkToNode
	w.startAction(p, p.job.gather.walking, walkTicks(p))
	return true
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
	w.releaseNode(p)
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
	if j.gather != nil {
		w.gatherNext(p) // back to the same node if it's still free
	} else {
		w.collectInput(p)
	}
}

// --- crafting ---

// collectInput takes a crafting trip's input from the bank: as much as the
// backpack holds, but no more than the goal still needs.
func (w *World) collectInput(p *player) {
	j := p.job
	r := j.recipe
	n := min(p.backpack.FreeSlots(), p.bank.Count(r.input))
	if j.action.spec.Goal == resources.GoalAmount {
		need := j.action.spec.Amount - p.bank.Count(r.output)
		n = min(n, (need+r.perInput-1)/r.perInput)
	}
	if n <= 0 {
		w.finishAction(p, r.noInput)
		return
	}
	p.bank.Remove(r.input, n)
	p.backpack.Add(r.input, n)
	p.invDirty = true
	j.state = stateWalkToStation
	if !w.moveTo(p, j.station) {
		w.failAction(p, activityNoPath)
		return
	}
	w.startAction(p, r.walking, walkTicks(p))
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
	case stateWalkToNode:
		if !arrived {
			return
		}
		if !j.node.freeFor(p) { // emptied while we walked
			w.gatherNext(p)
			return
		}
		j.state = stateGathering
		j.countdown = ticks(j.gather.swing)
		w.startAction(p, j.gather.working, j.countdown)

	case stateGathering:
		j.countdown--
		if j.countdown > 0 {
			return
		}
		n, item := j.node, j.gather.item
		n.amount--
		p.backpack.Add(item, 1)
		p.invDirty = true
		if n.amount == 0 {
			w.empty(n)
		}
		goal := j.action.spec
		enough := goal.Goal == resources.GoalAmount && p.bank.Count(item)+p.backpack.Count(item) >= goal.Amount
		switch {
		case p.backpack.Full(item) || enough:
			w.goBank(p)
		case n.amount == 0:
			w.gatherNext(p)
		default:
			j.countdown = ticks(j.gather.swing)
			w.startAction(p, j.gather.working, j.countdown)
		}

	case stateWaitForNode:
		j.countdown--
		if j.countdown > 0 {
			return
		}
		if n := w.nearestFreeNode(p, j.node, j.action.spec.Place, j.gather); n != nil {
			w.gatherNext(p)
		} else {
			j.countdown = ticks(rockWaitRetry)
		}

	case stateWalkToBank:
		if arrived {
			w.atBank(p)
		}

	case stateWalkToStation:
		if !arrived {
			return
		}
		j.state = stateCrafting
		j.countdown = ticks(j.recipe.step)
		w.startAction(p, j.recipe.working, j.countdown)

	case stateCrafting:
		j.countdown--
		if j.countdown > 0 {
			return
		}
		r := j.recipe
		p.backpack.Remove(r.input, 1)
		p.backpack.Add(r.output, r.perInput) // the input's slot is free, so this fits
		p.invDirty = true
		if p.backpack.Count(r.input) == 0 {
			w.goBank(p)
			return
		}
		j.countdown = ticks(r.step)
		w.startAction(p, r.working, j.countdown)
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

// --- nodes ---

func (w *World) setObjectTile(t tile, gid int) {
	w.wm.Layers[resources.LayerObject].Data[t.y*w.wm.Width+t.x] = gid
	w.broadcast(resources.NewObjectUpdateMessage(resources.ObjectUpdateContents{X: t.x, Y: t.y, Tile: gid}))
}

// empty turns a gathered-out node into its spent sprite until it refills.
func (w *World) empty(n *node) {
	n.claimedBy = nil
	n.respawnAt = w.ticks + ticks(n.kind.respawn)
	w.setObjectTile(n.t, n.kind.spent)
}

func (w *World) tickNodes() {
	for _, n := range w.nodes {
		if n.amount == 0 && w.ticks >= n.respawnAt {
			n.refill()
			w.setObjectTile(n.t, n.live)
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
