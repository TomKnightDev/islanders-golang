package gui

import (
	"image/color"
	"log"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/inkyblackness/imgui-go/v4"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	mplusNormalFont font.Face
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

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func NewMainMenu(screenWidth, screenHeight int, mgr *renderer.Manager) *MainMenu {

	mm := &MainMenu{
		mgr:      mgr,
		Connect:  make(chan string),
		username: "Tom",
		server:   "192.168.1.128:8285",
	}

	return mm
}

func (mm *MainMenu) Update() error {
	if mm.Connected {
		return nil
	}

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

	return nil
}

func (mm *MainMenu) Draw(screen *ebiten.Image) {

	text.Draw(screen, "Dungeon Crawl", mplusNormalFont, 80, 80, color.White)

	mm.mgr.Draw(screen)
}
