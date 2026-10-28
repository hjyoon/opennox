package legacy

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// Execute instruction probes in a child so the parent Go/engine FP state is
// untouched. ARM64 skips the provider; x86 native CI executes the same opcodes.
func TestSentryGlobeNativeX87Endpoint510E60(t *testing.T) {
	if os.Getenv("OPENNOX_SENTRY_X87_CHILD") == "" {
		if !sentryGlobeX87Available() {
			t.Skip("x87 instruction provider unavailable on this architecture")
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestSentryGlobeNativeX87Endpoint510E60$", "-test.count=1")
		cmd.Env = append(os.Environ(), "OPENNOX_SENTRY_X87_CHILD=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("native x87 child: %v\n%s", err, out)
		}
		if !strings.Contains(string(out), "SENTRY_NATIVE_X87_OK") {
			t.Fatal("native x87 marker")
		}
		return
	}
	angles := []uint32{0, 0x80000000, 1, 0x80000001, 0x3e99999a, 0x4b000001, 0x4f000000, 0x5effffff, 0x5f000000, 0x5f000001, 0xdf000000, 0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000, 0x7fc12345}
	for _, bits := range angles {
		data := &server.SentryUpdateData{Field0: bits}
		source := &server.Object{ObjFlags: object.FlagEnabled, PosVec: types.Ptf(100.1, 50.7), UpdateData: unsafe.Pointer(data)}
		var head *server.Object
		var actual types.Pointf
		h := server.SentryGlobeRawRuntime510E60{List: server.SentryGlobeListRuntime510E60{Head: func() *server.Object { return head }, SetHead: func(o *server.Object) { head = o }}, TraceRay: func(r *types.Rectf, _ *types.Pointf, _ uint32) int32 { actual = r.Max; return 1 }, EachInRect: func(*types.Rectf, *server.SentryGlobeRayContext511020) {}}
		server.SentryGlobeUpdateRaw510E60(source, h)
		x, ok := sentryGlobeX87EndpointProbe(bits, math.Float32bits(source.PosVec.X), true)
		if !ok {
			t.Fatal("child lacks x87")
		}
		y, _ := sentryGlobeX87EndpointProbe(bits, math.Float32bits(source.PosVec.Y), false)
		if math.Float32bits(actual.X) != x || math.Float32bits(actual.Y) != y {
			t.Fatalf("angle=%08x portable=(%08x,%08x) x87=(%08x,%08x)", bits, math.Float32bits(actual.X), math.Float32bits(actual.Y), x, y)
		}
	}
	fmt.Println("SENTRY_NATIVE_X87_OK")
}
