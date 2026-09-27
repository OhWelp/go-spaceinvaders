package i8080

func (c *CPU) evalCondition(b byte) bool {
	switch b {
	case 0:
		return c.Zero == false
	case 1:
		return c.Zero == true
	case 2:
		return c.Carry == false
	case 3:
		return c.Carry == true
	case 4:
		return c.Parity == false
	case 5:
		return c.Parity == true
	case 6:
		return c.Sign == false
	case 7:
		return c.Sign == true
	}
	return false
}

func opsCond() {

	// JCC implementation
	for cc := range byte(8) {
		ops[0xC2+8*cc] = func(c *CPU) int {
			addr := c.fetchWord()
			if c.evalCondition(cc) {
				c.PC = addr
			}
			return 10
		}
	}

	// JMP a16
	ops[0xC3] = func(c *CPU) int { c.PC = c.fetchWord(); return 10 }

	// PCHL
	ops[0xE9] = func(c *CPU) int { c.PC = c.hl(); return 5 }

}
