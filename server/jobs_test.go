package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

const (
	testVillage = "Testville"
	testMine    = "Test Mine"
)

// jobWorld is an open 20x20 field. Testville (centre 4,4) has a bank at
// (2,2) and a blacksmith at (6,2); players spawn there. Test Mine (centre
// 12,10) has rocks at (10,10) and (14,10).
func jobWorld(t *testing.T) *World {
	t.Helper()
	wm := testMap(20, 20)
	add := func(id uint16, kind resources.ObjectKind, x, y int) {
		wm.Layers[resources.LayerObject].Data[y*wm.Width+x] = resources.ObjectTiles[kind]
		wm.Objects = append(wm.Objects, resources.WorldObject{ID: id, Kind: kind, X: x, Y: y})
	}
	add(1, resources.ObjectBank, 2, 2)
	add(2, resources.ObjectBlacksmith, 6, 2)
	add(3, resources.ObjectIronRock, 10, 10)
	add(4, resources.ObjectIronRock, 14, 10)
	wm.Places = []resources.Place{
		{Name: testVillage, Kind: resources.PlaceVillage, X: 4, Y: 4, Radius: 4},
		{Name: testMine, Kind: resources.PlaceMine, X: 12, Y: 10, Radius: 4},
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
	return p
}

func mineFor(amount int) resources.ActionSpec {
	return resources.ActionSpec{Kind: resources.ActionMine, Place: testMine, Goal: resources.GoalAmount, Amount: amount}
}

func smeltFor(amount int) resources.ActionSpec {
	return resources.ActionSpec{Kind: resources.ActionSmelt, Place: testVillage, Goal: resources.GoalAmount, Amount: amount}
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

func ore(inv resources.Inventory) int  { return inv.Count(resources.ItemIronOre) }
func bars(inv resources.Inventory) int { return inv.Count(resources.ItemIronIngot) }

func TestMineUntilBankHasAmount(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
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
	if w.rocks[tile{10, 10}].claimedBy != nil {
		t.Error("rock should be released")
	}
}

func TestMineForTime(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionMine, Place: testMine, Goal: resources.GoalTime, Minutes: 1})
	run(w, p)

	runUntil(t, w, ticks(swingTime)*40, "time up", func() bool { return p.job != nil && p.job.finishing })
	if p.job.state != stateWalkToBank {
		t.Fatal("time up should send the miner to bank what's carried")
	}
	runUntil(t, w, 5000, "finished", func() bool { return len(p.actions) == 0 })
	// One minute of 2-second swings is at most 30 ore, less the time spent
	// walking to the bank when the backpack filled up.
	if got := ore(p.bank); got < 24 || got > 30 || !p.backpack.Empty() {
		t.Errorf("bank %d ore, backpack %v", got, p.backpack.Stacks)
	}
}

func TestSmeltUntilBankHasAmount(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Add(resources.ItemIronOre, 30)
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
	p.bank.Add(resources.ItemIronOre, 7)
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
	p.bank.Add(resources.ItemIronOre, 3)
	addAction(w, p, smeltFor(3))
	run(w, p)
	runUntil(t, w, 2000, "smelting", func() bool { return p.job != nil && p.job.state == stateSmelting })
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
	w.rocks[tile{10, 10}].ore = 100
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
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, mineFor(50))
	run(w, p)
	runUntil(t, w, 100, "mining", func() bool { return p.job != nil && p.job.state == stateMining })

	j := p.job
	w.click(p, tile{1, 18})
	w.tick()
	if p.job != j || p.job.state != stateMining || len(p.path) != 0 {
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
		{Kind: resources.ActionMine, Place: testVillage, Goal: resources.GoalAmount, Amount: 5}, // not a mine
		{Kind: resources.ActionSmelt, Place: testMine, Goal: resources.GoalAmount, Amount: 5},   // no blacksmith
		{Kind: resources.ActionMine, Place: testMine, Goal: resources.GoalAmount, Amount: 0},
		{Kind: resources.ActionMine, Place: testMine, Goal: resources.GoalTime, Minutes: 0},
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
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, mineFor(50))
	addAction(w, p, smeltFor(5))
	run(w, p)
	runUntil(t, w, 100, "mining", func() bool { return p.job != nil && p.job.state == stateMining })

	// Moving smelting to the top stops mining and starts smelting.
	w.actionCommand(p, resources.ActionCommandContents{Op: resources.ActionOpUp, Index: 1})
	if p.actions[0].spec.Kind != resources.ActionSmelt || p.job != nil {
		t.Fatal("reordering should cancel the old first action")
	}
	if w.rocks[tile{10, 10}].claimedBy != nil {
		t.Error("cancelled miner should free the rock")
	}
	w.actionCommand(p, resources.ActionCommandContents{Op: resources.ActionOpRemove, Index: 0})
	if len(p.actions) != 1 || p.actions[0].spec.Kind != resources.ActionMine {
		t.Errorf("after removing: %+v", p.actions)
	}
}

