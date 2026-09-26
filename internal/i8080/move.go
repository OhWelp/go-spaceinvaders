package i8080

func opsMove() {
	for s, getSrc := range src { // This implements MOV in its entirety.
		for d, put := range dst {
			if s == 6 && d == 6 {
				continue // This is HLT.
			}
			getSrc, put := getSrc, put
			cycles := 5
			if s == 6 || d == 6 {
				cycles = 7
			}
			ops[0x40+8*d+s] = func(c *CPU) int { put(c, getSrc(c)); return cycles }
		}
	}
}
