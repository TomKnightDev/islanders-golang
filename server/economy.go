package main

import (
	"math/rand"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

// Economy tuning. See docs/ECONOMY.md.
const (
	demandDropPerUnit = 0.98 // each unit a store buys lowers its price for that item by 2%
	demandFloor       = 0.2  // a store never pays less than 20% of its full price
	demandRecovery    = 0.03 // share of the gap back to full price recovered each minute
	stableSlowdown    = 2    // resting in the stable takes this many times as long
)

const activityStable = "Resting in the stable"

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

// --- the shop ---

// shopCommand shows the store's prices, buys from it or sells to it. Only
// allowed standing at the store; purchases go into the backpack and sales
// come out of it.
func (w *World) shopCommand(p *player, c resources.ShopCommandContents) {
	v := w.visiting(p)
	st := w.storeIn(v.village)
	if v.at != resources.BuildingStore || st == nil {
		return
	}
	switch c.Op {
	case resources.ShopOpBuy:
		w.shopBuy(p, c.Item, c.Count)
	case resources.ShopOpSell:
		w.shopSell(p, st, c)
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
	added := w.give(&p.backpack, k, count, quality)
	p.coins -= added * price
	p.invDirty = true
}

// shopSell sells a piece by ID, or Count of a material, from the backpack.
// Each unit is priced as the store's demand falls.
func (w *World) shopSell(p *player, st *store, c resources.ShopCommandContents) {
	var sold []resources.Stack
	if c.ID != 0 {
		if piece, ok := p.backpack.TakeID(c.ID); ok {
			sold = append(sold, piece)
		}
	} else if c.Count > 0 && c.Quality > 0 {
		sold = p.backpack.TakeQuality(c.Item, c.Quality, c.Count)
	} else if c.Count > 0 {
		sold = p.backpack.Take(c.Item, c.Count)
	}
	for _, part := range sold {
		if !resources.Sellable(part.Kind) {
			p.backpack.AddStack(part)
			continue
		}
		for i := 0; i < part.Count; i++ {
			p.coins += st.buy(part.Kind, part.Quality)
		}
	}
	p.invDirty = true
}

// storeContents is a store's prices: what it sells, and what it pays for
// everything the player carries (gear and tools piece by piece).
func (w *World) storeContents(p *player, st *store) resources.StoreContents {
	c := resources.StoreContents{Village: st.village}
	for _, s := range resources.StoreStock {
		c.Prices = append(c.Prices, resources.StorePrices{Item: s.Item, Buy: resources.StorePrice(s)})
	}
	seen := map[resources.ItemKind]bool{}
	for _, s := range p.backpack.Stacks {
		if !resources.Sellable(s.Kind) {
			continue
		}
		if resources.IsPiece(s.Kind) {
			c.Prices = append(c.Prices, resources.StorePrices{Item: s.Kind, ID: s.ID, Pays: st.pays(s.Kind, s.Quality)})
		} else if !seen[s.Kind] {
			seen[s.Kind] = true
			c.Prices = append(c.Prices, resources.StorePrices{Item: s.Kind, Pays: st.pays(s.Kind, p.backpack.AverageQuality(s.Kind))})
		}
	}
	return c
}

// --- coins elsewhere ---

func (w *World) dropCoins(p *player, e resources.EnemyKind) {
	p.coins += e.Coins[0] + rand.Intn(e.Coins[1]-e.Coins[0]+1)
	p.invDirty = true
}
