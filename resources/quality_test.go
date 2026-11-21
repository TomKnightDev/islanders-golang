package resources

import (
	"math/rand"
	"testing"
)

func TestStacksKeepTheirQuality(t *testing.T) {
	bank := NewInventory(10)
	bank.AddStack(Stack{Kind: ItemIronOre, Count: 4, Quality: 20})
	bank.AddStack(Stack{Kind: ItemIronOre, Count: 4, Quality: 40})
	if len(bank.Stacks) != 2 || bank.AverageQuality(ItemIronOre) != 30 {
		t.Errorf("stacks %+v, want q20 and q40 kept apart", bank.Stacks)
	}
}

func TestQualityMovesWithItems(t *testing.T) {
	bank := NewInventory(10)
	bank.AddStack(Stack{Kind: ItemIronOre, Count: 5, Quality: 60})
	bank.AddStack(Stack{Kind: ItemIronSword, Count: 1, Quality: 33, ID: 7})
	pack := NewInventory(20)
	pack.AddAll(bank.Take(ItemIronOre, 3))
	if pack.Count(ItemIronOre) != 3 || pack.AverageQuality(ItemIronOre) != 60 {
		t.Errorf("taken ore %+v", pack.Stacks)
	}
	pack.TransferAll(&bank)
	if bank.Count(ItemIronOre) != 5 || bank.AverageQuality(ItemIronOre) != 60 {
		t.Errorf("ore back in the bank %+v", bank.Stacks)
	}
	sword, ok := bank.TakeID(7)
	if !ok || sword.Quality != 33 || bank.Count(ItemIronSword) != 0 {
		t.Errorf("sword by ID %+v %v", sword, ok)
	}
}

func TestGearAndToolsScaleWithQuality(t *testing.T) {
	low := EquipmentBonuses(map[EquipSlot]Stack{SlotWeapon: {Kind: ItemMithrilSword, Count: 1, Quality: 1}})
	high := EquipmentBonuses(map[EquipSlot]Stack{SlotWeapon: {Kind: ItemMithrilSword, Count: 1, Quality: 100}})
	if float64(high.Strength) < float64(low.Strength)*1.7 {
		t.Errorf("quality 100 should be far stronger than 1: %d vs %d", high.Strength, low.Strength)
	}
	damaged := EquipmentBonuses(map[EquipSlot]Stack{SlotWeapon: {Kind: ItemMithrilSword, Count: 1, Quality: 100, Damage: 100}})
	if damaged.Strength != low.Strength && damaged.Strength != int(float64(Equipment[ItemMithrilSword].Strength)*0.7+0.5) {
		t.Errorf("a fully damaged sword works like quality 0: %d", damaged.Strength)
	}

	copper := Stack{Kind: ToolItem("Copper", SlotPickaxe), Count: 1, Quality: 10}
	mithril := Stack{Kind: ToolItem("Mithril", SlotPickaxe), Count: 1, Quality: 100}
	if ToolSpeed(nil) != 2 || ToolSpeed(&copper) >= 1 || ToolSpeed(&mithril) > 0.71 {
		t.Errorf("tool speeds: none %v, copper q10 %v, mithril q100 %v", ToolSpeed(nil), ToolSpeed(&copper), ToolSpeed(&mithril))
	}
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 50; i++ {
		if q := GatheredQuality(99, nil, r); q > 10 {
			t.Fatal("without a tool gathered quality must stay at 10 or below")
		}
	}
	if GatheredQuality(60, &mithril, r) < GatheredQuality(1, &copper, r)+20 {
		t.Error("skill and a good tool should gather much better materials")
	}
	if QualityValue(ItemIronSword, 100) != 5*QualityValue(ItemIronSword, 0) || QualityValue(ItemIronOre, 90) != Values[ItemIronOre] {
		t.Error("pieces scale in value with quality; materials don't")
	}
}

func TestToolsExist(t *testing.T) {
	for _, slot := range ToolSlots {
		k := ToolItem("Copper", slot)
		d, ok := Equipment[k]
		if !ok || d.Slot != slot || Values[k] == 0 || Items[k].Name == "" {
			t.Errorf("copper tool for %s: %q %+v", slot, k, d)
		}
	}
	if ToolItem("Mithril", SlotKnife) != "mithrilKnife" {
		t.Errorf("tool naming: %q", ToolItem("Mithril", SlotKnife))
	}
}

func TestRepairLoss(t *testing.T) {
	if RepairLoss(0, 1) != 0 || RepairLoss(1, 99) != 1 {
		t.Error("nothing to repair costs nothing; any repair costs at least 1")
	}
	if RepairLoss(60, 1) != 3 || RepairLoss(60, 99) != 1 {
		t.Errorf("60%% wear: %d at level 1, %d at 99", RepairLoss(60, 1), RepairLoss(60, 99))
	}
}
