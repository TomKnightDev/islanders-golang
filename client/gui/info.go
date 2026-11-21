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
	Open   bool // shown; the menu bar toggles it
	mgr    *renderer.Manager
	player *entities.Player
	// SendChat posts a chat message; used by "Share location".
	SendChat func(string)
	// OpenGuide shows the player guide.
	OpenGuide func()
}

func NewInfo(screenWidth, screenHeight int, player *entities.Player, mgr *renderer.Manager) *Info {

	info := Info{
		Open:   true,
		mgr:    mgr,
		player: player,
	}

	return &info
}

func (info *Info) Update() error {
	if !info.Open {
		infoBottom = topMargin()
		return nil
	}
	placeWindow("Info", imgui.Vec2{X: windowMargin, Y: topMargin()}, imgui.Vec2{})
	imgui.BeginV("Info", &info.Open, imgui.WindowFlagsAlwaysAutoResize)
	imgui.Text(fmt.Sprintf("Name: %s", info.player.Username))
	if mh := info.player.MaxHealth; mh > 0 {
		imgui.Text("Health")
		imgui.SameLine()
		centredBar(float32(info.player.Health)/float32(mh), 260, fmt.Sprintf("%d / %d", info.player.Health, mh))
	}
	if ms := info.player.MaxStamina; ms > 0 {
		imgui.Text("Stamina")
		imgui.SameLine()
		centredBar(float32(info.player.Stamina)/float32(ms), 260, fmt.Sprintf("%d / %d", info.player.Stamina, ms))
		if imgui.IsItemHovered() {
			imgui.SetTooltip("Used up by work and fighting. When it runs out, your character rests at the nearest inn, then carries on.")
		}
	}

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
	if info.OpenGuide != nil {
		imgui.SameLine()
		if imgui.Button("Guide (F1)") {
			info.OpenGuide()
		}
	}
	infoBottom = imgui.WindowPos().Y + imgui.WindowSize().Y
	imgui.End()

	return nil
}

func (info *Info) Draw(screen *ebiten.Image) {
	info.mgr.Draw(screen)
}
