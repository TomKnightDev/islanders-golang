package main

import (
	"container/heap"
	"math"
	"math/rand"

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

// Tiles the generator paints, beyond those in resources.
const (
	tileRiver       = 115 // fresh water, lighter than the sea
	tileRoad        = 191 // dirt road
	tileBridge      = 13  // wooden planks
	tilePlaza       = 108 // cobblestone village square
	tileForestFloor = 73
	tileStones      = 118 // stony tufts, drawn over grass at mines
	tileAutumnTree  = 75
	tilePine        = 80
	tileBush        = 79
	tileCottage     = 95
	tileTower       = 94
	tileFlowersW    = 111
	tileFlowersP    = 112
	tilePebbles     = 119
	tileMushroom    = 109
	tileRubble      = 123
)

var (
	villageNames = []string{"Oakhaven", "Saltmere", "Brindlewick", "Fernhollow", "Gullcrest"}
	mineNames    = []string{"Ironridge Mine", "Old Quarry", "Deepdelve Mine", "Greystone Pit"}
)

const (
	villageCount   = 4
	mineCount      = 3
	riverCount     = 3
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
	g.buildRoads(villages, mines)
	g.growPlants()
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
			g.ground[i] = tileRiver
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
	for i, s := range sites {
		g.buildMine(s, mineNames[i%len(mineNames)])
	}
	return sites
}

func (g *worldGen) addObject(t tile, kind resources.ObjectKind) {
	g.objects[g.at(t)] = resources.ObjectTiles[kind]
	g.objs = append(g.objs, resources.WorldObject{ID: uint16(len(g.objs) + 1), Kind: kind, X: t.x, Y: t.y})
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
//	C . B . T . S A C      B bank, S blacksmith, A anvil
//	# # # # # # # # #      C cottage, T tower (first village only)
//	# # # # # # # # #      # cobblestone square
//	# # # # # # # # #
//	# # # # # # # # #
//	# # # # # # # # #
//	. C . . C . . C .
func (g *worldGen) buildVillage(c tile, name string, capital bool) {
	g.reserve(c, 7)
	for y := c.y - 2; y <= c.y+2; y++ {
		for x := c.x - 4; x <= c.x+4; x++ {
			g.ground[g.at(tile{x, y})] = tilePlaza
		}
	}
	north := c.y - 3
	g.addObject(tile{c.x - 2, north}, resources.ObjectBank)
	g.addObject(tile{c.x + 2, north}, resources.ObjectBlacksmith)
	g.addObject(tile{c.x + 3, north}, resources.ObjectAnvil)
	g.objects[g.at(tile{c.x - 4, north})] = tileCottage
	g.objects[g.at(tile{c.x + 4, north})] = tileCottage
	if capital {
		g.objects[g.at(tile{c.x, north})] = tileTower
	}
	for _, dx := range []int{-3, 0, 3} {
		g.objects[g.at(tile{c.x + dx, c.y + 3})] = tileCottage
	}
	g.places = append(g.places, resources.Place{Name: name, Kind: resources.PlaceVillage, X: c.x, Y: c.y, Radius: 6})
}

func (g *worldGen) buildMine(c tile, name string) {
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
	for _, off := range mineRockOffsets {
		t := c.add(off)
		if g.land(t) && !g.river[g.at(t)] {
			g.decor[g.at(t)] = resources.TileNone
			g.addObject(t, resources.ObjectIronRock)
		}
	}
	g.places = append(g.places, resources.Place{Name: name, Kind: resources.PlaceMine, X: c.x, Y: c.y, Radius: 5})
}

// --- roads ---

// buildRoads joins the villages with a minimum spanning tree of roads and
// gives every mine a road to its nearest village.
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
	case !g.land(t), g.objects[i] != resources.TileNone && !isTree(g.objects[i]):
		return -1
	case g.ground[i] == tileRoad, g.ground[i] == tileBridge, g.ground[i] == tilePlaza:
		return 0.5
	case g.river[i]:
		return 12
	case g.ground[i] == resources.TileSand:
		return 3
	default:
		return 2
	}
}

func isTree(gid int) bool {
	switch gid {
	case resources.TileTree1, resources.TileTree2, tileAutumnTree, tilePine, tileBush:
		return true
	}
	return false
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
	if isTree(g.objects[i]) {
		g.objects[i] = resources.TileNone
	}
	switch {
	case g.ground[i] == tilePlaza:
	case g.river[i]:
		g.ground[i] = tileBridge
		g.river[i] = false
	default:
		g.ground[i] = tileRoad
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
	greens := []int{resources.TileTree1, resources.TileTree2, tilePine}
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
						g.objects[i] = tileAutumnTree
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
					g.objects[i] = tileBush
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
