package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

// Check the preserved C functions and linked Go entries against the five
// scalar records, including duplicate IDs and all frame boundary values.
func clientAbilityStateNative461090(t *testing.T, mode string) {
	t.Helper()
	fixture, err := os.ReadFile("testdata/client_ability_state_test.c")
	if err != nil {
		t.Fatal(err)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	bin := filepath.Join(t.TempDir(), "client-ability-state")
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-fsigned-char", "-fno-strict-aliasing",
		"-Wall", "-Wextra", "-Werror", "-Wno-unused-function", "-Itestdata", "-x", "c", "-", "-o", bin)
	cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
	cmd.Stdin = strings.NewReader(string(fixture))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile actual client ability-state subsystem: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin, mode).CombinedOutput(); err != nil {
		t.Fatalf("actual client ability %s contracts: %v\n%s", mode, err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestClientAbility461090NativeStateAllRecords(t *testing.T) {
	clientAbilityStateNative461090(t, "state")
	testClientAbilityStateLinked(t, "state")
}

func TestClientAbility4610D0NativeResetAllRecords(t *testing.T) {
	clientAbilityStateNative461090(t, "reset")
	testClientAbilityStateLinked(t, "reset")
}

func TestClientAbility461120NativeActiveAllRecords(t *testing.T) {
	clientAbilityStateNative461090(t, "active")
	testClientAbilityStateLinked(t, "active")
}

func TestClientAbility461160NativeLookupAllRecords(t *testing.T) {
	clientAbilityStateNative461090(t, "lookup")
	testClientAbilityStateLinked(t, "lookup")
}

func testClientAbilityStateLinked(t *testing.T, mode string) {
	t.Helper()
	table := memmap.PtrT[[48]uint32](0x5D4594, 1047764)
	original, oldFrame := *table, gameFrameHook
	defer func() { *table, gameFrameHook = original, oldFrame }()

	var frame uint32
	gameFrameHook = func() uint32 { return frame }
	for layout := range 11 {
		for key := range 7 {
			for value, availability := range []uint32{0, 1, 255, 0xffffffff} {
				for clock, tick := range []uint32{0, 0x7fffffff, 0x80000000, 0xffffffff} {
					frame = tick
					var want [48]uint32
					for i := range want {
						want[i] = 0x5a5a5a5a
					}
					want[0] = 0xabcdef00
					for row := 1; row <= 5; row++ {
						id := uint32(0)
						if layout < 5 {
							id = uint32((row+layout-1)%5 + 1)
						} else if layout < 10 {
							id = uint32(layout - 4)
						}
						base := row * 6
						want[base], want[base+2] = id, 0x80000000+uint32(row)
						want[base+3] = [4]uint32{0, 0xffffffff, 1 << id, 0xffffff80}[value]
						want[base+4], want[base+5] = 0x12345600+uint32(row), 0xfedcba00+uint32(row)
					}
					*table = want
					var result uintptr
					var wantOffset uintptr
					switch mode {
					case "state":
						for base := 6; base <= 30; base += 6 {
							if want[base] == uint32(key) {
								want[base+2], want[base+5] = availability, 0
								if availability == 0 {
									want[base+5] = frame
								}
							}
						}
						result, wantOffset = cfixture.ClientAbilityState461090(key, availability), 164
					case "reset":
						for base := 6; base <= 30; base += 6 {
							if key == 6 || want[base] == want[key*6] {
								want[base+2], want[base+5] = 1, 0
							}
						}
						result, wantOffset = cfixture.ClientAbilityReset4610D0(key), 164
					case "active":
						for base := 6; base <= 30; base += 6 {
							if want[base] == uint32(key) {
								if clock == 0 {
									want[base+3] &^= 1 << key
								} else {
									want[base+3] |= 1 << key
								}
							}
						}
						input := [4]int{0, 1, -1, 1}[clock]
						result, wantOffset = cfixture.ClientAbilitySetActive461120(key, input), 156
					}
					if wantOffset != 0 && result != uintptr(unsafe.Pointer(table))+wantOffset {
						t.Fatalf("%s layout %d key %d: returned %#x", mode, layout, key, result)
					}
					if mode == "active" || mode == "lookup" {
						active := false
						for base := 6; base <= 30; base += 6 {
							if want[base] == uint32(key) {
								active = want[base+3]&(1<<key) != 0
								break
							}
						}
						if got := cfixture.ClientAbilityActive461160(key); got != active {
							t.Fatalf("%s layout %d key %d: active = %t, want %t", mode, layout, key, got, active)
						}
					}
					if *table != want {
						t.Fatalf("%s layout %d key %d value %d clock %d changed the ability table", mode, layout, key, value, clock)
					}
				}
			}
		}
	}
}
