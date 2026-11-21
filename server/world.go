package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

const (
	tickInterval = 25 * time.Millisecond
	walkSpeed    = 1.5 // tiles per second

	// Pixels moved per tick.
	walkStep = walkSpeed * resources.TileSize * float64(tickInterval) / float64(time.Second)

	sessionBuffer = 256
)

// session is one websocket connection. Only its writer goroutine writes to
// the connection; everything else queues messages with push.
type session struct {
	conn   *websocket.Conn
	send   chan *resources.Message
	mu     sync.Mutex
	closed bool
}

func newSession(conn *websocket.Conn) *session {
	s := &session{conn: conn, send: make(chan *resources.Message, sessionBuffer)}
	go s.writeLoop()
	return s
}

func (s *session) writeLoop() {
	for m := range s.send {
		if err := s.conn.WriteJSON(m); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
	s.conn.Close()
}

// push queues a message without blocking. A client too slow to keep up is
// disconnected rather than allowed to stall the world.
func (s *session) push(m *resources.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.send <- m:
	default:
		log.Println("Send buffer full, closing session")
		s.closed = true
		close(s.send)
	}
}

// close flushes queued messages, then closes the connection.
func (s *session) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		close(s.send)
	}
}

type player struct {
	id       uint16
	username string
	password string
	pos      f64.Vec2 // top-left pixel of the sprite
	path     []tile   // tiles still to walk through; path[0] is the next one
	sess     *session // nil while offline; the character stays in the world
	class    resources.ClassID
	sprite   f64.Vec2 // pixel position in characters.png, from the class

	backpack resources.Inventory
	bank     resources.Inventory // this player's own bank account
	job      *job
	activity string
	actions  []*queuedAction // the action list; [0] is the one that runs
	running  bool            // working through the action list
	found    map[string]bool // places this player has found
	xp       map[resources.Skill]int
	xpCarry  map[resources.Skill]int // XP fractions left over from xpRate

	sentDest *tile // the path destination last sent to the client; nil if stopped

	health       int
	stamina      int  // used up by work; recovered only at an inn
	coins        int  // in the bank account; never lost on defeat
	fightShown   *mob // the enemy the client's fight panel shows, if any
	healthDirty  bool // changed since last broadcast, for everyone's health bars
	regenCounter int
	equipment    map[resources.EquipSlot]resources.Stack
	// ticks is this character's own clock, advancing every world tick.
	// Job and action timing use it.
	ticks int
	// The current action's timing, for the client's progress bar, on the
	// character's clock. actionTicks is 0 for actions with no set length.
	actionStart int
	actionTicks int
	invDirty    bool  // backpack, bank or action changed since last sent
	visit       visit // the building the client was last told p is at
}

// currentTile is the tile the player stands on, or is walking onto.
func (p *player) currentTile() tile { return walkerTile(p.pos, p.path) }

// walkerTile is the tile a walker stands on, or is stepping onto.
func walkerTile(pos f64.Vec2, path []tile) tile {
	if len(path) > 0 {
		return path[0]
	}
	return tile{
		int(math.Round(pos[0] / resources.TileSize)),
		int(math.Round(pos[1] / resources.TileSize)),
	}
}

func (p *player) updateMessage() *resources.Message {
	return resources.NewUpdateMessage(p.id, resources.UpdateContents{
		Pos:       p.pos,
		Tile:      p.sprite,
		Username:  p.username,
		Health:    p.health,
		MaxHealth: p.maxHealth(),
	})
}

type command func(w *World)

// World owns all game state. Only its Run goroutine touches it; connection
// handlers send it commands with Do.
type World struct {
	wm         resources.WorldMap
	players    map[uint16]*player
	byName     map[string]*player
	nextID     uint16
	nodes      map[tile]*node // rocks and trees to gather from
	banks      []tile
	inns       []tile
	stores     map[string]*store                        // village name -> its general store
	stations   map[resources.ObjectKind]map[string]tile // crafting stations by kind, then village
	mobs       []*mob
	ticks      int
	nextItemID uint32 // for giving every piece of gear and tool its own ID
	rng        *rand.Rand
	dev        bool // started with -dev: dev tools such as game speed are allowed
	speed      int  // world ticks per real tick; 1 is normal
	cmds       chan command
}

func NewWorld(wm resources.WorldMap) *World {
	w := &World{
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		wm:      wm,
		players: make(map[uint16]*player),
		byName:  make(map[string]*player),
		cmds:    make(chan command, 256),
	}
	w.initNodes()
	w.initMobs()
	return w
}

func (w *World) Run() {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	for {
		select {
		case c := <-w.cmds:
			c(w)
		case <-ticker.C:
			// At game speed N the world advances N ticks for each real one.
			for i := 0; i < max(w.speed, 1); i++ {
				w.tick()
			}
		}
	}
}

// devState is what clients are told about dev tools.
func (w *World) devState() *resources.Message {
	return resources.NewDevStateMessage(resources.DevStateContents{Enabled: w.dev, Speed: max(w.speed, 1)})
}

// devCommand changes the game speed, on servers started with -dev only.
func (w *World) devCommand(c resources.DevCommandContents) {
	if !w.dev {
		return
	}
	w.speed = min(max(c.Speed, 1), resources.MaxGameSpeed)
	w.broadcast(w.devState())
	for _, p := range w.players {
		p.invDirty = true // re-send action timings at the new speed
	}
}

// Do runs c on the world goroutine.
func (w *World) Do(c command) { w.cmds <- c }

func (w *World) broadcast(m *resources.Message) {
	for _, p := range w.players {
		if p.sess != nil {
			p.sess.push(m)
		}
	}
}

