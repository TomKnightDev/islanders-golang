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
	if !inv.Full() {
		t.Error("should be full")
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

func TestOnlyEqualQualityStacks(t *testing.T) {
	inv := NewInventory(5)
	inv.AddStack(Stack{Kind: ItemIronOre, Count: 3, Quality: 20})
	inv.AddStack(Stack{Kind: ItemIronOre, Count: 2, Quality: 20})
	inv.AddStack(Stack{Kind: ItemIronOre, Count: 4, Quality: 35})
	if len(inv.Stacks) != 2 || inv.Stacks[0].Count != 5 || inv.Stacks[1].Quality != 35 {
		t.Errorf("stacks %+v, want 5 at q20 and 4 at q35", inv.Stacks)
	}
}

func TestCombineOnlyThePickedStacks(t *testing.T) {
	inv := NewInventory(10)
	inv.AddStack(Stack{Kind: ItemIronOre, Count: 6, Quality: 8})
	inv.AddStack(Stack{Kind: ItemIronOre, Count: 6, Quality: 4})
	inv.AddStack(Stack{Kind: ItemIronOre, Count: 3, Quality: 50})
	inv.AddStack(Stack{Kind: ItemCopperOre, Count: 1, Quality: 6})
	// Combine the ore under 10: the q8 and q4 stacks.
	if !inv.Combine(ItemIronOre, []int{8, 4}) {
		t.Fatal("combine did nothing")
	}
	// 12 ore at q6 (a full 10 and a 2); the q50 iron and the copper untouched.
	n6 := 0
	for _, s := range inv.Stacks {
		switch {
		case s.Kind == ItemIronOre && s.Quality == 6:
			n6 += s.Count
		case s.Kind == ItemIronOre && s.Quality != 50:
			t.Errorf("unexpected stack %+v", s)
		}
	}
	if n6 != 12 || inv.Count(ItemIronOre) != 15 || inv.Count(ItemCopperOre) != 1 || len(inv.Stacks) != 4 {
		t.Errorf("after combining: %+v", inv.Stacks)
	}
	if inv.Combine(ItemIronOre, []int{50}) || inv.Combine(ItemIronOre, []int{99, 98}) {
		t.Error("one stack, or none, has nothing to combine")
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
	pack := NewInventory(30)
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

func TestTakeQualityAndSlot(t *testing.T) {
	inv := NewInventory(5)
	inv.AddStack(Stack{Kind: ItemIronOre, Count: 5, Quality: 20})
	inv.AddStack(Stack{Kind: ItemIronOre, Count: 5, Quality: 40})
	got := inv.TakeQuality(ItemIronOre, 20, 3)
	if len(got) != 1 || got[0].Count != 3 || got[0].Quality != 20 || inv.Count(ItemIronOre) != 7 {
		t.Errorf("took %+v, left %+v", got, inv.Stacks)
	}
	part, ok := inv.TakeSlot(1, 99)
	if !ok || part.Count != 5 || part.Quality != 40 || len(inv.Stacks) != 1 {
		t.Errorf("took slot %+v, left %+v", part, inv.Stacks)
	}
}

func TestArrowsStack(t *testing.T) {
	pack := NewInventory(20)
	pack.Add(ItemOakLogs, 3)
	if got := pack.Add(ItemOakArrows, 120); got != 120 || len(pack.Stacks) != 4 {
		t.Errorf("3 logs + 120 arrows should take 4 slots (logs, then 50+50+20 arrows), got %d stacks", len(pack.Stacks))
	}
}
