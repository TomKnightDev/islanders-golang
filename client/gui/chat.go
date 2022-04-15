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

func NewChat(screenWidth, screenHeight int, mgr *renderer.Manager) *Chat {

	chat := Chat{
		mgr:      mgr,
		SendChan: make(chan string),
	}

	return &chat
}

func (chat *Chat) Update() error {
	imgui.Begin("Chat")
	imgui.InputText("Message", &chat.message)
	if imgui.Button("Send") {
		chat.SendChan <- chat.message
	}

	for _, m := range chat.RecvMessages {
		imgui.Text(m)
	}
	imgui.End()

	return nil
}

func (chat *Chat) Draw(screen *ebiten.Image) {
	chat.mgr.Draw(screen)
}
