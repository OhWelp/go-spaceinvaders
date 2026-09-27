# Intel 8080 single-step tests

`i8080_tests.json` contains 1,323,952 single-instruction tests for the Intel 8080 in the style of
[SingleStepTests](https://github.com/SingleStepTests) (Tom Harte / raddad772). Each test records a
complete CPU state, the state after executing exactly one instruction (or accepting one interrupt),
the number of T-states consumed, and any port I/O. There was no 8080 set in the SingleStepTests
organisation when this was made, so this file fills that gap.

## Provenance and validation

- **Reference emulator:** [superzazu/8080](https://github.com/superzazu/8080), commit
  `274ffd700b81baabea99b0963bc1260b67132185`, a ~800-line C99 core.
- **Reference validation:** before generation the core was run against the four CP/M test ROMs and
  passed all of them: `TST8080.COM`, `CPUTEST.COM` (including its timing test), `8080PRE.COM` and
  `8080EXM.COM`. 8080EXM is the 8080 port of zexall; its expected CRCs were recorded from real 8080
  silicon, so the flag behaviour that most 8080 emulators get wrong (auxiliary carry on subtraction,
  the ANA auxiliary-carry rule, DAA, the fixed bits of F) is hardware-anchored.
- **Cycle counts** follow the Intel 8080 datasheet. No test ROM checks them precisely; CPUTEST's
  timing test is coarse. This is the least externally anchored part of the file.
- **Independent replay:** every test was replayed by a second 8080 model written from the datasheet
  using opcode-field decoding, its own cycle rules and its own flag formulas. Zero mismatches. The
  checker was itself verified by planting errors in copies of the file and confirming they were caught.
- **Deterministic:** generated with a splitmix64 RNG, seed 32896, by a small C program that wraps the
  reference core in a sparse, access-tracking 64 KB memory. Regenerating yields a byte-identical file.

## File layout

The file is standard JSON, laid out so line-oriented tools work on it:

```
{
  "meta": { ... a few lines ... },
  "tests": {
    "00": [
      {"name": "00 r0000", "initial": {...}, "final": {...}, "cycles": 4},
      ...
    ],
    ...
    "ff": [ ... ]
  },
  "interrupts": [
    {"name": "int c7 taken 0000", "vector": 199, "initial": {...}, "final": {...}, "cycles": 11},
    ...
  ]
}
```

- Every test is exactly one line, indented six spaces, ending in `},` or `}`.
- Each opcode array opens with a line `    "xx": [` and closes with a line `    ]`.
- The interrupt array opens with `  "interrupts": [`.
- `meta.counts.total` is the number of tests in the file.

Size: about 480 MB, 1.32 million lines. Editors will not open it comfortably; see the shell recipes below.

## Schema

### Test object

| Field | Type | Meaning |
|---|---|---|
| `name` | string | `xx eNNNNNN` (exhaustive sweep), `xx rNNNN` (random) or `int vv <case> NNNN` (interrupt). |
| `vector` | int | Interrupt tests only. The byte on the data bus during the interrupt acknowledge. Always an RST opcode. |
| `initial` | state | CPU state before the step. |
| `final` | state | CPU state after the step. |
| `cycles` | int | T-states consumed by the step. |
| `ports` | list | Only present when the instruction performs I/O. `[[port, value, "r"]]` for IN, `[[port, value, "w"]]` for OUT, in execution order. `port` is the 8-bit port number from the instruction operand. |

### State object

| Field | Type | Meaning |
|---|---|---|
| `pc`, `sp` | 16-bit | Program counter and stack pointer. |
| `a` `b` `c` `d` `e` `h` `l` | 8-bit | Registers. |
| `f` | 8-bit | Flags. Bit 7 S, bit 6 Z, bit 5 always 0, bit 4 AC, bit 3 always 0, bit 2 P (1 = even parity), bit 1 always 1, bit 0 CY. Every initial and final value is normalised this way. |
| `inte` | 0/1 | Interrupt-enable flip-flop. |
| `ei_pending` | 0/1 | 1 when the previous instruction was EI. Interrupts are inhibited until after this instruction completes. EI sets it; the next executed instruction clears it. |
| `halted` | 0/1 | 1 after HLT. Opcode tests always start with 0. |
| `ram` | list | `[[address, value], ...]` sorted by address. Lists exactly the addresses the instruction reads or writes, including the opcode and operand bytes at PC. Addresses not listed are never accessed and are unconstrained. Initial and final states list the same addresses. |

Access order inside an instruction is not asserted. Final values of every listed address are.

### Cycle counts

| Instruction group | T-states |
|---|---|
| MOV r,r | 5 |
| MOV r,M / MOV M,r / ALU M / MVI r | 7 |
| ALU r / rotates / DAA CMA STC CMC / XCHG / EI DI / NOP | 4 |
| INR r / DCR r / INX / DCX / PCHL / SPHL | 5 |
| INR M / DCR M / MVI M / LXI / DAD / POP / JMP / Jcc (always) / RET / IN / OUT | 10 |
| STA / LDA | 13 |
| SHLD / LHLD | 16 |
| XTHL | 18 |
| PUSH / RST / interrupt acknowledge | 11 |
| CALL | 17 |
| Ccc | 11 not taken, 17 taken |
| Rcc | 5 not taken, 11 taken |
| HLT | 7 |

### Undocumented opcodes

`0x08 0x10 0x18 0x20 0x28 0x30 0x38` execute as NOP (4 cycles). `0xcb` is JMP, `0xd9` is RET,
`0xdd 0xed 0xfd` are CALL, all with the documented instruction's timing. All 256 opcodes have tests.

## Interrupt tests

Each interrupt test asserts INT with `vector` on the data bus, then performs one step.

- If `inte` is 1 and `ei_pending` is 0, the CPU acknowledges: `inte` becomes 0, a halted CPU
  leaves the halt state, and the vector executes as an instruction without any fetch from memory.
  PC is not incremented before being pushed. Only RST vectors are used, so the effect is a push of PC
  and a jump to `n * 8`, costing 11 cycles.
- Otherwise the instruction at PC executes normally and the pending interrupt has no visible effect.

Cases, 8 RST vectors each:

| Case | Count | Initial state |
|---|---|---|
| `taken` | 100 | `inte` 1, not halted. |
| `halted` | 100 | `inte` 1, halted; the pushed PC already points past the HLT. |
| `ei` | 60 | `inte` 1, `ei_pending` 1. Instruction at PC runs instead: 10 EI, 10 DI, 40 random. |
| `disabled` | 50 | `inte` 0. Instruction at PC runs instead: 10 EI, 40 random. |

A halted CPU with interrupts disabled is never tested because it makes no observable progress.

If your emulator models the EI delay as "INTE becomes set after the next instruction" instead of an
inhibit flag, map it in the harness: report `inte = INTE || pending` and `ei_pending = pending`, and
load `inte=1, ei_pending=1` as `INTE=0, pending=1`. Note that EI followed by DI must end with
interrupts disabled.

## Coverage tiers

| Tier | Opcodes | Inputs swept | Tests |
|---|---|---|---|
| ALU immediate | ADI ACI SUI SBI ANI XRI ORI CPI | every A × operand × CY | 8 × 131,072 = 1,048,576 |
| DAA | 27 | every A × AC × CY | 1,024 |
| INR / DCR | all 16 forms including M | every value × CY | 16 × 512 = 8,192 |
| A-only | RLC RRC RAL RAR CMA STC CMC, ADD/ADC/SUB/SBB/ANA/XRA/ORA/CMP A | every A × CY | 15 × 512 = 7,680 |
| Random | all 256 opcodes | random state | 256 × 1,000 = 256,000 |
| Interrupts | 8 RST vectors | see above | 2,480 |

In every sweep the registers not being swept, PC, SP, the remaining flag bits and `inte` are random.
Within an opcode's array the exhaustive tests come first, then the random ones. The first 32 random
tests of every opcode are seeded with edge cases: PC, SP, HL, BC and DE at 0x0000, 0x0001, 0xfffe,
0xffff and neighbours, so 16-bit wraparound of fetches, stack operations and memory operands is
exercised for every instruction.

## Running a test

1. Load `initial`: registers, `f`, `inte`, `ei_pending`, `halted`, and each `ram` pair into memory.
   All other memory can hold anything.
2. If the test has `vector`, assert an interrupt with that byte, then step once. Otherwise step once.
3. On IN, return the value from `ports` whose port number matches. On OUT, record port and value.
4. Compare every field of `final`, every `ram` pair, `cycles`, and the recorded ports.
5. Optionally treat any read or write of an address not in `ram` as a failure. The listed set is
   exactly what the instruction touches.

## Shell recipes

```bash
head -n 9  i8080_tests.json                                  # the meta block
grep '"name": "c6 e000512"' i8080_tests.json                 # one test
sed -n '/^    "db": \[/,/^    \]/p' i8080_tests.json         # every test for opcode 0xdb
grep -c '"name": "27 ' i8080_tests.json                      # how many DAA tests
grep -n '"interrupts": \[' i8080_tests.json                  # where the interrupt tests start
```

## Loading in Go

These types decode the whole file with `encoding/json` in about 4.5 seconds using roughly 0.9 GB of
heap. Load once (for example in `TestMain`) and run subtests per opcode.

```go
type State struct {
	PC        uint16      `json:"pc"`
	SP        uint16      `json:"sp"`
	A         uint8       `json:"a"`
	B         uint8       `json:"b"`
	C         uint8       `json:"c"`
	D         uint8       `json:"d"`
	E         uint8       `json:"e"`
	F         uint8       `json:"f"`
	H         uint8       `json:"h"`
	L         uint8       `json:"l"`
	INTE      uint8       `json:"inte"`
	EIPending uint8       `json:"ei_pending"`
	Halted    uint8       `json:"halted"`
	RAM       [][2]uint16 `json:"ram"` // [address, value], sorted by address
}

type Port struct {
	Port  uint8
	Value uint8
	Dir   string // "r" (IN) or "w" (OUT)
}

func (p *Port) UnmarshalJSON(b []byte) error {
	var raw [3]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	p.Port = uint8(raw[0].(float64))
	p.Value = uint8(raw[1].(float64))
	p.Dir = raw[2].(string)
	return nil
}

type Test struct {
	Name    string `json:"name"`
	Vector  *uint8 `json:"vector,omitempty"` // interrupt tests only
	Initial State  `json:"initial"`
	Final   State  `json:"final"`
	Cycles  uint64 `json:"cycles"`
	Ports   []Port `json:"ports,omitempty"`
}

type Suite struct {
	Meta       json.RawMessage   `json:"meta"`
	Tests      map[string][]Test `json:"tests"` // key: opcode as two lowercase hex digits
	Interrupts []Test            `json:"interrupts"`
}

func loadSuite(path string) (*Suite, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var s Suite
	if err := json.NewDecoder(bufio.NewReaderSize(f, 1<<20)).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}
```

Because every test is one line, a streaming loader is also easy: scan the file line by line, and for
each line that starts with six spaces and `{`, trim the trailing comma and unmarshal it into `Test`.
Lines of the form `    "xx": [` announce the opcode of the tests that follow, and
`  "interrupts": [` announces the interrupt section. This takes the same time as the full decode but
keeps memory flat.

## What is not covered

- T-state level bus traces. The reference is not cycle-by-cycle accurate at the bus level and the
  Space Invaders hardware has no mid-instruction side effects, so only per-instruction totals are given.
- A halted CPU with interrupts disabled.
- Interrupt vectors other than RST n.
- Precise hardware validation of cycle counts (see Provenance).
