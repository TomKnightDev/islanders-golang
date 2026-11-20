package main

import (
	"encoding/json"
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

func TestWorldObjectsBlockAndAreReachable(t *testing.T) {
	wm := GenerateWorldMap()

	if len(wm.Objects) != len(townObjects) {
		t.Fatalf("got %d objects, want %d", len(wm.Objects), len(townObjects))
	}
	for _, o := range wm.Objects {
		if got := wm.Layers[resources.LayerObject].Data[o.Y*wm.Width+o.X]; got != resources.ObjectTiles[o.Kind] {
			t.Errorf("%s at (%d,%d): object layer has tile %d", o.Kind, o.X, o.Y, got)
		}
		if wm.TilePassable(o.X, o.Y) {
			t.Errorf("%s at (%d,%d) should block movement", o.Kind, o.X, o.Y)
		}
		reachable := false
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if wm.TilePassable(o.X+d[0], o.Y+d[1]) {
				reachable = true
			}
		}
		if !reachable {
			t.Errorf("%s at (%d,%d) has no walkable neighbour", o.Kind, o.X, o.Y)
		}
	}
}

func TestSpawnIsClear(t *testing.T) {
	wm := GenerateWorldMap()
	for y := spawnMin; y <= spawnMax; y++ {
		for x := spawnMin; x <= spawnMax; x++ {
			if !wm.TilePassable(x, y) {
				t.Errorf("spawn tile (%d,%d) is blocked", x, y)
			}
		}
	}
	if !wm.BoxPassable(400, 400) {
		t.Error("player spawn position (400,400) is blocked")
	}
}

func TestTreesAreOnObjectLayer(t *testing.T) {
	wm := GenerateWorldMap()
	trees := 0
	for _, gid := range wm.Layers[resources.LayerGround].Data {
		if gid == resources.TileTree1 || gid == resources.TileTree2 {
			t.Fatal("found a tree on the ground layer")
		}
	}
	for _, gid := range wm.Layers[resources.LayerObject].Data {
		if gid == resources.TileTree1 || gid == resources.TileTree2 {
			trees++
		}
	}
	if trees == 0 {
		t.Error("no trees placed")
	}
}

func TestBoxPassableEdges(t *testing.T) {
	wm := GenerateWorldMap()
	if wm.BoxPassable(-1, 400) || wm.BoxPassable(400, float64(mapHeight*resources.TileSize-7)) {
		t.Error("boxes leaving the map should be blocked")
	}
	if wm.BoxPassable(0, 0) {
		t.Error("deep water corner should be blocked")
	}
}

// The client receives the map as interface{} from websocket ReadJSON and
// decodes it with DecodeContents; check nothing is lost on the way.
func TestWorldMapSurvivesWireDecoding(t *testing.T) {
	wm := GenerateWorldMap()
	msg := resources.NewConnectResponseMessage(resources.ConnectResponseContents{WorldMap: wm})
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var received resources.Message
	if err := json.Unmarshal(b, &received); err != nil {
		t.Fatal(err)
	}
	var got resources.WorldMap
	if err := resources.DecodeContents(received.Contents.(map[string]interface{})["world"], &got); err != nil {
		t.Fatal(err)
	}
	if got.Width != wm.Width || len(got.Layers) != 2 || len(got.Objects) != len(wm.Objects) {
		t.Fatalf("decoded map lost data: width %d, %d layers, %d objects", got.Width, len(got.Layers), len(got.Objects))
	}
	if got.Objects[0] != wm.Objects[0] {
		t.Errorf("object mismatch: %+v vs %+v", got.Objects[0], wm.Objects[0])
	}
}
