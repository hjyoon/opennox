package legacy

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestPointShapeNativeCBefore547DB0(t *testing.T) {
	if os.Getenv("OPENNOX_POINT_SHAPE_C_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_POINT_SHAPE_C_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestPointShapeNativeCBefore547DB0$", "-test.v")
		cmd.Env = append(os.Environ(), "OPENNOX_POINT_SHAPE_C_CHILD=1")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err == nil || !bytes.Contains(out, []byte("POINT_SHAPE_BEFORE_VALID_NATIVE_CONTROLS")) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_point_shape_public_547db0")) {
			t.Fatal("actual legacy-int C must fault only after valid independent native controls", err)
		}
		return
	}
	TestMonsterDodgeablesNativeLayout547C50(t)
	o := server.Object{}
	want := [5]uintptr{unsafe.Offsetof(o.NewPos), unsafe.Offsetof(o.NewPos) + 4, unsafe.Offsetof(o.Shape), unsafe.Offsetof(o.Shape) + unsafe.Offsetof(server.Shape{}.Circle) + 4, unsafe.Sizeof(types.Pointf{})}
	if got := pointShapeOriginalLayout547DB0(); got != want {
		t.Fatal("independent point/shape C/Go layout", got, want)
	}
	w := newDodgeWorld547C50(t, dodgeOriginalRows547C50(t)[0])
	w.Unit.NewPos = types.Ptf(3, 4)
	w.Unit.Shape.Kind = server.ShapeKindCircle
	w.Unit.Shape.Circle.R2 = 100
	w.Data.AIStackInd = 0
	w.Data.AIStack[0].Action = 0
	w.Data.SoundSet122 = unsafe.Pointer(w.Pushes[0])
	point := (*types.Pointf)(unsafe.Pointer(&w.Heap[24000]))
	*point = w.Unit.NewPos
	if uintptr(unsafe.Pointer(point)) <= math.MaxUint32 {
		t.Fatal("point must exceed4GiB")
	}
	before := bytes.Clone(w.Heap)
	set, action := dodgeOriginalControls547C50(w.Unit)
	if set != uintptr(w.Data.SoundSet122) || action != 0 || !w.Unit.Sub547DB0(point) || !bytes.Equal(before, w.Heap) {
		t.Fatal("actual typed getter, positive native geometry and complete initial allocation controls")
	}
	if pointShapeOriginalABI547DB0() != 32 {
		t.Fatal("unchanged legacy int ABI required")
	}
	t.Logf("POINT_SHAPE_BEFORE_VALID_NATIVE_CONTROLS unit=%p point=%p kind=%d radius2=%08x publicABI=legacy32-explicit-narrowing", w.Unit, point, w.Unit.Shape.Kind, math.Float32bits(w.Unit.Shape.Circle.R2))
	pointShapeOriginalC547DB0(w.Unit, point)
	t.Fatal("old C unexpectedly returned")
}
