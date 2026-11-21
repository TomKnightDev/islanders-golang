package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/client/entities"
)

// CameraControls shows a button to follow the character again after the
// player has dragged the map away.
type CameraControls struct {
	player *entities.Player
}

func NewCameraControls(player *entities.Player) *CameraControls {
	return &CameraControls{player: player}
}

func (c *CameraControls) Update() error {
	if c.player == nil || c.player.FollowCam {
		return nil
	}
	w, _ := screenSize()
	imgui.SetNextWindowPosV(imgui.Vec2{X: w / 2, Y: windowMargin}, imgui.ConditionAlways, imgui.Vec2{X: 0.5, Y: 0})
	imgui.BeginV("##camera", nil, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoResize|
		imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoSavedSettings)
	if imgui.Button("Follow my character") {
		c.player.FollowCam = true
	}
	imgui.End()
	return nil
}

// Draw does nothing: the button is drawn with the other imgui windows.
func (c *CameraControls) Draw(screen *ebiten.Image) {}
