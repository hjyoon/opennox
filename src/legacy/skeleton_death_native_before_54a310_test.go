package legacy

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSkeletonDeathNativeLayout54A310(t *testing.T) {
	o := server.Object{}
	a := server.ModifierInitData{}
	want := [11]uintptr{unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.ObjSubClass), unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.PosVec) + 4, unsafe.Offsetof(o.Field27), unsafe.Offsetof(o.Field29), unsafe.Offsetof(o.UseData), unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup), unsafe.Offsetof(a.Field16), unsafe.Sizeof(a)}
	if got := skeletonOriginalLayout54A310(); got != want {
		t.Fatal("independent complete C/Go object/modifier layout", got, want)
	}
}

func skeletonActualConstructorFixture54A310(t *testing.T) (*server.Server, *server.Object) {
	t.Helper()
	TestSkeletonDeathNativeLayout54A310(t)
	s := npcReflectLegacyServer4E17B0(t)
	s.FreeObjectTypes()
	if !s.Objs.Init(16) {
		t.Fatal("object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	t.Cleanup(s.FreeObjectTypes)
	for _, name := range []string{"Skull", "ArmBone", "LegBone", "TestItem"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	for _, index := range []int{0, 5, 6, 7} {
		obj := skeletonOriginalNew54A310(skeletonOriginalName54A310(index))
		if obj == nil {
			t.Fatal("actual C constructor", index)
		}
		obj.PosVec = types.Ptf(7, 8)
		data, _ := alloc.MallocHigh(8)
		obj.UseData.SetPtr(data)
		before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(obj)), int(unsafe.Sizeof(*obj))))
		pointers, words := skeletonOriginalControl54A310(obj)
		if pointers != [3]uintptr{uintptr(unsafe.Pointer(obj)), uintptr(data), uintptr(unsafe.Pointer(&obj.PosVec))} || words != [2]uint32{math.Float32bits(7), math.Float32bits(8)} {
			t.Fatal("real C native constructor/control identity", pointers, words)
		}
		if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(obj)), int(unsafe.Sizeof(*obj)))) {
			t.Fatal("complete C constructor control frame")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(data) <= math.MaxUint32) {
			t.Fatal("native constructor/use pointers must exceed4GiB")
		}
		s.Objs.FreeObject(obj)
	}
	unit, free := alloc.NewHigh(server.Object{})
	t.Cleanup(free)
	unit.PosVec = types.Ptf(300, 300)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatal("source pointer must exceed4GiB")
	}
	return s, unit
}

func TestSkeletonDeathActualCBefore54A310(t *testing.T) {
	if os.Getenv("OPENNOX_SKELETON_C_BEFORE") != "1" {
		return
	}
	if text := os.Getenv("OPENNOX_SKELETON_C_CHILD"); text != "" {
		entry, err := strconv.Atoi(text)
		if err != nil {
			t.Fatal(err)
		}
		_, unit := skeletonActualConstructorFixture54A310(t)
		if unsafe.Sizeof(uintptr(0)) != 8 || skeletonOriginalABI54A310(entry) != 32 || skeletonOriginalRuntimeOverride54A310 != nil || skeletonOriginalStateOverride54A310 != nil {
			t.Fatal("unchanged actual old C required")
		}
		noxflags.SetGame(noxflags.GameModeCoop)
		t.Logf("SKELETON_REAL_C_CONTROLS entry=%d source=%p real-C-constructors-native-use-pointers-complete-layout-and-whole-records=true", entry, unit)
		skeletonOriginalC54A310(entry, unit, skeletonOriginalName54A310(0), [4]unsafe.Pointer{}, 0)
		t.Fatal("old signed32 C unexpectedly returned")
	}
	t.Run("real-constructor-controls", func(t *testing.T) { skeletonActualConstructorFixture54A310(t) })
	for entry := range 4 {
		t.Run(fmt.Sprintf("entry-%d", entry), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestSkeletonDeathActualCBefore54A310$", "-test.v")
			cmd.Env = append(os.Environ(), fmt.Sprintf("OPENNOX_SKELETON_C_CHILD=%d", entry))
			out, err := cmd.CombinedOutput()
			t.Logf("old C result %v\n%s", err, out)
			if err == nil || !bytes.Contains(out, []byte("SKELETON_REAL_C_CONTROLS")) || bytes.Contains(out, []byte("unexpectedly returned")) {
				t.Fatal("positive native controls before old C fault", err)
			}
			if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
				t.Fatal("actual native pointer fault required")
			}
		})
	}
}
