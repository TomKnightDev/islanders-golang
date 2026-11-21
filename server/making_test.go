package main

import (
	"strings"
	"testing"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

func makeAction(kind resources.ActionKind, item resources.ItemKind, n int) resources.ActionSpec {
	return resources.ActionSpec{Kind: kind, Material: string(item), Place: testVillage, Amount: n}
}

func TestSmithSwordsFromBars(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.xp[resources.SkillSmithing] = resources.XPForLevel(40)
	p.bank.AddStack(resources.Stack{Kind: resources.ItemCopperBar, Count: 7, Quality: 60})
	before := p.xp[resources.SkillSmithing]
	addAction(w, p, makeAction(resources.ActionSmith, resources.ItemCopperSword, 10))
	runUntil(t, w, 3*ticks(time.Minute)*10, "out of bars", func() bool { return len(p.actions) == 0 })

	// 7 bars at 2 a sword: 3 swords, one bar left.
	if got := p.bank.Count(resources.ItemCopperSword); got != 3 || p.bank.Count(resources.ItemCopperBar) != 1 {
		t.Fatalf("swords %d, bars left %d", got, p.bank.Count(resources.ItemCopperBar))
	}
	if p.activity != activityNoBars || p.xp[resources.SkillSmithing] <= before {
		t.Errorf("activity %q, xp %d -> %d", p.activity, before, p.xp[resources.SkillSmithing])
	}
	ids := map[uint32]bool{}
	for _, s := range p.bank.Stacks {
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

func TestStavesNeedLogsAndABar(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.xp[resources.SkillFletching] = resources.XPForLevel(20)
	p.bank.AddStack(resources.Stack{Kind: resources.ItemOakLogs, Count: 6, Quality: 50})
	p.bank.AddStack(resources.Stack{Kind: resources.ItemCopperBar, Count: 1, Quality: 50})
	addAction(w, p, makeAction(resources.ActionFletch, resources.ItemApprenticeStaff, 10))
	runUntil(t, w, 3*ticks(time.Minute)*5, "finished", func() bool { return len(p.actions) == 0 })
	// Only one bar, so only one staff; the other four logs stay banked.
	if p.bank.Count(resources.ItemApprenticeStaff) != 1 || p.bank.Count(resources.ItemOakLogs) != 4 {
		t.Errorf("staves %d, logs left %d", p.bank.Count(resources.ItemApprenticeStaff), p.bank.Count(resources.ItemOakLogs))
	}
}

func TestMakingNeedsTheLevel(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	addAction(w, p, makeAction(resources.ActionSmith, resources.ItemMithrilSword, 10))
	if len(p.actions) != 0 {
		t.Errorf("accepted %d actions above the player's level", len(p.actions))
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
	w.improveCommand(p, resources.ImproveCommandContents{Op: resources.ImproveOpStart, ID: id})
}

func TestImproveAtTheAnvil(t *testing.T) {
	w := jobWorld(t)
	p, sword := improver(t, w, 20, 5, 70)
	before := p.xp[resources.SkillSmithing]
	improve(w, p, sword.ID)
	if p.improving == nil {
		t.Fatalf("improving refused: %q", p.improveNote)
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
	if !strings.Contains(p.improveNote, "better than quality") && !strings.Contains(p.improveNote, "can't get any better") {
		t.Errorf("stop reason %q", p.improveNote)
	}
}

func TestImproveNeedsBetterMaterial(t *testing.T) {
	w := jobWorld(t)
	p, sword := improver(t, w, 50, 5, 40)
	improve(w, p, sword.ID)
	if p.improving != nil || !strings.Contains(p.improveNote, "better than quality 50") {
		t.Errorf("improving %v, note %q", p.improving, p.improveNote)
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
	if p.improving != nil || !strings.Contains(p.improveNote, "walked away") {
		t.Errorf("improving %v, note %q", p.improving, p.improveNote)
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
	if _, ok := p.equipment[resources.SlotHammer]; ok || !strings.Contains(p.improveNote, "broke") {
		t.Errorf("hammer %v, note %q", p.equipment[resources.SlotHammer], p.improveNote)
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
