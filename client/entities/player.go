package entities

import "github.com/hajimehoshi/ebiten/v2"

type player struct {
}

func NewPlayer() *player {
	p := &player{}

	return p
}

func (p *player) Update() error {
	return nil
}

func (p *player) Draw(screen *ebiten.Image) {

}
