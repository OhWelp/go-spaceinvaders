package i8080

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

// Implement ANA/ANI
func (c *CPU) ana(v byte) {
	c.Aux = ((c.A|v)>>3)&1 == 1
	c.A = c.A & v
	c.Carry = false
	c.setSZP(c.A)
}

// Implement XRA/XRI
func (c *CPU) xra(v byte) {
	c.A = c.A ^ v
	c.Aux = false
	c.Carry = false
	c.setSZP(c.A)
}

// Implement ORA/ORI
func (c *CPU) ora(v byte) {
	c.A = c.A | v
	c.Aux = false
	c.Carry = false
	c.setSZP(c.A)
}

// Implement CMP/CPI
func (c *CPU) cmp(v byte) {
	a := c.A
	c.sub(v)
	c.A = a
}

// INR
func (c *CPU) inr(v byte) byte {

	c.Aux = v&0x0f == 0xF

	result := v + byte(1)
	c.setSZP(result)
	return result
}

// DCR
func (c *CPU) dcr(v byte) byte {

	c.Aux = v&0x0f != 0

	result := v - byte(1)
	c.setSZP(result)
	return result
}

// DAD
func (c *CPU) dad(v uint16) uint16 {
	c.Carry = int(c.hl())+int(v) > 65535
	return (c.hl() + v)
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
		ops[0xA0+s] = func(c *CPU) int { c.ana(get(c)); return cycles } // ANA
		ops[0xA8+s] = func(c *CPU) int { c.xra(get(c)); return cycles } // XRA
		ops[0xB0+s] = func(c *CPU) int { c.ora(get(c)); return cycles } // ORA
		ops[0xB8+s] = func(c *CPU) int { c.cmp(get(c)); return cycles } // CMP
	}

	// INR/DCR
	for s, get := range src {
		put := dst[s]
		cycles := 5
		if s == 6 {
			cycles = 10
		}
		ops[0x04+(8*s)] = func(c *CPU) int { put(c, c.inr(get(c))); return cycles }
		ops[0x05+(8*s)] = func(c *CPU) int { put(c, c.dcr(get(c))); return cycles }
	}

	ops[0xC6] = func(c *CPU) int { c.add(c.fetchByte()); return 7 } // ADI
	ops[0xCE] = func(c *CPU) int { c.adc(c.fetchByte()); return 7 } // ACI
	ops[0xD6] = func(c *CPU) int { c.sub(c.fetchByte()); return 7 } // SUI
	ops[0xDE] = func(c *CPU) int { c.sbb(c.fetchByte()); return 7 } // SBI
	ops[0xE6] = func(c *CPU) int { c.ana(c.fetchByte()); return 7 } // ANI
	ops[0xEE] = func(c *CPU) int { c.xra(c.fetchByte()); return 7 } // XRI
	ops[0xF6] = func(c *CPU) int { c.ora(c.fetchByte()); return 7 } // ORI
	ops[0xFE] = func(c *CPU) int { c.cmp(c.fetchByte()); return 7 } // CPI

	// INX/DCX
	ops[0x03] = func(c *CPU) int { c.setBC(c.bc() + 1); return 5 }
	ops[0x13] = func(c *CPU) int { c.setDE(c.de() + 1); return 5 }
	ops[0x23] = func(c *CPU) int { c.setHL(c.hl() + 1); return 5 }
	ops[0x33] = func(c *CPU) int { c.SP += 1; return 5 }
	ops[0x0B] = func(c *CPU) int { c.setBC(c.bc() - 1); return 5 }
	ops[0x1B] = func(c *CPU) int { c.setDE(c.de() - 1); return 5 }
	ops[0x2B] = func(c *CPU) int { c.setHL(c.hl() - 1); return 5 }
	ops[0x3B] = func(c *CPU) int { c.SP -= 1; return 5 }

	// DAD
	ops[0x09] = func(c *CPU) int { c.setHL(c.dad(c.bc())); return 10 }
	ops[0x19] = func(c *CPU) int { c.setHL(c.dad(c.de())); return 10 }
	ops[0x29] = func(c *CPU) int { c.setHL(c.dad(c.hl())); return 10 }
	ops[0x39] = func(c *CPU) int { c.setHL(c.dad(c.SP)); return 10 }

}
