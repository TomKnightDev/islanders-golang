package main

import (
	"strings"
	"testing"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// atFletcher is next to the test village's fletcher's hut, at (8,2).
var atFletcher = tile{8, 3}

func makeGear(w *World, p *player, item resources.ItemKind, n int) {
	w.stationCommand(p, resources.StationCommandContents{Op: resources.StationOpMake, Item: item, Count: n})
}

// runStation ticks until the player stops working at the station.
func runStation(t *testing.T, w *World, p *player, steps int) {
	t.Helper()
	runUntil(t, w, steps*4*ticks(smithTime), "done at the station", func() bool { return p.making == nil && p.improving == nil })
}

func TestMakeSwordsAtTheAnvil(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", atAnvil)
	p.xp[resources.SkillSmithing] = resources.XPForLevel(40)
	p.backpack.AddStack(resources.Stack{Kind: resources.ItemCopperBar, Count: 7, Quality: 60})
	before := p.xp[resources.SkillSmithing]
	makeGear(w, p, resources.ItemCopperSword, 10)
	if p.making == nil {
		t.Fatalf("making refused: %q", p.stationNote)
	}
	runStation(t, w, p, 10)

	// 7 bars at 2 a sword: 3 swords, one bar left.
	if got := p.backpack.Count(resources.ItemCopperSword); got != 3 || p.backpack.Count(resources.ItemCopperBar) != 1 {
		t.Fatalf("swords %d, bars left %d", got, p.backpack.Count(resources.ItemCopperBar))
	}
	if !strings.Contains(p.stationNote, "Out of materials") || p.xp[resources.SkillSmithing] <= before {
		t.Errorf("note %q, xp %d -> %d", p.stationNote, before, p.xp[resources.SkillSmithing])
	}
	ids := map[uint32]bool{}
	for _, s := range p.backpack.Stacks {
		if s.Kind != resources.ItemCopperSword {
			continue
		}
		// Level 40 with a quality-10 copper hammer: about 20 + 1, never above the bars' 60.
		if s.Quality < 15 || s.Quality > 25 || s.ID == 0 || ids[s.ID] {
			t.Errorf("made sword %+v", s)
		}
		ids[s.ID] = true
	}
}

func TestMakeAStaffAtTheFletcher(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", atFletcher)
	p.xp[resources.SkillFletching] = resources.XPForLevel(20)
	p.backpack.AddStack(resources.Stack{Kind: resources.ItemOakLogs, Count: 6, Quality: 50})
	p.backpack.AddStack(resources.Stack{Kind: resources.ItemCopperBar, Count: 1, Quality: 50})
	makeGear(w, p, resources.ItemApprenticeStaff, 5)
	runStation(t, w, p, 5)
	// Only one bar, so only one staff; the other four logs are left.
	if p.backpack.Count(resources.ItemApprenticeStaff) != 1 || p.backpack.Count(resources.ItemOakLogs) != 4 {
		t.Errorf("staves %d, logs left %d", p.backpack.Count(resources.ItemApprenticeStaff), p.backpack.Count(resources.ItemOakLogs))
	}
	makeGear(w, p, resources.ItemCopperSword, 1)
	if p.making != nil {
		t.Error("swords are made at the anvil, not the fletcher")
	}
}

func TestMakingNeedsTheLevelAndMaterials(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", atAnvil)
	p.backpack.AddStack(resources.Stack{Kind: resources.ItemMithrilBar, Count: 2, Quality: 50})
	makeGear(w, p, resources.ItemMithrilSword, 1)
	if p.making != nil || !strings.Contains(p.stationNote, "level") {
		t.Errorf("made a mithril sword at level 1: note %q", p.stationNote)
	}
	makeGear(w, p, resources.ToolItem("Copper", resources.SlotPickaxe), 1)
	if p.making != nil || !strings.Contains(p.stationNote, "copper bar") {
		t.Errorf("made a pickaxe with no copper bars: note %q", p.stationNote)
	}
	p.pos[0], p.pos[1] = tile{12, 12}.pixel()
	p.backpack.AddStack(resources.Stack{Kind: resources.ItemCopperBar, Count: 1, Quality: 50})
	makeGear(w, p, resources.ToolItem("Copper", resources.SlotPickaxe), 1)
	if p.making != nil {
		t.Error("made a pickaxe away from the anvil")
	}
}

func TestGearIsntAListAction(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	addAction(w, p, resources.ActionSpec{Kind: "smith", Material: string(resources.ItemCopperSword), Place: testVillage, Amount: 10})
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionFletch, Material: string(resources.ItemOakShortbow), Place: testVillage, Amount: 10})
	if len(p.actions) != 0 {
		t.Errorf("gear made from the action list: %d actions", len(p.actions))
	}
}

// atAnvil is next to the test village's anvil, at (3,8).
var atAnvil = tile{4, 9}

