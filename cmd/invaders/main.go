package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/OhWelp/go-spaceinvaders/internal/machine"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type game struct {
	m   *machine.Machine
	img *ebiten.Image
}

type sounds struct {
	ctx     *audio.Context
	samples map[int][]byte
	ufo     *audio.Player
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
	snd, err := newSounds("game-data")
	if err != nil {
		log.Fatal(err)
	}
	m, err := machine.New(rom, snd)
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

func newSounds(dir string) (*sounds, error) {
	s := &sounds{ctx: audio.NewContext(11025), samples: map[int][]byte{}}
	for id, file := range map[int]string{1: "1.wav", 2: "2.wav", 3: "3.wav",
		8: "4.wav", 9: "5.wav", 10: "6.wav", 11: "7.wav", 12: "8.wav"} {
		b, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return nil, fmt.Errorf("loading %s, ^%w", file, err)
		}
		d, err := wav.DecodeWithSampleRate(11025, bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("decoding %s, %w", file, err)
		}
		pcm, _ := io.ReadAll(d)
		s.samples[id] = pcm
	}
	b, _ := os.ReadFile(filepath.Join(dir, "0.wav"))
	d, _ := wav.DecodeWithSampleRate(11025, bytes.NewReader(b))
	pcm, _ := io.ReadAll(d)
	s.ufo, _ = s.ctx.NewPlayer(audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm))))
	return s, nil
}

func (s *sounds) Play(id int) {
	if pcm, ok := s.samples[id]; ok {
		p := s.ctx.NewPlayerFromBytes(pcm)
		p.Play()
	}
}

func (s *sounds) SetUFO(on bool) {
	if on && !s.ufo.IsPlaying() {
		s.ufo.Rewind()
		s.ufo.Play()
	} else if !on && s.ufo.IsPlaying() {
		s.ufo.Pause()
	}
}
