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
	if contents.Pos[0] != 400 || contents.Pos[1] != 400 {
		t.Fatalf("spawned at %v", contents.Pos)
	}

	// Walk three tiles east of spawn (still inside the clearing).
	conn.WriteJSON(resources.NewMoveRequestMessage(contents.ClientId, resources.MoveRequestContents{X: 53, Y: 50}))
	readUntil(t, conn, 3*time.Second, func(m *resources.Message) bool {
		if m.MessageType != resources.UpdateMessage || m.ClientId != contents.ClientId {
			return false
		}
		var u resources.UpdateContents
		resources.DecodeContents(m.Contents, &u)
		return u.Pos[0] == 53*8 && u.Pos[1] == 50*8
	})

	// Wrong password on a second connection is refused.
	bad := dial(t, ts.URL)
	defer bad.Close()
	bad.WriteJSON(resources.NewConnectRequestMessage(resources.ConnectRequestContents{Username: "e2e", Password: "nope"}))
	readUntil(t, bad, 2*time.Second, func(m *resources.Message) bool {
		return m.MessageType == resources.FailedToConnectMessage
	})
}

func TestEndToEndMining(t *testing.T) {
	startWorld()
	ts := httptest.NewServer(httpHandler())
	defer ts.Close()

	conn := dial(t, ts.URL)
	defer conn.Close()
	conn.WriteJSON(resources.NewConnectRequestMessage(resources.ConnectRequestContents{Username: "miner", Password: "pw"}))
	resp := readUntil(t, conn, 2*time.Second, func(m *resources.Message) bool {
		return m.MessageType == resources.ConnectResponseMessage
	})
	var contents resources.ConnectResponseContents
	resources.DecodeContents(resp.Contents, &contents)

	var rockAt *resources.WorldObject
	for _, o := range contents.WorldMap.Objects {
		if o.Kind == resources.ObjectIronRock {
			rockAt = &o
			break
		}
	}
	if rockAt == nil {
		t.Fatal("no iron rock in the world")
	}

	conn.WriteJSON(resources.NewMoveRequestMessage(contents.ClientId, resources.MoveRequestContents{X: rockAt.X, Y: rockAt.Y}))
	readUntil(t, conn, 15*time.Second, func(m *resources.Message) bool {
		if m.MessageType != resources.InventoryMessage {
			return false
		}
		var inv resources.InventoryContents
		resources.DecodeContents(m.Contents, &inv)
		return inv.Backpack.Count(resources.ItemIronOre) >= 1 && inv.Activity == activityMining
	})
}
