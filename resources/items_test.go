package resources

import "testing"

func TestAddStacksUpToLimit(t *testing.T) {
	inv := NewInventory(3, true)
	if got := inv.Add(ItemIronOre, 25); got != 25 {
		t.Fatalf("added %d, want 25", got)
	}
	// 10 + 10 + 5
	if len(inv.Stacks) != 3 || inv.Stacks[2].Count != 5 {
		t.Fatalf("stacks %v", inv.Stacks)
	}
	// Tops up the partial stack, then has no free slot.
	if got := inv.Add(ItemIronOre, 10); got != 5 {
		t.Fatalf("added %d, want 5", got)
	}
	if !inv.Full(ItemIronOre) {
		t.Error("should be full of ore")
	}
}

func TestDifferentLimitsPerItem(t *testing.T) {
	inv := NewInventory(1, true)
	if got := inv.Add(ItemIronIngot, 25); got != 20 {
		t.Errorf("ingots: added %d, want 20", got)
	}
	inv = NewInventory(1, true)
	if got := inv.Add(ItemIronOre, 25); got != 10 {
		t.Errorf("ore: added %d, want 10", got)
	}
}

func TestFullIsPerItem(t *testing.T) {
	inv := NewInventory(2, true)
	inv.Add(ItemIronOre, 10)
	inv.Add(ItemIronIngot, 5)
	if !inv.Full(ItemIronOre) {
		t.Error("no room for ore: both slots used and the ore stack is at its limit")
	}
	if inv.Full(ItemIronIngot) {
		t.Error("the ingot stack still has room")
	}
}

func TestRemoveDropsEmptyStacks(t *testing.T) {
	inv := NewInventory(5, true)
	inv.Add(ItemIronOre, 23)
	if got := inv.Remove(ItemIronOre, 13); got != 13 {
		t.Fatalf("removed %d", got)
	}
	if inv.Count(ItemIronOre) != 10 || len(inv.Stacks) != 1 {
		t.Errorf("left %v", inv.Stacks)
	}
	if got := inv.Remove(ItemIronOre, 50); got != 10 || !inv.Empty() {
		t.Errorf("removed %d, stacks %v", got, inv.Stacks)
	}
}

func TestTransferAllIntoNearlyFullBank(t *testing.T) {
	pack := NewInventory(30, false)
	pack.Add(ItemIronOre, 30)
	bank := NewInventory(2, true)
	bank.Add(ItemIronOre, 5)

	if pack.TransferAll(&bank) {
		t.Error("bank can only take 15 more, backpack should not be empty")
	}
	if bank.Count(ItemIronOre) != 20 || pack.Count(ItemIronOre) != 15 {
		t.Errorf("bank %d, pack %d", bank.Count(ItemIronOre), pack.Count(ItemIronOre))
	}
	// Nothing created or lost.
	if bank.Count(ItemIronOre)+pack.Count(ItemIronOre) != 35 {
		t.Error("items were lost or duplicated")
	}
}

func TestBackpackDoesNotStack(t *testing.T) {
	pack := NewInventory(20, false)
	if got := pack.Add(ItemIronOre, 25); got != 20 {
		t.Fatalf("backpack took %d ore, want 20", got)
	}
	if len(pack.Stacks) != 20 || !pack.Full(ItemIronOre) {
		t.Errorf("each ore should take a slot: %d stacks", len(pack.Stacks))
	}
	if pack.Add(ItemIronIngot, 1) != 0 {
		t.Error("a full backpack should not take an ingot either")
	}
	// Banked, the same 20 ore fit in 2 slots.
	bank := NewInventory(100, true)
	if !pack.TransferAll(&bank) || len(bank.Stacks) != 2 {
		t.Errorf("bank should stack 20 ore into 2 slots, got %d", len(bank.Stacks))
	}
}

func TestArrowsStackInTheBackpack(t *testing.T) {
	pack := NewInventory(20, false)
	pack.Add(ItemLogs, 3)
	if got := pack.Add(ItemArrows, 120); got != 120 || len(pack.Stacks) != 6 {
		t.Errorf("3 logs + 120 arrows should take 6 slots (50+50+20), got %d stacks", len(pack.Stacks))
	}
}
