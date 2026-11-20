package main

import (
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

// testMap is an open grass field with optional blocked tiles on the object layer.
func testMap(w, h int, blocked ...tile) resources.WorldMap {
	ground := make([]int, w*h)
	objects := make([]int, w*h)
	for i := range ground {
		ground[i] = resources.TileGrass
	}
	for _, b := range blocked {
		objects[b.y*w+b.x] = resources.TileIronRock
	}
	return resources.WorldMap{
		Width:  w,
		Height: h,
		Layers: []resources.Layer{{Data: ground}, {Data: objects}},
	}
}

func testSession() *session {
	return &session{send: make(chan *resources.Message, sessionBuffer)}
}

func drain(s *session) []*resources.Message {
	var out []*resources.Message
	for {
		select {
		case m := <-s.send:
			out = append(out, m)
		default:
			return out
		}
	}
}

func TestFindPathStraight(t *testing.T) {
	wm := testMap(10, 10)
	path := findPath(&wm, tile{1, 1}, tile{4, 1})
	if len(path) != 3 || path[2] != (tile{4, 1}) {
		t.Fatalf("got %v", path)
	}
}

func TestFindPathAroundWall(t *testing.T) {
	// Wall at x=3 from y=0..3, gap at y=4.
	wm := testMap(10, 10, tile{3, 0}, tile{3, 1}, tile{3, 2}, tile{3, 3})
	path := findPath(&wm, tile{1, 1}, tile{5, 1})
	if path == nil {
		t.Fatal("no path")
	}
	prev := tile{1, 1}
	for _, p := range path {
		if !wm.TilePassable(p.x, p.y) {
			t.Fatalf("path goes through blocked tile %v", p)
		}
		if abs(p.x-prev.x)+abs(p.y-prev.y) != 1 {
			t.Fatalf("path jumps from %v to %v", prev, p)
		}
		prev = p
	}
	if prev != (tile{5, 1}) {
		t.Fatalf("path ends at %v", prev)
	}
	// Shortest is down to y=4, across, and back up: 3 + 4 + 3 = 10 steps.
	if len(path) != 10 {
		t.Errorf("path length %d, want 10", len(path))
	}
}

func TestFindPathToBlockedTileStopsBesideIt(t *testing.T) {
	wm := testMap(10, 10, tile{5, 5})
	path := findPath(&wm, tile{1, 5}, tile{5, 5})
	if len(path) == 0 {
		t.Fatal("no path")
	}
	if end := path[len(path)-1]; end != (tile{4, 5}) {
		t.Errorf("ended at %v, want the tile beside the rock (4,5)", end)
	}
}

func TestFindPathUnreachable(t *testing.T) {
	// Target boxed in on all sides, and itself blocked.
	wm := testMap(10, 10, tile{5, 5}, tile{4, 5}, tile{6, 5}, tile{5, 4}, tile{5, 6})
	if path := findPath(&wm, tile{1, 1}, tile{5, 5}); path != nil {
		t.Errorf("expected nil, got %v", path)
	}
	if path := findPath(&wm, tile{1, 1}, tile{-5, 3}); path != nil {
		t.Errorf("far off-map target: expected nil, got %v", path)
	}
}

func TestFindPathAlreadyThere(t *testing.T) {
	wm := testMap(10, 10)
	if path := findPath(&wm, tile{2, 2}, tile{2, 2}); path == nil || len(path) != 0 {
		t.Errorf("expected empty path, got %v", path)
	}
}

func TestPlayerWalksToClickedTile(t *testing.T) {
	w := NewWorld(testMap(20, 20))
	w.npcs = nil
	s := testSession()
	p, err := w.login("tom", "pw", s)
	if err != nil {
		t.Fatal(err)
	}
	p.pos[0], p.pos[1] = tile{2, 2}.pixel()

	w.moveTo(p, tile{6, 2})
	ticks := 0
	for len(p.path) > 0 && ticks < 1000 {
		w.tick()
		ticks++
	}
	if x, y := (tile{6, 2}).pixel(); p.pos[0] != x || p.pos[1] != y {
		t.Fatalf("player at %v, want (%v,%v)", p.pos, x, y)
	}
	// 4 tiles at walkSpeed tiles/s with 40 ticks/s.
	want := int(4 / walkSpeed * 40)
	if ticks < want-1 || ticks > want+1 {
		t.Errorf("took %d ticks, want about %d", ticks, want)
	}
	// The owner is told about their own position now.
	sawSelf := false
	for _, m := range drain(s) {
		if m.MessageType == resources.UpdateMessage && m.ClientId == p.id {
			sawSelf = true
		}
	}
	if !sawSelf {
		t.Error("owner never received their own position update")
	}
}

func TestRedirectMidStepDoesNotSnap(t *testing.T) {
	w := NewWorld(testMap(20, 20))
	w.npcs = nil
	p, _ := w.login("tom", "pw", testSession())
	p.pos[0], p.pos[1] = tile{5, 5}.pixel()

	w.moveTo(p, tile{8, 5})
	w.tick() // now part-way to (6,5)
	before := p.pos
	w.moveTo(p, tile{5, 8})
	if p.pos != before {
		t.Fatal("redirect moved the player")
	}
	if p.path[0] != (tile{6, 5}) {
		t.Errorf("should finish the step to (6,5) first, path starts at %v", p.path[0])
	}
}

func TestReconnectKeepsCharacterAndKicksOldSession(t *testing.T) {
	w := NewWorld(testMap(20, 20))
	w.npcs = nil
	s1 := testSession()
	p1, _ := w.login("tom", "pw", s1)
	p1.pos[0], p1.pos[1] = tile{7, 7}.pixel()

	if _, err := w.login("tom", "wrong", testSession()); err == nil {
		t.Error("wrong password accepted")
	}

	s2 := testSession()
	p2, err := w.login("tom", "pw", s2)
	if err != nil {
		t.Fatal(err)
	}
	if p2 != p1 || p2.sess != s2 {
		t.Fatal("reconnect should reuse the character with the new session")
	}
	if !s1.closed {
		t.Error("old session should be closed")
	}
	// The old connection dropping must not detach the new one.
	w.disconnect(p1, s1)
	if p1.sess != s2 {
		t.Error("stale disconnect detached the new session")
	}
	w.disconnect(p1, s2)
	if p1.sess != nil {
		t.Error("player should be offline")
	}
	if _, ok := w.byName["tom"]; !ok {
		t.Error("offline character should stay in the world")
	}
}

func TestOfflineCharacterKeepsWalking(t *testing.T) {
	w := NewWorld(testMap(20, 20))
	w.npcs = nil
	s := testSession()
	p, _ := w.login("tom", "pw", s)
	p.pos[0], p.pos[1] = tile{2, 2}.pixel()
	w.moveTo(p, tile{10, 2})
	w.disconnect(p, s)
	for i := 0; i < 200; i++ {
		w.tick()
	}
	if x, _ := (tile{10, 2}).pixel(); p.pos[0] != x {
		t.Errorf("offline player stopped at %v", p.pos)
	}
}

func TestSlowClientIsDisconnectedNotBlocking(t *testing.T) {
	s := &session{send: make(chan *resources.Message, 2)}
	for i := 0; i < 5; i++ {
		s.push(resources.NewChatMessage(0, "hi")) // must not block
	}
	if !s.closed {
		t.Error("overflowing session should be closed")
	}
}
