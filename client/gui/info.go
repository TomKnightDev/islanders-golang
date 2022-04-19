package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/client/entities"
)

type Info struct {
	mgr    *renderer.Manager
	player *entities.Player
}

func NewInfo(screenWidth, screenHeight int, player *entities.Player, mgr *renderer.Manager) *Info {

	info := Info{
		mgr:    mgr,
		player: player,
	}

	return &info
}

func (info *Info) Update() error {
	imgui.Begin("Info")
	imgui.Text(fmt.Sprintf("Name: %s", info.player.Username))
	imgui.Text(fmt.Sprintf("Position: %v", info.player.Position))
	imgui.End()

	return nil
}

func (info *Info) Draw(screen *ebiten.Image) {
	info.mgr.Draw(screen)
}
