package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Shop is a village general store: what it sells, delivered to the bank,
// and what it pays right now for what's in the bank. Selling itself is a
// Sell action, since goods have to be carried there.
type Shop struct {
	mgr    *renderer.Manager
	places map[string]resources.Place
	send   func(resources.ShopCommandContents)

	open    bool
	village string
	found   []string
	coins   int
	store   resources.StoreContents
}

func NewShop(mgr *renderer.Manager, places []resources.Place, send func(resources.ShopCommandContents)) *Shop {
	s := &Shop{mgr: mgr, places: map[string]resources.Place{}, send: send}
	for _, p := range places {
		s.places[p.Name] = p
	}
	return s
}

// Open shows the shop, at the first village found if none is picked yet.
func (s *Shop) Open() {
	s.open = true
	if s.village == "" {
		for _, v := range s.found {
			if s.places[v].Kind == resources.PlaceVillage {
				s.village = v
				break
			}
		}
	}
	s.view()
}

func (s *Shop) view() {
	if s.village != "" {
		s.send(resources.ShopCommandContents{Op: resources.ShopOpView, Village: s.village})
	}
}

// SetState keeps the shop's list of found villages and coins current.
func (s *Shop) SetState(c resources.InventoryContents) {
	s.found, s.coins = c.Found, c.Coins
}

// SetStore takes a store's prices from the server.
func (s *Shop) SetStore(c resources.StoreContents) {
	if c.Village == s.village {
		s.store = c
	}
}

func (s *Shop) Update() error {
	if !s.open {
		return nil
	}
	w, h := screenSize()
	imgui.SetNextWindowPosV(imgui.Vec2{X: w / 2, Y: h / 2}, imgui.ConditionAppearing, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.BeginV("Shop", &s.open, imgui.WindowFlagsAlwaysAutoResize)

	if imgui.BeginCombo("Village", s.village) {
		for _, v := range s.found {
			if s.places[v].Kind != resources.PlaceVillage {
				continue
			}
			if imgui.SelectableV(v, v == s.village, 0, imgui.Vec2{}) && v != s.village {
				s.village = v
				s.store = resources.StoreContents{}
				s.view()
			}
		}
		imgui.EndCombo()
	}
	imgui.Text(fmt.Sprintf("Coins: %d", s.coins))
	imgui.Separator()

	if s.store.Village != s.village {
		dimmed("Loading prices...")
		imgui.End()
		return nil
	}

	imgui.Text("For sale (goes to your bank)")
	for _, p := range s.store.Prices {
		if p.Buy == 0 {
			continue
		}
		imgui.PushID(string(p.Item))
		imgui.Text(fmt.Sprintf("%-18s %4d coins", resources.ItemNames[p.Item], p.Buy))
		imgui.SameLine()
		if button("Buy 1", s.coins >= p.Buy) {
			s.send(resources.ShopCommandContents{Op: resources.ShopOpBuy, Village: s.village, Item: p.Item, Count: 1})
		}
		imgui.SameLine()
		if button("Buy 10", s.coins >= p.Buy) {
			s.send(resources.ShopCommandContents{Op: resources.ShopOpBuy, Village: s.village, Item: p.Item, Count: 10})
		}
		imgui.PopID()
	}

	imgui.Separator()
	imgui.Text("The store pays (each)")
	for _, p := range s.store.Prices {
		if p.Pays > 0 {
			imgui.Text(fmt.Sprintf("%-18s %4d coins", resources.ItemNames[p.Item], p.Pays))
		}
	}
	dimmed("To sell, add a Sell action: your character carries the goods here.")
	dimmed("Prices drop as the store buys more, and recover over a couple of hours.")
	imgui.End()
	return nil
}

func (s *Shop) Draw(screen *ebiten.Image) {
	s.mgr.Draw(screen)
}
