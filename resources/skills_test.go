package resources

import "testing"

func TestXPCurve(t *testing.T) {
	if XPForLevel(1) != 0 || XPForLevel(2) != 83 {
		t.Errorf("level 2 at %d XP, want 83", XPForLevel(2))
	}
	if got := XPForLevel(99); got < 13_000_000 || got > 13_100_000 {
		t.Errorf("level 99 at %d XP, want about 13 million", got)
	}
	for l := 2; l <= MaxLevel; l++ {
		if XPForLevel(l) <= XPForLevel(l-1) {
			t.Fatalf("XP doesn't rise from level %d to %d", l-1, l)
		}
		if LevelForXP(XPForLevel(l)) != l || LevelForXP(XPForLevel(l)-1) != l-1 {
			t.Fatalf("LevelForXP is off around level %d", l)
		}
	}
	if LevelForXP(1<<30) != MaxLevel {
		t.Error("XP past level 99 should stay at 99")
	}
}

func TestSpeedFactor(t *testing.T) {
	if SpeedFactor(1) != 1 {
		t.Error("level 1 should be full time")
	}
	if f := SpeedFactor(99); f < 0.6 || f > 0.62 {
		t.Errorf("level 99 factor %v", f)
	}
}

func TestMaterialsAreConsistent(t *testing.T) {
	for _, k := range ActionKinds {
		if IsCombat(k) || IsVisit(k) {
			continue // fighting's materials are enemy kinds, visiting's buildings
		}
		mats := ActionMaterials[k]
		if len(mats) == 0 {
			t.Fatalf("%s has no materials", k)
		}
		if mats[0].Level != 1 {
			t.Errorf("%s: the first tier must be open at level 1", k)
		}
		for i, m := range mats {
			if _, ok := Items[m.Item]; !ok {
				t.Errorf("%s %s gives unknown item %q", k, m.ID, m.Item)
			}
			if IsGathering(k) != (m.Node != 0) || IsGathering(k) == (len(m.Inputs) > 0) {
				t.Errorf("%s %s: gathering needs a node, crafting an input", k, m.ID)
			}
			for _, in := range m.Inputs {
				if _, ok := Items[in.Item]; !ok || in.Count < 1 {
					t.Errorf("%s %s: bad ingredient %+v", k, m.ID, in)
				}
			}
			// Tiers of materials get harder; gear recipes come several to a tier.
			if i > 0 && !IsPiece(m.Item) && !IsPiece(mats[i-1].Item) && (m.Level <= mats[i-1].Level || m.XP <= mats[i-1].XP) {
				t.Errorf("%s %s: higher tiers should need more and give more", k, m.ID)
			}
		}
	}
	if k, m, ok := NodeMaterial(TileCopperRock); !ok || k != ActionMine || m.ID != "copper" {
		t.Error("copper rock sprite should be mineable copper")
	}
	if info, ok := NodeInfo(TilePine); !ok || info.Name != "Pine tree" {
		t.Errorf("pine info %+v", info)
	}
}

func TestEveryGearPieceCanBeMadeAndImproved(t *testing.T) {
	for k := range Equipment {
		if !IsPiece(k) || k == ItemTrainingSword {
			continue
		}
		kind, m, ok := ImproveRecipe(k)
		if !ok {
			t.Errorf("no recipe makes %s", k)
			continue
		}
		if m.Level > 99 || len(m.Inputs) == 0 || (kind != ActionSmith && kind != ActionFletch) {
			t.Errorf("%s: recipe %s %+v", k, kind, m)
		}
	}
	if _, _, ok := ImproveRecipe(ItemIronOre); ok {
		t.Error("materials aren't improved")
	}
}
