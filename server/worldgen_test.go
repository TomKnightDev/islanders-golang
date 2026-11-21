package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/tomknightdev/islanders-golang/resources"
)

var testIsland = GenerateWorldMap()

func countKind(wm resources.WorldMap, kind resources.ObjectKind) int {
	n := 0
	for _, o := range wm.Objects {
		if o.Kind == kind {
			n++
		}
	}
	return n
}

func TestIslandIsTheSameEveryTime(t *testing.T) {
	again := GenerateWorldMap()
	if !reflect.DeepEqual(testIsland, again) {
		t.Fatal("the island must be identical on every start so shared locations stay valid")
	}
}

func TestIslandHasVillagesAndMines(t *testing.T) {
	villages, mines := 0, 0
	for _, p := range testIsland.Places {
		switch p.Kind {
		case resources.PlaceVillage:
			villages++
		case resources.PlaceMine:
			mines++
		}
	}
	if villages != villageCount || mines != mineCount {
		t.Errorf("%d villages and %d mines, want %d and %d", villages, mines, villageCount, mineCount)
	}
	if got := countKind(testIsland, resources.ObjectBank); got != villageCount {
		t.Errorf("%d banks, want one per village", got)
	}
	if got := countKind(testIsland, resources.ObjectIronRock); got < mineCount*4 {
		t.Errorf("only %d iron rocks", got)
	}
}

func TestEverythingIsReachableFromSpawn(t *testing.T) {
	wm := testIsland
	spawn := tile{wm.SpawnX, wm.SpawnY}
	if !wm.TilePassable(spawn.x, spawn.y) {
		t.Fatalf("spawn %v is blocked", spawn)
	}
	reach := reachable(&wm, spawn)
	for _, o := range wm.Objects {
		ok := false
		for _, d := range neighbours {
			if reach[tile{o.X, o.Y}.add(d)] {
				ok = true
			}
		}
		if !ok {
			t.Errorf("%s at (%d,%d) can't be reached from spawn", o.Kind, o.X, o.Y)
		}
	}
	for _, p := range wm.Places {
		if p.Kind != resources.PlaceForest && !reach[tile{p.X, p.Y}] {
			t.Errorf("can't walk to the middle of %s", p.Name)
		}
	}
}

func TestForestsHaveTreesToChop(t *testing.T) {
	w := NewWorld(testIsland)
	reach := reachable(&w.wm, tile{w.wm.SpawnX, w.wm.SpawnY})
	forests := 0
	for _, pl := range w.wm.Places {
		if pl.Kind != resources.PlaceForest {
			continue
		}
		forests++
		choppable := 0
		for _, n := range w.nodes {
			if n.place != pl.Name {
				continue
			}
			for _, d := range neighbours {
				if reach[n.t.add(d)] {
					choppable++
					break
				}
			}
		}
		if choppable < 10 {
			t.Errorf("%s has only %d trees you can walk up to", pl.Name, choppable)
		}
	}
	if forests != forestCount {
		t.Errorf("%d named forests, want %d", forests, forestCount)
	}
	if got := countKind(testIsland, resources.ObjectFletcher); got != villageCount {
		t.Errorf("%d fletcher's huts, want one per village", got)
	}
}

func TestIslandShape(t *testing.T) {
	wm := testIsland
	ground := wm.Layers[resources.LayerGround].Data
	// Open sea all round the edge.
	for x := 0; x < wm.Width; x++ {
		for _, y := range []int{0, wm.Height - 1} {
			if g := ground[y*wm.Width+x]; g != resources.TileWaterDeep {
				t.Fatalf("edge tile (%d,%d) is %d, want deep water", x, y, g)
			}
		}
	}
	counts := map[int]int{}
	for _, g := range ground {
		counts[g]++
	}
	land := len(ground) - counts[resources.TileWaterDeep] - counts[resources.TileWaterShallow]
	if share := float64(land) / float64(len(ground)); share < 0.3 || share > 0.7 {
		t.Errorf("land is %.0f%% of the map", share*100)
	}
	for _, want := range []struct {
		gid  int
		what string
	}{
		{resources.TileRiver, "rivers"}, {resources.TileBridge, "bridges"}, {resources.TileRoad, "roads"},
		{resources.TilePlaza, "village squares"}, {resources.TileSand, "beaches"}, {tileForestFloor, "forests"},
	} {
		if counts[want.gid] == 0 {
			t.Errorf("no %s", want.what)
		}
	}
}

func TestTreesAreOnObjectLayerAndDecorNeverBlocks(t *testing.T) {
	wm := testIsland
	trees := 0
	for _, gid := range wm.Layers[resources.LayerObject].Data {
		if isPlant(gid) {
			trees++
		}
	}
	if trees < 500 {
		t.Errorf("only %d trees", trees)
	}
	decor := wm.Layers[resources.LayerDecor].Data
	for i, gid := range decor {
		if gid == resources.TileNone {
			continue
		}
		x, y := i%wm.Width, i/wm.Width
		if wm.Layers[resources.LayerObject].Data[i] == resources.TileNone && !wm.TilePassable(x, y) &&
			wm.Layers[resources.LayerGround].Data[i] == resources.TileGrass {
			t.Fatalf("decoration at (%d,%d) blocks movement", x, y)
		}
	}
}

func TestBoxPassableEdges(t *testing.T) {
	wm := testIsland
	if wm.BoxPassable(-1, 400) || wm.BoxPassable(400, float64(wm.Height*resources.TileSize-7)) {
		t.Error("boxes leaving the map should be blocked")
	}
	if wm.BoxPassable(0, 0) {
		t.Error("deep water corner should be blocked")
	}
}

// The client receives the map as interface{} from websocket ReadJSON and
// decodes it with DecodeContents; check nothing is lost on the way.
func TestWorldMapSurvivesWireDecoding(t *testing.T) {
	wm := testIsland
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
	if !reflect.DeepEqual(got.Layers[2].Data, wm.Layers[2].Data) || !reflect.DeepEqual(got.Places, wm.Places) ||
		got.SpawnX != wm.SpawnX || len(got.Objects) != len(wm.Objects) {
		t.Fatal("decoded map lost data")
	}
}
