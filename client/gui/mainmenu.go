package gui

import (
	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type MainMenu struct {
	mgr        *renderer.Manager
	playerName string
	Connect    chan string
	connected  bool
	server     string
}

func NewMainMenu(screenWidth, screenHeight int) *MainMenu {
	mgr := renderer.New(nil)

	mm := &MainMenu{
		mgr:        mgr,
		Connect:    make(chan string),
		playerName: "Tom",
		server:     "192.168.1.98:8285",
	}

	mm.mgr.SetDisplaySize(float32(screenWidth), float32(screenHeight))

	return mm
}

func (mm *MainMenu) Update() error {
	if mm.connected {
		return nil
	}

	mm.mgr.Update(1.0 / 60.0)
	mm.mgr.BeginFrame()
	{
		imgui.InputText("Server", &mm.server)
		imgui.InputText("Name", &mm.playerName)
		if imgui.Button("Connect") {
			mm.Connect <- mm.server
			mm.Connect <- mm.playerName
			mm.connected = true
		}
	}
	mm.mgr.EndFrame()

	return nil
}

func (mm *MainMenu) Draw(screen *ebiten.Image) {
	mm.mgr.Draw(screen)
}
