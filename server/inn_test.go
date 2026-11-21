package main

import (
	"strings"
	"testing"
	"time"

	"github.com/tomknightdev/islanders-golang/resources"
)

func visitAction(b resources.Building) resources.ActionSpec {
	return resources.ActionSpec{Kind: resources.ActionVisit, Material: string(b), Place: testVillage, Amount: 1}
}

func TestVisitWalksThereAndStops(t *testing.T) {
	w := jobWorld(t)
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.found[testMine] = true
	p.pos[0], p.pos[1] = tile{12, 12}.pixel()
	addAction(w, p, visitAction(resources.BuildingStore))
	addAction(w, p, mineFor(10))
	run(w, p)
	runUntil(t, w, 30*ticks(time.Second), "arrived", func() bool { return len(p.actions) == 1 })
	w.tick()
	if p.running || w.visiting(p).at != resources.BuildingStore {
		t.Fatalf("on arrival the list should stop at the store: running %v, at %+v", p.running, w.visiting(p))
	}
	if inv := lastInventory(t, s); inv.At != resources.BuildingStore || inv.AtVillage != testVillage {
		t.Errorf("client told at %q in %q", inv.At, inv.AtVillage)
	}
	if p.actions[0].spec.Kind != resources.ActionMine {
		t.Error("the rest of the list should wait for Start")
	}
}

func TestVisitValidation(t *testing.T) {
	w := jobWorld(t)
	p := joinAt(t, w, "tom", tile{4, 5})
	bad := visitAction("castle")
	elsewhere := visitAction(resources.BuildingBank)
	elsewhere.Place = testMine
	addAction(w, p, bad)
	addAction(w, p, elsewhere)
	if len(p.actions) != 0 {
		t.Errorf("invalid visits accepted: %d", len(p.actions))
	}
	for _, b := range resources.Buildings {
		if !w.validAction(p, visitAction(b)) {
			t.Errorf("a visit to the village's %s was refused", b)
		}
	}
}

func TestRumoursPointToUnfoundPlaces(t *testing.T) {
	w := jobWorld(t)
	p, _ := w.login("tom", "pw", testSession()) // has only found the village
	got := w.rumours(p, testVillage)
	if len(got) != 2 {
		t.Fatalf("rumours %q", got)
	}
	// The mine (12,10) is south-east of the village (4,4) and nearer than the forest.
	if !strings.Contains(got[0], testMine) || !strings.Contains(got[0], "south-east") || !strings.Contains(got[0], "copper") {
		t.Errorf("mine rumour %q", got[0])
	}
	p.found[testMine] = true
	if got := w.rumours(p, testVillage); len(got) != 1 || !strings.Contains(got[0], testWood) {
		t.Errorf("found places shouldn't be rumoured: %q", got)
	}
}

func TestCompass(t *testing.T) {
	for _, c := range []struct {
		dx, dy float64
		want   string
	}{{1, 0, "east"}, {0, -1, "north"}, {-1, 1, "south-west"}, {1, -1, "north-east"}, {-1, 0, "west"}} {
		if got := compass(c.dx, c.dy); got != c.want {
			t.Errorf("compass(%v, %v) = %q, want %q", c.dx, c.dy, got, c.want)
		}
	}
}
