package resources

import "encoding/json"

// Tile indices are Tiled-style gids into environments.png: row*16 + col + 1.
// 0 means "no tile" and is only used in the object layer.
const (
	TileNone             = 0
	TileGrass            = 71
	TileSand             = 190
	TileWaterShallow     = 189
	TileWaterDeep        = 205
	TileTree1            = 74
	TileTree2            = 78
	TileIronRock         = 90
	TileIronRockDepleted = 107
	TileBank             = 92
	TileBlacksmith       = 93
	TileAnvil            = 175
	TileFletcher         = 125
	TileRiver            = 115 // rivers and lakes; lighter than the sea
	TileRoad             = 191 // dirt road
	TileBridge           = 13  // wooden planks over a river
	TilePlaza            = 108 // cobbled village square
	TilePine             = 80
	TileAutumnTree       = 75
	TileFelledTree       = 76
	TileBush             = 79
	TileCottage          = 95
	TileTower            = 94
)

const (
	TileSize    = 8
	LayerGround = 0 // opaque terrain; water blocks movement
	LayerObject = 1 // trees, buildings, rocks; anything here blocks movement
	LayerDecor  = 2 // flowers, pebbles and the like; drawn, never blocks
)

type ObjectKind string

const (
	ObjectIronRock   ObjectKind = "ironRock"
	ObjectBank       ObjectKind = "bank"
	ObjectBlacksmith ObjectKind = "blacksmith"
	ObjectAnvil      ObjectKind = "anvil"
	ObjectFletcher   ObjectKind = "fletcher"
)

// ObjectTiles maps each interactive object kind to its sprite.
var ObjectTiles = map[ObjectKind]int{
	ObjectIronRock:   TileIronRock,
	ObjectBank:       TileBank,
	ObjectBlacksmith: TileBlacksmith,
	ObjectAnvil:      TileAnvil,
	ObjectFletcher:   TileFletcher,
}

// WorldObject is something on the object layer a player can interact with.
type WorldObject struct {
	ID   uint16     `json:"id"`
	Kind ObjectKind `json:"kind"`
	X    int        `json:"x"`
	Y    int        `json:"y"`
}

// IsPassable reports whether a ground tile can be walked on. All water
// blocks: sea, shallows, rivers and lakes.
func IsPassable(tileIndex int) bool {
	switch tileIndex {
	case TileWaterShallow, TileWaterDeep, TileRiver:
		return false
	}
	return true
}

// IsTree reports whether an object tile is a choppable kind of tree.
func IsTree(gid int) bool {
	switch gid {
	case TileTree1, TileTree2, TilePine, TileAutumnTree:
		return true
	}
	return false
}

// IsRoad reports whether a ground tile is a road, bridge or village square,
// which walkers prefer.
func IsRoad(gid int) bool {
	return gid == TileRoad || gid == TileBridge || gid == TilePlaza
}

type WorldMap struct {
	Compressionlevel int     `json:"compressionlevel"`
	Height           int     `json:"height"`
	Infinite         bool    `json:"infinite"`
	Layers           []Layer `json:"layers"`
	Nextlayerid      int     `json:"nextlayerid"`
	Nextobjectid     int     `json:"nextobjectid"`
	Orientation      string  `json:"orientation"`
	Renderorder      string  `json:"renderorder"`
	Tiledversion     string  `json:"tiledversion"`
	Tileheight       int     `json:"tileheight"`
	Tilesets         []struct {
		Firstgid int    `json:"firstgid"`
		Source   string `json:"source"`
	} `json:"tilesets"`
	Tilewidth int           `json:"tilewidth"`
	Type      string        `json:"type"`
	Version   string        `json:"version"`
	Width     int           `json:"width"`
	Objects   []WorldObject `json:"objects"`
	Places    []Place       `json:"places"`
	SpawnX    int           `json:"spawnX"`
	SpawnY    int           `json:"spawnY"`
}

type PlaceKind string

const (
	PlaceVillage PlaceKind = "village"
	PlaceMine    PlaceKind = "mine"
	PlaceForest  PlaceKind = "forest"
)

// Place is a named area of the island, such as a village or a mine.
type Place struct {
	Name   string    `json:"name"`
	Kind   PlaceKind `json:"kind"`
	X      int       `json:"x"`
	Y      int       `json:"y"`
	Radius int       `json:"radius"` // tiles from the centre that count as inside
}

// Clone returns a deep copy of the map's layers and objects, safe to send
// while the original keeps changing.
func (wm WorldMap) Clone() WorldMap {
	layers := make([]Layer, len(wm.Layers))
	for i, l := range wm.Layers {
		l.Data = append([]int(nil), l.Data...)
		layers[i] = l
	}
	wm.Layers = layers
	wm.Objects = append([]WorldObject(nil), wm.Objects...)
	wm.Places = append([]Place(nil), wm.Places...)
	return wm
}

// TilePassable reports whether tile (tx, ty) is inside the map and walkable on every layer.
func (wm *WorldMap) TilePassable(tx, ty int) bool {
	if tx < 0 || ty < 0 || tx >= wm.Width || ty >= wm.Height {
		return false
	}
	i := ty*wm.Width + tx
	for l, layer := range wm.Layers {
		if i >= len(layer.Data) {
			continue
		}
		t := layer.Data[i]
		if l == LayerGround && !IsPassable(t) {
			return false
		}
		if l == LayerObject && t != TileNone {
			return false
		}
	}
	return true
}

// BoxPassable reports whether an 8x8 sprite with its top-left corner at pixel (px, py) fits.
func (wm *WorldMap) BoxPassable(px, py float64) bool {
	if len(wm.Layers) == 0 {
		return true
	}
	for _, corner := range [4][2]float64{{px, py}, {px + TileSize - 1, py}, {px, py + TileSize - 1}, {px + TileSize - 1, py + TileSize - 1}} {
		if !wm.TilePassable(int(corner[0]/TileSize), int(corner[1]/TileSize)) {
			return false
		}
	}
	return true
}

type Layer struct {
	Data    []int  `json:"data"`
	Height  int    `json:"height"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Opacity int    `json:"opacity"`
	Type    string `json:"type"`
	Visible bool   `json:"visible"`
	Width   int    `json:"width"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

// DecodeContents converts a message's loosely-typed Contents (as produced by
// websocket ReadJSON into interface{}) into the typed struct v.
func DecodeContents(contents interface{}, v interface{}) error {
	b, err := json.Marshal(contents)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
