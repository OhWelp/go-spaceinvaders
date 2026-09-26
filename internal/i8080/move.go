package i8080

func opsMove() {
	for s, getSrc := range src { // This implements MOV in its entirety.
		for d, put := range dst {
			if s == 6 && d == 6 {
				continue // This is HLT.
			}
			cycles := 5
			if s == 6 || d == 6 {
				cycles = 7
			}
			ops[0x40+8*d+s] = func(c *CPU) int { put(c, getSrc(c)); return cycles }
		}
	}

	// HLT
	ops[0x76] = func(c *CPU) int { c.halted = true; return 7 }

	// MVI implementation below
	for d, put := range dst {
		cycles := 7
		if d == 6 {
			cycles = 10 // M = 10
		}
		ops[0x06+8*d] = func(c *CPU) int { put(c, c.fetchByte()); return cycles }
	}

	// LXI implementation
	ops[0x01] = func(c *CPU) int { c.setBC(c.fetchWord()); return 10 }
	ops[0x11] = func(c *CPU) int { c.setDE(c.fetchWord()); return 10 }
	ops[0x21] = func(c *CPU) int { c.setHL(c.fetchWord()); return 10 }
	ops[0x31] = func(c *CPU) int { c.SP = c.fetchWord(); return 10 }

	// LDA
	ops[0x3a] = func(c *CPU) int { c.A = c.bus.Read(c.fetchWord()); return 13 }

	// STA
	ops[0x32] = func(c *CPU) int { c.bus.Write(c.fetchWord(), c.A); return 13 }

	// LHLD
	ops[0x2a] = func(c *CPU) int { c.setHL(c.readWord(c.fetchWord())); return 16 }

	// SHLD
	ops[0x22] = func(c *CPU) int { c.writeWord(c.fetchWord(), c.hl()); return 16 }

	// LDAX
	ops[0x0A] = func(c *CPU) int { c.A = c.bus.Read(c.bc()); return 7 }
	ops[0x1A] = func(c *CPU) int { c.A = c.bus.Read(c.de()); return 7 }

	// STAX
	ops[0x02] = func(c *CPU) int { c.bus.Write(c.bc(), c.A); return 7 }
	ops[0x12] = func(c *CPU) int { c.bus.Write(c.de(), c.A); return 7 }

	// XCHG
	ops[0xEB] = func(c *CPU) int { store := c.hl(); c.setHL(c.de()); c.setDE(store); return 4 }
}
