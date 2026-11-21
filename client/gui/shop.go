package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// Shop is a village general store's window. It opens when the character
// stops at the store: buy into the backpack, or sell what's in it.
type Shop struct {
	mgr   *renderer.Manager
	send  func(resources.ShopCommandContents)
	state resources.InventoryContents
	store resources.StoreContents

	open    bool
	shownAt string // the village whose store the window was last opened for
}

func NewShop(mgr *renderer.Manager, send func(resources.ShopCommandContents)) *Shop {
	return &Shop{mgr: mgr, send: send}
}

// SetState opens the window on arriving at a store and closes it on leaving.
func (s *Shop) SetState(c resources.InventoryContents) {
	s.state = c
	at := ""
	if c.At == resources.BuildingStore {
		at = c.AtVillage
	}
	if at != s.shownAt {
		s.open = at != ""
		s.shownAt = at
	}
}

// SetStore takes the store's prices from the server.
func (s *Shop) SetStore(c resources.StoreContents) { s.store = c }

func (s *Shop) Update() error {
	if !s.open || s.shownAt == "" {
		return nil
	}
	w, h := screenSize()
	placeWindow("Store", imgui.Vec2{X: w / 2, Y: h / 2}, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.BeginV("Store", &s.open, imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text(fmt.Sprintf("%s store  ·  Coins: %d", s.shownAt, s.state.Coins))
	imgui.Separator()
	if s.store.Village != s.shownAt {
		dimmed("Loading prices...")
		imgui.End()
		return nil
	}

	imgui.BeginGroup()
	imgui.Text("For sale (into your backpack)")
	for _, p := range s.store.Prices {
		if p.Buy == 0 {
			continue
		}
		imgui.PushID("buy" + string(p.Item))
		imgui.Text(fmt.Sprintf("%-18s %4d coins", resources.ItemNames[p.Item], p.Buy))
		itemTooltip(p.Item)
		imgui.SameLine()
		if button("Buy 1", s.state.Coins >= p.Buy) {
			s.send(resources.ShopCommandContents{Op: resources.ShopOpBuy, Village: s.shownAt, Item: p.Item, Count: 1})
		}
		imgui.SameLine()
		if button("Buy 10", s.state.Coins >= p.Buy) {
			s.send(resources.ShopCommandContents{Op: resources.ShopOpBuy, Village: s.shownAt, Item: p.Item, Count: 10})
		}
		imgui.PopID()
	}
	imgui.EndGroup()

	imgui.SameLineV(0, 40)

	imgui.BeginGroup()
	imgui.Text("Sell from your backpack (each)")
	// What the store pays: per piece for gear and tools, per kind for
	// materials, at any quality.
	pays := map[resources.ItemKind]int{}
	piecePays := map[uint32]int{}
	for _, p := range s.store.Prices {
		if p.ID != 0 {
			piecePays[p.ID] = p.Pays
		} else if p.Pays > 0 {
			pays[p.Item] = p.Pays
		}
	}
	any := false
	for _, i := range slotOrder(s.state.Backpack) {
		st := s.state.Backpack.Stacks[i]
		price := pays[st.Kind]
		if st.ID != 0 {
			price = piecePays[st.ID]
		}
		if price == 0 {
			continue
		}
		any = true
		imgui.PushID(fmt.Sprintf("sell%d", i))
		imgui.Text(fmt.Sprintf("%s  %d coins", stackLabel(st), price))
		itemTooltip(st.Kind)
		imgui.SameLine()
		if st.ID != 0 {
			if imgui.Button("Sell") {
				s.send(resources.ShopCommandContents{Op: resources.ShopOpSell, Village: s.shownAt, Item: st.Kind, ID: st.ID})
			}
		} else {
			if imgui.Button("Sell 1") {
				s.send(resources.ShopCommandContents{Op: resources.ShopOpSell, Village: s.shownAt, Item: st.Kind, Quality: st.Quality, Count: 1})
			}
			imgui.SameLine()
			if imgui.Button("Sell all") {
				s.send(resources.ShopCommandContents{Op: resources.ShopOpSell, Village: s.shownAt, Item: st.Kind, Quality: st.Quality, Count: st.Count})
			}
		}
		imgui.PopID()
	}
	if !any {
		dimmed("Nothing in your backpack to sell.")
	}
	dimmed("Prices drop as the store buys more,")
	dimmed("and recover over a couple of hours.")
	imgui.EndGroup()
	imgui.End()
	return nil
}

func (s *Shop) Draw(screen *ebiten.Image) {}
