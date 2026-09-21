package i8080

func opsControl() {
	ops[0x00] = func(c *CPU) int { return 4 } // NOP
	ops[0xC3] = func(c *CPU) int {            // JMP a16
		c.PC = c.fetchWord()
		return 10
	}
}
