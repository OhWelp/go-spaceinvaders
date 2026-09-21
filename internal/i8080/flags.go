package i8080

import "math/bits"

// set sign, zero, and parity from result byte
func (c *CPU) setSZP(b byte) {
	c.Sign = b&0x80 != 0
	c.Zero = b == 0
	c.Parity = bits.OnesCount8(b)%2 == 0
}

// materializes the flag byte from our flags
func (c *CPU) psw() byte {
	b := byte(0x02) // bit 1 is always set

	if c.Sign {
		b |= 0x80
	}
	if c.Zero {
		b |= 0x40
	}
	if c.Aux {
		b |= 0x10
	}
	if c.Parity {
		b |= 0x04
	}
	if c.Carry {
		b |= 0x01
	}
	return b
}

// scatter flags byte into our flag fields
func (c *CPU) setPSW(b byte) {
	c.Sign = b&0x80 != 0
	c.Zero = b&0x40 != 0
	c.Aux = b&0x10 != 0
	c.Parity = b&0x04 != 0
	c.Carry = b&0x01 != 0
}
