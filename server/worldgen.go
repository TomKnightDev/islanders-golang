package main

import (
	"container/heap"
	"math"
	"math/rand"
	"sort"

	"github.com/tomknightdev/islanders-golang/resources"
)

const (
	mapWidth  = 160
	mapHeight = 160

	// The island is generated from a fixed seed so it's the same after every
	// restart and deploy, and shared locations keep meaning the same place.
	worldSeed = 116
)

// Elevation bands. Elevation is noise shaped into an island: high in the
// middle, falling to sea at the edges.
const (
	seaLevel     = 0.34 // below: deep water
	shallowLevel = 0.37 // below: shallow water
	beachLevel   = 0.40 // below: sand
	highland     = 0.62 // rivers start above this
)

// Tiles only the generator paints. Shared ones live in resources.
const (
	tileForestFloor = 73
	tileStones      = 118 // stony tufts, drawn over grass at mines
	tileFlowersW    = 111
	tileFlowersP    = 112
	tilePebbles     = 119
	tileMushroom    = 109
	tileRubble      = 123
)

var (
	villageNames = []string{"Oakhaven", "Saltmere", "Brindlewick", "Fernhollow", "Gullcrest"}
	mineNames    = []string{"Ironridge Mine", "Old Quarry", "Deepdelve Mine", "Greystone Pit"}
	forestNames  = []string{"Whisperwood", "Elderglade", "Mossy Hollow", "Thornwood", "Amberwood"}
	farmNames    = []string{"Meadowbrook Farm", "Willowbank Farm", "Hillcrest Farm", "Sunfield Farm"}
)

// mineOres gives each mine's rocks, matched to mineRockOffsets. Mines are
// ranked by distance from the spawn, so the nearest has the easiest ore and
// the farthest the hardest.
var mineOres = [][]int{
	{resources.TileCopperRock, resources.TileCopperRock, resources.TileIronRock, resources.TileCopperRock, resources.TileCopperRock, resources.TileIronRock},
	{resources.TileIronRock, resources.TileIronRock, resources.TileGoldRock, resources.TileIronRock, resources.TileGoldRock, resources.TileIronRock},
	{resources.TileGoldRock, resources.TileMithrilRock, resources.TileGoldRock, resources.TileMithrilRock, resources.TileGoldRock, resources.TileMithrilRock},
}

const (
	villageCount   = 4
	mineCount      = 3
	riverCount     = 3
	farmCount      = 3
	forestCount    = 4  // the largest woods get a name and can be chopped
	minForestSize  = 60 // tiles of forest floor before a wood can be named
	villageSpacing = 34 // tiles between village centres
	mineSpacing    = 26
)

// rock positions in a mine, relative to its centre. Spaced so each rock has
// free tiles around it.
var mineRockOffsets = []tile{{-2, -2}, {1, -3}, {3, 0}, {-3, 1}, {0, 2}, {2, 3}}

type worldGen struct {
	r       *rand.Rand
	w, h    int
	elev    []float64
	ground  []int
	objects []int
	decor   []int
	river   []bool
	keepOut []bool // villages, mines and roads: no trees or decoration
	objs    []resources.WorldObject
	places  []resources.Place
	spawn   tile
}

func GenerateWorldMap() resources.WorldMap {
	return generateWorld(worldSeed)
}

func generateWorld(seed int64) resources.WorldMap {
	n := mapWidth * mapHeight
	g := &worldGen{
		r:       rand.New(rand.NewSource(seed)),
		w:       mapWidth,
		h:       mapHeight,
		elev:    make([]float64, n),
		ground:  make([]int, n),
		objects: make([]int, n),
		decor:   make([]int, n),
		river:   make([]bool, n),
		keepOut: make([]bool, n),
	}
	g.shapeIsland()
	g.carveRivers()
	g.paintTerrain()
	mainland := g.mainland()
	villages := g.placeVillages(mainland)
	mines := g.placeMines(mainland, villages)
	farms := g.placeFarms(mainland, villages, mines)
	g.buildRoads(villages, append(mines, farms...))
	g.growPlants()
	g.nameForests()
	g.ensureReachable()

	return resources.WorldMap{
		Width:  g.w,
		Height: g.h,
		Layers: []resources.Layer{
			newLayer(1, "Ground", g.ground),
			newLayer(2, "Objects", g.objects),
			newLayer(3, "Decor", g.decor),
		},
		Objects: g.objs,
		Places:  g.places,
		SpawnX:  g.spawn.x,
		SpawnY:  g.spawn.y,
	}
}

