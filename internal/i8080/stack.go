package i8080

func (c *CPU) push(v uint16) {
	c.SP -= 2
	c.writeWord(c.SP, v)
}

func (c *CPU) pop() uint16 {
	word := c.readWord(c.SP)
	c.SP += 2
	return word
}

func opsStack() {
	// PUSH operations
	ops[0xc5] = func(c *CPU) int { c.push(c.bc()); return 11 }
	ops[0xd5] = func(c *CPU) int { c.push(c.de()); return 11 }
	ops[0xe5] = func(c *CPU) int { c.push(c.hl()); return 11 }
	ops[0xf5] = func(c *CPU) int { c.push(uint16(c.A)<<8 | uint16(c.psw())); return 11 }

	// POP operations
	ops[0xc1] = func(c *CPU) int { c.setBC(c.pop()); return 10 }
	ops[0xd1] = func(c *CPU) int { c.setDE(c.pop()); return 10 }
	ops[0xe1] = func(c *CPU) int { c.setHL(c.pop()); return 10 }
	ops[0xf1] = func(c *CPU) int { v := c.pop(); c.A = byte(v >> 8); c.setPSW(byte(v)); return 10 }

	// XTHL
	ops[0xe3] = func(c *CPU) int { swap := c.hl(); c.setHL(c.readWord(c.SP)); c.writeWord(c.SP, swap); return 18 }

	// SPHL
	ops[0xf9] = func(c *CPU) int { c.SP = c.hl(); return 5 }
}
