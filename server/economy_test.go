package main

import (
	"testing"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// atStore is next to the test village's store, at (1,6).
var atStore = tile{2, 6}

func TestStartingCoins(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	if p.coins != resources.StartingCoins {
		t.Errorf("new characters start with %d coins, got %d", resources.StartingCoins, p.coins)
	}
}

func TestSellingAtTheStore(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", atStore)
	p.backpack.Add(resources.ItemIronOre, 10)
	sword := w.newPiece(resources.ItemIronSword, 90)
	p.backpack.AddStack(sword)
	st := w.storeIn(testVillage)
	fullPrice := st.pays(resources.ItemIronOre, resources.DefaultQuality)
	swordPrice := st.pays(resources.ItemIronSword, 90)

	w.shopCommand(p, resources.ShopCommandContents{Op: resources.ShopOpSell, Item: resources.ItemIronOre, Count: 10})
	earned := p.coins - resources.StartingCoins
	if p.backpack.Count(resources.ItemIronOre) != 0 || earned <= 0 || earned > 10*fullPrice {
		t.Errorf("earned %d for 10 iron ore at up to %d each; %d left", earned, fullPrice, p.backpack.Count(resources.ItemIronOre))
	}
	if st.demandFor(resources.ItemIronOre) >= 1 {
		t.Error("selling should lower the store's demand")
	}
	before := p.coins
	w.shopCommand(p, resources.ShopCommandContents{Op: resources.ShopOpSell, ID: sword.ID})
	if p.coins-before != swordPrice || p.backpack.Count(resources.ItemIronSword) != 0 {
		t.Errorf("sword sold for %d, want %d", p.coins-before, swordPrice)
	}

	// Not from afar.
	p.backpack.Add(resources.ItemIronOre, 3)
	p.pos[0], p.pos[1] = tile{12, 12}.pixel()
	w.shopCommand(p, resources.ShopCommandContents{Op: resources.ShopOpSell, Item: resources.ItemIronOre, Count: 3})
	if p.backpack.Count(resources.ItemIronOre) != 3 {
		t.Error("sold away from the store")
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
	p := joinAt(t, w, "tom", atStore)
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
	if p.backpack.Count(resources.ItemCabbage) != 4 || p.coins != resources.StartingCoins-4*price {
		t.Errorf("bought 4 cabbages: backpack %d, coins %d", p.backpack.Count(resources.ItemCabbage), p.coins)
	}
	buy(testVillage, resources.ItemIronSword, 1)
	if p.backpack.Count(resources.ItemIronSword) != 0 {
		t.Error("the store doesn't sell iron swords")
	}
	p.coins = price*2 + 1
	buy(testVillage, resources.ItemCabbage, 10)
	if p.backpack.Count(resources.ItemCabbage) != 6 || p.coins != 1 {
		t.Errorf("should only buy what's affordable: backpack %d, coins %d", p.backpack.Count(resources.ItemCabbage), p.coins)
	}
	p.pos[0], p.pos[1] = tile{12, 12}.pixel()
	p.coins = 100
	buy(testVillage, resources.ItemCabbage, 1)
	if p.coins != 100 {
		t.Error("can't shop away from the store")
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
