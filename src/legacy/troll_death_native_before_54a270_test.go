package legacy

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestTrollDeathNativeLayout54A270(t *testing.T) {
	o := server.Object{}
	want := [4]uintptr{unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(o.PosVec.X), unsafe.Offsetof(o.PosVec) + unsafe.Offsetof(o.PosVec.Y), unsafe.Offsetof(o.UpdateData), unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup)}
	if got := trollDeathOriginalLayout54A270(); got != want {
		t.Fatal("independent complete native C/Go layout", got, want)
	}
}

func trollDeathActualConstructorFixture54A270(t *testing.T) (*server.Server, *server.Object, uint32) {
	t.Helper()
	TestTrollDeathNativeLayout54A270(t)
	s := npcReflectLegacyServer4E17B0(t)
	s.FreeObjectTypes()
	if !s.Objs.Init(4) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "SmallToxicCloud", OnUpdate: &things.ProcFunc{Name: "SmallToxicCloudUpdate"}}); err != nil {
		t.Fatal(err)
	}
	id := trollDeathOriginalType54A270()
	if id == 0 || s.Types.ByID("SmallToxicCloud").Ind() != int(id) {
		t.Fatal("real C name lookup")
	}
	cloud := trollDeathOriginalNew54A270(id)
	if cloud == nil || cloud.UpdateData == nil || cloud.Update == nil {
		t.Fatal("real C constructor and registered native update data")
	}
	t.Cleanup(func() { s.Objs.FreeObject(cloud) })
	cloud.PosVec = types.Ptf(7, 8)
	data := (*server.ToxicCloudUpdateData)(cloud.UpdateData)
	data.Duration = -123
	frame := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(cloud)), int(unsafe.Sizeof(*cloud))))
	dataFrame := bytes.Clone(unsafe.Slice((*byte)(cloud.UpdateData), int(unsafe.Sizeof(*data))))
	pointers, bits := trollDeathOriginalControl54A270(cloud)
	if pointers != [4]uintptr{uintptr(unsafe.Pointer(cloud)), uintptr(cloud.UpdateData), uintptr(unsafe.Pointer(&cloud.PosVec.X)), uintptr(unsafe.Pointer(&cloud.PosVec.Y))} || bits != [2]uint32{math.Float32bits(7), math.Float32bits(8)} {
		t.Fatal("actual C native constructor result/position/update identity", pointers, bits)
	}
	if !bytes.Equal(frame, unsafe.Slice((*byte)(unsafe.Pointer(cloud)), int(unsafe.Sizeof(*cloud)))) || !bytes.Equal(dataFrame, unsafe.Slice((*byte)(cloud.UpdateData), int(unsafe.Sizeof(*data)))) {
		t.Fatal("whole actual constructor records changed in typed C controls")
	}
	unit, free := alloc.NewHigh(server.Object{})
	t.Cleanup(free)
	unit.PosVec = types.Ptf(300, 300)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(cloud), cloud.UpdateData} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatal("all valid constructor/source pointers must exceed4GiB", p)
			}
		}
	}
	return s, unit, id
}

func TestTrollDeathActualCBefore54A270(t *testing.T) {
	if os.Getenv("OPENNOX_TROLL_C_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_TROLL_C_CHILD") == "1" {
		_, unit, id := trollDeathActualConstructorFixture54A270(t)
		if unsafe.Sizeof(uintptr(0)) != 8 || trollDeathOriginalABI54A270() != 32 || trollDeathOriginalRuntimeOverride54A270 != nil || trollDeathOriginalStateOverride54A270 != nil {
			t.Fatal("unchanged old C native64 fixture required")
		}
		cache := memmap.PtrUint32(0x5D4594, 2491584)
		old := *cache
		*cache = id
		defer func() { *cache = old }()
		t.Logf("TROLL_REAL_C_NATIVE_CONTROLS source=%p type=%d full-C-Go-layout-real-C-constructor-native-update-and-whole-records=true", unit, id)
		trollDeathOriginalC54A270(unit)
		t.Fatal("old signed32 C entry unexpectedly returned")
	}
	t.Run("real-constructor-control", func(t *testing.T) {
		trollDeathActualConstructorFixture54A270(t)
	})
	t.Run("native-source-and-real-constructor", func(t *testing.T) {
		cmd := exec.Command(os.Args[0], "-test.run=^TestTrollDeathActualCBefore54A270$", "-test.v")
		cmd.Env = append(os.Environ(), "OPENNOX_TROLL_C_CHILD=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual old C result: %v\n%s", err, out)
		if err == nil || !bytes.Contains(out, []byte("TROLL_REAL_C_NATIVE_CONTROLS")) || bytes.Contains(out, []byte("unexpectedly returned")) {
			t.Fatal("old C failure requires valid complete native positive controls", err)
		}
		if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
			t.Fatal("old C result is not a native pointer fault")
		}
	})
}
