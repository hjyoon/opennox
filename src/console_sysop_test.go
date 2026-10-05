package opennox

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

func TestConsoleSysopPasswordBounds(t *testing.T) {
	c, capture := consoleCommandTestConsole(t)
	old := legacy.Nox_xxx_sysopGetPass_40A630()
	// Preserve both neighbouring native strings/scalars, without printing
	// any original password in test output.
	neighbour := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 3560)), 32)
	saved := append([]byte(nil), neighbour...)
	t.Cleanup(func() { legacy.Nox_xxx_sysopSetPass_40A610(old); copy(neighbour, saved) })
	for _, password := range []string{"AbCdEfGHi", "한글암호가나다라마", "🐺🐺🐺🐺x", ""} {
		if !consoleCommandTestExec(c, "SET SYSOP \""+password+"\"") || legacy.Nox_xxx_sysopGetPass_40A630() != password {
			t.Fatal("valid native password did not round trip")
		}
		for i := range neighbour {
			if neighbour[i] != saved[i] {
				t.Fatal("password corrupted adjacent native state")
			}
		}
	}
	legacy.Nox_xxx_sysopSetPass_40A610("fixture")
	for _, password := range []string{strings.Repeat("x", 10), strings.Repeat("🐺", 5), "x\x00y", "bad\xc0"} {
		consoleSetSysop(nil, c, []string{password})
		if legacy.Nox_xxx_sysopGetPass_40A630() != "fixture" {
			t.Fatal("invalid/oversized input changed native password")
		}
	}
	for _, line := range capture.lines {
		if strings.Contains(line, "fixture") || strings.Contains(line, "AbCdEfGHi") {
			t.Fatal("password leaked into console output")
		}
	}
}
