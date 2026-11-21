package gui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/resources"
)

// devSpeeds are the quick-pick game speeds.
var devSpeeds = []int{1, 2, 5, 10, 20}

// Dev is a developer window, toggled with F12, for testing: it sets the
// game speed on servers started with -dev.
type Dev struct {
	Open  bool
	send  func(resources.DevCommandContents)
	state resources.DevStateContents
	speed int32
}

func NewDev(send func(resources.DevCommandContents)) *Dev {
	return &Dev{send: send, speed: 1}
}

func (d *Dev) SetState(c resources.DevStateContents) {
	d.state = c
	if !imgui.IsAnyItemActive() {
		d.speed = int32(c.Speed)
	}
}

func (d *Dev) setSpeed(s int) {
	if s != d.state.Speed {
		d.send(resources.DevCommandContents{Speed: s})
	}
}

func (d *Dev) Update() error {
	if !d.Open {
		return nil
	}
	imgui.BeginV("Dev", &d.Open, imgui.WindowFlagsAlwaysAutoResize)
	if !d.state.Enabled {
		dimmed("Dev tools are off on this server.")
		dimmed("Start it with -dev (make run-server does) to use them.")
		imgui.End()
		return nil
	}
	imgui.Text(fmt.Sprintf("Game speed: %dx", d.state.Speed))
	imgui.PushItemWidth(300)
	if imgui.SliderInt("##speed", &d.speed, 1, resources.MaxGameSpeed) {
		d.setSpeed(int(d.speed))
	}
	imgui.PopItemWidth()
	for i, s := range devSpeeds {
		if i > 0 {
			imgui.SameLine()
		}
		if imgui.Button(fmt.Sprintf("%dx", s)) {
			d.speed = int32(s)
			d.setSpeed(s)
		}
	}
	dimmed("Speeds up the whole world for everyone on this server:")
	dimmed("walking, actions, fights, respawns and regrowth.")
	imgui.End()
	return nil
}

// Draw does nothing: the window is drawn with the other imgui windows.
func (d *Dev) Draw(screen *ebiten.Image) {}
