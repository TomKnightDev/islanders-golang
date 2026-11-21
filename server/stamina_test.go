package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

func TestTiredMinerRestsAtInnThenCarriesOn(t *testing.T) {
	w := jobWorld(t)
	rock := w.nodes[tile{10, 10}]
	rock.amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	p.stamina = 3
	addAction(w, p, mineFor(10))
	run(w, p)

	runUntil(t, w, ticks(swingTime)*10, "worn out", func() bool { return p.job != nil && p.job.state == stateWalkToInn })
	if ore(p.backpack) != 3 || p.stamina != 0 {
		t.Errorf("should work until stamina runs out: %d ore, %d stamina", ore(p.backpack), p.stamina)
	}
	if rock.claimedBy != nil {
		t.Error("a tired miner should let go of their rock")
	}
	runUntil(t, w, 5000, "relaxing", func() bool { return p.job != nil && p.job.state == stateRelaxing })
	runUntil(t, w, ticks(innRestTime)*(p.maxStamina()+5), "rested", func() bool { return p.stamina == p.maxStamina() })
	runUntil(t, w, 40000, "back to finish the job", func() bool { return len(p.actions) == 0 })
	if ore(p.bank) != 10 {
		t.Errorf("should finish the goal after resting: bank %d ore", ore(p.bank))
	}
}

func TestInnTimeDoesntCountTowardsTimeGoals(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	p.stamina = 1
	addAction(w, p, resources.ActionSpec{Kind: resources.ActionMine, Material: "copper", Place: testMine, Goal: resources.GoalTime, Minutes: 5})
	run(w, p)
	runUntil(t, w, ticks(swingTime)*3, "to the inn", func() bool { return p.job != nil && p.job.state == stateWalkToInn })
	ran := p.actions[0].ran
	runUntil(t, w, 40000, "rested", func() bool { return p.job == nil || !p.job.atInn() })
	if p.actions[0].ran != ran {
		t.Errorf("the time goal advanced while at the inn: %d -> %d", ran, p.actions[0].ran)
	}
}

func TestTiredFighterLetsGoAndComesBack(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 40)
	addAction(w, p, fightGoblins(50))
	run(w, p)
	runUntil(t, w, 5000, "fighting", func() bool { return p.job != nil && p.job.state == stateFighting })
	m := p.job.mob
	p.stamina = 0
	runUntil(t, w, ticks(playerAttackTime)+5, "heads for the inn", func() bool { return p.job.state == stateWalkToInn })
	if m.fighting != nil || p.job.mob != nil {
		t.Error("a tired fighter should let the enemy go")
	}
	runUntil(t, w, 40000, "back fighting", func() bool { return p.job != nil && p.job.state == stateFighting })
	if p.stamina < p.maxStamina()-staminaPerAttack*2 {
		t.Errorf("should come back rested, stamina %d", p.stamina)
	}
}

func TestStaminaTrains(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	if p.stamina != 100 || p.maxStamina() != 100 {
		t.Fatalf("new characters start with 100 stamina, got %d/%d", p.stamina, p.maxStamina())
	}
	w.spendStamina(p, 83)
	if p.level(resources.SkillStamina) != 2 || p.maxStamina() != 105 {
		t.Errorf("83 stamina spent: level %d, max %d", p.level(resources.SkillStamina), p.maxStamina())
	}
	if resources.MaxStamina(99) != 590 {
		t.Errorf("level 99 max %d", resources.MaxStamina(99))
	}
}
