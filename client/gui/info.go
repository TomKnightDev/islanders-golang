package gui

import (
	"fmt"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/client/entities"
	"github.com/tomknightdev/islanders-golang/resources"
)

type Info struct {
	mgr    *renderer.Manager
	player *entities.Player
	// SendChat posts a chat message; used by "Share location".
	SendChat func(string)
}

func NewInfo(screenWidth, screenHeight int, player *entities.Player, mgr *renderer.Manager) *Info {

	info := Info{
		mgr:    mgr,
		player: player,
	}

	return &info
}

func (info *Info) Update() error {
	imgui.BeginV("Info", nil, imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text(fmt.Sprintf("Name: %s", info.player.Username))

	// A grid square plus exact tile. It only changes when the player steps
	// onto a new tile, and it's short enough to read out to someone.
	tx, ty := resources.TileAt(info.player.Position[0], info.player.Position[1])
	location := resources.Location(tx, ty)
	imgui.Text("Location: " + location)
	landmark := ""
	if info.player.WorldMap != nil {
		landmark = info.player.WorldMap.Landmark(tx, ty)
	}
	if landmark != "" {
		imgui.Text("Near " + landmark)
	}
	if imgui.Button("Share location") && info.SendChat != nil {
		msg := "I'm at " + location
		if landmark != "" {
			msg += ", near " + landmark
		}
		info.SendChat(msg)
	}
	imgui.End()

	return nil
}

func (info *Info) Draw(screen *ebiten.Image) {
	info.mgr.Draw(screen)
}
