package resources

// Repairing: any worn piece of gear or tool can be repaired anywhere. It
// takes the wear back to 0 but costs a little quality, less at a higher
// Repair level. See docs/DESIGN.md.

// RepairLoss is the quality a repair costs: about 1 for every 20% of wear
// at level 1, down to a third of that at level 99, and at least 1.
func RepairLoss(damage, level int) int {
	if damage <= 0 {
		return 0
	}
	loss := float64(damage) / 20 * (1 - float64(min(level, MaxLevel)-1)/147)
	return max(1, int(loss+0.5))
}

// RepairXP is the Repair XP for mending this much wear.
func RepairXP(damage int) int { return max(1, damage/2) }

// RepairCommandContents repairs a piece: the one in an equipment Slot, or
// the one with ID in the backpack (or in the bank, standing at one).
type RepairCommandContents struct {
	Slot EquipSlot `json:"slot"`
	ID   uint32    `json:"id"`
}

func NewRepairCommandMessage(c RepairCommandContents) *Message {
	return &Message{MessageType: RepairCommandMessage, Contents: c}
}
