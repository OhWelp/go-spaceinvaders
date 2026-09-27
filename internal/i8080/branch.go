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

func (c *CPU) callCore() {
	addr := c.fetchWord()
	c.push(c.PC)
	c.PC = addr
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

	// CALL
	ops[0xCD] = func(c *CPU) int { c.callCore(); return 17 }

	// Cccc implementation
	for cc := range byte(8) {
		ops[0xc4+8*cc] = func(c *CPU) int {
			if c.evalCondition(cc) {
				c.callCore()
				return 17
			}
			c.PC += 2 // Consume operand if we don't take the branch
			return 11
		}
	}

	// RET
	ops[0xc9] = func(c *CPU) int { c.PC = c.pop(); return 10 }

	for cc := range byte(8) {
		ops[0xc0+8*cc] = func(c *CPU) int {
			if c.evalCondition(cc) {
				c.PC = c.pop()
				return 11
			}
			return 5
		}
	}

	// RST
	for nn := range byte(8) {
		ops[0xc7+8*nn] = func(c *CPU) int {
			c.push(c.PC)
			c.PC = uint16(nn * 8)
			return 11
		}
	}
}
