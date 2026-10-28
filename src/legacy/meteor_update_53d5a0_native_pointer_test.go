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

	"github.com/opennox/opennox/v1/server"
)

func TestMeteorCNativePointers53D5A0(t *testing.T) {
	if value := os.Getenv("OPENNOX_METEOR_NATIVE_PROBE"); value != "" {
		kind, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&frame[32]))
		*source = server.Object{Field32: 1, Field34: 11, ZVal: 1}
		before := append([]byte(nil), frame...)
		gameFrameHook = func() uint32 { return 10 }
		gameFPSHook = func() uint32 { return 30 }
		meteorShowerRawRuntime53D5A0 = func() server.MeteorShowerRawRuntime53D5A0 {
			return server.MeteorShowerRawRuntime53D5A0{Frame: gameFrameHook, FPS: gameFPSHook}
		}
		fmt.Printf("NATIVE_METEOR_BEGIN kind=%d source=%#x\n", kind, uintptr(unsafe.Pointer(source)))
		meteorNativeEntryProbe53D5A0(kind, source)
		if !bytes.Equal(before, frame) {
			t.Fatal("waiting/airborne native frame or canary changed")
		}
		fmt.Println("NATIVE_METEOR_OK")
		return
	}
	for kind, name := range []string{"waiting-shower", "airborne-meteor"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestMeteorCNativePointers53D5A0$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENNOX_METEOR_NATIVE_PROBE="+strconv.Itoa(kind))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "NATIVE_METEOR_OK") {
				t.Fatal("native completion marker missing")
			}
		})
	}
}
