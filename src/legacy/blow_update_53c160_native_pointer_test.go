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

func TestBlowCNativePointer53C160(t *testing.T) {
	if os.Getenv("OPENNOX_BLOW_CABI_PROBE") == "1" {
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&frame[32]))
		*source = server.Object{}
		fmt.Printf("NATIVE_BLOW_CABI_BEGIN source=%#x\n", uintptr(unsafe.Pointer(source)))
		blowCProbe53C160(nil, source, false)
		fmt.Println("NATIVE_BLOW_CABI_OK")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestBlowCNativePointer53C160$", "-test.count=1")
	cmd.Env = append(os.Environ(), "OPENNOX_BLOW_CABI_PROBE=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("native child:\n%s", out)
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "NATIVE_BLOW_CABI_OK") {
		t.Fatal("native child success marker missing")
	}
}
