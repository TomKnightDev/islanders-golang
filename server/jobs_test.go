package main

import (
	"strings"
	"testing"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

const (
	testVillage = "Testville"
	testMine    = "Test Mine"
	testWood    = "Test Wood"
)

// jobWorld is an open 20x20 field. Testville (centre 4,4) has a bank at
// (2,2), a blacksmith at (6,2), a fletcher at (8,2) and an inn at (7,7);
// players spawn there.
// Test Mine (centre 12,10) has rocks at (10,10) and (14,10). Test Wood
// (centre 16,16) has trees at (15,15), (17,15) and (16,18).
func jobWorld(t *testing.T) *World {
	t.Helper()
	wm := testMap(20, 20)
	add := func(id uint16, kind resources.ObjectKind, x, y int) {
		gid := resources.ObjectTiles[kind]
		if kind == resources.ObjectRock {
			gid = resources.TileCopperRock
		}
		wm.Layers[resources.LayerObject].Data[y*wm.Width+x] = gid
		wm.Objects = append(wm.Objects, resources.WorldObject{ID: id, Kind: kind, X: x, Y: y})
	}
	add(1, resources.ObjectBank, 2, 2)
	add(2, resources.ObjectBlacksmith, 6, 2)
	add(3, resources.ObjectRock, 10, 10)
	add(4, resources.ObjectRock, 14, 10)
	add(5, resources.ObjectFletcher, 8, 2)
	add(6, resources.ObjectInn, 7, 7)
	for _, tr := range []tile{{15, 15}, {17, 15}, {16, 18}} {
		wm.Layers[resources.LayerObject].Data[tr.y*wm.Width+tr.x] = resources.TileTree1
	}
	wm.Places = []resources.Place{
		{Name: testVillage, Kind: resources.PlaceVillage, X: 4, Y: 4, Radius: 4},
		{Name: testMine, Kind: resources.PlaceMine, X: 12, Y: 10, Radius: 4, Materials: []string{"copper"}},
		{Name: testWood, Kind: resources.PlaceForest, X: 16, Y: 16, Radius: 3, Materials: []string{"oak"}},
	}
	wm.SpawnX, wm.SpawnY = 4, 5
	return NewWorld(wm)
}

// joinAt logs a player in at a tile, already knowing both places.
func joinAt(t *testing.T, w *World, name string, at tile) *player {
	t.Helper()
	p, err := w.login(name, "pw", testSession())
	if err != nil {
		t.Fatal(err)
	}
	p.pos[0], p.pos[1] = at.pixel()
	p.found[testMine] = true
	p.found[testWood] = true
	return p
}

func mineFor(amount int) resources.ActionSpec {
	return resources.ActionSpec{Kind: resources.ActionMine, Material: "copper", Place: testMine, Goal: resources.GoalAmount, Amount: amount}
}

func smeltFor(amount int) resources.ActionSpec {
	return resources.ActionSpec{Kind: resources.ActionSmelt, Material: "copper", Place: testVillage, Goal: resources.GoalAmount, Amount: amount}
}

func addAction(w *World, p *player, a resources.ActionSpec) {
	w.actionCommand(p, resources.ActionCommandContents{Op: resources.ActionOpAdd, Action: a})
}

func run(w *World, p *player) {
	w.actionCommand(p, resources.ActionCommandContents{Op: resources.ActionOpStart})
}

// runUntil ticks the world until cond holds, failing after max ticks.
func runUntil(t *testing.T, w *World, max int, what string, cond func() bool) {
	t.Helper()
	for i := 0; i < max; i++ {
		if cond() {
			return
		}
		w.tick()
	}
	t.Fatalf("timed out waiting for: %s", what)
}

func ore(inv resources.Inventory) int  { return inv.Count(resources.ItemCopperOre) }
func bars(inv resources.Inventory) int { return inv.Count(resources.ItemCopperBar) }

func TestMineUntilBankHasAmount(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{8, 10})
	addAction(w, p, mineFor(25))
	run(w, p)

	runUntil(t, w, 20000, "action finished", func() bool { return len(p.actions) == 0 })
	// A full backpack of 20, then just the 5 still needed.
	if ore(p.bank) != 25 || !p.backpack.Empty() {
		t.Errorf("bank %d ore, backpack %v", ore(p.bank), p.backpack.Stacks)
	}
	if p.running || p.activity != activityIdle {
		t.Errorf("should be idle after the last action: running %v, %q", p.running, p.activity)
	}
	if w.nodes[tile{10, 10}].claimedBy != nil {
		t.Error("rock should be released")
	}
}

