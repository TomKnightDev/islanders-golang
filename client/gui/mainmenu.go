package gui

import (
	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type MainMenu struct {
	mgr             *renderer.Manager
	username        string
	password        string
	Connect         chan string
	Connected       bool
	server          string
	FailedToConnect []string
}

func NewMainMenu(screenWidth, screenHeight int) *MainMenu {
	mgr := renderer.New(nil)

	mm := &MainMenu{
		mgr:      mgr,
		Connect:  make(chan string),
		username: "Tom",
		server:   "192.168.1.98:8285",
	}

	mm.mgr.SetDisplaySize(float32(screenWidth), float32(screenHeight))

	return mm
}

func (mm *MainMenu) Update() error {
	if mm.Connected {
		return nil
	}

	mm.mgr.Update(1.0 / 60.0)
	mm.mgr.BeginFrame()
	{
		flags := imgui.WindowFlagsNone
		// flags |= imgui.WindowFlagsNoTitleBar
		flags |= imgui.WindowFlagsNoResize
		flags |= imgui.WindowFlagsNoCollapse

		imgui.SetNextWindowPos(imgui.Vec2{100, 100})
		imgui.SetNextWindowSize(imgui.Vec2{600, 400})
		imgui.BeginV("Main Menu", nil, flags)

		imgui.InputText("Server", &mm.server)
		imgui.InputText("Username", &mm.username)
		imgui.InputText("Password", &mm.password)
		if imgui.Button("Connect") {
			mm.Connect <- mm.server
			mm.Connect <- mm.username
			mm.Connect <- mm.password
		}
		for _, m := range mm.FailedToConnect {
			imgui.Text(m)
		}
		imgui.End()
	}
	mm.mgr.EndFrame()

	return nil
}

func (mm *MainMenu) Draw(screen *ebiten.Image) {
	mm.mgr.Draw(screen)
}
