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
	"github.com/opennox/opennox/v1/server"
)

func TestStateUpdateCNativePointers53DA60(t *testing.T) {
	if value := os.Getenv("OPENNOX_STATE_NATIVE_PROBE"); value != "" {
		kind, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&frame[32]))
		*source = server.Object{ObjClass: object.ClassImmobile, ObjFlags: 0x8000, Field5: 2, Field32: 1, Field34: 11}
		want := append([]byte(nil), frame...)
		gameFrameHook = func() uint32 { return 10 }
		gameFPSHook = func() uint32 { return 30 }
		arachnaphobiaRawRuntime53DA60 = func() server.ArachnaphobiaRawRuntime53DA60 {
			return server.ArachnaphobiaRawRuntime53DA60{Frame: gameFrameHook, FPS: gameFPSHook}
		}
		stateUpdateRawRuntime53DB30 = func() server.StateUpdateRawRuntime53DB30 {
			return server.StateUpdateRawRuntime53DB30{
				Frame: gameFrameHook, FPS: gameFPSHook,
				UnsetXStatus: func(o *server.Object, v uint32) { o.Field5 &^= v },
				SetXStatus:   func(o *server.Object, v uint32) { o.Field5 |= v },
			}
		}
		fmt.Printf("NATIVE_STATE_BEGIN kind=%d source=%#x\n", kind, uintptr(unsafe.Pointer(source)))
		stateUpdateNativeProbe53DA60(kind, source)
		if kind != 0 {
			binary.LittleEndian.PutUint32(want[32+unsafe.Offsetof(server.Object{}.Field5):], 4)
			binary.LittleEndian.PutUint32(want[32+unsafe.Offsetof(server.Object{}.Field34):], 70)
			if kind != 2 {
				binary.LittleEndian.PutUint32(want[32+unsafe.Offsetof(server.Object{}.ObjFlags):], 0x8040)
			}
		}
		if !bytes.Equal(want, frame) {
			t.Fatalf("native state/frame/canary differs: status=%#x flags=%#x next=%d", source.Field5, source.ObjFlags, source.Field34)
		}
		fmt.Println("NATIVE_STATE_OK")
		return
	}
	for kind, name := range []string{"arachnaphobia-wait", "break", "open", "break-remove"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestStateUpdateCNativePointers53DA60$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENNOX_STATE_NATIVE_PROBE="+strconv.Itoa(kind))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "NATIVE_STATE_OK") {
				t.Fatal("native marker missing")
			}
		})
	}
}
