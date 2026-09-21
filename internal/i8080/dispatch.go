package i8080

import "fmt"

// handler executes an opcode and returns cycle count
type handler func(c *CPU) int

var ops [256]handler

func init() {
	opsControl() // NOP, JMP, HLT... (control.go)
	opsArith()   // ADD family (arith.go)

	for i, h := range ops {
		if h == nil {
			ops[i] = unimplemented(byte(i))
		}
	}
}

func unimplemented(op byte) handler {
	return func(c *CPU) int {
		panic(fmt.Sprintf("unimplemented opcode %02X at %04X", op, c.PC-1))
	}
}