func TestMineForTime(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionMine, Material: "copper", Place: testMine, Goal: resources.GoalTime, Minutes: 1})
	run(w, p)

	runUntil(t, w, ticks(swingTime)*40, "time up", func() bool { return p.job != nil && p.job.finishing })
	if p.job.state != stateWalkToBank {
		t.Fatal("time up should send the miner to bank what's carried")
	}
	runUntil(t, w, 5000, "finished", func() bool { return len(p.actions) == 0 })
	// One minute of swings is at most 60s/swingTime ore.
	maxOre := int(time.Minute / swingTime)
	if got := ore(p.bank); got < maxOre-3 || got > maxOre || !p.backpack.Empty() {
		t.Errorf("bank %d ore, backpack %v", got, p.backpack.Stacks)
	}
}

func TestSmeltUntilBankHasAmount(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Add(resources.ItemCopperOre, 30)
	addAction(w, p, smeltFor(25))
	run(w, p)

	runUntil(t, w, 20000, "action finished", func() bool { return len(p.actions) == 0 })
	// 20 ore on the first trip, then only the 5 more needed.
	if bars(p.bank) != 25 || ore(p.bank) != 5 || !p.backpack.Empty() {
		t.Errorf("bank %d bars %d ore, backpack %v", bars(p.bank), ore(p.bank), p.backpack.Stacks)
	}
}

func TestSmeltStopsWhenOreRunsOut(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Add(resources.ItemCopperOre, 7)
	addAction(w, p, smeltFor(100))
	run(w, p)

	runUntil(t, w, 20000, "action finished", func() bool { return len(p.actions) == 0 })
	if bars(p.bank) != 7 || ore(p.bank) != 0 {
		t.Errorf("bank %d bars %d ore", bars(p.bank), ore(p.bank))
	}
	if p.activity != activityNoOre {
		t.Errorf("activity %q, want %q", p.activity, activityNoOre)
	}
}

func TestSmeltTiming(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Add(resources.ItemCopperOre, 3)
	addAction(w, p, smeltFor(3))
	run(w, p)
	runUntil(t, w, 2000, "smelting", func() bool { return p.job != nil && p.job.state == stateCrafting })
	for i := 0; i < ticks(smeltTime)-1; i++ {
		w.tick()
	}
	if bars(p.backpack) != 0 {
		t.Fatal("bar made before the smelt finished")
	}
	w.tick()
	if bars(p.backpack) != 1 || ore(p.backpack) != 2 {
		t.Errorf("after one smelt: %d bars, %d ore", bars(p.backpack), ore(p.backpack))
	}
}

func TestListRunsInOrder(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{8, 10})
	addAction(w, p, mineFor(20))
	addAction(w, p, smeltFor(20))
	run(w, p)

	runUntil(t, w, 40000, "list finished", func() bool { return len(p.actions) == 0 })
	if bars(p.bank) != 20 || ore(p.bank) != 0 {
		t.Errorf("bank %d bars %d ore", bars(p.bank), ore(p.bank))
	}
}

func TestClickDoesNotCancelARunningList(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, mineFor(50))
	run(w, p)
	runUntil(t, w, 100, "mining", func() bool { return p.job != nil && p.job.state == stateGathering })

	j := p.job
	w.click(p, tile{1, 18})
	w.tick()
	if p.job != j || p.job.state != stateGathering || len(p.path) != 0 {
		t.Error("a click while the list runs should change nothing")
	}
}