func newLayer(id int, name string, data []int) resources.Layer {
	return resources.Layer{
		Data:    data,
		Height:  mapHeight,
		ID:      id,
		Name:    name,
		Opacity: 1,
		Type:    "tilelayer",
		Visible: true,
		Width:   mapWidth,
	}
}

func (g *worldGen) at(t tile) int      { return t.y*g.w + t.x }
func (g *worldGen) inside(t tile) bool { return t.x >= 0 && t.y >= 0 && t.x < g.w && t.y < g.h }
func (g *worldGen) land(t tile) bool   { return g.elev[g.at(t)] >= shallowLevel }

// --- terrain ---

func (g *worldGen) shapeIsland() {
	height := newValueNoise(g.r)
	cx, cy := float64(g.w)/2, float64(g.h)/2
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			dx, dy := (float64(x)-cx)/cx, (float64(y)-cy)/cy
			d := math.Sqrt(dx*dx + dy*dy)
			e := 0.55*height.fbm(float64(x), float64(y), 42, 4) + 0.45 - 0.62*d*d
			if d > 0.95 {
				e = 0 // always open sea at the map edge
			}
			g.elev[y*g.w+x] = e
		}
	}
}

// carveRivers runs rivers from the highlands downhill to the sea. They're
// one tile wide up high and two wide in their lower reaches.
func (g *worldGen) carveRivers() {
	meander := newValueNoise(g.r)
	var sources []tile
	for tries := 0; len(sources) < riverCount && tries < 5000; tries++ {
		t := tile{g.r.Intn(g.w), g.r.Intn(g.h)}
		if g.elev[g.at(t)] < highland || nearAny(sources, t, 30) {
			continue
		}
		sources = append(sources, t)
	}
	for _, src := range sources {
		widen := steps[g.r.Intn(4)] // which side the lower river spreads to
		visited := map[tile]bool{}
		cur := src
		for i := 0; i < 600; i++ {
			visited[cur] = true
			g.river[g.at(cur)] = true
			if g.elev[g.at(cur)] < 0.56 {
				if w := cur.add(widen); g.inside(w) && g.land(w) {
					g.river[g.at(w)] = true
				}
			}
			if g.elev[g.at(cur)] < shallowLevel {
				break // reached the sea
			}
			var next tile
			best := math.Inf(1)
			for _, d := range neighbours {
				nb := cur.add(d)
				if !g.inside(nb) || visited[nb] {
					continue
				}
				// Noise on top of the slope makes rivers wander instead of
				// running straight to the coast.
				if e := g.elev[g.at(nb)] + meander.fbm(float64(nb.x), float64(nb.y), 6, 2)*0.06; e < best {
					best, next = e, nb
				}
			}
			if math.IsInf(best, 1) {
				break
			}
			cur = next
		}
	}
}

func (g *worldGen) paintTerrain() {
	for i, e := range g.elev {
		switch {
		case e < seaLevel:
			g.ground[i] = resources.TileWaterDeep
		case e < shallowLevel:
			g.ground[i] = resources.TileWaterShallow
		case g.river[i]:
			g.ground[i] = resources.TileRiver
		case e < beachLevel:
			g.ground[i] = resources.TileSand
		default:
			g.ground[i] = resources.TileGrass
		}
	}
}

// mainland marks the land connected to the middle of the island. Villages
// and mines only go here, so roads can reach all of them.
func (g *worldGen) mainland() []bool {
	main := make([]bool, len(g.elev))
	start := tile{g.w / 2, g.h / 2}
	queue := []tile{start}
	main[g.at(start)] = true
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range neighbours {
			nb := cur.add(d)
			if g.inside(nb) && !main[g.at(nb)] && g.land(nb) {
				main[g.at(nb)] = true
				queue = append(queue, nb)
			}
		}
	}
	return main
}

