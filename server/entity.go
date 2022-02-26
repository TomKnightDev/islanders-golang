package main

import "golang.org/x/image/math/f64"

type Entity struct {
	id   uint16
	tile f64.Vec2
	pos  f64.Vec2
}

func NewEntity(tile f64.Vec2, pos f64.Vec2) *Entity {
	e := &Entity{
		id:   100 + uint16(len(ServerInstance.enemies)),
		tile: tile,
		pos:  pos,
	}

	return e
}

func (e *Entity) Move(targetPos f64.Vec2) {
	if targetPos[0] > e.pos[0] {
		e.pos[0] += 0.5
	} else if targetPos[0] < e.pos[0] {
		e.pos[0] -= 0.5
	}
	if targetPos[1] > e.pos[1] {
		e.pos[1] += 0.5
	} else if targetPos[1] < e.pos[1] {
		e.pos[1] -= 0.5
	}
}
