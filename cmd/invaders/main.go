package main

import (
	"image/color"
	"log"

	"github.com/OhWelp/go-spaceinvaders/internal/machine"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type game struct {
	m   *machine.Machine
	img *ebiten.Image
}

func (g *game) Update() error {
	g.m.SetInputs(
		inpututil.IsKeyJustPressed(ebiten.KeyC),
		ebiten.IsKeyPressed(ebiten.Key1),
		ebiten.IsKeyPressed(ebiten.Key2),
		ebiten.IsKeyPressed(ebiten.KeySpace),
		ebiten.IsKeyPressed(ebiten.KeyArrowLeft),
		ebiten.IsKeyPressed(ebiten.KeyArrowRight),
	)
	g.m.RunFrame()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	g.img.Clear()
	fb := g.m.Framebuffer()
	for x := 0; x < 224; x++ {
		for byteY := 0; byteY < 32; byteY++ {
			b := fb[x*32+byteY]
			for bit := 0; bit < 8; bit++ {
				if b&(1<<bit) != 0 {
					g.img.Set(x, 255-(byteY*8+bit), color.White)
				}
			}
		}
	}
	screen.DrawImage(g.img, nil)
}

func (g *game) Layout(_, _ int) (int, int) { return 224, 256 }

func main() {
	rom, err := machine.LoadROM("game-data")
	if err != nil {
		log.Fatal(err)
	}
	m, err := machine.New(rom)
	if err != nil {
		log.Fatal(err)
	}

	const scale = 4
	g := &game{m: m, img: ebiten.NewImage(224, 256)}
	ebiten.SetWindowSize(224*scale, 256*scale)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Space Invaders")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