// waterDistance is each tile's distance in tiles to the nearest sea or river.
func (g *worldGen) waterDistance() []int {
	dist := make([]int, len(g.elev))
	var queue []tile
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			t := tile{x, y}
			if !g.land(t) || g.river[g.at(t)] {
				queue = append(queue, t)
			} else {
				dist[g.at(t)] = -1
			}
		}
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range neighbours {
			nb := cur.add(d)
			if g.inside(nb) && dist[g.at(nb)] == -1 {
				dist[g.at(nb)] = dist[g.at(cur)] + 1
				queue = append(queue, nb)
			}
		}
	}
	return dist
}

// --- villages and mines ---

func nearAny(points []tile, t tile, within int) bool {
	for _, p := range points {
		if max(abs(p.x-t.x), abs(p.y-t.y)) < within {
			return true
		}
	}
	return false
}

// pickSites chooses up to n spread-out tiles from the candidates, starting
// with first if it's valid.
func (g *worldGen) pickSites(candidates []tile, n, spacing int, avoid []tile, avoidSpacing int) []tile {
	g.r.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
	var sites []tile
	for _, c := range candidates {
		if len(sites) == n {
			break
		}
		if nearAny(sites, c, spacing) || nearAny(avoid, c, avoidSpacing) {
			continue
		}
		sites = append(sites, c)
	}
	return sites
}

func (g *worldGen) placeVillages(mainland []bool) []tile {
	water := g.waterDistance()
	var candidates []tile
	centre := tile{g.w / 2, g.h / 2}
	var first tile
	firstDist := math.MaxInt
	for y := 8; y < g.h-8; y += 2 {
		for x := 8; x < g.w-8; x += 2 {
			t := tile{x, y}
			if !mainland[g.at(t)] || water[g.at(t)] < 9 {
				continue
			}
			candidates = append(candidates, t)
			if d := octile(t, centre); d < firstDist {
				first, firstDist = t, d
			}
		}
	}
	if len(candidates) == 0 {
		panic("worldgen: no room for a village")
	}
	// The village nearest the middle is where everyone starts.
	sites := append([]tile{first}, g.pickSites(candidates, villageCount-1, villageSpacing, []tile{first}, villageSpacing)...)
	for i, s := range sites {
		g.buildVillage(s, villageNames[i%len(villageNames)], i == 0)
	}
	g.spawn = tile{sites[0].x, sites[0].y + 1}
	return sites
}

func (g *worldGen) placeMines(mainland []bool, villages []tile) []tile {
	water := g.waterDistance()
	var candidates []tile
	for y := 8; y < g.h-8; y += 2 {
		for x := 8; x < g.w-8; x += 2 {
			t := tile{x, y}
			if mainland[g.at(t)] && water[g.at(t)] >= 6 && !g.keepOut[g.at(t)] && g.elev[g.at(t)] > 0.5 {
				candidates = append(candidates, t)
			}
		}
	}
	sites := g.pickSites(candidates, mineCount, mineSpacing, villages, 22)
	sort.SliceStable(sites, func(i, j int) bool { return octile(sites[i], villages[0]) < octile(sites[j], villages[0]) })
	for i, s := range sites {
		g.buildMine(s, mineNames[i%len(mineNames)], mineOres[min(i, len(mineOres)-1)])
	}
	return sites
}

// placeFarms puts farms on open lowland a short walk from the villages.
func (g *worldGen) placeFarms(mainland []bool, villages, mines []tile) []tile {
	water := g.waterDistance()
	var candidates []tile
	for y := 8; y < g.h-8; y += 2 {
		for x := 8; x < g.w-8; x += 2 {
			t := tile{x, y}
			if !mainland[g.at(t)] || water[g.at(t)] < 7 || g.keepOut[g.at(t)] || g.elev[g.at(t)] > 0.62 {
				continue
			}
			near := math.MaxInt
			for _, v := range villages {
				near = min(near, max(abs(v.x-x), abs(v.y-y)))
			}
			if near >= 12 && near <= 22 {
				candidates = append(candidates, t)
			}
		}
	}
	sites := g.pickSites(candidates, farmCount, 22, append(append([]tile(nil), villages...), mines...), 12)
	for i, s := range sites {
		g.buildFarm(s, farmNames[i%len(farmNames)])
	}
	return sites
}