func TestClickWalksWhenStopped(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{5, 5})
	w.click(p, tile{7, 5})
	if p.activity != activityWalking || len(p.path) == 0 {
		t.Fatalf("activity %q, path %v", p.activity, p.path)
	}
	runUntil(t, w, 200, "arrive", func() bool { return len(p.path) == 0 })
	w.tick()
	if p.activity != activityIdle {
		t.Errorf("after arriving: %q", p.activity)
	}
}

func TestPlacesMustBeFound(t *testing.T) {
	w := jobWorld(t)
	p, _ := w.login("new", "pw", testSession())
	if !p.found[testVillage] || p.found[testMine] {
		t.Fatalf("new players know only their spawn village: %v", p.found)
	}
	addAction(w, p, mineFor(10))
	if len(p.actions) != 0 {
		t.Fatal("can't add an action at a place not yet found")
	}

	// Walk into the mine to find it.
	w.click(p, tile{12, 13})
	runUntil(t, w, 2000, "found the mine", func() bool { return p.found[testMine] })
	addAction(w, p, mineFor(10))
	if len(p.actions) != 1 {
		t.Error("found place should accept actions")
	}
}

func TestActionValidation(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	bad := []resources.ActionSpec{
		{Kind: resources.ActionMine, Material: "copper", Place: testVillage, Goal: resources.GoalAmount, Amount: 5}, // not a mine
		{Kind: resources.ActionSmelt, Material: "copper", Place: testMine, Goal: resources.GoalAmount, Amount: 5},   // no blacksmith
		{Kind: resources.ActionMine, Material: "copper", Place: testMine, Goal: resources.GoalAmount, Amount: 0},
		{Kind: resources.ActionMine, Material: "copper", Place: testMine, Goal: resources.GoalTime, Minutes: 0},
		{Kind: "fish", Place: testMine, Goal: resources.GoalAmount, Amount: 5},
	}
	for _, a := range bad {
		addAction(w, p, a)
	}
	if len(p.actions) != 0 {
		t.Errorf("invalid actions were accepted: %d", len(p.actions))
	}
	for i := 0; i < resources.MaxActions+2; i++ {
		addAction(w, p, mineFor(10))
	}
	if len(p.actions) != resources.MaxActions {
		t.Errorf("list holds %d, want at most %d", len(p.actions), resources.MaxActions)
	}
}

func TestReorderAndRemove(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, mineFor(50))
	addAction(w, p, smeltFor(5))
	run(w, p)
	runUntil(t, w, 100, "mining", func() bool { return p.job != nil && p.job.state == stateGathering })

	// Moving smelting to the top stops mining and starts smelting.
	w.actionCommand(p, resources.ActionCommandContents{Op: resources.ActionOpUp, Index: 1})
	if p.actions[0].spec.Kind != resources.ActionSmelt || p.job != nil {
		t.Fatal("reordering should cancel the old first action")
	}
	if w.nodes[tile{10, 10}].claimedBy != nil {
		t.Error("cancelled miner should free the rock")
	}
	w.actionCommand(p, resources.ActionCommandContents{Op: resources.ActionOpRemove, Index: 0})
	if len(p.actions) != 1 || p.actions[0].spec.Kind != resources.ActionMine {
		t.Errorf("after removing: %+v", p.actions)
	}
}

func TestStopPausesTimer(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionMine, Material: "copper", Place: testMine, Goal: resources.GoalTime, Minutes: 5})
	run(w, p)
	for i := 0; i < 100; i++ {
		w.tick()
	}
	w.actionCommand(p, resources.ActionCommandContents{Op: resources.ActionOpStop})
	ran := p.actions[0].ran
	for i := 0; i < 1000; i++ {
		w.tick()
	}
	if p.actions[0].ran != ran || p.job != nil || p.running {
		t.Errorf("stopped list kept running: ran %d -> %d", ran, p.actions[0].ran)
	}
}

