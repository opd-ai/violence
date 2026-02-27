package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/opd-ai/violence/pkg/config"
)

// Game implements ebiten.Game for the VIOLENCE raycasting FPS.
type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "VIOLENCE")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.C.InternalWidth, config.C.InternalHeight
}

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(config.C.WindowWidth, config.C.WindowHeight)
	ebiten.SetVsyncEnabled(config.C.VSync)
	ebiten.SetFullscreen(config.C.FullScreen)
	ebiten.SetWindowTitle("VIOLENCE")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
