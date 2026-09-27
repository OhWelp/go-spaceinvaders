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

// Builds our dispatch table on initialization
func init() {
	opsControl() // NOP, JMP, HLT... (control.go)
	opsArith()   // Math family (arith.go)
	opsMove()    // MOV family (move.go)
	opsAcc()     // Accumulator bit operations (acc.go)
	opsCond()    // JMP, calls/returns, etc (branch.go)
	opsStack()   // Stack operations (stack.go)

	// Undocumented aliases

	//NOP
	for _, op := range []byte{0x08, 0x10, 0x18, 0x20, 0x28, 0x30, 0x38} {
		ops[op] = ops[0x00]
	}

	ops[0xCB] = ops[0xC3] // JMP a16
	ops[0xD9] = ops[0xC9] // RET
	for _, op := range []byte{0xDD, 0xED, 0xFD} {
		ops[op] = ops[0xCD] // CALL a16
	}

	for i, h := range ops {
		if h == nil {
			ops[i] = unimplemented(byte(i))
		}
	}
}

type errUnimplemented byte

func (op errUnimplemented) Error() string {
	return fmt.Sprintf("unimplemented opcode %02X", byte(op))
}

func unimplemented(op byte) handler {
	return func(c *CPU) int {
		panic(errUnimplemented(op))
	}
}
