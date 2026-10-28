package legacy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellFlightNativePointer53B940(t *testing.T) {
	if os.Getenv("OPENNOX_SPELL_FLIGHT_CABI_PROBE") == "1" {
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		missile := (*server.Object)(unsafe.Pointer(&frame[32]))
		updateFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.SpellProjectileUpdateData{}))
		update := (*server.SpellProjectileUpdateData)(unsafe.Pointer(&updateFrame[32]))
		*update = server.SpellProjectileUpdateData{Target: missile}
		*missile = server.Object{Field32: 10, Field34: 15, SpeedCur: 10, VelVec: types.Ptf(3, 4), UpdateData: unsafe.Pointer(update)}
		spellProjectileUpdateRuntime53B940 = func() server.SpellProjectileUpdateRuntime53B940 {
			return server.SpellProjectileUpdateRuntime53B940{
				Frame: func() uint32 { return 20 }, Lifetime: func(string) float64 { return 100 },
				TickRate: func() uint32 { t.Fatal("unexpected target search"); return 0 },
			}
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(missile)) <= 0xffffffff {
			t.Fatal("C-owned missile must exercise an address above 4 GiB")
		}
		fmt.Printf("NATIVE_SPELL_FLIGHT_CABI_BEGIN ptr=%#x\n", uintptr(unsafe.Pointer(missile)))
		spellFlightCProbe53B940(missile)
		if update.Target != missile || missile.ForceVec != (types.Pointf{}) || missile.Float28 == 0 {
			t.Fatalf("native missile/update state = %p/%v/%v", update.Target, missile.ForceVec, missile.Float28)
		}
		fmt.Println("NATIVE_SPELL_FLIGHT_CABI_OK")
		return
	}
	command := exec.Command(os.Args[0], "-test.run=^TestSpellFlightNativePointer53B940$", "-test.count=1")
	command.Env = append(os.Environ(), "OPENNOX_SPELL_FLIGHT_CABI_PROBE=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Logf("native child output:\n%s", output)
		t.Fatalf("native spell-flight C pointer boundary: %v", err)
	}
	if !strings.Contains(string(output), "NATIVE_SPELL_FLIGHT_CABI_OK") {
		t.Fatal("native child success marker missing")
	}
}
