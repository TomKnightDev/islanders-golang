package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tomknightdev/islanders-golang/resources"
	"golang.org/x/image/math/f64"
)

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

func initNPCs() []*NPC {
	npcs := make([]*NPC, len(npcNames))
	for i, name := range npcNames {
		pos := randomMapPos()
		npcs[i] = &NPC{
			id:               uint16(200 + i),
			username:         name,
			position:         pos,
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

func npcCanMoveTo(px, py float64) bool {
	wm := ServerInstance.worldMap
	if len(wm.Layers) == 0 {
		return true
	}
	w := wm.Width
	data := wm.Layers[0].Data
	for _, corner := range [4][2]float64{{px, py}, {px + 7, py}, {px, py + 7}, {px + 7, py + 7}} {
		tx, ty := int(corner[0]/8), int(corner[1]/8)
		if tx < 0 || ty < 0 || tx >= w || ty >= wm.Height {
			return false
		}
		if !resources.IsPassable(data[ty*w+tx]) {
			return false
		}
	}
	return true
}

func npcLoop(npcs []*NPC) {
	for {
		time.Sleep(25 * time.Millisecond)

		contents := make([]resources.ServerEntityUpdateContents, len(npcs))
		for i, npc := range npcs {
			dx := npc.target[0] - npc.position[0]
			dy := npc.target[1] - npc.position[1]
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 1 {
				const speed = 0.3
				newX := npc.position[0] + (dx/dist)*speed
				newY := npc.position[1] + (dy/dist)*speed
				if npcCanMoveTo(newX, newY) {
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
				broadcastNPCChat(npc, "Hello!")
				npc.ticksToHello = 600 + rand.Intn(1200)
			}

			contents[i] = resources.ServerEntityUpdateContents{
				EntityId: npc.id,
				Pos:      npc.position,
				Tile:     npc.tile,
				Username: npc.username,
			}
		}

		broadcastNPCPositions(contents)
	}
}

func broadcastNPCPositions(contents []resources.ServerEntityUpdateContents) {
	message := resources.NewServerEntityUpdateMessage(contents)
	for _, c := range ServerInstance.clientsById {
		if c.conn == nil {
			continue
		}
		c.mu.Lock()
		if err := c.conn.WriteJSON(message); err != nil {
			log.Println("NPC position write error:", err)
		}
		c.mu.Unlock()
	}
}

func sendNPCSnapshot(conn *websocket.Conn) {
	contents := make([]resources.ServerEntityUpdateContents, len(ServerInstance.npcs))
	for i, npc := range ServerInstance.npcs {
		contents[i] = resources.ServerEntityUpdateContents{
			EntityId: npc.id,
			Pos:      npc.position,
			Tile:     npc.tile,
			Username: npc.username,
		}
	}
	conn.WriteJSON(resources.NewServerEntityUpdateMessage(contents))
}

func broadcastNPCChat(npc *NPC, msg string) {
	m := resources.NewChatMessage(npc.id, fmt.Sprintf("%s: %s", npc.username, msg))
	for _, c := range ServerInstance.clientsById {
		if c.conn == nil {
			continue
		}
		c.mu.Lock()
		if err := c.conn.WriteJSON(m); err != nil {
			log.Println("NPC chat write error:", err)
		}
		c.mu.Unlock()
	}
}