func TestStopPausesTimer(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionMine, Place: testMine, Goal: resources.GoalTime, Minutes: 5})
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
	if a.job.rock == nil || b.job.rock == nil || a.job.rock == b.job.rock {
		t.Fatal("two miners should be on different rocks")
	}
}

func TestWaitWhenNoRockIsFree(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{14, 10}].ore = 0 // depleted, and not respawning during the test
	w.rocks[tile{14, 10}].respawnAt = 1 << 30
	a := joinAt(t, w, "a", tile{9, 9})
	b := joinAt(t, w, "b", tile{9, 11})
	addAction(w, a, mineFor(50))
	run(w, a)
	w.tick()
	addAction(w, b, mineFor(50))
	run(w, b)
	w.tick()
	if b.job.state != stateWaitForRock || b.activity != activityWaiting {
		t.Fatalf("b should wait, state %v activity %q", b.job.state, b.activity)
	}
	w.actionCommand(a, resources.ActionCommandContents{Op: resources.ActionOpStop})
	runUntil(t, w, 200, "b claims the freed rock", func() bool { return w.rocks[tile{10, 10}].claimedBy == b })
}

func TestRockDepletesMovesOnAndRespawns(t *testing.T) {
	w := jobWorld(t)
	r1, r2 := w.rocks[tile{10, 10}], w.rocks[tile{14, 10}]
	r1.ore, r2.ore = 1, 50
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.found[testMine] = true
	p.pos[0], p.pos[1] = tile{9, 10}.pixel()
	addAction(w, p, mineFor(50))
	run(w, p)

	runUntil(t, w, 500, "rock depleted", func() bool { return r1.ore == 0 })
	if got := w.wm.Layers[resources.LayerObject].Data[10*w.wm.Width+10]; got != resources.TileIronRockDepleted {
		t.Errorf("depleted rock tile %d", got)
	}
	if p.job.rock != r2 || r2.claimedBy != p {
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
	runUntil(t, w, ticks(rockRespawn)+10, "respawn", func() bool { return r1.ore > 0 })
	if r1.ore < rockOreMin || r1.ore > rockOreMax {
		t.Errorf("respawned with %d ore", r1.ore)
	}
}

func TestBankFullStopsTheList(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	p.bank = resources.NewInventory(1, true)
	p.bank.Add(resources.ItemIronOre, 5) // room for 5 more only
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
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	addAction(w, p, mineFor(100))
	run(w, p)
	runUntil(t, w, 5000, "head to the bank", func() bool { return p.job != nil && p.job.state == stateWalkToBank })
	if got := ore(p.backpack); got != backpackSlots {
		t.Errorf("went to the bank carrying %d ore, want %d", got, backpackSlots)
	}
}

func TestBankAccountsArePrivate(t *testing.T) {
	w := jobWorld(t)
	a := joinAt(t, w, "a", tile{4, 5})
	b := joinAt(t, w, "b", tile{4, 6})
	a.bank.Add(resources.ItemIronOre, 3)
	addAction(w, a, smeltFor(3))
	run(w, a)
	runUntil(t, w, 5000, "a done", func() bool { return len(a.actions) == 0 })
	if bars(b.bank) != 0 || bars(a.bank) != 3 {
		t.Error("smelting leaked into another player's account")
	}
}

func TestMiningContinuesOfflineAtTenthSpeed(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.found[testMine] = true
	p.pos[0], p.pos[1] = tile{9, 10}.pixel()
	addAction(w, p, mineFor(100))
	run(w, p)
	w.tick() // start: walk to the rock
	w.tick() // arrive and start the first swing while online
	w.disconnect(p, s)

	// Ten swings' worth of time offline yields one swing.
	for i := 0; i < ticks(swingTime)*offlineSlowdown; i++ {
		w.tick()
	}
	if got := ore(p.backpack); got != 1 {
		t.Errorf("offline for 10 swings' time: mined %d ore, want 1", got)
	}

	// Back online, full speed again.
	w.login("tom", "pw", testSession())
	for i := 0; i < ticks(swingTime); i++ {
		w.tick()
	}
	if got := ore(p.backpack); got != 2 {
		t.Errorf("after one online swing: %d ore, want 2", got)
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
	p.bank.Add(resources.ItemIronIngot, 4)
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionSmelt, Place: testVillage, Goal: resources.GoalTime, Minutes: 2})
	addAction(w, p, smeltFor(10))
	w.tick()
	inv = lastInventory(t, s)
	if len(inv.Actions) != 2 || inv.Actions[0].SecondsLeft != 120 || inv.Actions[1].Banked != 4 {
		t.Errorf("actions %+v", inv.Actions)
	}
}

func TestActionTimingForProgressBar(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
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
