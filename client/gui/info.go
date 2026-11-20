package gui

import (
	"fmt"
	"strings"

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
	where := "" // "in Oakhaven", "near Ironridge Mine", ...
	if info.player.WorldMap != nil {
		where = info.player.WorldMap.Describe(tx, ty)
	}
	if where != "" {
		imgui.Text(strings.ToUpper(where[:1]) + where[1:])
	}
	if imgui.Button("Share location") && info.SendChat != nil {
		msg := "I'm at " + location
		if where != "" {
			msg += ", " + where
		}
		info.SendChat(msg)
	}
	imgui.End()

	return nil
}

func (info *Info) Draw(screen *ebiten.Image) {
	info.mgr.Draw(screen)
}
