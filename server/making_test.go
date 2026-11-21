package main

import (
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
	addAction(w, p, makeAction(resources.ActionImprove, resources.ItemMithrilSword, 10))
	addAction(w, p, makeAction(resources.ActionImprove, resources.ItemIronOre, 10))
	if len(p.actions) != 0 {
		t.Errorf("accepted %d actions above the player's level or for materials", len(p.actions))
	}
}

func TestImproveRaisesQuality(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.xp[resources.SkillSmithing] = resources.XPForLevel(60)
	sword := w.newPiece(resources.ItemCopperSword, 20)
	p.bank.AddStack(sword)
	p.bank.AddStack(resources.Stack{Kind: resources.ItemCopperBar, Count: 20, Quality: 70})
	before := p.xp[resources.SkillSmithing]
	addAction(w, p, makeAction(resources.ActionImprove, resources.ItemCopperSword, 20))
	runUntil(t, w, 30*ticks(time.Minute), "improved", func() bool { return len(p.actions) == 0 })

	got, ok := p.bank.TakeID(sword.ID)
	if !ok {
		t.Fatalf("the sword should be back in the bank: %+v", p.bank.Stacks)
	}
	if got.Quality <= 40 || got.Quality > 70 {
		t.Errorf("20 attempts with quality-70 bars at level 60 took the sword from 20 to %d", got.Quality)
	}
	if p.bank.Count(resources.ItemCopperBar) != 0 || p.xp[resources.SkillSmithing] <= before {
		t.Errorf("bars left %d, xp %d -> %d", p.bank.Count(resources.ItemCopperBar), before, p.xp[resources.SkillSmithing])
	}
}

func TestImproveNeedsBetterMaterial(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.xp[resources.SkillSmithing] = resources.XPForLevel(10)
	p.bank.AddStack(w.newPiece(resources.ItemCopperSword, 50))
	p.bank.AddStack(resources.Stack{Kind: resources.ItemCopperBar, Count: 5, Quality: 40})
	addAction(w, p, makeAction(resources.ActionImprove, resources.ItemCopperSword, 10))
	if len(p.actions) != 1 {
		t.Fatal("improve refused")
	}
	runUntil(t, w, 5*ticks(time.Minute), "finished", func() bool { return len(p.actions) == 0 })
	if p.activity != activityNoBetter || p.bank.Count(resources.ItemCopperBar) != 5 {
		t.Errorf("activity %q, bars %d", p.activity, p.bank.Count(resources.ItemCopperBar))
	}
}