func TestOneMinerPerRock(t *testing.T) {
	w := jobWorld(t)
	a := joinAt(t, w, "a", tile{9, 9})
	b := joinAt(t, w, "b", tile{9, 11})
	for _, p := range []*player{a, b} {
		addAction(w, p, mineFor(50))
		run(w, p)
	}
	w.tick()
	if a.job.node == nil || b.job.node == nil || a.job.node == b.job.node {
		t.Fatal("two miners should be on different rocks")
	}
}

func TestWaitWhenNoRockIsFree(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{14, 10}].amount = 0 // depleted, and not respawning during the test
	w.nodes[tile{14, 10}].respawnAt = 1 << 30
	a := joinAt(t, w, "a", tile{9, 9})
	b := joinAt(t, w, "b", tile{9, 11})
	addAction(w, a, mineFor(50))
	run(w, a)
	w.tick()
	addAction(w, b, mineFor(50))
	run(w, b)
	w.tick()
	if b.job.state != stateWaitForNode || b.activity != activityWaiting {
		t.Fatalf("b should wait, state %v activity %q", b.job.state, b.activity)
	}
	w.actionCommand(a, resources.ActionCommandContents{Op: resources.ActionOpStop})
	runUntil(t, w, 200, "b claims the freed rock", func() bool { return w.nodes[tile{10, 10}].claimedBy == b })
}

func TestRockDepletesMovesOnAndRespawns(t *testing.T) {
	w := jobWorld(t)
	r1, r2 := w.nodes[tile{10, 10}], w.nodes[tile{14, 10}]
	r1.amount, r2.amount = 1, 50
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.found[testMine] = true
	p.pos[0], p.pos[1] = tile{9, 10}.pixel()
	addAction(w, p, mineFor(50))
	run(w, p)

	runUntil(t, w, 500, "rock depleted", func() bool { return r1.amount == 0 })
	if got := w.wm.Layers[resources.LayerObject].Data[10*w.wm.Width+10]; got != resources.TileIronRockDepleted {
		t.Errorf("depleted rock tile %d", got)
	}
	if p.job.node != r2 || r2.claimedBy != p {
		t.Error("miner should move on to the other rock")
	}
	sawUpdate := false
	for _, m := range drain(s) {
		if m.MessageType == resources.ObjectUpdateMessage {
			sawUpdate = true
		}
	}
	if !sawUpdate {
		t.Error("clients should be told the rock changed")
	}
	runUntil(t, w, ticks(rockRespawn)+10, "respawn", func() bool { return r1.amount > 0 })
	if r1.amount < rockOreMin || r1.amount > rockOreMax {
		t.Errorf("respawned with %d ore", r1.amount)
	}
}

func TestBankFullStopsTheList(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	p.bank = resources.NewInventory(1, true)
	p.bank.Add(resources.ItemCopperOre, 5) // room for 5 more only
	addAction(w, p, mineFor(50))
	run(w, p)

	runUntil(t, w, 20000, "list stops", func() bool { return !p.running })
	if p.activity != activityBankFull {
		t.Errorf("activity %q", p.activity)
	}
	if ore(p.bank) != 10 || ore(p.backpack) != 15 {
		t.Errorf("bank %d, backpack %d", ore(p.bank), ore(p.backpack))
	}
	if len(p.actions) != 1 {
		t.Error("the action should stay in the list so the player can fix things and restart")
	}
}

func TestFullBackpackIsTwentyOre(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, mineFor(100))
	run(w, p)
	runUntil(t, w, ticks(swingTime)*backpackSlots*2, "head to the bank", func() bool { return p.job != nil && p.job.state == stateWalkToBank })
	if got := ore(p.backpack); got != backpackSlots {
		t.Errorf("went to the bank carrying %d ore, want %d", got, backpackSlots)
	}
}

