package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

// jobWorld is an open field with a bank at (2,2) and rocks at (10,10) and (14,10).
func jobWorld(t *testing.T) *World {
	t.Helper()
	wm := testMap(20, 20)
	add := func(id uint16, kind resources.ObjectKind, x, y int) {
		wm.Layers[resources.LayerObject].Data[y*wm.Width+x] = resources.ObjectTiles[kind]
		wm.Objects = append(wm.Objects, resources.WorldObject{ID: id, Kind: kind, X: x, Y: y})
	}
	add(1, resources.ObjectBank, 2, 2)
	add(2, resources.ObjectIronRock, 10, 10)
	add(3, resources.ObjectIronRock, 14, 10)
	w := NewWorld(wm)
	return w
}

func joinAt(t *testing.T, w *World, name string, at tile) *player {
	t.Helper()
	p, err := w.login(name, "pw", testSession())
	if err != nil {
		t.Fatal(err)
	}
	p.pos[0], p.pos[1] = at.pixel()
	return p
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

func TestMiningLoopBanksAndReturns(t *testing.T) {
	w := jobWorld(t)
	r := w.rocks[tile{10, 10}]
	r.ore = 100
	p := joinAt(t, w, "tom", tile{8, 10})
	p.backpack = resources.NewInventory(10, false) // full at 10 ore, to keep the test short

	w.click(p, tile{10, 10})
	if r.claimedBy != p {
		t.Fatal("clicked rock should be claimed")
	}
	runUntil(t, w, 5000, "first deposit", func() bool { return p.bank.Count(resources.ItemIronOre) == 10 })
	if !p.backpack.Empty() {
		t.Error("backpack should be emptied into the bank")
	}
	runUntil(t, w, 5000, "back to mining", func() bool { return p.job.state == stateMining })
	if p.job.rock != r || r.claimedBy != p {
		t.Error("should return to the same rock")
	}
	if p.activity != activityMining {
		t.Errorf("activity %q", p.activity)
	}
}

func TestSwingTiming(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10}) // already beside the rock
	w.click(p, tile{10, 10})
	w.tick() // arrive, start swinging
	for i := 0; i < ticks(swingTime)-1; i++ {
		w.tick()
	}
	if got := p.backpack.Count(resources.ItemIronOre); got != 0 {
		t.Fatalf("ore before the swing finished: %d", got)
	}
	w.tick()
	if got := p.backpack.Count(resources.ItemIronOre); got != 1 {
		t.Fatalf("after one swing: %d ore", got)
	}
}

func TestOneMinerPerRock(t *testing.T) {
	w := jobWorld(t)
	a := joinAt(t, w, "a", tile{9, 9})
	b := joinAt(t, w, "b", tile{9, 11})
	w.click(a, tile{10, 10})
	w.click(b, tile{10, 10})
	if b.job.rock != w.rocks[tile{14, 10}] {
		t.Fatal("second player should be sent to the other free rock")
	}
	if w.rocks[tile{10, 10}].claimedBy != a || w.rocks[tile{14, 10}].claimedBy != b {
		t.Error("each rock should be claimed by one player")
	}
}

func TestWaitWhenNoRockIsFree(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{14, 10}].ore = 0 // depleted
	a := joinAt(t, w, "a", tile{9, 9})
	b := joinAt(t, w, "b", tile{9, 11})
	w.click(a, tile{10, 10})
	w.click(b, tile{10, 10})
	if b.job.state != stateWaitForRock || b.activity != activityWaiting {
		t.Fatalf("b should wait, state %v activity %q", b.job.state, b.activity)
	}
	// a walks away; b takes the rock.
	w.click(a, tile{2, 10})
	runUntil(t, w, 200, "b claims the freed rock", func() bool { return w.rocks[tile{10, 10}].claimedBy == b })
}

func TestRockDepletesMovesOnAndRespawns(t *testing.T) {
	w := jobWorld(t)
	r1, r2 := w.rocks[tile{10, 10}], w.rocks[tile{14, 10}]
	r1.ore, r2.ore = 1, 50
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.pos[0], p.pos[1] = tile{9, 10}.pixel()

	w.click(p, tile{10, 10})
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
	if got := w.wm.Layers[resources.LayerObject].Data[10*w.wm.Width+10]; got != resources.TileIronRock {
		t.Errorf("respawned rock tile %d", got)
	}
}

func TestBankFullStopsTheLoop(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	p.backpack = resources.NewInventory(10, false)
	p.bank = resources.NewInventory(1, true)
	p.bank.Add(resources.ItemIronOre, 5) // room for 5 more only

	w.click(p, tile{10, 10})
	runUntil(t, w, 5000, "loop stops", func() bool { return p.job == nil })
	if p.activity != activityBankFull {
		t.Errorf("activity %q", p.activity)
	}
	if p.bank.Count(resources.ItemIronOre) != 10 || p.backpack.Count(resources.ItemIronOre) != 5 {
		t.Errorf("bank %d, backpack %d", p.bank.Count(resources.ItemIronOre), p.backpack.Count(resources.ItemIronOre))
	}
	if w.rocks[tile{10, 10}].claimedBy != nil {
		t.Error("rock should be released")
	}
}

