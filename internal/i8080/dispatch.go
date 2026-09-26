package i8080

import "fmt"

// handler executes an opcode and returns cycle count
type handler func(c *CPU) int

var ops [256]handler

type reader func(*CPU) byte
type writer func(*CPU, byte)

var src = []reader{
	func(c *CPU) byte { return c.B },                // 000
	func(c *CPU) byte { return c.C },                // 001
	func(c *CPU) byte { return c.D },                // 010
	func(c *CPU) byte { return c.E },                // 011
	func(c *CPU) byte { return c.H },                // 100
	func(c *CPU) byte { return c.L },                // 101
	func(c *CPU) byte { return c.bus.Read(c.hl()) }, // 110 = M
	func(c *CPU) byte { return c.A },                // 111
}

var dst = []writer{
	func(c *CPU, v byte) { c.B = v },
	func(c *CPU, v byte) { c.C = v },
	func(c *CPU, v byte) { c.D = v },
	func(c *CPU, v byte) { c.E = v },
	func(c *CPU, v byte) { c.H = v },
	func(c *CPU, v byte) { c.L = v },
	func(c *CPU, v byte) { c.bus.Write(c.hl(), v) }, // M
	func(c *CPU, v byte) { c.A = v },
}

func init() {
	opsControl() // NOP, JMP, HLT... (control.go)
	opsArith()   // Math family (arith.go)
	opsMove()    // MOV family
	opsAcc()     // Accumulator bit operations

	for i, h := range ops {
		if h == nil {
			ops[i] = unimplemented(byte(i))
		}
	}
}

type errUnimplemented byte

func unimplemented(op byte) handler {
	return func(c *CPU) int {
		panic(fmt.Sprintf("unimplemented opcode %02X at %04X", op, c.PC-1))
	}
}