func TestBankAccountsArePrivate(t *testing.T) {
	w := jobWorld(t)
	a := joinAt(t, w, "a", tile{4, 5})
	b := joinAt(t, w, "b", tile{4, 6})
	a.bank.Add(resources.ItemCopperOre, 3)
	addAction(w, a, smeltFor(3))
	run(w, a)
	runUntil(t, w, 5000, "a done", func() bool { return len(a.actions) == 0 })
	if bars(b.bank) != 0 || bars(a.bank) != 3 {
		t.Error("smelting leaked into another player's account")
	}
}

func TestMiningContinuesOfflineAtFullSpeed(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.found[testMine] = true
	p.pos[0], p.pos[1] = tile{9, 10}.pixel()
	addAction(w, p, mineFor(100))
	run(w, p)
	w.tick() // start: walk to the rock
	w.tick() // arrive and start the first swing while online
	w.disconnect(p, s)

	// Three swings' time offline gives three ore, the same as online.
	for i := 0; i < ticks(swingTime)*3; i++ {
		w.tick()
	}
	if got := ore(p.backpack); got != 3 {
		t.Errorf("offline for three swings' time: mined %d ore, want 3", got)
	}
}

// lastInventory returns the most recent inventory message queued for s.
func lastInventory(t *testing.T, s *session) resources.InventoryContents {
	t.Helper()
	var last *resources.Message
	for _, m := range drain(s) {
		if m.MessageType == resources.InventoryMessage {
			last = m
		}
	}
	if last == nil {
		t.Fatal("no inventory message sent")
	}
	return last.Contents.(resources.InventoryContents)
}

func TestInventoryReportsListAndPlaces(t *testing.T) {
	w := jobWorld(t)
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	inv := lastInventory(t, s)
	if len(inv.Found) != 1 || inv.Found[0] != testVillage {
		t.Errorf("found %v", inv.Found)
	}
	p.bank.Add(resources.ItemCopperBar, 4)
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionSmelt, Material: "copper", Place: testVillage, Goal: resources.GoalTime, Minutes: 2})
	addAction(w, p, smeltFor(10))
	w.tick()
	inv = lastInventory(t, s)
	if len(inv.Actions) != 2 || inv.Actions[0].SecondsLeft != 120 || inv.Actions[1].Banked != 4 {
		t.Errorf("actions %+v", inv.Actions)
	}
}

func TestActionTimingForProgressBar(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.found[testMine] = true
	p.pos[0], p.pos[1] = tile{2, 10}.pixel()
	addAction(w, p, mineFor(100))
	run(w, p)

	// Walking 7 tiles to stand beside the rock at (9,10) takes 7/walkSpeed seconds.
	w.tick()
	inv := lastInventory(t, s)
	if inv.Activity != activityToRock || inv.ActionSeconds != 7/walkSpeed {
		t.Fatalf("walk: %q for %vs, want %q for %vs", inv.Activity, inv.ActionSeconds, activityToRock, 7/walkSpeed)
	}

	// Each swing restarts the bar at 0 for swingTime.
	runUntil(t, w, 500, "first ore", func() bool { return ore(p.backpack) == 1 })
	inv = lastInventory(t, s)
	if inv.Activity != activityMining || inv.ActionSeconds != swingTime.Seconds() || inv.ActionElapsed != 0 {
		t.Errorf("swing: %q %vs elapsed %vs", inv.Activity, inv.ActionSeconds, inv.ActionElapsed)
	}
}

func logs(inv resources.Inventory) int   { return inv.Count(resources.ItemOakLogs) }
func arrows(inv resources.Inventory) int { return inv.Count(resources.ItemOakArrows) }

