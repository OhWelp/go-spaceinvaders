package i8080

var src = []func(*CPU) byte{
	func(c *CPU) byte { return c.B },                // 000
	func(c *CPU) byte { return c.C },                // 001
	func(c *CPU) byte { return c.D },                // 010
	func(c *CPU) byte { return c.E },                // 011
	func(c *CPU) byte { return c.H },                // 100
	func(c *CPU) byte { return c.L },                // 101
	func(c *CPU) byte { return c.bus.Read(c.hl()) }, // 110 = M
	func(c *CPU) byte { return c.A },                // 111
}

// implement ADD/ADI/ADC
func (c *CPU) addCore(v byte, carry byte) {
	sum := uint16(c.A) + uint16(v) + uint16(carry)
	c.Carry = sum > 0xFF
	c.Aux = (c.A&0x0F)+(v&0x0F)+carry > 0x0F
	c.A = byte(sum)
	c.setSZP(c.A)
}

func (c *CPU) add(v byte) {
	c.addCore(v, 0)
}

func (c *CPU) adc(v byte) {
	var carry byte
	if c.Carry {
		carry = 1
	}
	c.addCore(v, carry)
}

// Implement SUB/SBI/SBB/SBI
func (c *CPU) subCore(v byte, carry byte) {
	result := byte(uint16(c.A) - (uint16(v) + uint16(carry)))
	c.Carry = uint16(v)+uint16(carry) > uint16(c.A)
	c.Aux = c.A&0x0F >= (v&0x0F)+carry
	c.A = byte(result)
	c.setSZP(c.A)
}

func (c *CPU) sbb(v byte) {
	var carry byte
	if c.Carry {
		carry = 1
	}
	c.subCore(v, carry)
}

func (c *CPU) sub(v byte) {
	c.subCore(v, 0)
}

func opsArith() {
	for s, get := range src {
		cycles := 4
		if s == 6 {
			cycles = 7 // 6 = M = is a memory read
		}
		ops[0x80+s] = func(c *CPU) int { c.add(get(c)); return cycles } // ADD
		ops[0x88+s] = func(c *CPU) int { c.adc(get(c)); return cycles } // ADC
		ops[0x90+s] = func(c *CPU) int { c.sub(get(c)); return cycles } // SUB
		ops[0x98+s] = func(c *CPU) int { c.sbb(get(c)); return cycles } // SBB
	}
	ops[0xC6] = func(c *CPU) int { c.add(c.fetchByte()); return 7 } // ADI
	ops[0xCE] = func(c *CPU) int { c.adc(c.fetchByte()); return 7 } // ACI
	ops[0xD6] = func(c *CPU) int { c.sub(c.fetchByte()); return 7 } // SUI
	ops[0xDE] = func(c *CPU) int { c.sbb(c.fetchByte()); return 7 } // SBI
}