func TestClickInterruptsMiningAndReleasesRock(t *testing.T) {
	w := jobWorld(t)
	r := w.rocks[tile{10, 10}]
	r.ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	w.click(p, tile{10, 10})
	runUntil(t, w, 50, "mining", func() bool { return p.job != nil && p.job.state == stateMining })

	w.click(p, tile{5, 5})
	if p.job != nil || r.claimedBy != nil {
		t.Error("walking away should end the job and free the rock")
	}
	if p.activity != activityWalking {
		t.Errorf("activity %q", p.activity)
	}
}

func TestBankTripDepositsOnce(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{8, 8})
	p.backpack.Add(resources.ItemIronOre, 15)
	w.click(p, tile{2, 2})
	runUntil(t, w, 2000, "deposit", func() bool { return p.job == nil })
	if p.bank.Count(resources.ItemIronOre) != 15 || !p.backpack.Empty() {
		t.Errorf("bank %d, backpack %v", p.bank.Count(resources.ItemIronOre), p.backpack.Stacks)
	}
}

func TestBankAccountsArePrivate(t *testing.T) {
	w := jobWorld(t)
	a := joinAt(t, w, "a", tile{3, 3})
	b := joinAt(t, w, "b", tile{3, 4})
	a.backpack.Add(resources.ItemIronOre, 7)
	w.click(a, tile{2, 2})
	runUntil(t, w, 500, "deposit", func() bool { return a.job == nil })
	if b.bank.Count(resources.ItemIronOre) != 0 {
		t.Error("deposit leaked into another player's account")
	}
}

func TestMiningContinuesOfflineAtTenthSpeed(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.pos[0], p.pos[1] = tile{9, 10}.pixel()
	w.click(p, tile{10, 10})
	w.tick() // arrive and start the first swing while online
	w.disconnect(p, s)

	// Ten swings' worth of time offline yields one swing.
	for i := 0; i < ticks(swingTime)*offlineSlowdown; i++ {
		w.tick()
	}
	if got := p.backpack.Count(resources.ItemIronOre); got != 1 {
		t.Errorf("offline for 10 swings' time: mined %d ore, want 1", got)
	}

	// Back online, full speed again.
	s2 := testSession()
	w.login("tom", "pw", s2)
	for i := 0; i < ticks(swingTime); i++ {
		w.tick()
	}
	if got := p.backpack.Count(resources.ItemIronOre); got != 2 {
		t.Errorf("after one online swing: %d ore, want 2", got)
	}
}

func TestLoginSendsInventory(t *testing.T) {
	w := jobWorld(t)
	s := testSession()
	w.login("tom", "pw", s)
	for _, m := range drain(s) {
		if m.MessageType == resources.InventoryMessage {
			return
		}
	}
	t.Error("login should send the inventory")
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

func TestActionTimingForProgressBar(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.pos[0], p.pos[1] = tile{2, 10}.pixel()

	// Walking 7 tiles to stand beside the rock at (9,10) takes 7/walkSpeed seconds.
	w.click(p, tile{10, 10})
	w.tick()
	inv := lastInventory(t, s)
	if inv.Activity != activityToRock || inv.ActionSeconds != 7/walkSpeed {
		t.Fatalf("walk: %q for %vs, want %q for %vs", inv.Activity, inv.ActionSeconds, activityToRock, 7/walkSpeed)
	}

	// Each swing restarts the bar at 0 for swingTime.
	runUntil(t, w, 500, "mining", func() bool { return p.job.state == stateMining })
	w.tick()
	drain(s)
	runUntil(t, w, 500, "next swing", func() bool { return p.backpack.Count(resources.ItemIronOre) == 1 })
	inv = lastInventory(t, s)
	if inv.Activity != activityMining || inv.ActionSeconds != swingTime.Seconds() || inv.ActionElapsed != 0 {
		t.Errorf("swing: %q %vs elapsed %vs", inv.Activity, inv.ActionSeconds, inv.ActionElapsed)
	}

	// Waiting and idle have no set length.
	w.click(p, tile{5, 5})
	w.stopJob(p)
	w.tick()
	if inv = lastInventory(t, s); inv.ActionSeconds != 0 {
		t.Errorf("idle should have no duration, got %v", inv.ActionSeconds)
	}
}

func TestWalkEndsIdle(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{5, 5})
	w.click(p, tile{7, 5})
	if p.activity != activityWalking {
		t.Fatalf("activity %q", p.activity)
	}
	runUntil(t, w, 200, "arrive", func() bool { return len(p.path) == 0 })
	w.tick()
	if p.activity != activityIdle {
		t.Errorf("after arriving: %q", p.activity)
	}
}

func TestFullBackpackIsTwentyOre(t *testing.T) {
	w := jobWorld(t)
	w.rocks[tile{10, 10}].ore = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	w.click(p, tile{10, 10})
	runUntil(t, w, 5000, "head to the bank", func() bool { return p.job.state == stateWalkToBank })
	if got := p.backpack.Count(resources.ItemIronOre); got != backpackSlots {
		t.Errorf("went to the bank carrying %d ore, want %d", got, backpackSlots)
	}
}