func (g *worldGen) addObject(t tile, kind resources.ObjectKind) {
	g.addObjectTile(t, kind, resources.ObjectTiles[kind])
}

func (g *worldGen) addObjectTile(t tile, kind resources.ObjectKind, gid int) {
	g.objects[g.at(t)] = gid
	g.objs = append(g.objs, resources.WorldObject{ID: uint16(len(g.objs) + 1), Kind: kind, X: t.x, Y: t.y})
}

// materialsOf lists the gatherable material IDs among the given sprites, in
// tier order.
func materialsOf(kind resources.ActionKind, gids map[int]bool) []string {
	var ids []string
	for _, m := range resources.ActionMaterials[kind] {
		if gids[m.Node] {
			ids = append(ids, m.ID)
		}
	}
	return ids
}

func (g *worldGen) reserve(c tile, radius int) {
	for y := c.y - radius; y <= c.y+radius; y++ {
		for x := c.x - radius; x <= c.x+radius; x++ {
			if t := (tile{x, y}); g.inside(t) {
				g.keepOut[g.at(t)] = true
			}
		}
	}
}

// buildVillage lays out a cobbled square with the bank and blacksmith on its
// north side and houses around it.
//
//	C . B . T . S A C      B bank, S blacksmith, A anvil, F fletcher
//	# # # # # # # # #      C cottage, T tower (first village only)
//	# # # # # # # # #      # cobblestone square
//	# # # # # # # # #
//	# # # # # # # # #
//	# # # # # # # # #
//	. C . . F . . C .
func (g *worldGen) buildVillage(c tile, name string, capital bool) {
	g.reserve(c, 7)
	for y := c.y - 2; y <= c.y+2; y++ {
		for x := c.x - 4; x <= c.x+4; x++ {
			g.ground[g.at(tile{x, y})] = resources.TilePlaza
		}
	}
	north := c.y - 3
	g.addObject(tile{c.x - 2, north}, resources.ObjectBank)
	g.addObject(tile{c.x + 2, north}, resources.ObjectBlacksmith)
	g.addObject(tile{c.x + 3, north}, resources.ObjectAnvil)
	g.objects[g.at(tile{c.x - 4, north})] = resources.TileCottage
	g.objects[g.at(tile{c.x + 4, north})] = resources.TileCottage
	if capital {
		g.objects[g.at(tile{c.x, north})] = resources.TileTower
	}
	g.objects[g.at(tile{c.x - 3, c.y + 3})] = resources.TileCottage
	g.addObject(tile{c.x, c.y + 3}, resources.ObjectFletcher)
	g.objects[g.at(tile{c.x + 3, c.y + 3})] = resources.TileCottage
	g.places = append(g.places, resources.Place{Name: name, Kind: resources.PlaceVillage, X: c.x, Y: c.y, Radius: 6})
}

func (g *worldGen) buildMine(c tile, name string, ores []int) {
	g.reserve(c, 5)
	for y := c.y - 4; y <= c.y+4; y++ {
		for x := c.x - 4; x <= c.x+4; x++ {
			t := tile{x, y}
			dx, dy := x-c.x, y-c.y
			if dx*dx+dy*dy > 18 || !g.land(t) || g.river[g.at(t)] {
				continue
			}
			if g.r.Float64() < 0.2 {
				g.decor[g.at(t)] = tileRubble
			} else {
				g.decor[g.at(t)] = tileStones
			}
		}
	}
	present := map[int]bool{}
	for i, off := range mineRockOffsets {
		t := c.add(off)
		if g.land(t) && !g.river[g.at(t)] {
			g.decor[g.at(t)] = resources.TileNone
			g.addObjectTile(t, resources.ObjectRock, ores[i])
			present[ores[i]] = true
		}
	}
	g.places = append(g.places, resources.Place{
		Name: name, Kind: resources.PlaceMine, X: c.x, Y: c.y, Radius: 5,
		Materials: materialsOf(resources.ActionMine, present),
	})
}

