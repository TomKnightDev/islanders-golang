package gui

import (
	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type Chat struct {
	mgr          *renderer.Manager
	message      string
	SendChan     chan string
	RecvMessages []string
}

func NewChat(screenWidth, screenHeight int) *Chat {
	mgr := renderer.New(nil)

	chat := &Chat{
		mgr:      mgr,
		SendChan: make(chan string),
	}

	chat.mgr.SetDisplaySize(float32(screenWidth), float32(screenHeight))

	return chat
}

func (chat *Chat) Update() error {
	chat.mgr.Update(1.0 / 60.0)
	chat.mgr.BeginFrame()
	{
		imgui.InputText("Message", &chat.message)
		if imgui.Button("Send") {
			chat.SendChan <- chat.message
		}

		for _, m := range chat.RecvMessages {
			imgui.Text(m)
		}
	}
	chat.mgr.EndFrame()

	return nil
}

func (chat *Chat) Draw(screen *ebiten.Image) {
	chat.mgr.Draw(screen)
}
