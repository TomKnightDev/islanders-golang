package main

import (
	"log"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/tomknightdev/islanders-golang/client/entities"
	"github.com/tomknightdev/islanders-golang/client/gui"
)

type Entity interface {
	Update() error
	Draw(*ebiten.Image)
}

// uiFontSize is the UI text size in pixels. The UI uses a real TrueType
// font at this size; scaling imgui's small built-in bitmap font up instead
// makes it blurry.
const uiFontSize = 22

type Game struct {
	username             string
	password             string
	serverAddr           string
	Gui                  []Entity
	Environment          []Entity
	Entities             map[uint16]Entity
	Player               Entity
	screenWidth          int
	screenHeight         int
	ConnectFailedMessage chan string
	renderMgr            *renderer.Manager
	uiFont               imgui.Font
}

func (g *Game) Update() error {
	handleMessages(g)

	if g.Player != nil {
		if err := g.Player.Update(); err != nil {
			log.Print(err)
		}
	}

	g.renderMgr.Update(1.0 / 60.0)
	g.renderMgr.BeginFrame()
	imgui.PushFont(g.uiFont)
	for _, gui := range g.Gui {
		gui.Update()
	}
	imgui.PopFont()
	g.renderMgr.EndFrame()

	for _, e := range g.Entities {
		e.Update()
	}
	for _, e := range g.Environment {
		e.Update()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, e := range g.Environment {
		e.Draw(screen)
	}
	for _, e := range g.Entities {
		e.Draw(screen)
	}
	if g.Player != nil {
		g.Player.Draw(screen)
	}

	// Everything above drew into the camera's world image; show it once,
	// zoomed, then draw names on top at screen resolution.
	if cam != nil {
		cam.Blit(screen)
		if pl, ok := g.Player.(*entities.Player); ok {
			pl.DrawPath(screen)
		}
		for _, e := range g.Environment {
			if l, ok := e.(entities.Labeled); ok {
				l.DrawLabel(screen)
			}
		}
		for _, e := range g.Entities {
			if l, ok := e.(entities.Labeled); ok {
				l.DrawLabel(screen)
			}
		}
		if l, ok := g.Player.(entities.Labeled); ok {
			l.DrawLabel(screen)
		}
	}

	for _, gui := range g.Gui {
		gui.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if cam != nil && (outsideWidth != cam.Width || outsideHeight != cam.Height) {
		cam.Resize(outsideWidth, outsideHeight) // keep the player centred in the window
	}
	g.screenWidth = outsideWidth
	g.screenHeight = outsideHeight
	g.renderMgr.SetDisplaySize(float32(outsideWidth), float32(outsideHeight))
	return outsideWidth, outsideHeight
}

func main() {
	game := &Game{
		screenWidth:          1920,
		screenHeight:         1080,
		Entities:             make(map[uint16]Entity),
		ConnectFailedMessage: make(chan string),
		renderMgr:            renderer.New(nil),
	}

	// Load the UI font, rebuild imgui's font texture with it, and have the
	// renderer pick up the new texture. Nearest filtering keeps the glyphs
	// crisp, since they're drawn at their native size.
	game.uiFont = imgui.CurrentIO().Fonts().AddFontFromMemoryTTF(fonts.MPlus1pRegular_ttf, uiFontSize)
	imgui.CurrentIO().Fonts().TextureDataRGBA32()
	game.renderMgr.Cache.ResetFontAtlasCache(ebiten.FilterNearest)

	mm := gui.NewMainMenu(game.screenWidth, game.screenHeight, game.renderMgr)

	go func() {
		for {
			m := <-game.ConnectFailedMessage
			mm.FailedToConnect = append(mm.FailedToConnect, m)
		}
	}()

	go func() {
		for {
			game.serverAddr = <-mm.Connect
			game.username = <-mm.Connect
			game.password = <-mm.Connect
			err := connectToServer(game)
			if err != nil {
				log.Printf("Failed to connect %s to server: %s", game.username, err)
			} else {
				return
			}
		}
	}()

	game.Gui = append(game.Gui, mm)

	ebiten.SetWindowSize(game.screenWidth, game.screenHeight)
	ebiten.SetWindowTitle("Islanders")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
