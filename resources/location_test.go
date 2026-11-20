package resources

import "testing"

func TestGridRef(t *testing.T) {
	cases := []struct {
		x, y int
		want string
	}{
		{0, 0, "A1"},
		{9, 9, "A1"},
		{10, 0, "B1"},
		{52, 48, "F5"},
		{50, 50, "F6"},
		{99, 99, "J10"},
	}
	for _, c := range cases {
		if got := GridRef(c.x, c.y); got != c.want {
			t.Errorf("GridRef(%d,%d) = %q, want %q", c.x, c.y, got, c.want)
		}
	}
}

func TestLocationAndLandmark(t *testing.T) {
	wm := WorldMap{Objects: []WorldObject{
		{Kind: ObjectBank, X: 47, Y: 44},
		{Kind: ObjectIronRock, X: 59, Y: 57},
	}}
	if got := Location(50, 50); got != "F6 (50, 50)" {
		t.Errorf("Location = %q", got)
	}
	if got := wm.Describe(48, 46); got != "near the bank" {
		t.Errorf("beside the bank: %q", got)
	}
	if got := wm.Describe(50, 51); got != "" {
		t.Errorf("open ground should have no landmark, got %q", got)
	}
	if got := wm.Describe(60, 58); got != "near the iron rocks" {
		t.Errorf("by a rock: %q", got)
	}
}

func TestDescribePlaces(t *testing.T) {
	wm := WorldMap{
		Places: []Place{
			{Name: "Oakhaven", Kind: PlaceVillage, X: 50, Y: 50, Radius: 6},
			{Name: "Ironridge Mine", Kind: PlaceMine, X: 90, Y: 50, Radius: 5},
		},
		Objects: []WorldObject{{Kind: ObjectBank, X: 48, Y: 47}},
	}
	cases := []struct {
		x, y int
		want string
	}{
		{52, 53, "in Oakhaven"}, // places win over the bank inside them
		{91, 48, "at Ironridge Mine"},
		{62, 50, "near Oakhaven"},
		{120, 120, ""},
	}
	for _, c := range cases {
		if got := wm.Describe(c.x, c.y); got != c.want {
			t.Errorf("Describe(%d,%d) = %q, want %q", c.x, c.y, got, c.want)
		}
	}
}

func TestTileAtUsesSpriteCentre(t *testing.T) {
	// Standing exactly on tile (50,50) and halfway to (51,50).
	if x, y := TileAt(400, 400); x != 50 || y != 50 {
		t.Errorf("got %d,%d", x, y)
	}
	if x, _ := TileAt(404, 400); x != 51 {
		t.Errorf("past the halfway point should count as the next tile, got %d", x)
	}
}
