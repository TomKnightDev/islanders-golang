package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

// The test bank is at (2,2); (3,3) is next to it.

func TestBankOnlyOpenStandingAtIt(t *testing.T) {
	w := jobWorld(t)
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.pos[0], p.pos[1] = tile{12, 12}.pixel()
	p.bank.Add(resources.ItemIronOre, 7)
	w.tick()
	if inv := lastInventory(t, s); inv.At != "" || !inv.Bank.Empty() {
		t.Errorf("away from the bank the client must not see it: %q %+v", inv.At, inv.Bank.Stacks)
	}

	p.pos[0], p.pos[1] = tile{3, 3}.pixel()
	w.tick()
	inv := lastInventory(t, s)
	if inv.At != resources.BuildingBank || inv.AtVillage != testVillage || inv.Bank.Count(resources.ItemIronOre) != 7 {
		t.Errorf("at the bank: %q %q %+v", inv.At, inv.AtVillage, inv.Bank.Stacks)
	}

	// A running action list keeps the bank closed, even standing there.
	addAction(w, p, smeltFor(10))
	run(w, p)
	if _, ok := w.bankOpen(p); ok {
		t.Error("the bank should be closed while the list runs")
	}
}

func TestDepositAndWithdrawKeepQuality(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{3, 3})
	p.backpack.AddStack(resources.Stack{Kind: resources.ItemIronOre, Count: 5, Quality: 44})
	sword := w.newPiece(resources.ItemIronSword, 70)
	p.backpack.AddStack(sword)

	w.bankCommand(p, resources.BankCommandContents{Op: resources.BankOpDepositAll})
	if !p.backpack.Empty() || p.bank.AverageQuality(resources.ItemIronOre) != 44 {
		t.Fatalf("deposit all: backpack %+v", p.backpack.Stacks)
	}
	w.bankCommand(p, resources.BankCommandContents{Op: resources.BankOpWithdraw, Item: resources.ItemIronOre, Count: 2})
	w.bankCommand(p, resources.BankCommandContents{Op: resources.BankOpWithdraw, ID: sword.ID})
	if p.backpack.Count(resources.ItemIronOre) != 2 || p.backpack.AverageQuality(resources.ItemIronOre) != 44 {
		t.Errorf("withdrawn ore %+v", p.backpack.Stacks)
	}
	if got, ok := p.backpack.TakeID(sword.ID); !ok || got.Quality != 70 {
		t.Errorf("withdrawn sword %+v", got)
	}
	w.bankCommand(p, resources.BankCommandContents{Op: resources.BankOpDeposit, Item: resources.ItemIronOre, Count: 1})
	if p.backpack.Count(resources.ItemIronOre) != 1 || p.bank.Count(resources.ItemIronOre) != 4 {
		t.Error("deposit one")
	}
}

func TestNoBankingOrBankEquipFromAfar(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{12, 12})
	p.backpack.Add(resources.ItemIronOre, 3)
	w.bankCommand(p, resources.BankCommandContents{Op: resources.BankOpDepositAll})
	if p.bank.Count(resources.ItemIronOre) != 0 {
		t.Error("deposited without being at a bank")
	}
	p.equipment = map[resources.EquipSlot]resources.Stack{}
	p.bank.AddStack(w.newPiece(resources.ItemCopperSword, 50))
	w.equipCommand(p, resources.EquipCommandContents{Op: resources.EquipOpEquip, Item: resources.ItemCopperSword, FromBank: true})
	if _, ok := p.equipment[resources.SlotWeapon]; ok {
		t.Error("equipped from the bank while away from it")
	}

	p.pos[0], p.pos[1] = tile{3, 3}.pixel()
	w.equipCommand(p, resources.EquipCommandContents{Op: resources.EquipOpEquip, Item: resources.ItemCopperSword, FromBank: true})
	if p.equipment[resources.SlotWeapon].Kind != resources.ItemCopperSword {
		t.Error("equipping from the bank at the bank should work")
	}
}

func TestPurchasesGoToTheBackpack(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", atStore)
	w.shopCommand(p, resources.ShopCommandContents{Op: resources.ShopOpBuy, Village: testVillage, Item: resources.ItemCabbage, Count: 2})
	if p.backpack.Count(resources.ItemCabbage) != 2 {
		t.Errorf("backpack %+v", p.backpack.Stacks)
	}
}
