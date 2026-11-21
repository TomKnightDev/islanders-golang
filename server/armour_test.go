package main

import (
	"testing"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// atTailor is next to the test village's tailor, at (8,5).
var atTailor = tile{9, 5}

func TestArmourOnlyForItsClass(t *testing.T) {
	w := jobWorld(t)
	p, _ := w.createCharacter("robin", "pw", mustClass(t, resources.ClassArcher)), 0
	if p.equipment[resources.SlotBody].Kind != resources.ItemLeatherBody {
		t.Fatalf("archers start in leather: %+v", p.equipment[resources.SlotBody])
	}
	p.pos[0], p.pos[1] = tile{12, 12}.pixel()
	p.backpack.AddStack(w.newPiece(resources.ItemCopperHelmet, 30))
	p.backpack.AddStack(w.newPiece(resources.ItemLeatherCap, 30))
	w.equipCommand(p, resources.EquipCommandContents{Op: resources.EquipOpEquip, Item: resources.ItemCopperHelmet})
	if _, ok := p.equipment[resources.SlotHead]; ok {
		t.Fatal("an archer wore a metal helmet")
	}
	w.equipCommand(p, resources.EquipCommandContents{Op: resources.EquipOpEquip, Item: resources.ItemLeatherCap})
	if p.equipment[resources.SlotHead].Kind != resources.ItemLeatherCap {
		t.Error("an archer should wear a leather cap")
	}
}

func mustClass(t *testing.T, id resources.ClassID) resources.Class {
	t.Helper()
	c, ok := resources.FindClass(id)
	if !ok {
		t.Fatal("no class", id)
	}
	return c
}

func TestWeaveClothAtTheTailor(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.AddStack(resources.Stack{Kind: resources.ItemCotton, Count: 9, Quality: 30})
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionWeave, Material: "cloth", Place: testVillage, Amount: 10})
	if len(p.actions) != 1 {
		t.Fatal("weave refused")
	}
	runUntil(t, w, 20*ticks(time.Minute), "woven", func() bool { return len(p.actions) == 0 })
	// 9 cotton at 2 a bolt: 4 cloth, 1 cotton left.
	if p.bank.Count(resources.ItemCloth) != 4 || p.bank.Count(resources.ItemCotton) != 1 || p.xp[resources.SkillTailoring] == 0 {
		t.Errorf("cloth %d, cotton %d, tailoring xp %d", p.bank.Count(resources.ItemCloth), p.bank.Count(resources.ItemCotton), p.xp[resources.SkillTailoring])
	}
}

func TestMakeLeatherArmourAtTheTailor(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", atTailor)
	if v := w.visiting(p); v.at != resources.BuildingTailor {
		t.Fatalf("should be at the tailor, at %+v", v)
	}
	p.xp[resources.SkillTailoring] = resources.XPForLevel(10)
	p.backpack.AddStack(resources.Stack{Kind: resources.ItemLeather, Count: 4, Quality: 40})
	makeGear(w, p, resources.ItemLeatherBody, 1)
	if p.making == nil {
		t.Fatalf("making refused: %q", p.stationNote)
	}
	runStation(t, w, p, 4)
	if p.backpack.Count(resources.ItemLeatherBody) != 1 || p.backpack.Count(resources.ItemLeather) != 0 {
		t.Errorf("backpack %+v", p.backpack.Stacks)
	}
}

func TestBattleWearsGear(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 40)
	p.xp[resources.SkillDefence] = 0 // so goblins land hits
	p.equipment[resources.SlotWeapon] = w.newPiece(resources.ItemIronSword, 1)
	p.equipment[resources.SlotBody] = w.newPiece(resources.ItemCopperBody, 1)
	addAction(w, p, fightGoblins(100))
	run(w, p)
	runUntil(t, w, 400000, "worn", func() bool {
		return p.equipment[resources.SlotWeapon].Damage > 0 && p.equipment[resources.SlotBody].Damage > 0
	})
}

func TestWornOutGearBreaks(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{12, 12})
	sword := w.newPiece(resources.ItemIronSword, 30)
	sword.Damage = 99
	p.equipment[resources.SlotWeapon] = sword
	if !w.wearPiece(p, resources.SlotWeapon, 1) {
		t.Fatal("at 100% wear it should break")
	}
	if _, ok := p.equipment[resources.SlotWeapon]; ok {
		t.Error("a broken piece is gone")
	}
}

func TestHuntDeer(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{14, 14})
	deer := 0
	for _, m := range w.mobs {
		if m.kind.ID == resources.Deer.ID && m.camp == testWood {
			deer++
		}
	}
	if deer != deerPerForest {
		t.Fatalf("%d deer in the wood, want %d", deer, deerPerForest)
	}
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionHunt, Material: resources.Deer.ID, Place: testWood, Amount: 3})
	if len(p.actions) != 1 {
		t.Fatal("hunt refused")
	}
	runUntil(t, w, 30*ticks(time.Minute), "hunted", func() bool { return len(p.actions) == 0 })
	if p.bank.Count(resources.ItemLeather) < 3 || p.bank.Count(resources.ItemVenison) != 3 || p.xp[resources.SkillHunting] == 0 {
		t.Errorf("leather %d, venison %d, hunting xp %d", p.bank.Count(resources.ItemLeather), p.bank.Count(resources.ItemVenison), p.xp[resources.SkillHunting])
	}
	if p.health < p.maxHealth() {
		t.Error("deer don't fight back")
	}
}

func TestFightersTakeAnyFood(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 40)
	p.bank.Remove(resources.ItemCabbage, 100)
	p.bank.Add(resources.ItemVenison, 3)
	addAction(w, p, fightGoblins(1))
	run(w, p)
	runUntil(t, w, 5000, "heading out", func() bool { return p.job != nil && p.job.state == stateWalkToCamp })
	if p.backpack.Count(resources.ItemVenison) != 3 {
		t.Errorf("should carry the venison: %+v", p.backpack.Stacks)
	}
}
