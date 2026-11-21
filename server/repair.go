package main

import (
	"fmt"
	"slices"

	"github.com/tomknightdev/islanders-golang/resources"
)

// repairCommand repairs a worn piece, wherever the player is: equipped, in
// the backpack, or in the bank while standing at one. The wear goes back to
// 0 at a small cost in quality (less at a higher Repair level), and Repair
// trains a little.
func (w *World) repairCommand(p *player, c resources.RepairCommandContents) {
	var piece *resources.Stack
	if c.Slot != "" {
		if s, ok := p.equipment[c.Slot]; ok {
			defer func() { p.equipment[c.Slot] = s }()
			piece = &s
		}
	} else if c.ID != 0 {
		piece = findPiece(&p.backpack, c.ID)
		if _, atBank := w.bankOpen(p); piece == nil && atBank {
			piece = findPiece(&p.bank, c.ID)
		}
	}
	if piece == nil || piece.Damage == 0 || !resources.IsPiece(piece.Kind) {
		return
	}
	damage, before := piece.Damage, piece.Quality
	piece.Quality = max(1, piece.Quality-resources.RepairLoss(damage, p.level(resources.SkillRepair)))
	piece.Damage = 0
	w.addXP(p, resources.SkillRepair, resources.RepairXP(damage))
	p.invDirty = true
	if p.sess != nil {
		p.sess.push(resources.NewChatMessage(p.id, fmt.Sprintf("You repaired your %s (%d%% worn): quality %d → %d.",
			resources.ItemNames[piece.Kind], damage, before, piece.Quality)))
	}
}

// findPiece is the piece with an ID in an inventory, or nil.
func findPiece(inv *resources.Inventory, id uint32) *resources.Stack {
	i := slices.IndexFunc(inv.Stacks, func(s resources.Stack) bool { return s.ID == id })
	if i < 0 {
		return nil
	}
	return &inv.Stacks[i]
}
