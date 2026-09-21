package machine

import "github.com/OhWelp/go-spaceinvaders/internal/i8080"

type Machine struct {
	cpu         *i8080.CPU
	mem         [0x4000]byte
	shift       uint16
	shiftOffset byte

	port1 byte
	port2 byte

	prevOut3 byte
	prevOut5 byte
}
