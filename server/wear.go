package main

import (
	"fmt"
	"math/rand"

	"github.com/tomknightdev/islanders-golang/resources"
)

// Wear in battle: a weapon may wear with each attack, and a piece of armour
// with each hit taken. Better quality wears more slowly. At 100% a piece
// breaks and is gone; repair it first (server/repair.go).
const (
	weaponWearChance = 0.15 // per attack, at quality 50
	armourWearChance = 0.30 // per hit taken, at quality 50
	wearWarnAt       = 75   // chat warns when a piece passes this
)

// wearInBattle may wear the piece in a slot by 1%: chance at quality 50,
// half again as likely at quality 0, half as likely at 100.
func (w *World) wearInBattle(p *player, slot resources.EquipSlot, chance float64) {
	s, ok := p.equipment[slot]
	if !ok {
		return
	}
	if rand.Float64() < chance*(1.5-float64(s.Quality)/100) {
		w.wearPiece(p, slot, 1)
	}
}

// wearArmour may wear one of the armour pieces p has on, at random.
func (w *World) wearArmour(p *player) {
	var worn []resources.EquipSlot
	for _, slot := range []resources.EquipSlot{resources.SlotHead, resources.SlotBody, resources.SlotShield} {
		if _, ok := p.equipment[slot]; ok {
			worn = append(worn, slot)
		}
	}
	if len(worn) > 0 {
		w.wearInBattle(p, worn[rand.Intn(len(worn))], armourWearChance)
	}
}

// wearPiece adds wear to the piece in a slot, warning in chat as it nears
// breaking, and reports whether it broke (and is gone).
func (w *World) wearPiece(p *player, slot resources.EquipSlot, n int) bool {
	s, ok := p.equipment[slot]
	if !ok {
		return false
	}
	before := s.Damage
	s.Damage = min(100, s.Damage+n)
	p.invDirty = true
	say := func(text string) {
		if p.sess != nil {
			p.sess.push(resources.NewChatMessage(p.id, text))
		}
	}
	if s.Damage >= 100 {
		delete(p.equipment, slot)
		say(fmt.Sprintf("Your %s broke.", resources.ItemNames[s.Kind]))
		return true
	}
	if before < wearWarnAt && s.Damage >= wearWarnAt {
		say(fmt.Sprintf("Your %s is %d%% worn and will break soon. Repair it from the Equipment window.", resources.ItemNames[s.Kind], s.Damage))
	}
	p.equipment[slot] = s
	return false
}
