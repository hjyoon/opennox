package legacy

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestBlowCandidateCNativePointer53C240(t *testing.T) {
	if os.Getenv("OPENNOX_BLOW_CANDIDATE_CABI_PROBE") == "1" {
		sframe := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		cframe := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&sframe[32]))
		candidate := (*server.Object)(unsafe.Pointer(&cframe[32]))
		*source = server.Object{PosVec: types.Ptf(100, 200)}
		*candidate = server.Object{PosVec: types.Ptf(110, 202), Mass: 2}
		checked := false
		blowUpdateRuntime53C160 = func() server.BlowUpdateRuntime53C160 {
			return server.BlowUpdateRuntime53C160{
				IndexedDirection: func(int16) image.Point { return image.Pt(1, 0) },
				CanInteract: func(a, b *server.Object) bool {
					if a != source || b != candidate {
						t.Fatal("native interaction arguments changed")
					}
					checked = true
					return false
				},
			}
		}
		fmt.Printf("NATIVE_BLOW_CANDIDATE_CABI_BEGIN source=%#x candidate=%#x\n", uintptr(unsafe.Pointer(source)), uintptr(unsafe.Pointer(candidate)))
		blowCProbe53C160(candidate, source, true)
		if !checked {
			t.Fatal("native interaction not reached")
		}
		fmt.Println("NATIVE_BLOW_CANDIDATE_CABI_OK")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestBlowCandidateCNativePointer53C240$", "-test.count=1")
	cmd.Env = append(os.Environ(), "OPENNOX_BLOW_CANDIDATE_CABI_PROBE=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("native child:\n%s", out)
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "NATIVE_BLOW_CANDIDATE_CABI_OK") {
		t.Fatal("native child success marker missing")
	}
}