// exists reports whether a character with this name has been created.
func (w *World) exists(username string) bool {
	_, ok := w.byName[username]
	return ok
}

// createCharacter makes a new character of a class. The name must be free.
func (w *World) createCharacter(username, password string, class resources.Class) *player {
	x, y := tile{w.wm.SpawnX, w.wm.SpawnY}.pixel()
	p := &player{
		id:       w.nextID,
		username: username,
		password: password,
		pos:      f64.Vec2{x, y},
		backpack: resources.NewInventory(backpackSlots, false),
		bank:     resources.NewInventory(bankSlots, true),
		activity: activityIdle,
		found:    map[string]bool{},
		xp:       map[resources.Skill]int{},
		xpCarry:  map[resources.Skill]int{},
	}
	w.giveStartingKit(p, class)
	// Everyone knows the village they start in.
	if pl := w.placeAt(tile{w.wm.SpawnX, w.wm.SpawnY}, resources.PlaceVillage); pl != nil {
		p.found[pl.Name] = true
	}
	w.nextID++
	w.players[p.id] = p
	w.byName[username] = p
	return p
}

// login connects a session to a character. A name nobody has used yet gets
// a new character of defaultClass; players choose their class first through
// the connection handler, which calls createCharacter.
func (w *World) login(username, password string, sess *session) (*player, error) {
	p, found := w.byName[username]
	if found {
		if p.password != password {
			return nil, fmt.Errorf("incorrect password")
		}
		if p.sess != nil {
			p.sess.close() // logged in somewhere else; that connection loses
		}
	} else {
		class, _ := resources.FindClass(defaultClass)
		p = w.createCharacter(username, password, class)
	}
	p.sess = sess
	p.sentDest = &tile{-1, -1} // force the current path to be sent
	p.fightShown = nil         // and the current fight
	sess.push(resources.NewConnectResponseMessage(resources.ConnectResponseContents{
		ClientId: p.id,
		Pos:      p.pos,
		Tile:     p.sprite,
		WorldMap: w.wm.Clone(),
	}))
	sess.push(w.inventoryMessage(p))
	sess.push(w.mobSnapshot())
	sess.push(w.devState())
	for _, other := range w.players {
		if other != p {
			sess.push(other.updateMessage())
		}
	}

	// Everyone else learns about the new or returning player.
	m := p.updateMessage()
	for _, other := range w.players {
		if other != p && other.sess != nil {
			other.sess.push(m)
		}
	}
	return p, nil
}

// disconnect detaches sess from p. The character stays in the world and
// keeps walking; a newer session from the same player is left alone.
func (w *World) disconnect(p *player, sess *session) {
	sess.close()
	if p.sess == sess {
		p.sess = nil
	}
}

func (w *World) chat(p *player, text string) {
	w.broadcast(resources.NewChatMessage(p.id, fmt.Sprintf("%s: %s", p.username, text)))
}

// moveTo sends p walking to goal and reports whether there is a way there.
// A player already between tiles finishes that step first, so there is no
// snapping.
func (w *World) moveTo(p *player, goal tile) bool {
	start := p.currentTile()
	path := findPath(&w.wm, start, goal)
	if path == nil {
		return false
	}
	if len(p.path) > 0 {
		p.path = append([]tile{start}, path...)
	} else {
		p.path = path
	}
	return true
}

// step moves p one tick along its path and reports whether it moved.
func (p *player) step() bool { return stepAlong(&p.pos, &p.path, walkStep) }

// stepAlong moves a walker by speed pixels along its path and reports
// whether it moved.
func stepAlong(pos *f64.Vec2, path *[]tile, speed float64) bool {
	if len(*path) == 0 {
		return false
	}
	tx, ty := (*path)[0].pixel()
	dx, dy := tx-pos[0], ty-pos[1]
	dist := math.Hypot(dx, dy)
	if dist <= speed+1e-9 { // tolerance so float drift can't add a tick per tile
		*pos = f64.Vec2{tx, ty}
		*path = (*path)[1:]
	} else {
		pos[0] += dx / dist * speed
		pos[1] += dy / dist * speed
	}
	return true
}

func (w *World) tick() {
	w.ticks++
	for _, p := range w.players {
		p.ticks++
		moved := p.step()
		w.tickDiscovery(p)
		w.tickJob(p)
		w.tickVisit(p)
		w.tickHealth(p)
		w.sendFightState(p)
		if moved || p.healthDirty {
			w.broadcast(p.updateMessage())
			p.healthDirty = false
		}
		w.sendPath(p)
		if p.invDirty && p.sess != nil {
			p.sess.push(w.inventoryMessage(p))
			p.invDirty = false
		}
	}
	w.tickNodes()
	w.tickStores()
	w.tickMobs()
}

// sendPath tells the player their route whenever its destination changes,
// so their client can draw it. Nobody else is sent it.
func (w *World) sendPath(p *player) {
	if p.sess == nil {
		return
	}
	var dest *tile
	if len(p.path) > 0 {
		dest = &p.path[len(p.path)-1]
	}
	if (dest == nil) == (p.sentDest == nil) && (dest == nil || *dest == *p.sentDest) {
		return
	}
	tiles := make([][2]int, len(p.path))
	for i, t := range p.path {
		tiles[i] = [2]int{t.x, t.y}
	}
	p.sess.push(resources.NewPathMessage(resources.PathContents{Tiles: tiles}))
	if dest == nil {
		p.sentDest = nil
	} else {
		d := *dest
		p.sentDest = &d
	}
}
