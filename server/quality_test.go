package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

func TestStartingToolBelt(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	ids := map[uint32]bool{}
	for _, slot := range resources.ToolSlots {
		tool, ok := p.equipment[slot]
		if !ok || tool.Kind != resources.ToolItem("Copper", slot) || tool.Quality != resources.ToolStartQuality {
			t.Errorf("%s: %+v", slot, tool)
		}
		if tool.ID == 0 || ids[tool.ID] {
			t.Errorf("%s needs its own ID, got %d", slot, tool.ID)
		}
		ids[tool.ID] = true
	}
	if w := p.equipment[resources.SlotWeapon]; w.Quality != resources.StarterQuality || w.ID == 0 {
		t.Errorf("starting weapon %+v", w)
	}
}

// mineWith mines n copper ore with the given pickaxe (nil for none) at a
// Mining level, and returns the average quality banked and the swing time.
func mineWith(t *testing.T, pickaxe *resources.Stack, level, n int) (quality, swing int) {
	t.Helper()
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 1000
	p := joinAt(t, w, "tom", tile{9, 10})
	p.xp[resources.SkillMining] = resources.XPForLevel(level)
	if pickaxe == nil {
		delete(p.equipment, resources.SlotPickaxe)
	} else {
		p.equipment[resources.SlotPickaxe] = *pickaxe
	}
	addAction(w, p, mineFor(n))
	run(w, p)
	runUntil(t, w, 200000, "mined", func() bool { return len(p.actions) == 0 })
	return p.bank.AverageQuality(resources.ItemCopperOre), p.stepTicks(resources.ActionMine, p.skillTicks(swingTime, resources.SkillMining))
}

func TestToolsSetSpeedAndQuality(t *testing.T) {
	copper := resources.Stack{Kind: resources.ToolItem("Copper", resources.SlotPickaxe), Count: 1, Quality: 10}
	mithril := resources.Stack{Kind: resources.ToolItem("Mithril", resources.SlotPickaxe), Count: 1, Quality: 100}

	noneQ, noneSwing := mineWith(t, nil, 60, 10)
	copperQ, copperSwing := mineWith(t, &copper, 60, 10)
	mithrilQ, mithrilSwing := mineWith(t, &mithril, 60, 10)

	if noneQ > 10 {
		t.Errorf("without a pickaxe ore should be quality 10 or less, got %d", noneQ)
	}
	if !(mithrilQ > copperQ && copperQ > noneQ) {
		t.Errorf("better tools should mine better ore: none %d, copper %d, mithril %d", noneQ, copperQ, mithrilQ)
	}
	if !(noneSwing > copperSwing && copperSwing > mithrilSwing) {
		t.Errorf("better tools should be faster: none %d, copper %d, mithril %d ticks", noneSwing, copperSwing, mithrilSwing)
	}
	if noneSwing < copperSwing*17/10 {
		t.Errorf("no tool should be about half speed: %d vs %d ticks", noneSwing, copperSwing)
	}
}

func TestSmeltedBarsAreNoBetterThanTheirOre(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.xp[resources.SkillSmithing] = resources.XPForLevel(99)
	p.bank.AddStack(resources.Stack{Kind: resources.ItemCopperOre, Count: 5, Quality: 40})
	addAction(w, p, smeltFor(5))
	run(w, p)
	runUntil(t, w, 40000, "smelted", func() bool { return len(p.actions) == 0 })
	if q := p.bank.AverageQuality(resources.ItemCopperBar); q < 35 || q > 42 {
		t.Errorf("bars from quality-40 ore by a master smith: quality %d", q)
	}
}

func TestEquipASpecificPiece(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.xp[resources.SkillMelee] = resources.XPForLevel(15)
	rough := w.newPiece(resources.ItemIronSword, 20)
	fine := w.newPiece(resources.ItemIronSword, 80)
	p.backpack.AddStack(rough)
	p.backpack.AddStack(fine)

	w.equipCommand(p, resources.EquipCommandContents{Op: resources.EquipOpEquip, ID: rough.ID})
	if got := p.equipment[resources.SlotWeapon]; got.ID != rough.ID || got.Quality != 20 {
		t.Fatalf("equipped %+v, want the quality-20 sword", got)
	}
	w.equipCommand(p, resources.EquipCommandContents{Op: resources.EquipOpEquip, Item: resources.ItemIronSword})
	if got := p.equipment[resources.SlotWeapon]; got.ID != fine.ID {
		t.Errorf("equipping by kind should pick the best: %+v", got)
	}
	if _, ok := p.backpack.TakeID(rough.ID); !ok {
		t.Error("the swapped-out sword should be back in the backpack with its ID")
	}
}

func TestLootGearIsRoughAndUnique(t *testing.T) {
	w := jobWorld(t)
	ids := map[uint32]bool{}
	inv := resources.NewInventory(50)
	for i := 0; i < 30; i++ {
		w.give(&inv, resources.ItemCopperSword, 1, lootGearQuality+w.rng.Intn(21))
	}
	for _, s := range inv.Stacks {
		if s.Quality < 5 || s.Quality > 25 || s.ID == 0 || ids[s.ID] {
			t.Fatalf("loot sword %+v", s)
		}
		ids[s.ID] = true
	}
}

func TestStoresPayForQuality(t *testing.T) {
	w := jobWorld(t)
	st := w.storeIn(testVillage)
	if st.pays(resources.ItemIronSword, 90) <= st.pays(resources.ItemIronSword, 10)*3 {
		t.Error("a quality-90 sword should fetch several times a quality-10 one")
	}
	if st.pays(resources.ItemIronOre, 90) != st.pays(resources.ItemIronOre, 10) {
		t.Error("materials sell at the same price whatever their quality")
	}
}

func TestQualitySurvivesBankTrips(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 1000
	p := joinAt(t, w, "tom", tile{9, 10})
	p.xp[resources.SkillMining] = resources.XPForLevel(70)
	addAction(w, p, mineFor(25)) // more than a backpack, so it banks twice
	run(w, p)
	runUntil(t, w, 200000, "mined", func() bool { return len(p.actions) == 0 })
	if q := p.bank.AverageQuality(resources.ItemCopperOre); q <= resources.DefaultQuality+20 {
		t.Errorf("a level-70 miner's ore should keep its quality in the bank, got %d", q)
	}
}
