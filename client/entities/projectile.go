package entities

import (
	"golang.org/x/image/math/f64"
)

type Projectile struct {
	startPos  f64.Vec2
	direction f64.Vec2
	velocity  float32
}

func NewProjectile(startPos f64.Vec2, targetPos f64.Vec2) *Projectile {
	return &Projectile{
		startPos:  startPos,
		direction: getDir(startPos, targetPos),
		velocity:  1,
	}
}

func getDir(startPos f64.Vec2, targetPos f64.Vec2) f64.Vec2 {
	return f64.Vec2{targetPos[0] - startPos[0], targetPos[1] - startPos[1]}
}
