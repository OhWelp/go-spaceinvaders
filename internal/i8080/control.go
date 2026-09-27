package i8080

func opsControl() {
	ops[0x00] = func(c *CPU) int { return 4 } // NOP
}
