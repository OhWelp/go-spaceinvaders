package i8080

func opsControl() {
	ops[0x00] = func(c *CPU) int { return 4 } // NOP

	// IN/OUT
	ops[0xDB] = func(c *CPU) int { c.A = c.bus.In(c.fetchByte()); return 10 }
	ops[0xD3] = func(c *CPU) int { c.bus.Out(c.fetchByte(), c.A); return 10 }
}