// buildFarm lays out tilled fields with rows of crops, each row next to a
// walkable row so every plant can be reached.
//
//	. . . . . . . . . . .      : tilled soil (walkable)
//	c c c c c c c c c . .      c cabbages, w cotton
//	: : : : : : : : : : :
//	w w w w w w w w w : :
//	: : : : : : : : : : :
//	c c c c c c c c c : :
//	: : : : : : : : : : :
func (g *worldGen) buildFarm(c tile, name string) {
	g.reserve(c, 6)
	for y := c.y - 3; y <= c.y+3; y++ {
		for x := c.x - 5; x <= c.x+5; x++ {
			g.ground[g.at(tile{x, y})] = resources.TileField
		}
	}
	rows := map[int]int{-2: resources.TileCabbage, 0: resources.TileCotton, 2: resources.TileCabbage}
	for dy, crop := range rows {
		for x := c.x - 4; x <= c.x+3; x++ {
			g.objects[g.at(tile{x, c.y + dy})] = crop
		}
	}
	g.places = append(g.places, resources.Place{
		Name: name, Kind: resources.PlaceFarm, X: c.x, Y: c.y, Radius: 5,
		Materials: materialsOf(resources.ActionFarm, map[int]bool{resources.TileCabbage: true, resources.TileCotton: true}),
	})
}

// --- roads ---

// buildRoads joins the villages with a minimum spanning tree of roads and
// gives every outpost (mine, farm) a road to its nearest village.
func (g *worldGen) buildRoads(villages, mines []tile) {
	joined := []tile{villages[0]}
	rest := append([]tile(nil), villages[1:]...)
	for len(rest) > 0 {
		bi, bj, best := 0, 0, math.MaxInt
		for i, a := range joined {
			for j, b := range rest {
				if d := octile(a, b); d < best {
					bi, bj, best = i, j, d
				}
			}
		}
		g.road(joined[bi], rest[bj])
		joined = append(joined, rest[bj])
		rest = append(rest[:bj], rest[bj+1:]...)
	}
	for _, m := range mines {
		nearest := villages[0]
		for _, v := range villages[1:] {
			if octile(m, v) < octile(m, nearest) {
				nearest = v
			}
		}
		g.road(nearest, m)
	}
}

// roadCost is the cost of laying road on t, or -1 where roads can't go.
// Existing roads are cheap so roads merge; rivers are allowed but costly,
// and become bridges.
func (g *worldGen) roadCost(t tile) float64 {
	i := g.at(t)
	switch {
	case !g.land(t), g.objects[i] != resources.TileNone && !isPlant(g.objects[i]):
		return -1
	case g.ground[i] == resources.TileField && g.objects[i] != resources.TileNone:
		return -1 // crops aren't cleared like wild plants
	case g.ground[i] == resources.TileRoad, g.ground[i] == resources.TileBridge, g.ground[i] == resources.TilePlaza:
		return 0.5
	case g.river[i]:
		return 12
	case g.ground[i] == resources.TileSand:
		return 3
	default:
		return 2
	}
}

// isPlant is anything roads may clear: trees and bushes.
func isPlant(gid int) bool {
	return resources.IsTree(gid) || gid == resources.TileBush
}

