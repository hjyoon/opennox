package legacy

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPointOccupancyNativeCBefore54EF00(t *testing.T) {
	if os.Getenv("OPENNOX_POINT_OCCUPANCY_C_BEFORE") != "1" {
		return
	}
	mode := os.Getenv("OPENNOX_POINT_OCCUPANCY_C_CHILD")
	if mode == "" {
		for _, kind := range []string{"callback", "parent"} {
			t.Run(kind, func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestPointOccupancyNativeCBefore54EF00$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_POINT_OCCUPANCY_C_CHILD="+kind)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err == nil || !bytes.Contains(out, []byte("OCCUPANCY_BEFORE_VALID_NATIVE_CONTROLS "+kind)) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_occupancy_")) {
					t.Fatal("actual old native call must fault after controls", err)
				}
			})
		}
		return
	}
	TestMonsterDodgeablesNativeLayout547C50(t)
	o := server.Object{}
	if got, want := occupancyOriginalLayout54EF60(), [2]uintptr{unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.Field5)}; got != want {
		t.Fatal("independent native class/Field5 layouts", got, want)
	}
	srv, unit, update := monsterLookAtFixture5125A0(t)
	old := GetServer
	GetServer = func() Server { return &monsterMainLegacyServer547210{srv: srv} }
	t.Cleanup(func() { GetServer = old })
	point, freePoint := alloc.NewHigh(types.Ptf(300, 300))
	*point = types.Ptf(300, 300)
	t.Cleanup(freePoint)
	sound, freeSound := alloc.NewHigh(server.AIStackItem{})
	t.Cleanup(freeSound)
	unit.ObjClass, unit.ObjFlags, unit.Field5 = object.ClassMonster, object.FlagActive, 0
	unit.PosVec, unit.NewPos = *point, *point
	unit.Shape.Kind, unit.Shape.Circle.R, unit.Shape.Circle.R2 = server.ShapeKindCircle, 10, 100
	update.AIStackInd, update.AIStack[0].Action = 0, 0
	update.SoundSet122 = unsafe.Pointer(sound)
	if mode == "parent" {
		srv.Map.Init()
		t.Cleanup(srv.Map.Free)
		srv.Map.AddObjectToIndex(unit)
		found := 0
		srv.Map.EachObjInRect(types.Rectf{Min: types.Ptf(285, 285), Max: types.Ptf(315, 315)}, func(it *server.Object) bool {
			if it == unit {
				found++
			}
			return true
		})
		if found != 1 {
			t.Fatal("real native map must enumerate the same admitted original candidate once", found)
		}
	}
	if unit.ObjFlags&0x800 != 0 {
		t.Fatal("old C alias must also admit this positive native candidate")
	}
	if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(point)) <= math.MaxUint32 {
		t.Fatal("both native addresses must exceed4GiB")
	}
	before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
	set, action := dodgeOriginalControls547C50(unit)
	if set != uintptr(update.SoundSet122) || action != 0 || !unit.Sub547DB0(point) || !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(unit)), len(before))) {
		t.Fatal("typed native getter, positive geometry and whole initial object controls")
	}
	if occupancyOriginalCandidateABI54EF60() != 32 || occupancyOriginalParentABI54EF00() != 32 {
		t.Fatal("unchanged legacy context signatures required")
	}
	t.Logf("OCCUPANCY_BEFORE_VALID_NATIVE_CONTROLS %s unit=%p point=%p class=%08x flags=%08x", mode, unit, point, uint32(unit.ObjClass), unit.Field5)
	if mode == "callback" {
		occupancyOriginalCandidateC54EF60(unit, point)
	} else {
		occupancyOriginalParentC54EF00(point)
	}
	t.Fatal("old C unexpectedly returned")
}
