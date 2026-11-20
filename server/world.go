package main

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

const (
	tickInterval = 25 * time.Millisecond
	walkSpeed    = 4.0 // tiles per second

	// Pixels moved per tick.
	walkStep = walkSpeed * resources.TileSize * float64(tickInterval) / float64(time.Second)

	sessionBuffer = 256
)

var spawnTile = tile{50, 50}

// playerSprite is the sprite's pixel position in characters.png.
var playerSprite = f64.Vec2{0, 0}

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
}

// currentTile is the tile the player stands on, or is walking onto.
func (p *player) currentTile() tile {
	if len(p.path) > 0 {
		return p.path[0]
	}
	return tile{
		int(math.Round(p.pos[0] / resources.TileSize)),
		int(math.Round(p.pos[1] / resources.TileSize)),
	}
}

func (p *player) updateMessage() *resources.Message {
	return resources.NewUpdateMessage(p.id, resources.UpdateContents{
		Pos:      p.pos,
		Tile:     playerSprite,
		Username: p.username,
	})
}

type command func(w *World)

// World owns all game state. Only its Run goroutine touches it; connection
// handlers send it commands with Do.
type World struct {
	wm      resources.WorldMap
	players map[uint16]*player
	byName  map[string]*player
	nextID  uint16
	npcs    []*NPC
	cmds    chan command
}

func NewWorld(wm resources.WorldMap) *World {
	w := &World{
		wm:      wm,
		players: make(map[uint16]*player),
		byName:  make(map[string]*player),
		cmds:    make(chan command, 256),
	}
	w.npcs = w.initNPCs()
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
			w.tick()
		}
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
		x, y := spawnTile.pixel()
		p = &player{
			id:       w.nextID,
			username: username,
			password: password,
			pos:      f64.Vec2{x, y},
		}
		w.nextID++
		w.players[p.id] = p
		w.byName[username] = p
	}
	p.sess = sess

	sess.push(resources.NewConnectResponseMessage(resources.ConnectResponseContents{
		ClientId: p.id,
		Pos:      p.pos,
		Tile:     playerSprite,
		WorldMap: w.wm,
	}))
	for _, other := range w.players {
		if other != p {
			sess.push(other.updateMessage())
		}
	}
	sess.push(w.npcSnapshot())

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

// moveTo sends p walking to goal. A player already between tiles finishes
// that step first, so there is no snapping.
func (w *World) moveTo(p *player, goal tile) {
	start := p.currentTile()
	path := findPath(&w.wm, start, goal)
	if path == nil {
		return
	}
	if len(p.path) > 0 {
		p.path = append([]tile{start}, path...)
	} else {
		p.path = path
	}
}

// step moves p one tick along its path and reports whether it moved.
func (p *player) step() bool {
	if len(p.path) == 0 {
		return false
	}
	tx, ty := p.path[0].pixel()
	dx, dy := tx-p.pos[0], ty-p.pos[1]
	dist := math.Hypot(dx, dy)
	if dist <= walkStep+1e-9 { // tolerance so float drift can't add a tick per tile
		p.pos = f64.Vec2{tx, ty}
		p.path = p.path[1:]
	} else {
		p.pos[0] += dx / dist * walkStep
		p.pos[1] += dy / dist * walkStep
	}
	return true
}

func (w *World) tick() {
	for _, p := range w.players {
		if p.step() {
			w.broadcast(p.updateMessage())
		}
	}
	w.tickNPCs()
}
