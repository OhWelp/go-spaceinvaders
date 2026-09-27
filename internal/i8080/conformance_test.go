package i8080

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
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
	Ports   []port    `json:"ports,omitempty"`
}

type port struct {
	Num, Value byte
	Dir        string
}

func (p *port) UnmarshalJSON(b []byte) error {
	var raw [3]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	for i, dst := range []any{&p.Num, &p.Value, &p.Dir} {
		if err := json.Unmarshal(raw[i], dst); err != nil {
			return err
		}
	}
	return nil
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

type flatBus struct {
	mem [65536]byte
	in  map[byte]byte
	log []port
}

func (b *flatBus) Read(a uint16) byte     { return b.mem[a] }
func (b *flatBus) Write(a uint16, v byte) { b.mem[a] = v }
func (b *flatBus) In(p byte) byte {
	v := b.in[p]
	b.log = append(b.log, port{p, v, "r"})
	return v
}
func (b *flatBus) Out(p, v byte) { b.log = append(b.log, port{p, v, "w"}) }

func loadPorts(bus *flatBus, ports []port) {
	for _, p := range ports {
		if p.Dir == "r" {
			if bus.in == nil {
				bus.in = make(map[byte]byte, len(ports))
			}
			bus.in[p.Num] = p.Value
		}
	}
}

func checkPorts(t *testing.T, got, want []port) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("ports: got %v, want %v", got, want)
	}
}

func loadState(c *CPU, bus *flatBus, s testState) {
	c.PC, c.SP = s.PC, s.SP
	c.A, c.B, c.C, c.D, c.E, c.H, c.L = s.A, s.B, s.C, s.D, s.E, s.H, s.L
	c.setPSW(s.F)
	// The suite models the EI delay as an inhibit flag: EI sets inte at once and
	// ei_pending suppresses acknowledgement for one instruction. This CPU defers
	// instead (eiPending promotes to intEnabled after the next instruction), so
	// inte=1,ei_pending=1 loads as intEnabled=false. See testdata/tests_readme.md.
	c.eiPending = s.EIPending != 0
	c.intEnabled = s.Inte != 0 && !c.eiPending
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
	// Inverse of the load mapping above: a deferred EI still reads as inte set.
	inte := c.intEnabled || c.eiPending
	if inte != (want.Inte != 0) || c.eiPending != (want.EIPending != 0) || c.halted != (want.Halted != 0) {
		t.Errorf("inte/ei_pending/halted: got %v/%v/%v, want %v/%v/%v",
			inte, c.eiPending, c.halted,
			want.Inte != 0, want.EIPending != 0, want.Halted != 0)
	}
	for _, r := range want.RAM {
		if bus.mem[r[0]] != byte(r[1]) {
			t.Errorf("ram[%04X]: got %02X, want %02X", r[0], bus.mem[r[0]], byte(r[1]))
		}
	}
}

// Subtest outcome counts, kept per opcode and summed for the whole run.
type tally struct{ pass, skip, fail int }

// record classifies a finished subtest. Failed is checked first so a case that
// errors and then skips still counts against us.
func (ta *tally) record(t *testing.T) {
	switch {
	case t.Failed():
		ta.fail++
	case t.Skipped():
		ta.skip++
	default:
		ta.pass++
	}
}

func (ta *tally) add(o tally) {
	ta.pass += o.pass
	ta.skip += o.skip
	ta.fail += o.fail
}

func (ta tally) total() int { return ta.pass + ta.skip + ta.fail }

func TestConformance(t *testing.T) {
	data, err := os.ReadFile("testdata/i8080_tests.json")
	if err != nil {
		t.Fatalf("reading tests: %v", err)
	}
	var tf testFile
	if err := json.Unmarshal(data, &tf); err != nil {
		t.Fatalf("parsing tests: %v", err)
	}

	var total tally
	var skippedOps, failedOps []string

	for opcode, cases := range tf.Tests {
		var counts tally
		t.Run(opcode, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.Name, func(t *testing.T) {
					t.Cleanup(func() { counts.record(t) })
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
					loadPorts(bus, tc.Ports)

					cycles := c.Step()

					if cycles != tc.Cycles {
						t.Errorf("cycles: got %d, want %d", cycles, tc.Cycles)
					}
					checkState(t, c, bus, tc.Final)
					checkPorts(t, bus.log, tc.Ports)
				})
			}
		})

		total.add(counts)
		switch {
		case counts.fail > 0:
			failedOps = append(failedOps, fmt.Sprintf("%s(%d)", opcode, counts.fail))
		case len(cases) > 0 && counts.skip == len(cases):
			skippedOps = append(skippedOps, opcode)
		}
	}

	// Printed rather than logged: t.Log output is hidden without -v, and the
	// summary is the point of running the suite.
	fmt.Printf("conformance: %d cases, %d passed, %d skipped, %d failed\n",
		total.total(), total.pass, total.skip, total.fail)
	if len(skippedOps) > 0 {
		slices.Sort(skippedOps)
		fmt.Printf("  unimplemented: %s\n", strings.Join(skippedOps, " "))
	}
	if len(failedOps) > 0 {
		slices.Sort(failedOps)
		fmt.Printf("  failing: %s\n", strings.Join(failedOps, " "))
	}
}
