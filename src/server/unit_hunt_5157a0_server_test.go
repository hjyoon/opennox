package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestUnitHuntNative5157A0StackAndUnrelatedFlags(t *testing.T) {
	s := New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(1) {
		t.Fatal("cannot initialize server object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	unit := s.Objs.NewObject(&ObjectType{})
	update, free := alloc.New(MonsterUpdateData{})
	t.Cleanup(free)
	unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(update)
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unit.UpdateData) <= math.MaxUint32) {
		t.Fatalf("native pointers below 4 GiB: object=%p update=%p", unit, update)
	}
	s.SetFrame(901)
	for _, flags := range []object.Flags{0, object.FlagNoUpdate, object.FlagDestroyed} {
		unit.ObjFlags = flags
		*update = MonsterUpdateData{
			AIStackInd: 1, Field2: 2, Field67: 67, Field74: 74, Field91: 91,
			Field120_0: 7, Field120_1: 1, Field120_2: 2, Field120_3: 3,
			Field124: 124, Field137: 137,
		}
		update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{1, 2, 3, 4}}
		update.AIStack[1].Action = uint32(ai.DEPENDENCY_TIME)
		s.AI.StackChanged = false
		s.UnitHunt5157A0(unit)
		if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_HUNT ||
			update.AIStack[0].Args != ([4]uintptr{1, 2, 3, 4}) || update.AIStack[0].Field5 != 0 || !s.AI.StackChanged {
			t.Fatalf("flags=%#x stack=%+v index=%d changed=%t", flags, update.AIStack[0], update.AIStackInd, s.AI.StackChanged)
		}
		if update.Field2 != 0 || update.Field67 != 0 || update.Field74 != 0 || update.Field91 != 0 ||
			update.Field120_0 != 7 || update.Field120_1 != 0 || update.Field120_2 != 0 || update.Field120_3 != 0 ||
			update.Field124 != 901 || update.Field137 != 901 {
			t.Fatal("Hunt did not preserve the real action-reset state")
		}
	}
}

func TestUnitHuntNative5157A0GatesBeforeUpdateData(t *testing.T) {
	UnitHunt5157A0(nil)
	UnitHunt5157A0(&Object{ObjClass: object.ClassPlayer})
	UnitHunt5157A0(&Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagDead})
}

func TestUnitHuntNative5157A0DoesNotHideMissingUpdateData(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("eligible monster's missing UpdateData fault was hidden")
		}
	}()
	UnitHunt5157A0(&Object{ObjClass: object.ClassMonster})
}

func TestUnitHuntNativeLayout5157A0(t *testing.T) {
	wantSize, wantClass, wantFlags := uintptr(780), uintptr(8), uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize, wantClass, wantFlags = 928, 12, 20
	}
	if unsafe.Sizeof(Object{}) != wantSize || unsafe.Offsetof(Object{}.ObjClass) != wantClass ||
		unsafe.Offsetof(Object{}.ObjFlags) != wantFlags {
		t.Fatalf("%s/%s native Object layout: size=%d class=%d flags=%d", runtime.GOOS, runtime.GOARCH,
			unsafe.Sizeof(Object{}), unsafe.Offsetof(Object{}.ObjClass), unsafe.Offsetof(Object{}.ObjFlags))
	}
}
