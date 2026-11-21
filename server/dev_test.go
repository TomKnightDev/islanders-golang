package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

func TestGameSpeedNeedsDevMode(t *testing.T) {
	w := jobWorld(t)
	w.devCommand(resources.DevCommandContents{Speed: 10})
	if w.speed > 1 {
		t.Fatal("a server without -dev must ignore speed changes")
	}

	w.dev = true
	s := testSession()
	w.login("tom", "pw", s)
	drain(s)
	w.devCommand(resources.DevCommandContents{Speed: 10})
	if w.speed != 10 {
		t.Fatalf("speed %d, want 10", w.speed)
	}
	told := false
	for _, m := range drain(s) {
		if m.MessageType == resources.DevStateMessage && m.Contents.(resources.DevStateContents).Speed == 10 {
			told = true
		}
	}
	if !told {
		t.Error("players should be told the new speed")
	}
	w.devCommand(resources.DevCommandContents{Speed: 500})
	if w.speed != resources.MaxGameSpeed {
		t.Errorf("speed should be capped at %d, got %d", resources.MaxGameSpeed, w.speed)
	}
}

func TestActionTimingsAreRealTime(t *testing.T) {
	w := jobWorld(t)
	w.dev = true
	p := joinAt(t, w, "tom", tile{4, 5})
	w.startAction(p, activityMining, ticks(swingTime))
	normal := w.inventoryMessage(p).Contents.(resources.InventoryContents).ActionSeconds
	w.devCommand(resources.DevCommandContents{Speed: 4})
	fast := w.inventoryMessage(p).Contents.(resources.InventoryContents).ActionSeconds
	if normal != swingTime.Seconds() || fast != normal/4 {
		t.Errorf("a swing shows as %vs at 1x and %vs at 4x", normal, fast)
	}
}
