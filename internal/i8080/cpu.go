package i8080

type Bus interface {
	Read(addr uint16) byte
	Write(addr uint16, b byte)
	In(port byte) byte
	Out(port byte, b byte)
}

type CPU struct {
	A, B, C, D, E, H, L            byte
	SP, PC                         uint16
	Sign, Zero, Parity, Carry, Aux bool
	halted                         bool
	intEnabled                     bool
	bus                            Bus
	eiPending                      bool
}

// Return a CPU in power-on state.
func New(bus Bus) *CPU {
	return &CPU{bus: bus}
}

func (c *CPU) fetchByte() byte {
	b := c.bus.Read(c.PC)
	c.PC++
	return b
}

func (c *CPU) fetchWord() uint16 {
	lo := uint16(c.fetchByte())
	hi := uint16(c.fetchByte())
	return hi<<8 | lo
}

// Read word lo byte first.
func (c *CPU) readWord(addr uint16) uint16 {
	return uint16(c.bus.Read(addr+1))<<8 | uint16(c.bus.Read(addr))
}

func (c *CPU) writeWord(addr uint16, v uint16) {
	c.bus.Write(addr, byte(v))
	c.bus.Write(addr+1, byte(v>>8))
}

// Execute instruction at PC and return cycles count.
func (c *CPU) Step() int {
	if c.halted {
		return 4
	}
	op := c.fetchByte()
	return ops[op](c)
}

func (c *CPU) setBC(v uint16) { c.B, c.C = byte(v>>8), byte(v) }
func (c *CPU) setDE(v uint16) { c.D, c.E = byte(v>>8), byte(v) }
func (c *CPU) setHL(v uint16) { c.H, c.L = byte(v>>8), byte(v) }
func (c *CPU) hl() uint16     { return uint16(c.H)<<8 | uint16(c.L) }
func (c *CPU) bc() uint16     { return uint16(c.B)<<8 | uint16(c.C) }
func (c *CPU) de() uint16     { return uint16(c.D)<<8 | uint16(c.E) }
