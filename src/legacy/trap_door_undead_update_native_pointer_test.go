package legacy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestTrapDoorUndeadCNativePointers(t *testing.T) {
	if value := os.Getenv("OPENNOX_TRAP_UNDEAD_NATIVE_PROBE"); value != "" {
		kind, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&frame[32]))
		*source = server.Object{ObjClass: object.ClassImmobile, ObjFlags: object.FlagEnabled, Field5: 4, Field34: 10}
		data := radialCFrameOriginal(t, unsafe.Sizeof(server.TrapDoorCollideData{}))
		typed := (*server.TrapDoorCollideData)(unsafe.Pointer(&data[32]))
		*typed = server.TrapDoorCollideData{NextFrame: 11, Activated: 7}
		if kind == 1 {
			*(*server.UndeadKillerCollideData)(unsafe.Pointer(typed)) = server.UndeadKillerCollideData{}
		}
		source.CollideData = unsafe.Pointer(typed)
		wantObject, wantData := append([]byte(nil), frame...), append([]byte(nil), data...)
		gameFrameHook = func() uint32 { return 10 }
		gameFPSHook = func() uint32 { return 30 }
		trapDoorUpdateRawRuntime53DE80 = func() server.TrapDoorUpdateRawRuntime53DE80 {
			return server.TrapDoorUpdateRawRuntime53DE80{Frame: gameFrameHook, FPS: gameFPSHook}
		}
		undeadKillerUpdateRawRuntime53E190 = func() server.UndeadKillerUpdateRawRuntime53E190 {
			return server.UndeadKillerUpdateRawRuntime53E190{Frame: gameFrameHook}
		}

		fmt.Printf("NATIVE_TRAP_UNDEAD_BEGIN kind=%d source=%#x data=%#x\n", kind, uintptr(unsafe.Pointer(source)), uintptr(unsafe.Pointer(typed)))
		trapDoorUndeadNativeProbe(kind, source)
		if !bytes.Equal(wantObject, frame) || !bytes.Equal(wantData, data) {
			t.Fatalf("native object/data/canary differs: flags=%#x status=%#x next=%d", uint32(source.ObjFlags), source.Field5, typed.NextFrame)
		}
		fmt.Println("NATIVE_TRAP_UNDEAD_OK")
		return
	}
	for kind, name := range []string{"trapdoor-wait", "undead-live"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestTrapDoorUndeadCNativePointers$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENNOX_TRAP_UNDEAD_NATIVE_PROBE="+strconv.Itoa(kind))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "NATIVE_TRAP_UNDEAD_OK") {
				t.Fatal("native marker missing")
			}
		})
	}
}
