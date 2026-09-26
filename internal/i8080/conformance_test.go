package i8080

import (
	"encoding/json"
	"os"
	"testing"
)

type testFile struct {
	Meta  json.RawMessage       `json:"meta"`
	Tests map[string][]testCase `json:"tests"`
}

type testCase struct {
	Name    string    `json:"name"`
	Initial testState `json:"initial"`
	Final   testState `json:"final"`
	Cycles  int       `json:"cycles"`
}

type testState struct {
	PC               uint16 `json:"pc"`
	SP               uint16 `json:"sp"`
	A, B, C, D, E, F byte
	H, L             byte
	Inte             int      `json:"inte"`
	EIPending        int      `json:"ei_pending"`
	Halted           int      `json:"halted"`
	RAM              [][2]int `json:"ram"`
}

type flatBus struct{ mem [65536]byte }

func (b *flatBus) Read(a uint16) byte     { return b.mem[a] }
func (b *flatBus) Write(a uint16, v byte) { b.mem[a] = v }
func (b *flatBus) In(p byte) byte         { return 0 }
func (b *flatBus) Out(p, v byte)          {}

func loadState(c *CPU, bus *flatBus, s testState) {
	c.PC, c.SP = s.PC, s.SP
	c.A, c.B, c.C, c.D, c.E, c.H, c.L = s.A, s.B, s.C, s.D, s.E, s.H, s.L
	c.setPSW(s.F)
	c.intEnabled = s.Inte != 0
	c.eiPending = s.EIPending != 0
	c.halted = s.Halted != 0
	for _, r := range s.RAM {
		bus.mem[r[0]] = byte(r[1])
	}
}

func checkState(t *testing.T, c *CPU, bus *flatBus, want testState) {
	t.Helper()
	if c.PC != want.PC || c.SP != want.SP {
		t.Errorf("pc/sp: got %04X/%04X, want %04X/%04X", c.PC, c.SP, want.PC, want.SP)
	}
	got := [7]byte{c.A, c.B, c.C, c.D, c.E, c.H, c.L}
	exp := [7]byte{want.A, want.B, want.C, want.D, want.E, want.H, want.L}
	if got != exp {
		t.Errorf("regs: got % 02X, want % 02X", got, exp)
	}
	if c.psw() != want.F {
		t.Errorf("flags: got %08b, want %08b", c.psw(), want.F)
	}
	if c.intEnabled != (want.Inte != 0) || c.halted != (want.Halted != 0) {
		t.Errorf("inte/halted: got %v/%v, want %v/%v",
			c.intEnabled, c.halted, want.Inte != 0, want.Halted != 0)
	}
	for _, r := range want.RAM {
		if bus.mem[r[0]] != byte(r[1]) {
			t.Errorf("ram[%04X]: got %02X, want %02X", r[0], bus.mem[r[0]], byte(r[1]))
		}
	}
}

func TestConformance(t *testing.T) {
	data, err := os.ReadFile("testdata/i8080_tests.json")
	if err != nil {
		t.Fatalf("reading tests: %v", err)
	}
	var tf testFile
	if err := json.Unmarshal(data, &tf); err != nil {
		t.Fatalf("parsing tests: %v", err)
	}

	for opcode, cases := range tf.Tests {
		t.Run(opcode, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.Name, func(t *testing.T) {
					defer func() {
						if r := recover(); r != nil {
							if op, ok := r.(errUnimplemented); ok {
								t.Skipf("unimplemented opcode %02X", byte(op))
							}
							panic(r) // anything else is a real bug; let it fly
						}
					}()
					bus := &flatBus{}
					c := New(bus)
					loadState(c, bus, tc.Initial)

					cycles := c.Step()

					if cycles != tc.Cycles {
						t.Errorf("cycles: got %d, want %d", cycles, tc.Cycles)
					}
					checkState(t, c, bus, tc.Final)
				})
			}
		})
	}

}
