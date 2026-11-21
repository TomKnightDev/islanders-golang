package main

import (
	"math/rand"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// Economy tuning. See docs/ECONOMY.md.
const (
	sellTime          = 3 * time.Second // handing goods over at a store
	demandDropPerUnit = 0.98            // each unit a store buys lowers its price for that item by 2%
	demandFloor       = 0.2             // a store never pays less than 20% of its full price
	demandRecovery    = 0.03            // share of the gap back to full price recovered each minute
	stableSlowdown    = 2               // resting in the stable takes this many times as long
)

const (
	activityToStore = "Walking to the store"
	activitySelling = "Selling"
	activityNoGoods = "Nothing left to sell"
	activityStable  = "Resting in the stable"
)

// store is a village general store. Its demand for each item sets how much
// it pays: buying lowers demand, and demand recovers over time.
type store struct {
	village string
	t       tile
	demand  map[resources.ItemKind]float64 // 1 = full price; missing means 1
}

func (s *store) demandFor(k resources.ItemKind) float64 {
	if d, ok := s.demand[k]; ok {
		return d
	}
	return 1
}

// pays is what the store gives for one of an item of a quality right now.
// Higher-quality gear and tools fetch more; materials pay the same at any
// quality.
func (s *store) pays(k resources.ItemKind, quality int) int {
	return max(1, int(float64(resources.QualityValue(k, quality))*resources.StoreBuysAt*s.demandFor(k)))
}

// buy pays for one item and lowers the store's demand for it.
func (s *store) buy(k resources.ItemKind, quality int) int {
	price := s.pays(k, quality)
	s.demand[k] = max(demandFloor, s.demandFor(k)*demandDropPerUnit)
	return price
}

func (w *World) storeIn(village string) *store { return w.stores[village] }

// tickStores lets every store's demand recover towards full price, once a
// minute.
func (w *World) tickStores() {
	if w.ticks%ticks(time.Minute) != 0 {
		return
	}
	for _, s := range w.stores {
		for k, d := range s.demand {
			d += (1 - d) * demandRecovery
			if d > 0.999 {
				delete(s.demand, k)
			} else {
				s.demand[k] = d
			}
		}
	}
}

// --- the sell trip ---

// collectForSale takes the item to sell from the bank, as much as the
// backpack holds and the goal still needs, and heads for the store.
func (w *World) collectForSale(p *player) {
	j := p.job
	k := resources.ItemKind(j.action.spec.Material)
	want := p.bank.Count(k)
	if j.action.spec.Goal == resources.GoalAmount {
		want = min(want, j.action.spec.Amount-j.action.count)
	}
	added := moveItems(&p.bank, &p.backpack, k, want)
	if added == 0 {
		w.finishAction(p, activityNoGoods)
		return
	}
	p.invDirty = true
	j.state = stateWalkToStore
	if !w.moveTo(p, j.station) {
		w.failAction(p, activityNoPath)
		return
	}
	w.startAction(p, activityToStore, walkTicks(p))
}

// sellCarried sells everything of the action's item in the backpack to the
// store. Each unit is priced as the store's demand falls.
func (w *World) sellCarried(p *player) {
	j := p.job
	k := resources.ItemKind(j.action.spec.Material)
	st := w.storeIn(j.action.spec.Place)
	n, total := 0, 0
	for _, part := range p.backpack.Take(k, p.backpack.Count(k)) {
		for i := 0; i < part.Count; i++ {
			total += st.buy(k, part.Quality)
		}
		n += part.Count
	}
	p.coins += total
	j.action.count += n
	p.invDirty = true
	if amountMet(p, j.action) {
		w.wrapUp(p)
		return
	}
	w.goBank(p) // for the next load
}

// --- the shop ---

// shopCommand shows a store's prices or buys from it. Purchases go straight
// to the bank; the player must have found the village.
func (w *World) shopCommand(p *player, c resources.ShopCommandContents) {
	st := w.storeIn(c.Village)
	if st == nil || !p.found[c.Village] {
		return
	}
	if c.Op == resources.ShopOpBuy {
		w.shopBuy(p, c.Item, c.Count)
	}
	if p.sess != nil {
		p.sess.push(resources.NewStoreMessage(w.storeContents(p, st)))
	}
}

func (w *World) shopBuy(p *player, k resources.ItemKind, count int) {
	var stocked *resources.StoreItem
	for i := range resources.StoreStock {
		if resources.StoreStock[i].Item == k {
			stocked = &resources.StoreStock[i]
		}
	}
	if stocked == nil || count < 1 {
		return
	}
	price := resources.StorePrice(*stocked)
	count = min(count, p.coins/price)
	quality := resources.DefaultQuality
	if resources.IsPiece(k) {
		quality = resources.StarterQuality
	}
	added := w.give(&p.bank, k, count, quality)
	p.coins -= added * price
	p.invDirty = true
}

// storeContents is a store's prices: what it sells and, for anything the
// player has in the bank, what it pays.
func (w *World) storeContents(p *player, st *store) resources.StoreContents {
	c := resources.StoreContents{Village: st.village}
	seen := map[resources.ItemKind]bool{}
	for _, s := range resources.StoreStock {
		seen[s.Item] = true
		q := p.bank.AverageQuality(s.Item)
		if q == 0 {
			q = resources.StarterQuality
		}
		c.Prices = append(c.Prices, resources.StorePrices{Item: s.Item, Buy: resources.StorePrice(s), Pays: st.pays(s.Item, q)})
	}
	for _, s := range p.bank.Stacks {
		if !seen[s.Kind] && resources.Sellable(s.Kind) {
			seen[s.Kind] = true
			c.Prices = append(c.Prices, resources.StorePrices{Item: s.Kind, Pays: st.pays(s.Kind, p.bank.AverageQuality(s.Kind))})
		}
	}
	return c
}

// --- coins elsewhere ---

func (w *World) dropCoins(p *player, e resources.EnemyKind) {
	p.coins += e.Coins[0] + rand.Intn(e.Coins[1]-e.Coins[0]+1)
	p.invDirty = true
}
