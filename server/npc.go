package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

// NPC ids start well above player ids; both share one id space on the client.
const npcIDBase = 1000

type NPC struct {
	id               uint16
	username         string
	position         f64.Vec2
	tile             f64.Vec2
	target           f64.Vec2
	ticksToNewTarget int
	ticksToHello     int
}

var npcNames = []string{"Bob", "Alice", "Charlie"}

func (w *World) initNPCs() []*NPC {
	npcs := make([]*NPC, len(npcNames))
	for i, name := range npcNames {
		npcs[i] = &NPC{
			id:               uint16(npcIDBase + i),
			username:         name,
			position:         w.randomWalkablePos(),
			tile:             f64.Vec2{0, 0},
			target:           randomMapPos(),
			ticksToNewTarget: 300 + rand.Intn(300),
			ticksToHello:     600 + rand.Intn(1200),
		}
	}
	return npcs
}

func randomMapPos() f64.Vec2 {
	return f64.Vec2{
		50 + rand.Float64()*700,
		50 + rand.Float64()*700,
	}
}

// randomWalkablePos picks a spawn point an NPC can actually move from.
func (w *World) randomWalkablePos() f64.Vec2 {
	for {
		p := randomMapPos()
		if w.wm.BoxPassable(p[0], p[1]) {
			return p
		}
	}
}

func (w *World) tickNPCs() {
	contents := make([]resources.ServerEntityUpdateContents, len(w.npcs))
	for i, npc := range w.npcs {
		dx := npc.target[0] - npc.position[0]
		dy := npc.target[1] - npc.position[1]
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > 1 {
			const speed = 0.3
			newX := npc.position[0] + (dx/dist)*speed
			newY := npc.position[1] + (dy/dist)*speed
			if w.wm.BoxPassable(newX, newY) {
				npc.position[0] = newX
				npc.position[1] = newY
			} else {
				npc.target = randomMapPos()
				npc.ticksToNewTarget = 300 + rand.Intn(300)
			}
		}

		npc.ticksToNewTarget--
		if npc.ticksToNewTarget <= 0 {
			npc.target = randomMapPos()
			npc.ticksToNewTarget = 300 + rand.Intn(300)
		}

		npc.ticksToHello--
		if npc.ticksToHello <= 0 {
			w.broadcast(resources.NewChatMessage(npc.id, fmt.Sprintf("%s: %s", npc.username, "Hello!")))
			npc.ticksToHello = 600 + rand.Intn(1200)
		}

		contents[i] = npc.updateContents()
	}

	w.broadcast(resources.NewServerEntityUpdateMessage(contents))
}

func (npc *NPC) updateContents() resources.ServerEntityUpdateContents {
	return resources.ServerEntityUpdateContents{
		EntityId: npc.id,
		Pos:      npc.position,
		Tile:     npc.tile,
		Username: npc.username,
	}
}

func (w *World) npcSnapshot() *resources.Message {
	contents := make([]resources.ServerEntityUpdateContents, len(w.npcs))
	for i, npc := range w.npcs {
		contents[i] = npc.updateContents()
	}
	return resources.NewServerEntityUpdateMessage(contents)
}
