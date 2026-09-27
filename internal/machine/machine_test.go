package machine

import (
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
