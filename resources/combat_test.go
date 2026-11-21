package resources

import "testing"

func TestHitChance(t *testing.T) {
	even := HitChance(AttackRoll(10, 0), DefenceRoll(10, 0))
	if even < 0.4 || even > 0.55 {
		t.Errorf("evenly matched hit chance %.2f, want about a half", even)
	}
	if HitChance(AttackRoll(50, 30), DefenceRoll(1, 0)) < 0.9 {
		t.Error("a strong attacker should almost always hit a weak defender")
	}
	if HitChance(AttackRoll(1, 0), DefenceRoll(50, 50)) > 0.1 {
		t.Error("a weak attacker should rarely hit a strong defender")
	}
}

func TestMaxHit(t *testing.T) {
	// A new character with a copper sword hits around 10.
	if got := MaxHit(1, Equipment[ItemCopperSword].Strength); got < 9 || got > 12 {
		t.Errorf("level 1 copper sword max hit %d", got)
	}
	if MaxHit(45, Equipment[ItemMithrilSword].Strength) <= MaxHit(1, Equipment[ItemCopperSword].Strength)*3 {
		t.Error("strength and better weapons should raise max hit a lot")
	}
}

func TestEquipmentBonuses(t *testing.T) {
	b := EquipmentBonuses(map[EquipSlot]Stack{
		SlotWeapon: {ItemOakShortbow, 1},
		SlotAmmo:   {ItemPineArrows, 40},
		SlotHead:   {ItemCopperHelmet, 1},
	})
	if !b.Ranged || b.Accuracy != 8 || b.Strength != 14 || b.Defence != 4 {
		t.Errorf("bow + pine arrows + helmet: %+v", b)
	}
	// Arrows don't add strength to a sword.
	b = EquipmentBonuses(map[EquipSlot]Stack{SlotWeapon: {ItemCopperSword, 1}, SlotAmmo: {ItemOakArrows, 10}})
	if b.Ranged || b.Strength != 6 {
		t.Errorf("sword with arrows: %+v", b)
	}
}

func TestEnemies(t *testing.T) {
	for i, e := range Enemies {
		if i > 0 && e.Level <= Enemies[i-1].Level {
			t.Errorf("%s should be tougher than %s", e.Name, Enemies[i-1].Name)
		}
		for _, l := range e.Loot {
			if _, ok := Items[l.Item]; !ok {
				t.Errorf("%s drops unknown item %q", e.Name, l.Item)
			}
		}
		if _, ok := FindMaterial(ActionFight, e.ID); !ok {
			t.Errorf("can't add a fight against %s", e.Plural)
		}
	}
	for k, d := range Equipment {
		if _, ok := Items[k]; !ok {
			t.Errorf("equipment %q isn't an item", k)
		}
		if d.Level < 1 || SkillNames[d.Skill] == "" {
			t.Errorf("equipment %q has a bad requirement", k)
		}
	}
}
