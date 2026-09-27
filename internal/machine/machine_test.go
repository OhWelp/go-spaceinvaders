package machine

import (
	"hash/crc32"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}
	return filepath.Dir(strings.TrimSpace(string(out)))
}

func loadROM(t *testing.T) []byte {
	t.Helper()
	dir := filepath.Join(moduleRoot(t), "game-data")
	var rom []byte
	// Load order is h,g,f,e — h maps to 0x0000. Reverse-alphabetical!
	for _, part := range []string{"invaders.h", "invaders.g", "invaders.f", "invaders.e"} {
		b, err := os.ReadFile(filepath.Join(dir, part))
		if err != nil {
			t.Skipf("ROM part %s not available: %v", part, err)
		}
		rom = append(rom, b...)
	}
	return rom
}

func TestAttractMode(t *testing.T) {
	m, err := New(loadROM(t))
	if err != nil {
		t.Fatal(err)
	}

	for range 600 { // ~10 seconds of attract mode
		m.RunFrame()
	}

	vram := m.mem[0x2400:0x4000]
	lit := 0
	for _, b := range vram {
		if b != 0 {
			lit++
		}
	}
	if lit < 500 {
		t.Fatalf("VRAM nearly empty after 600 frames (%d nonzero bytes) — emulation likely stalled", lit)
	}
	t.Logf("VRAM: %d/%d bytes lit, checksum %08X", lit, len(vram), crc32.ChecksumIEEE(vram))
}
