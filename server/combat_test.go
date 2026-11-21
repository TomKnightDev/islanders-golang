package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

const testCamp = "Test Camp"

// campWorld is jobWorld plus a goblin camp centred on (15,4), with goblins
// living at (14,3) and (16,5).
func campWorld(t *testing.T) *World {
	t.Helper()
	w := jobWorld(t)
	for i, s := range []tile{{14, 3}, {16, 5}} {
		w.wm.Objects = append(w.wm.Objects, resources.WorldObject{ID: uint16(100 + i), Kind: resources.ObjectSpawn, X: s.x, Y: s.y, Variant: "goblin"})
	}
	w.wm.Places = append(w.wm.Places, resources.Place{Name: testCamp, Kind: resources.PlaceCamp, X: 15, Y: 4, Radius: 3, Materials: []string{"goblin"}})
	w.initMobs()
	return w
}

// fighter is a player who knows the camp, with combat levels set.
func fighter(t *testing.T, w *World, name string, level int) *player {
	t.Helper()
	p := joinAt(t, w, name, tile{12, 4})
	p.found[testCamp] = true
	for _, s := range []resources.Skill{resources.SkillAttack, resources.SkillStrength, resources.SkillDefence, resources.SkillRanged} {
		p.xp[s] = resources.XPForLevel(level)
	}
	return p
}

func fightGoblins(kills int) resources.ActionSpec {
	return resources.ActionSpec{Kind: resources.ActionFight, Material: "goblin", Place: testCamp, Goal: resources.GoalAmount, Amount: kills}
}

func TestStartingKit(t *testing.T) {
	w := jobWorld(t)
	p, _ := w.login("new", "pw", testSession())
	if p.equipment[resources.SlotWeapon].Kind != resources.ItemTrainingSword {
		t.Errorf("weapon %v", p.equipment)
	}
	if p.bank.Count(resources.ItemCabbage) != startingFood {
		t.Errorf("%d cabbages", p.bank.Count(resources.ItemCabbage))
	}
	if p.level(resources.SkillHitpoints) != resources.StartingHitpoints || p.health != 100 || p.maxHealth() != 100 {
		t.Errorf("hitpoints %d, health %d/%d", p.level(resources.SkillHitpoints), p.health, p.maxHealth())
	}
}

func TestFightUntilKills(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 40)
	addAction(w, p, fightGoblins(3))
	if len(p.actions) != 1 {
		t.Fatal("fight action refused")
	}
	run(w, p)
	runUntil(t, w, 40000, "3 kills", func() bool { return len(p.actions) == 0 })
	if p.running || p.activity != activityIdle {
		t.Errorf("should finish idle: running %v, %q", p.running, p.activity)
	}
	if p.xp[resources.SkillAttack] == 0 || p.xp[resources.SkillStrength] == 0 || p.xp[resources.SkillHitpoints] <= resources.XPForLevel(10) {
		t.Error("melee should train Attack, Strength and Hitpoints")
	}
	dead := 0
	for _, m := range w.mobs {
		if m.dead {
			dead++
		}
		if m.fighting != nil {
			t.Error("no enemy should still be held")
		}
	}
	if dead == 0 {
		t.Fatal("expected a goblin to be dead")
	}
	runUntil(t, w, ticks(enemyRespawn)+10, "goblins respawn", func() bool {
		for _, m := range w.mobs {
			if m.dead {
				return false
			}
		}
		return true
	})
}

func TestEatsWhenHurt(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 1)
	addAction(w, p, fightGoblins(50))
	run(w, p)
	runUntil(t, w, 5000, "fighting", func() bool { return p.job != nil && p.job.state == stateFighting })
	food := p.backpack.Count(resources.ItemCabbage)
	if food != startingFood {
		t.Fatalf("took %d cabbages, want all %d", food, startingFood)
	}
	p.health = 40
	w.tick()
	if p.health < 70 || p.backpack.Count(resources.ItemCabbage) != food-1 {
		t.Errorf("should eat below half health: health %d, %d cabbages", p.health, p.backpack.Count(resources.ItemCabbage))
	}
}

func TestRetreatsAndRestsWithoutFood(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 1)
	p.bank.Remove(resources.ItemCabbage, startingFood)
	addAction(w, p, fightGoblins(50))
	run(w, p)
	runUntil(t, w, 5000, "fighting", func() bool { return p.job != nil && p.job.state == stateFighting })
	held := p.job.mob
	p.health = 20
	w.tick()
	if p.job.state != stateWalkToBank || held.fighting != nil {
		t.Fatalf("should retreat to the bank and let the goblin go, state %v", p.job.state)
	}
	runUntil(t, w, 5000, "resting", func() bool { return p.job.state == stateResting })
	runUntil(t, w, ticks(restRegenTime)*100, "rested", func() bool { return p.job.state == stateWalkToCamp })
	if p.health < 90 {
		t.Errorf("left the bank at %d health", p.health)
	}
}

