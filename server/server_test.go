package main

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomknightdev/islanders-golang/resources"
)

var worldOnce sync.Once

// startWorld runs the package's world once for all end-to-end tests; only
// one goroutine may ever own it.
func startWorld() { worldOnce.Do(func() { go world.Run() }) }

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(url, "http")+"/connect", nil)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

// readUntil reads messages until match returns true or the deadline passes.
func readUntil(t *testing.T, conn *websocket.Conn, d time.Duration, match func(*resources.Message) bool) *resources.Message {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(d))
	for {
		m := &resources.Message{}
		if err := conn.ReadJSON(m); err != nil {
			t.Fatalf("no matching message: %v", err)
		}
		if match(m) {
			return m
		}
	}
}

func TestEndToEndLoginAndWalk(t *testing.T) {
	startWorld()
	ts := httptest.NewServer(httpHandler())
	defer ts.Close()

	conn := dial(t, ts.URL)
	defer conn.Close()
	conn.WriteJSON(resources.NewConnectRequestMessage(resources.ConnectRequestContents{Username: "e2e", Password: "pw"}))

	resp := readUntil(t, conn, 2*time.Second, func(m *resources.Message) bool {
		return m.MessageType == resources.ConnectResponseMessage
	})
	var contents resources.ConnectResponseContents
	if err := resources.DecodeContents(resp.Contents, &contents); err != nil {
		t.Fatal(err)
	}
	spawn := tile{contents.WorldMap.SpawnX, contents.WorldMap.SpawnY}
	if sx, sy := spawn.pixel(); contents.Pos[0] != sx || contents.Pos[1] != sy {
		t.Fatalf("spawned at %v, want the map's spawn %v", contents.Pos, spawn)
	}

	// Walk three tiles east of spawn (still on the village square).
	dest := spawn.add(tile{3, 0})
	conn.WriteJSON(resources.NewMoveRequestMessage(contents.ClientId, resources.MoveRequestContents{X: dest.x, Y: dest.y}))
	readUntil(t, conn, 3*time.Second, func(m *resources.Message) bool {
		if m.MessageType != resources.UpdateMessage || m.ClientId != contents.ClientId {
			return false
		}
		var u resources.UpdateContents
		resources.DecodeContents(m.Contents, &u)
		dx, dy := dest.pixel()
		return u.Pos[0] == dx && u.Pos[1] == dy
	})

	// Wrong password on a second connection is refused.
	bad := dial(t, ts.URL)
	defer bad.Close()
	bad.WriteJSON(resources.NewConnectRequestMessage(resources.ConnectRequestContents{Username: "e2e", Password: "nope"}))
	readUntil(t, bad, 2*time.Second, func(m *resources.Message) bool {
		return m.MessageType == resources.FailedToConnectMessage
	})
}

func TestEndToEndActionList(t *testing.T) {
	startWorld()
	ts := httptest.NewServer(httpHandler())
	defer ts.Close()

	conn := dial(t, ts.URL)
	defer conn.Close()
	conn.WriteJSON(resources.NewConnectRequestMessage(resources.ConnectRequestContents{Username: "planner", Password: "pw"}))

	// The first inventory lists the spawn village as found.
	var village string
	readUntil(t, conn, 2*time.Second, func(m *resources.Message) bool {
		if m.MessageType != resources.InventoryMessage {
			return false
		}
		var inv resources.InventoryContents
		resources.DecodeContents(m.Contents, &inv)
		if len(inv.Found) == 1 {
			village = inv.Found[0]
		}
		return village != ""
	})

	// Smelting with an empty bank runs, then stops with nothing to smelt.
	conn.WriteJSON(resources.NewActionCommandMessage(resources.ActionCommandContents{
		Op:     resources.ActionOpAdd,
		Action: resources.ActionSpec{Kind: resources.ActionSmelt, Material: "copper", Place: village, Goal: resources.GoalAmount, Amount: 10},
	}))
	conn.WriteJSON(resources.NewActionCommandMessage(resources.ActionCommandContents{Op: resources.ActionOpStart}))
	readUntil(t, conn, 10*time.Second, func(m *resources.Message) bool {
		if m.MessageType != resources.InventoryMessage {
			return false
		}
		var inv resources.InventoryContents
		resources.DecodeContents(m.Contents, &inv)
		return inv.Activity == activityNoOre && !inv.Running && len(inv.Actions) == 0
	})
}
