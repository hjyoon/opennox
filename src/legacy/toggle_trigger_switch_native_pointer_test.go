package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestToggleTriggerSwitchCNativePointers(t *testing.T) {
	if value := os.Getenv("OPENNOX_TTS_NATIVE_PROBE"); value != "" {
		kind, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&frame[32]))
		recordFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.TriggerUpdateData{}))
		record := (*server.TriggerUpdateData)(unsafe.Pointer(&recordFrame[32]))
		*record = server.TriggerUpdateData{Flags: 8, State: 5, Field4: 0xabbaabba}
		*source = server.Object{TypeInd: 17, ObjClass: object.ClassImmobile, ObjFlags: object.FlagEnabled, Field33: 9, Field34: 10, UpdateData: unsafe.Pointer(record)}
		if kind == 2 {
			source.ObjFlags |= object.FlagNoCollide
		}
		wantObject, wantRecord := append([]byte(nil), frame...), append([]byte(nil), recordFrame...)
		if kind == 2 {
			binary.LittleEndian.PutUint32(wantObject[32+unsafe.Offsetof(source.ObjFlags):], uint32(object.FlagEnabled))
		}
		*memmap.PtrUint32(0x5D4594, 2488680) = 17
		*memmap.PtrUint32(0x5D4594, 2488676) = 18
		gameFrameHook = func() uint32 { return 10 }
		gameFPSHook = func() uint32 { return 30 }
		fmt.Printf("NATIVE_TTS_BEGIN kind=%d source=%#x data=%#x\n", kind, uintptr(unsafe.Pointer(source)), uintptr(unsafe.Pointer(record)))
		got := toggleTriggerSwitchProbe(kind, source)
		want := uint8(8)
		if kind == 2 {
			want = 0
		}
		if got != want || !bytes.Equal(wantObject, frame) || !bytes.Equal(wantRecord, recordFrame) {
			t.Fatalf("native return/object/data/canary differs: got=%#x want=%#x flags=%#x", got, want, uint32(source.ObjFlags))
		}
		fmt.Println("NATIVE_TTS_OK")
		return
	}
	for kind, name := range []string{"toggle", "trigger", "switch"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestToggleTriggerSwitchCNativePointers$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENNOX_TTS_NATIVE_PROBE="+strconv.Itoa(kind))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "NATIVE_TTS_OK") {
				t.Fatal("native marker")
			}
		})
	}
}