func TestDefeatLosesBackpackKeepsEquipment(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 1)
	addAction(w, p, fightGoblins(50))
	run(w, p)
	runUntil(t, w, 5000, "fighting", func() bool { return p.job != nil && p.job.state == stateFighting })
	p.backpack.Add(resources.ItemCopperOre, 3)
	m := p.job.mob
	for i := 0; i < 1000 && p.running; i++ {
		p.health = 1
		w.enemyAttack(m, p)
	}
	if p.running || p.activity != activityDefeated {
		t.Fatalf("should be defeated: running %v, %q", p.running, p.activity)
	}
	if !p.backpack.Empty() || p.health != p.maxHealth() {
		t.Errorf("backpack %v, health %d", p.backpack.Stacks, p.health)
	}
	if sx, sy := (tile{w.wm.SpawnX, w.wm.SpawnY}).pixel(); p.pos[0] != sx || p.pos[1] != sy {
		t.Errorf("woke up at %v, want the spawn", p.pos)
	}
	if p.equipment[resources.SlotWeapon].Kind != resources.ItemTrainingSword {
		t.Error("equipment should be kept")
	}
	if m.fighting != nil || m.health != m.kind.Health {
		t.Error("the goblin should be released and heal")
	}
}

func TestEquipping(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	equip := func(k resources.ItemKind) {
		w.equipCommand(p, resources.EquipCommandContents{Op: resources.EquipOpEquip, Item: k})
	}
	p.bank.Add(resources.ItemCopperSword, 1)
	p.bank.Add(resources.ItemIronSword, 1)
	p.bank.Add(resources.ItemCopperShield, 1)
	p.bank.Add(resources.ItemOakShortbow, 1)
	p.bank.Add(resources.ItemOakArrows, 120)

	equip(resources.ItemCopperSword)
	if p.equipment[resources.SlotWeapon].Kind != resources.ItemCopperSword || p.bank.Count(resources.ItemTrainingSword) != 1 {
		t.Fatal("swapping swords should put the old one in the bank")
	}
	equip(resources.ItemIronSword)
	if p.equipment[resources.SlotWeapon].Kind != resources.ItemCopperSword {
		t.Error("iron sword needs Attack 15")
	}
	equip(resources.ItemCopperShield)
	equip(resources.ItemOakShortbow)
	if p.equipment[resources.SlotWeapon].Kind != resources.ItemOakShortbow {
		t.Fatal("bow not equipped")
	}
	if _, ok := p.equipment[resources.SlotShield]; ok || p.bank.Count(resources.ItemCopperShield) != 1 {
		t.Error("a two-handed bow should put the shield away")
	}
	equip(resources.ItemOakArrows)
	if p.equipment[resources.SlotAmmo].Count != 120 || p.bank.Count(resources.ItemOakArrows) != 0 {
		t.Errorf("all arrows should go to the ammo slot: %+v", p.equipment[resources.SlotAmmo])
	}
	w.equipCommand(p, resources.EquipCommandContents{Op: resources.EquipOpUnequip, Slot: resources.SlotAmmo})
	if p.bank.Count(resources.ItemOakArrows) != 120 {
		t.Error("unequipping should return the arrows")
	}
}

func TestRangedUsesArrows(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 40)
	p.equipment[resources.SlotWeapon] = resources.Stack{Kind: resources.ItemOakShortbow, Count: 1}
	p.equipment[resources.SlotAmmo] = resources.Stack{Kind: resources.ItemOakArrows, Count: 3}
	addAction(w, p, fightGoblins(50))
	run(w, p)
	runUntil(t, w, 20000, "out of arrows", func() bool { return !p.running })
	if p.activity != activityNoArrows {
		t.Errorf("activity %q, want %q", p.activity, activityNoArrows)
	}
	if _, ok := p.equipment[resources.SlotAmmo]; ok {
		t.Error("ammo slot should be empty")
	}
	if p.xp[resources.SkillRanged] == resources.XPForLevel(40) && p.actions[0].kills == 0 {
		t.Log("three arrows all missed; possible but rare")
	}
}

func TestOneFighterPerEnemy(t *testing.T) {
	w := campWorld(t)
	a := fighter(t, w, "a", 1)
	b := fighter(t, w, "b", 1)
	for _, p := range []*player{a, b} {
		addAction(w, p, fightGoblins(50))
		run(w, p)
	}
	runUntil(t, w, 5000, "both engaged", func() bool {
		return a.job != nil && b.job != nil && a.job.mob != nil && b.job.mob != nil
	})
	if a.job.mob == b.job.mob {
		t.Error("two players are fighting the same goblin")
	}
}

func TestCampsOnTheIsland(t *testing.T) {
	w := NewWorld(testIsland)
	spawn := tile{w.wm.SpawnX, w.wm.SpawnY}
	reach := reachable(&w.wm, spawn)
	var camps []resources.Place
	for _, pl := range w.wm.Places {
		if pl.Kind == resources.PlaceCamp {
			camps = append(camps, pl)
		}
	}
	if len(camps) != campCount {
		t.Fatalf("%d camps, want %d", len(camps), campCount)
	}
	nearest := camps[0]
	for _, c := range camps {
		if octile(tile{c.X, c.Y}, spawn) < octile(tile{nearest.X, nearest.Y}, spawn) {
			nearest = c
		}
	}
	if nearest.Materials[0] != "goblin" {
		t.Errorf("the nearest camp, %s, has %v; it should have the easiest enemies", nearest.Name, nearest.Materials)
	}
	perCamp := map[string]int{}
	for _, m := range w.mobs {
		perCamp[m.camp]++
		if !reach[m.home] {
			t.Errorf("%s at %v can't be reached", m.kind.Name, m.home)
		}
	}
	for _, c := range camps {
		if perCamp[c.Name] != len(campSpawns) {
			t.Errorf("%s has %d enemies", c.Name, perCamp[c.Name])
		}
	}
}
