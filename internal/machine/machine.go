package machine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/OhWelp/go-spaceinvaders/internal/i8080"
)

const cyclesPerHalfFrame = 2_000_000 / 60 / 2 // 16,666 cycles per half-frame

func (m *Machine) RunFrame() {
	m.runCycles(cyclesPerHalfFrame)
	m.cpu.Interrupt(0xCF)
	m.runCycles(cyclesPerHalfFrame)
	m.cpu.Interrupt(0xD7)
}

func (m *Machine) Framebuffer() []byte { return m.mem[0x2400:0x4000] }

func (m *Machine) runCycles(n int) {
	for n > 0 {
		n -= m.cpu.Step()
	}
}

type Machine struct {
	cpu         *i8080.CPU
	mem         [0x4000]byte
	shift       uint16
	shiftOffset byte

	port1 byte
	port2 byte

	prevOut3 byte
	prevOut5 byte

	sounds Sounds
}

func New(rom []byte, sounds Sounds) (*Machine, error) {
	if len(rom) > 0x2000 {
		return nil, fmt.Errorf("rom is %d bytes; max is 8192", len(rom))
	}
	m := &Machine{}
	copy(m.mem[:], rom)
	m.cpu = i8080.New(m)
	m.port1 = 0x08
	m.sounds = sounds
	return m, nil
}

func (m *Machine) Read(addr uint16) byte {
	if addr >= 0x4000 {
		addr = 0x2000 + (addr & 0x1FFF)
	}
	return m.mem[addr]
}

func (m *Machine) Write(addr uint16, b byte) {
	if addr >= 0x4000 {
		addr = 0x2000 + (addr & 0x1FFF)
	}
	if addr < 0x2000 {
		return
	}
	m.mem[addr] = b
}

func (m *Machine) In(port byte) byte {
	switch port {
	case 1:
		return m.port1
	case 2:
		return m.port2
	case 3:
		return byte(m.shift >> (8 - m.shiftOffset))
	}
	return 0
}

func (m *Machine) Out(port byte, b byte) {
	switch port {
	case 2:
		m.shiftOffset = b & 0x07
	case 3:
		if m.sounds != nil {
			m.sounds.SetUFO(b&0x01 != 0)
			for bit := 1; bit <= 3; bit++ {
				if b&(1<<bit) != 0 && m.prevOut3&(1<<bit) == 0 {
					m.sounds.Play(bit)
				}
			}
		}
		m.prevOut3 = b
	case 4:
		m.shift = uint16(b)<<8 | m.shift>>8
	case 5:
		if m.sounds != nil {
			for bit := 0; bit <= 4; bit++ {
				if b&(1<<bit) != 0 && m.prevOut5&(1<<bit) == 0 {
					m.sounds.Play(8 + bit)
				}
			}
		}
		m.prevOut5 = b
	}
}

func (m *Machine) SetInputs(coin, start1, start2, fire, left, right bool) {
	set := func(bit byte, on bool) {
		if on {
			m.port1 |= 1 << bit
		} else {
			m.port1 &^= 1 << bit
		}
	}
	set(0, coin)
	set(1, start2)
	set(2, start1)
	set(4, fire)
	set(5, left)
	set(6, right)
}

var romParts = []string{"invaders.h", "invaders.g", "invaders.f", "invaders.e"}

func LoadROM(dir string) ([]byte, error) {
	var rom []byte
	for _, part := range romParts {
		b, err := os.ReadFile(filepath.Join(dir, part))
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", part, err)
		}
		rom = append(rom, b...)
	}
	return rom, nil
}

type Sounds interface {
	Play(id int)
	SetUFO(on bool)
}