// road finds the cheapest route from a to b and paints it: dirt on land,
// planks over rivers, clearing any trees in the way.
func (g *worldGen) road(a, b tile) {
	cost := map[tile]float64{a: 0}
	from := map[tile]tile{}
	open := &costHeap{{t: a}}
	for open.Len() > 0 {
		cur := heap.Pop(open).(costNode)
		if cur.t == b {
			break
		}
		if cur.g > cost[cur.t] {
			continue
		}
		for _, d := range neighbours {
			nb := cur.t.add(d)
			if !g.inside(nb) {
				continue
			}
			c := g.roadCost(nb)
			if c < 0 && nb != b {
				continue
			}
			ng := cur.g + max(c, 0.5)
			if old, ok := cost[nb]; ok && ng >= old {
				continue
			}
			cost[nb] = ng
			from[nb] = cur.t
			heap.Push(open, costNode{t: nb, g: ng, f: ng + 0.5*float64(abs(nb.x-b.x)+abs(nb.y-b.y))})
		}
	}
	if _, ok := from[b]; !ok {
		return
	}
	for t := b; t != a; t = from[t] {
		g.paintRoad(t)
	}
}

func (g *worldGen) paintRoad(t tile) {
	i := g.at(t)
	g.keepOut[i] = true
	g.decor[i] = resources.TileNone
	if isPlant(g.objects[i]) {
		g.objects[i] = resources.TileNone
	}
	switch {
	case g.ground[i] == resources.TilePlaza:
	case g.river[i]:
		g.ground[i] = resources.TileBridge
		g.river[i] = false
	default:
		g.ground[i] = resources.TileRoad
	}
}

type costNode struct {
	t    tile
	g, f float64
}

type costHeap []costNode

func (h costHeap) Len() int           { return len(h) }
func (h costHeap) Less(i, j int) bool { return h[i].f < h[j].f }
func (h costHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *costHeap) Push(x any)        { *h = append(*h, x.(costNode)) }
func (h *costHeap) Pop() any {
	old := *h
	n := old[len(old)-1]
	*h = old[:len(old)-1]
	return n
}

// --- plants ---

// growPlants fills forests where the forest noise is high, scatters trees
// elsewhere, and adds flowers, pebbles and mushrooms.
func (g *worldGen) growPlants() {
	forest := newValueNoise(g.r)
	autumn := newValueNoise(g.r)
	flowers := newValueNoise(g.r)
	greens := []int{resources.TileTree1, resources.TileTree2, resources.TilePine}
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			i := y*g.w + x
			if g.ground[i] != resources.TileGrass || g.keepOut[i] {
				continue
			}
			fx, fy := float64(x), float64(y)
			f := forest.fbm(fx, fy, 16, 3)
			roll := g.r.Float64()
			switch {
			case f > 0.6: // deep forest
				g.ground[i] = tileForestFloor
				if roll < 0.5 {
					if autumn.fbm(fx, fy, 24, 2) > 0.6 {
						g.objects[i] = resources.TileAutumnTree
					} else {
						g.objects[i] = greens[g.r.Intn(len(greens))]
					}
				} else if roll < 0.56 {
					g.decor[i] = tileMushroom
				}
				continue
			case f > 0.52: // forest edge
				if roll < 0.12 {
					g.objects[i] = greens[g.r.Intn(len(greens))]
					continue
				}
			default: // open meadow
				if roll < 0.012 {
					g.objects[i] = greens[g.r.Intn(len(greens))]
					continue
				}
				if roll < 0.02 {
					g.objects[i] = resources.TileBush
					continue
				}
			}
			if fl := flowers.fbm(fx, fy, 10, 2); fl > 0.66 && g.r.Float64() < 0.15 {
				if fl > 0.72 {
					g.decor[i] = tileFlowersP
				} else {
					g.decor[i] = tileFlowersW
				}
			} else if g.r.Float64() < 0.008 {
				g.decor[i] = tilePebbles
			}
		}
	}
}

