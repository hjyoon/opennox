package legacy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestZombieEntryNativePointers534A10(t *testing.T) {
	names := []string{"plant", "zombie", "getup", "raise"}
	if child := os.Getenv("OPENNOX_ZOMBIE_CABI_PROBE"); child != "" {
		kind := -1
		for n, name := range names {
			if name == child {
				kind = n
			}
		}
		if kind < 0 {
			t.Fatal("unknown native zombie C probe")
		}
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{})+256)
		base := unsafe.Pointer(&frame[32])
		delta := (uintptr(0x80) - uintptr(base)) & 255
		unit := (*server.Object)(unsafe.Add(base, delta))
		*unit = server.Object{TypeInd: 100, ObjClass: 2}
		udFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.MonsterUpdateData{}))
		ud := (*server.MonsterUpdateData)(unsafe.Pointer(&udFrame[32]))
		*ud = server.MonsterUpdateData{}
		unit.UpdateData = unsafe.Pointer(ud)
		if kind == 3 {
			unit.TypeInd = 201
		}
		cache := [3]uint32{100, 100, 200}
		zombieEntryTypes534A10 = func() server.TypePredicateRuntime534A10 {
			return server.TypePredicateRuntime534A10{Read: func(i int) uint32 { return cache[i] }, Write: func(i int, v uint32) { cache[i] = v }, Lookup: func(string) uint32 { panic("warm native child must not look up types") }}
		}
		fmt.Printf("NATIVE_ZOMBIE_CABI_BEGIN kind=%s ptr=%#x\n", child, uintptr(unsafe.Pointer(unit)))
		got := zombieEntryCProbe534A10(kind, unit)
		want := []uint32{1, 1, 0x80, 0}[kind]
		if got != want {
			t.Fatalf("native C root %s returned %#x want %#x", child, got, want)
		}
		fmt.Printf("NATIVE_ZOMBIE_CABI_OK kind=%s result=%#x\n", child, got)
		return
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestZombieEntryNativePointers534A10$", "-test.count=1")
			command.Env = append(os.Environ(), "OPENNOX_ZOMBIE_CABI_PROBE="+name)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Logf("native child output:\n%s", output)
				t.Fatalf("native C %s pointer boundary: %v", name, err)
			}
			if !strings.Contains(string(output), "NATIVE_ZOMBIE_CABI_OK kind="+name) {
				t.Fatal("native child success marker missing")
			}
		})
	}
}
