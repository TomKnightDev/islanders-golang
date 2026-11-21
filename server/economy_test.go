package main

import (
	"testing"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

func sellIron(amount int) resources.ActionSpec {
	return resources.ActionSpec{Kind: resources.ActionSell, Material: string(resources.ItemIronOre), Place: testVillage, Goal: resources.GoalAmount, Amount: amount}
}

func TestStartingCoins(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	if p.coins != resources.StartingCoins {
		t.Errorf("new characters start with %d coins, got %d", resources.StartingCoins, p.coins)
	}
}

func TestSellTripPaysAndPricesFall(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Add(resources.ItemIronOre, 30)
	st := w.storeIn(testVillage)
	fullPrice := st.pays(resources.ItemIronOre)
	addAction(w, p, sellIron(30))
	if len(p.actions) != 1 {
		t.Fatal("sell action refused")
	}
	run(w, p)
	runUntil(t, w, 20000, "sold 30", func() bool { return len(p.actions) == 0 })

	earned := p.coins - resources.StartingCoins
	if p.bank.Count(resources.ItemIronOre) != 0 || !p.backpack.Empty() {
		t.Errorf("all 30 should be sold: bank %d, backpack %v", p.bank.Count(resources.ItemIronOre), p.backpack.Stacks)
	}
	if earned <= 0 || earned > 30*fullPrice {
		t.Errorf("earned %d for 30 iron ore at up to %d each", earned, fullPrice)
	}
	if st.demandFor(resources.ItemIronOre) >= 1 {
		t.Error("selling should lower the store's demand")
	}
}

func TestSellingStopsWhenNothingIsLeft(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Add(resources.ItemIronOre, 5)
	addAction(w, p, sellIron(100))
	run(w, p)
	runUntil(t, w, 20000, "finished", func() bool { return len(p.actions) == 0 })
	if p.activity != activityNoGoods || p.coins <= resources.StartingCoins {
		t.Errorf("activity %q, coins %d", p.activity, p.coins)
	}
}

func TestStorePricesRecover(t *testing.T) {
	w := jobWorld(t)
	st := w.storeIn(testVillage)
	st.demand[resources.ItemIronOre] = 0.5
	for i := 0; i < ticks(time.Minute)*30; i++ {
		w.tick()
	}
	d := st.demandFor(resources.ItemIronOre)
	if d <= 0.5 || d >= 1 {
		t.Errorf("after 30 minutes demand should be partly back: %v", d)
	}
}

func TestShop(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	p.bank.Remove(resources.ItemCabbage, startingFood)
	buy := func(village string, k resources.ItemKind, n int) {
		w.shopCommand(p, resources.ShopCommandContents{Op: resources.ShopOpBuy, Village: village, Item: k, Count: n})
	}
	price := 0
	for _, s := range resources.StoreStock {
		if s.Item == resources.ItemCabbage {
			price = resources.StorePrice(s)
		}
	}

	buy(testVillage, resources.ItemCabbage, 4)
	if p.bank.Count(resources.ItemCabbage) != 4 || p.coins != resources.StartingCoins-4*price {
		t.Errorf("bought 4 cabbages: bank %d, coins %d", p.bank.Count(resources.ItemCabbage), p.coins)
	}
	buy(testVillage, resources.ItemIronSword, 1)
	if p.bank.Count(resources.ItemIronSword) != 0 {
		t.Error("the store doesn't sell iron swords")
	}
	p.coins = price*2 + 1
	buy(testVillage, resources.ItemCabbage, 10)
	if p.bank.Count(resources.ItemCabbage) != 6 || p.coins != 1 {
		t.Errorf("should only buy what's affordable: bank %d, coins %d", p.bank.Count(resources.ItemCabbage), p.coins)
	}
	delete(p.found, testVillage)
	p.coins = 100
	buy(testVillage, resources.ItemCabbage, 1)
	if p.coins != 100 {
		t.Error("can't shop in a village you haven't found")
	}
}

func TestInnFeeAndStable(t *testing.T) {
	w := jobWorld(t)
	w.nodes[tile{10, 10}].amount = 100
	p := joinAt(t, w, "tom", tile{9, 10})
	p.stamina = 1
	addAction(w, p, mineFor(50))
	run(w, p)
	runUntil(t, w, 20000, "relaxing", func() bool { return p.job != nil && p.job.state == stateRelaxing })
	if p.coins != resources.StartingCoins-resources.InnFee || p.job.restTicks != ticks(innRestTime) || p.activity != activityRelaxing {
		t.Errorf("paying guest: coins %d, rest %d ticks, %q", p.coins, p.job.restTicks, p.activity)
	}

	poor := joinAt(t, w, "poor", tile{9, 9})
	poor.coins = 0
	poor.stamina = 1
	w.nodes[tile{14, 10}].amount = 100
	addAction(w, poor, mineFor(50))
	run(w, poor)
	runUntil(t, w, 20000, "in the stable", func() bool { return poor.job != nil && poor.job.state == stateRelaxing })
	if poor.job.restTicks != ticks(innRestTime)*stableSlowdown || poor.activity != activityStable || poor.coins != 0 {
		t.Errorf("stable: rest %d ticks, %q, coins %d", poor.job.restTicks, poor.activity, poor.coins)
	}
}

func TestKillsDropCoins(t *testing.T) {
	w := campWorld(t)
	p := fighter(t, w, "tom", 40)
	addAction(w, p, fightGoblins(3))
	run(w, p)
	runUntil(t, w, 40000, "3 kills", func() bool { return len(p.actions) == 0 })
	goblin, _ := resources.FindEnemy("goblin")
	earned := p.coins - resources.StartingCoins
	if earned < 3*goblin.Coins[0] || earned > 3*goblin.Coins[1] {
		t.Errorf("3 goblins dropped %d coins, want %d–%d", earned, 3*goblin.Coins[0], 3*goblin.Coins[1])
	}
}

func TestSellValidation(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	for _, a := range []resources.ActionSpec{
		{Kind: resources.ActionSell, Material: "nothing", Place: testVillage, Goal: resources.GoalAmount, Amount: 5},
		{Kind: resources.ActionSell, Material: string(resources.ItemIronOre), Place: testMine, Goal: resources.GoalAmount, Amount: 5},
		{Kind: resources.ActionSell, Material: string(resources.ItemIronOre), Place: testVillage, Goal: resources.GoalAmount, Amount: 0},
	} {
		addAction(w, p, a)
	}
	if len(p.actions) != 0 {
		t.Errorf("invalid sell actions accepted: %d", len(p.actions))
	}
}
