package gui

import (
	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type Chat struct {
	Open         bool // shown; the menu bar toggles it
	mgr          *renderer.Manager
	message      string
	SendChan     chan string
	RecvMessages []string
}

func NewChat(screenWidth, screenHeight int, mgr *renderer.Manager) *Chat {

	chat := Chat{
		Open:     true,
		mgr:      mgr,
		SendChan: make(chan string),
	}

	return &chat
}

func (chat *Chat) Update() error {
	if !chat.Open {
		return nil
	}
	_, h := screenSize()
	placeWindow("Chat", imgui.Vec2{X: windowMargin, Y: h - windowMargin}, imgui.Vec2{X: 0, Y: 1})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 560, Y: 260}, imgui.ConditionFirstUseEver)
	imgui.BeginV("Chat", &chat.Open, 0)

	// Messages fill the window above the input line and stay scrolled to
	// the newest one unless the player has scrolled up to read back.
	imgui.BeginChildV("messages", imgui.Vec2{X: 0, Y: -imgui.FrameHeightWithSpacing()}, false, 0)
	atBottom := imgui.ScrollY() >= imgui.ScrollMaxY()
	for _, m := range chat.RecvMessages {
		imgui.Text(m)
	}
	if atBottom {
		imgui.SetScrollHereY(1)
	}
	imgui.EndChild()

	imgui.PushItemWidth(imgui.ContentRegionAvail().X - 90)
	sent := imgui.InputTextV("##message", &chat.message, imgui.InputTextFlagsEnterReturnsTrue, nil)
	imgui.PopItemWidth()
	imgui.SameLine()
	if (imgui.Button("Send") || sent) && chat.message != "" {
		chat.SendChan <- chat.message
		chat.message = ""
		imgui.SetKeyboardFocusHere() // keep typing after Enter
	}
	imgui.End()

	return nil
}

func (chat *Chat) Draw(screen *ebiten.Image) {
	chat.mgr.Draw(screen)
}
