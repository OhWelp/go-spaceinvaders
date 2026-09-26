package i8080

import "math/bits"

func (c *CPU) ral() {
	var carryIn byte

	if c.Carry {
		carryIn = 1
	}
	c.Carry = c.A&0x80 != 0
	c.A = c.A<<1 | carryIn
}

func (c *CPU) rar() {
	var carryIn byte

	if c.Carry {
		carryIn = 0x80
	}

	c.Carry = c.A&0x01 != 0
	c.A = c.A>>1 | carryIn
}

func (c *CPU) daa() {
	overflow1 := false
	lowAdjust := c.A&0x0F > 9 || c.Aux
	if lowAdjust {
		overflow1 = uint16(c.A)+6 > 0xFF
		c.Aux = c.A&0x0F > 9
		c.A += 0x06
	} else {
		c.Aux = false
	}
	if c.A>>4 > 9 || c.Carry || overflow1 {
		if uint16(c.A)+0x60 > 0xFF || overflow1 {
			c.Carry = true
		}
		c.A += 0x60
	}
	c.setSZP(c.A)
}

func opsAcc() {

	// RLC
	ops[0x07] = func(c *CPU) int { c.Carry = (c.A>>7)&1 == 1; c.A = bits.RotateLeft8(c.A, 1); return 4 }
	// RRC
	ops[0x0F] = func(c *CPU) int { c.Carry = c.A&1 == 1; c.A = bits.RotateLeft8(c.A, -1); return 4 }

	// RAL and RAR
	ops[0x17] = func(c *CPU) int { c.ral(); return 4 }
	ops[0x1F] = func(c *CPU) int { c.rar(); return 4 }

	// DAA
	ops[0x27] = func(c *CPU) int { c.daa(); return 4 }

	// CMA
	ops[0x2F] = func(c *CPU) int { c.A = ^c.A; return 4 }

	// STC
	ops[0x37] = func(c *CPU) int { c.Carry = true; return 4 }

	// CMC
	ops[0x3F] = func(c *CPU) int { c.Carry = !c.Carry; return 4 }
}
