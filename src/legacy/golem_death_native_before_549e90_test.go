package legacy

import (
	"bytes"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"
)

func TestGolemDeathNativeLayout549E90(t *testing.T) {
	o := server.Object{}
	want := [8]uintptr{unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.PosVec) + 4, unsafe.Offsetof(o.Field27), unsafe.Offsetof(o.Field29), unsafe.Offsetof(o.UseData), unsafe.Offsetof(o.UpdateData), unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup)}
	if got := golemOriginalLayout549E90(); got != want {
		t.Fatal("independent complete C/Go layout", got, want)
	}
}
func golemActualConstructorFixture549E90(t *testing.T) (*server.Server, *server.Object) {
	t.Helper()
	TestGolemDeathNativeLayout549E90(t)
	s := npcReflectLegacyServer4E17B0(t)
	s.FreeObjectTypes()
	if !s.Objs.InitHigh(16) {
		t.Fatal("object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	t.Cleanup(s.FreeObjectTypes)
	for _, name := range []string{"MechGear", "MechHead", "MechArm", "MechLeg", "BigRock", "MediumRock", "SmallRock"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	for index := range 7 {
		obj := golemOriginalNew549E90(golemOriginalName549E90(index))
		if obj == nil {
			t.Fatal("actual C constructor", index)
		}
		obj.PosVec = types.Ptf(7, 8)
		obj.ObjFlags = 0x10203040
		obj.Field27 = -1.25
		obj.Field29 = math.Float32bits(3)
		data, _ := alloc.MallocHigh(8)
		obj.UseData.SetPtr(data)
		before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(obj)), int(unsafe.Sizeof(*obj))))
		pointers, words := golemOriginalControl549E90(obj)
		if pointers != [4]uintptr{uintptr(unsafe.Pointer(obj)), uintptr(data), uintptr(obj.UpdateData), uintptr(unsafe.Pointer(&obj.PosVec))} || words != [5]uint32{math.Float32bits(7), math.Float32bits(8), 0x10203040, math.Float32bits(-1.25), math.Float32bits(3)} {
			t.Fatal("actual C constructor/control pointers and scalars", pointers, words)
		}
		if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(obj)), int(unsafe.Sizeof(*obj)))) {
			t.Fatal("whole C constructor/control record")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(data) <= math.MaxUint32) {
			t.Fatal("native pointers must exceed4GiB")
		}
		s.Objs.FreeObject(obj)
	}
	unit, free := alloc.NewHigh(server.Object{})
	t.Cleanup(free)
	unit.PosVec = types.Ptf(300, 300)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatal("source must exceed4GiB")
	}
	return s, unit
}
func TestGolemDeathActualCBefore549E90(t *testing.T) {
	if os.Getenv("OPENNOX_GOLEM_C_BEFORE") != "1" {
		return
	}
	if entry := os.Getenv("OPENNOX_GOLEM_C_CHILD"); entry != "" {
		_, unit := golemActualConstructorFixture549E90(t)
		mech := entry == "mech"
		if unsafe.Sizeof(uintptr(0)) != 8 || golemOriginalABI549E90(mech) != 32 || golemOriginalRuntimeOverride549E90 != nil || golemOriginalStateOverride549E90 != nil {
			t.Fatal("unchanged actual old C required")
		}
		noxflags.SetGame(noxflags.GameModeCoop)
		t.Logf("GOLEM_REAL_C_CONTROLS entry=%s source=%p real-C-constructors-native-pointers-complete-layout-and-whole-records=true", entry, unit)
		golemOriginalC549E90(mech, unit)
		t.Fatal("old signed32 C unexpectedly returned")
	}
	t.Run("real-constructor-controls", func(t *testing.T) { golemActualConstructorFixture549E90(t) })
	for _, entry := range []string{"mech", "golem"} {
		t.Run(entry, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestGolemDeathActualCBefore549E90$", "-test.v")
			cmd.Env = append(os.Environ(), "OPENNOX_GOLEM_C_CHILD="+entry)
			out, err := cmd.CombinedOutput()
			t.Logf("old C %s %v\n%s", entry, err, out)
			if err == nil || !bytes.Contains(out, []byte("GOLEM_REAL_C_CONTROLS")) || bytes.Contains(out, []byte("unexpectedly returned")) {
				t.Fatal("positive native controls before old C fault", err)
			}
			if !bytes.Contains(out, []byte("SIGSEGV")) && !bytes.Contains(out, []byte("SIGBUS")) && !bytes.Contains(out, []byte("unexpected fault address")) {
				t.Fatal("actual pointer fault required")
			}
		})
	}
}