// nameForests gives the largest woods a name, making them places where
// players can chop wood. Each is centred on the forest tile nearest its
// middle and sized to cover most of it.
func (g *worldGen) nameForests() {
	seen := make([]bool, len(g.ground))
	var woods [][]tile
	for i, gid := range g.ground {
		if gid != tileForestFloor || seen[i] {
			continue
		}
		start := tile{i % g.w, i / g.w}
		seen[i] = true
		wood := []tile{start}
		for q := 0; q < len(wood); q++ {
			for _, d := range neighbours {
				nb := wood[q].add(d)
				if g.inside(nb) && !seen[g.at(nb)] && g.ground[g.at(nb)] == tileForestFloor {
					seen[g.at(nb)] = true
					wood = append(wood, nb)
				}
			}
		}
		if len(wood) >= minForestSize {
			woods = append(woods, wood)
		}
	}
	sort.SliceStable(woods, func(i, j int) bool { return len(woods[i]) > len(woods[j]) })
	for n, wood := range woods[:min(forestCount, len(woods))] {
		sx, sy := 0, 0
		for _, t := range wood {
			sx, sy = sx+t.x, sy+t.y
		}
		mid := tile{sx / len(wood), sy / len(wood)}
		centre := wood[0]
		for _, t := range wood {
			if octile(t, mid) < octile(centre, mid) {
				centre = t
			}
		}
		radius := min(10, max(5, int(math.Sqrt(float64(len(wood))))/2+1))
		trees := map[int]bool{}
		for y := centre.y - radius; y <= centre.y+radius; y++ {
			for x := centre.x - radius; x <= centre.x+radius; x++ {
				if t := (tile{x, y}); g.inside(t) && resources.IsTree(g.objects[g.at(t)]) {
					trees[g.objects[g.at(t)]] = true
				}
			}
		}
		g.places = append(g.places, resources.Place{
			Name: forestNames[n%len(forestNames)], Kind: resources.PlaceForest, X: centre.x, Y: centre.y, Radius: radius,
			Materials: materialsOf(resources.ActionChop, trees),
		})
	}
}

// ensureReachable makes sure everyone can walk from the spawn to every bank,
// blacksmith, anvil and rock. Anything cut off gets a road.
func (g *worldGen) ensureReachable() {
	wm := resources.WorldMap{Width: g.w, Height: g.h, Layers: []resources.Layer{{Data: g.ground}, {Data: g.objects}}}
	for _, o := range g.objs {
		reach := reachable(&wm, g.spawn)
		ok := false
		for _, d := range neighbours {
			if reach[tile{o.X, o.Y}.add(d)] {
				ok = true
				break
			}
		}
		if !ok {
			g.road(g.spawn, tile{o.X, o.Y})
		}
	}
}

// reachable is every tile walkable from start.
func reachable(wm *resources.WorldMap, start tile) map[tile]bool {
	seen := map[tile]bool{start: true}
	queue := []tile{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range neighbours {
			nb := cur.add(d)
			if !seen[nb] && wm.TilePassable(nb.x, nb.y) {
				seen[nb] = true
				queue = append(queue, nb)
			}
		}
	}
	return seen
}

// --- noise ---

// valueNoise is smooth random noise in [0, 1]: random values on a lattice,
// blended with smoothstep.
type valueNoise struct {
	vals [64 * 64]float64
}

func newValueNoise(r *rand.Rand) *valueNoise {
	n := &valueNoise{}
	for i := range n.vals {
		n.vals[i] = r.Float64()
	}
	return n
}

func (n *valueNoise) lattice(x, y int) float64 {
	return n.vals[(y&63)*64+(x&63)]
}

func (n *valueNoise) at(x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	tx, ty := x-x0, y-y0
	sx, sy := tx*tx*(3-2*tx), ty*ty*(3-2*ty)
	ix, iy := int(x0), int(y0)
	top := n.lattice(ix, iy)*(1-sx) + n.lattice(ix+1, iy)*sx
	bottom := n.lattice(ix, iy+1)*(1-sx) + n.lattice(ix+1, iy+1)*sx
	return top*(1-sy) + bottom*sy
}

// fbm layers octaves of noise; scale is the size in tiles of the largest
// features.
func (n *valueNoise) fbm(x, y, scale float64, octaves int) float64 {
	sum, amp, norm, freq := 0.0, 1.0, 0.0, 1/scale
	for o := 0; o < octaves; o++ {
		sum += amp * n.at(x*freq+float64(o)*17.3, y*freq+float64(o)*31.7)
		norm += amp
		amp /= 2
		freq *= 2
	}
	return sum / norm
}
