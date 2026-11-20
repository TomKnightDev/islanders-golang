package resources

import "testing"

func TestAddStacksUpToLimit(t *testing.T) {
	inv := NewInventory(3)
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
	inv := NewInventory(1)
	if got := inv.Add(ItemIronIngot, 25); got != 20 {
		t.Errorf("ingots: added %d, want 20", got)
	}
	inv = NewInventory(1)
	if got := inv.Add(ItemIronOre, 25); got != 10 {
		t.Errorf("ore: added %d, want 10", got)
	}
}

func TestFullIsPerItem(t *testing.T) {
	inv := NewInventory(2)
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
	inv := NewInventory(5)
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
	pack := NewInventory(20)
	pack.Add(ItemIronOre, 30)
	bank := NewInventory(2)
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