func TestChopWoodFellsAndRegrowsTrees(t *testing.T) {
	w := jobWorld(t)
	for _, n := range w.nodes {
		if n.place == testWood {
			n.amount = 2
		}
	}
	p := joinAt(t, w, "tom", tile{14, 16})
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionChop, Material: "oak", Place: testWood, Goal: resources.GoalAmount, Amount: 5})
	run(w, p)

	runUntil(t, w, 20000, "chopped 5 logs", func() bool { return len(p.actions) == 0 })
	if logs(p.bank) != 5 || !p.backpack.Empty() {
		t.Errorf("bank %d logs, backpack %v", logs(p.bank), p.backpack.Stacks)
	}
	// Two logs a tree, so at least two trees came down.
	felled := 0
	for _, tr := range []tile{{15, 15}, {17, 15}, {16, 18}} {
		if w.wm.Layers[resources.LayerObject].Data[tr.y*w.wm.Width+tr.x] == resources.TileFelledTree {
			felled++
		}
	}
	if felled < 2 {
		t.Errorf("%d trees felled, want at least 2", felled)
	}
	runUntil(t, w, ticks(treeRegrow)+10, "trees regrow", func() bool {
		return w.wm.Layers[resources.LayerObject].Data[15*w.wm.Width+15] == resources.TileTree1
	})
}

func TestFletchArrows(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Add(resources.ItemOakLogs, 25)
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionFletch, Material: "oak", Place: testVillage, Goal: resources.GoalAmount, Amount: 200})
	run(w, p)

	runUntil(t, w, 20000, "fletched", func() bool { return len(p.actions) == 0 })
	// 200 arrows needs 20 logs: one trip, 5 logs left over.
	if arrows(p.bank) != 200 || logs(p.bank) != 5 {
		t.Errorf("bank %d arrows %d logs", arrows(p.bank), logs(p.bank))
	}
	// 200 arrows at 50 a slot.
	slots := 0
	for _, s := range p.bank.Stacks {
		if s.Kind == resources.ItemOakArrows {
			slots++
		}
	}
	if slots != 4 {
		t.Errorf("200 arrows took %d bank slots, want 4", slots)
	}
}

func TestFletchingFillsBackpackWithStackedArrows(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Add(resources.ItemOakLogs, 20)
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionFletch, Material: "oak", Place: testVillage, Goal: resources.GoalAmount, Amount: 1000})
	run(w, p)
	runUntil(t, w, 20000, "heading back with arrows", func() bool {
		return p.job != nil && p.job.state == stateWalkToBank && arrows(p.backpack) > 0
	})
	if arrows(p.backpack) != 200 || len(p.backpack.Stacks) != 4 {
		t.Errorf("carrying %d arrows in %d slots, want 200 in 4", arrows(p.backpack), len(p.backpack.Stacks))
	}
}

func TestHigherTiersNeedLevels(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	iron := resources.ActionSpec{Kind: resources.ActionSmelt, Material: "iron", Place: testVillage, Goal: resources.GoalAmount, Amount: 5}
	addAction(w, p, iron)
	if len(p.actions) != 0 {
		t.Fatal("level 1 smith shouldn't be able to smelt iron")
	}
	p.xp[resources.SkillSmithing] = resources.XPForLevel(15)
	addAction(w, p, iron)
	if len(p.actions) != 1 {
		t.Fatal("level 15 smith should be able to smelt iron")
	}
	// A mine only offers the ore it has.
	p.xp[resources.SkillMining] = resources.XPForLevel(99)
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionMine, Material: "gold", Place: testMine, Goal: resources.GoalAmount, Amount: 5})
	if len(p.actions) != 1 {
		t.Error("Test Mine has no gold")
	}
}

func TestGatheringEarnsXPAndLevelsUp(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.found[testMine] = true
	p.pos[0], p.pos[1] = tile{9, 10}.pixel()
	addAction(w, p, mineFor(9))
	run(w, p)
	runUntil(t, w, 20000, "mined", func() bool { return len(p.actions) == 0 })

	copper, _ := resources.FindMaterial(resources.ActionMine, "copper")
	if got := p.xp[resources.SkillMining]; got != 9*copper.XP {
		t.Errorf("mining XP %d, want %d", got, 9*copper.XP)
	}
	if p.level(resources.SkillMining) != 2 {
		t.Errorf("90 XP should be level 2, got %d", p.level(resources.SkillMining))
	}
	congratulated := false
	for _, m := range drain(s) {
		if m.MessageType == resources.ChatMessage && strings.Contains(m.Contents.(string), "Mining level is now 2") {
			congratulated = true
		}
	}
	if !congratulated {
		t.Error("no level-up message")
	}
}

