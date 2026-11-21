package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

func TestRepairAnywhere(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{12, 12}) // nowhere near a building
	hammer := p.equipment[resources.SlotHammer]
	hammer.Damage, hammer.Quality = 60, 50
	p.equipment[resources.SlotHammer] = hammer
	sword := w.newPiece(resources.ItemIronSword, 40)
	sword.Damage = 20
	p.backpack.AddStack(sword)

	w.repairCommand(p, resources.RepairCommandContents{Slot: resources.SlotHammer})
	got := p.equipment[resources.SlotHammer]
	if got.Damage != 0 || got.Quality != 50-resources.RepairLoss(60, 1) || got.Quality >= 50 {
		t.Errorf("repaired hammer %+v", got)
	}
	if p.xp[resources.SkillRepair] == 0 {
		t.Error("repairing should train Repair")
	}

	w.repairCommand(p, resources.RepairCommandContents{ID: sword.ID})
	if s := findPiece(&p.backpack, sword.ID); s.Damage != 0 || s.Quality != 39 {
		t.Errorf("repaired sword %+v", s)
	}

	xp := p.xp[resources.SkillRepair]
	w.repairCommand(p, resources.RepairCommandContents{ID: sword.ID})
	if p.xp[resources.SkillRepair] != xp || findPiece(&p.backpack, sword.ID).Quality != 39 {
		t.Error("repairing an unworn piece should do nothing")
	}
}

func TestRepairInTheBankOnlyAtABank(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{12, 12})
	shield := w.newPiece(resources.ItemCopperShield, 30)
	shield.Damage = 40
	p.bank.AddStack(shield)
	w.repairCommand(p, resources.RepairCommandContents{ID: shield.ID})
	if findPiece(&p.bank, shield.ID).Damage != 40 {
		t.Error("repaired a piece in the bank from afar")
	}
	p.pos[0], p.pos[1] = tile{3, 3}.pixel()
	w.repairCommand(p, resources.RepairCommandContents{ID: shield.ID})
	if findPiece(&p.bank, shield.ID).Damage != 0 {
		t.Error("should repair a banked piece at the bank")
	}
}