// improver is a smith at the anvil with a copper sword of quality q and
// some copper bars of quality barQ in their backpack.
func improver(t *testing.T, w *World, q, bars, barQ int) (*player, resources.Stack) {
	t.Helper()
	p := joinAt(t, w, "tom", atAnvil)
	p.xp[resources.SkillSmithing] = resources.XPForLevel(60)
	sword := w.newPiece(resources.ItemCopperSword, q)
	p.backpack.AddStack(sword)
	p.backpack.AddStack(resources.Stack{Kind: resources.ItemCopperBar, Count: bars, Quality: barQ})
	if v := w.visiting(p); v.at != resources.BuildingAnvil {
		t.Fatalf("should be at the anvil, at %+v", v)
	}
	return p, sword
}

func improve(w *World, p *player, id uint32) {
	w.stationCommand(p, resources.StationCommandContents{Op: resources.StationOpImprove, ID: id})
}

func TestImproveAtTheAnvil(t *testing.T) {
	w := jobWorld(t)
	p, sword := improver(t, w, 20, 5, 70)
	before := p.xp[resources.SkillSmithing]
	improve(w, p, sword.ID)
	if p.improving == nil {
		t.Fatalf("improving refused: %q", p.stationNote)
	}
	// 5 bars last 4 attempts each.
	runUntil(t, w, 30*resources.ImproveUnitAttempts*5*ticks(improveTime), "stopped", func() bool { return p.improving == nil })

	got, _ := p.backpack.TakeID(sword.ID)
	if got.Quality <= 40 || got.Quality > 70 {
		t.Errorf("quality-70 bars at level 60 took the sword from 20 to %d", got.Quality)
	}
	if p.backpack.Count(resources.ItemCopperBar) == 5 || p.xp[resources.SkillSmithing] <= before {
		t.Errorf("bars left %d, xp %d -> %d", p.backpack.Count(resources.ItemCopperBar), before, p.xp[resources.SkillSmithing])
	}
	if p.equipment[resources.SlotHammer].Damage == 0 {
		t.Error("improving should wear the hammer")
	}
	if !strings.Contains(p.stationNote, "better than quality") && !strings.Contains(p.stationNote, "can't get any better") {
		t.Errorf("stop reason %q", p.stationNote)
	}
}

func TestImproveNeedsBetterMaterial(t *testing.T) {
	w := jobWorld(t)
	p, sword := improver(t, w, 50, 5, 40)
	improve(w, p, sword.ID)
	if p.improving != nil || !strings.Contains(p.stationNote, "better than quality 50") {
		t.Errorf("improving %v, note %q", p.improving, p.stationNote)
	}
	if p.backpack.Count(resources.ItemCopperBar) != 5 {
		t.Error("no bars should be used")
	}
}

func TestImproveStopsWhenWalkingAway(t *testing.T) {
	w := jobWorld(t)
	p, sword := improver(t, w, 20, 5, 70)
	improve(w, p, sword.ID)
	w.click(p, tile{12, 12})
	w.tick()
	if p.improving != nil || !strings.Contains(p.stationNote, "walked away") {
		t.Errorf("improving %v, note %q", p.improving, p.stationNote)
	}
}

func TestImprovingCanBreakTheTool(t *testing.T) {
	w := jobWorld(t)
	p, sword := improver(t, w, 20, 5, 70)
	hammer := p.equipment[resources.SlotHammer]
	hammer.Damage = 100 - resources.ImproveToolWear
	p.equipment[resources.SlotHammer] = hammer
	improve(w, p, sword.ID)
	runUntil(t, w, 2*ticks(improveTime), "stopped", func() bool { return p.improving == nil })
	if _, ok := p.equipment[resources.SlotHammer]; ok || !strings.Contains(p.stationNote, "broke") {
		t.Errorf("hammer %v, note %q", p.equipment[resources.SlotHammer], p.stationNote)
	}
	improve(w, p, sword.ID)
	if p.improving != nil {
		t.Error("can't improve without a hammer")
	}
}

func TestImproveOnlyAtTheRightBuilding(t *testing.T) {
	w := jobWorld(t)
	p, sword := improver(t, w, 20, 5, 70)
	p.pos[0], p.pos[1] = atStore.pixel()
	improve(w, p, sword.ID)
	if p.improving != nil {
		t.Error("improved a sword at the store")
	}
	bow := w.newPiece(resources.ItemOakShortbow, 20)
	p.backpack.AddStack(bow)
	p.pos[0], p.pos[1] = atAnvil.pixel()
	improve(w, p, bow.ID)
	if p.improving != nil {
		t.Error("bows are improved at the fletcher, not the anvil")
	}
}

func TestVisitTheAnvilAndFletcher(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{12, 12})
	addAction(w, p, visitAction(resources.BuildingAnvil))
	addAction(w, p, visitAction(resources.BuildingFletcher))
	if len(p.actions) != 2 {
		t.Fatalf("visits refused: %d", len(p.actions))
	}
	runUntil(t, w, 30*ticks(time.Second), "arrived", func() bool { return len(p.actions) == 1 })
	w.tick()
	if v := w.visiting(p); v.at != resources.BuildingAnvil {
		t.Errorf("at %+v, want the anvil", v)
	}
}
