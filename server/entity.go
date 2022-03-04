package main

import (
	"github.com/solarlune/resolv"
	"golang.org/x/image/math/f64"
)

type Entity struct {
	id       uint16
	tile     f64.Vec2
	pos      f64.Vec2
	collider *resolv.Object
}

func NewEntity(tile f64.Vec2, pos f64.Vec2) *Entity {
	e := &Entity{
		id:       100 + uint16(len(ServerInstance.enemies)),
		tile:     tile,
		pos:      pos,
		collider: resolv.NewObject(pos[0], pos[1], 8, 8),
	}

	e.collider.SetShape(resolv.NewCircle(8, 8, 8))

	return e
}

func (e *Entity) Move(targetPos f64.Vec2) {
	x := 0.0
	y := 0.0

	if targetPos[0] > e.pos[0] {
		x += 0.5
	} else if targetPos[0] < e.pos[0] {
		x -= 0.5
	}
	if targetPos[1] > e.pos[1] {
		y += 0.5
	} else if targetPos[1] < e.pos[1] {
		y -= 0.5
	}

	if !checkForCollision(e, x, y) {
		e.pos[0] += x
		e.pos[1] += y

		e.collider.X += x
		e.collider.Y += y
		e.collider.Update()
	}
}

func checkForCollision(e *Entity, x, y float64) bool {
	if collision := e.collider.Check(x, y); collision != nil {
		return true
	}

	return false
}