func TestHigherLevelsWorkFaster(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	base := ticks(swingTime)
	if got := p.skillTicks(swingTime, resources.SkillMining); got != base {
		t.Errorf("level 1: %d ticks, want %d", got, base)
	}
	p.xp[resources.SkillMining] = resources.XPForLevel(99)
	if got := float64(p.skillTicks(swingTime, resources.SkillMining)) / float64(base); got > 0.62 || got < 0.6 {
		t.Errorf("level 99 takes %.2f of the base time, want about 0.61", got)
	}
}

func TestFarming(t *testing.T) {
	w := jobWorld(t)
	// A small farm: two cabbages and one cotton plant, each with room around it.
	for _, c := range []struct {
		t   tile
		gid int
	}{{tile{3, 15}, resources.TileCabbage}, {tile{5, 15}, resources.TileCabbage}, {tile{3, 17}, resources.TileCotton}} {
		w.wm.Layers[resources.LayerObject].Data[c.t.y*w.wm.Width+c.t.x] = c.gid
	}
	w.wm.Places = append(w.wm.Places, resources.Place{Name: "Test Farm", Kind: resources.PlaceFarm, X: 4, Y: 16, Radius: 2, Materials: []string{"cabbage", "cotton"}})
	w.initNodes()
	p := joinAt(t, w, "tom", tile{4, 13})
	p.found["Test Farm"] = true
	p.bank.Remove(resources.ItemCabbage, startingFood) // start without the starting kit's food

	addAction(w, p, resources.ActionSpec{Kind: resources.ActionFarm, Material: "cotton", Place: "Test Farm", Goal: resources.GoalAmount, Amount: 3})
	if len(p.actions) != 0 {
		t.Fatal("cotton needs Farming 10")
	}
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionFarm, Material: "cabbage", Place: "Test Farm", Goal: resources.GoalAmount, Amount: 3})
	run(w, p)
	runUntil(t, w, 20000, "harvested", func() bool { return len(p.actions) == 0 })
	if got := p.bank.Count(resources.ItemCabbage); got != 3 {
		t.Errorf("bank has %d cabbages, want 3", got)
	}
	if p.xp[resources.SkillFarming] == 0 {
		t.Error("farming should train Farming")
	}
}

func TestPathIsSentOnlyToItsPlayer(t *testing.T) {
	w := jobWorld(t)
	sa, sb := testSession(), testSession()
	a, _ := w.login("a", "pw", sa)
	w.login("b", "pw", sb)
	a.pos[0], a.pos[1] = tile{5, 5}.pixel()
	drain(sa)
	drain(sb)

	w.click(a, tile{9, 5})
	w.tick()
	var path *resources.PathContents
	for _, m := range drain(sa) {
		if m.MessageType == resources.PathMessage {
			c := m.Contents.(resources.PathContents)
			path = &c
		}
	}
	if path == nil || len(path.Tiles) == 0 || path.Tiles[len(path.Tiles)-1] != [2]int{9, 5} {
		t.Fatalf("walker should get their path to (9,5): %+v", path)
	}
	for _, m := range drain(sb) {
		// b only ever hears about its own (empty) path.
		if m.MessageType == resources.PathMessage && len(m.Contents.(resources.PathContents).Tiles) > 0 {
			t.Fatal("another player was sent the path")
		}
	}
	// Arriving clears it.
	runUntil(t, w, 2000, "arrive", func() bool { return len(a.path) == 0 })
	w.tick()
	cleared := false
	for _, m := range drain(sa) {
		if m.MessageType == resources.PathMessage && len(m.Contents.(resources.PathContents).Tiles) == 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("arriving should send an empty path")
	}
}
