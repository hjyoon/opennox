package server

import (
	"math"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func unitFollowTestServer5158C0(t *testing.T) *Server {
	t.Helper()
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	return s
}

func unitFollowNativeFixture5158C0(t *testing.T) (*Server, *Object, *Object) {
	t.Helper()
	s := unitFollowTestServer5158C0(t)
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{
		Action: uint32(ai.ACTION_WAIT),
		Args:   [4]uintptr{1, 2, 3, 4},
		Field5: 5,
	}
	target := &Object{PosVec: types.Pointf{
		X: math.Float32frombits(0x7fa12345),
		Y: math.Float32frombits(0x80000000),
	}}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
		}
		if uintptr(unsafe.Pointer(target)) <= math.MaxUint32 {
			t.Fatalf("target pointer = %p, want native address above 4 GiB", target)
		}
	}
	return s, unit, target
}

func TestUnitFollowNative5158C0PreservesPointersAndExactPayload(t *testing.T) {
	s, unit, target := unitFollowNativeFixture5158C0(t)
	unitFollowNative5158C0(unit, target)

	update := unit.UpdateDataMonster()
	if update.AIStackInd != 0 {
		t.Fatalf("AIStackInd = %d, want 0", update.AIStackInd)
	}
	head := update.AIStackHead()
	if head == nil || head.Type() != ai.ACTION_ESCORT {
		t.Fatalf("head = %#v, want ACTION_ESCORT", head)
	}
	if got := uint32(head.Args[0]); got != 0x7fa12345 {
		t.Fatalf("X bits = %#08x, want 0x7fa12345", got)
	}
	if got := uint32(head.Args[1]); got != 0x80000000 {
		t.Fatalf("Y bits = %#08x, want 0x80000000", got)
	}
	if got := head.Args[2]; got != uintptr(unsafe.Pointer(target)) {
		t.Fatalf("target = %#x, want native pointer %#x", got, uintptr(unsafe.Pointer(target)))
	}
	if head.Args[3] != 0 || head.Field5 != 0 {
		t.Fatalf("untouched action tail = %#x/%d, want zero", head.Args[3], head.Field5)
	}
	if !s.AI.StackChanged {
		t.Fatal("native action-stack callbacks did not mark the stack changed")
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(target)
}

func TestUnitFollowNative5158C0PreservesOriginalGates(t *testing.T) {
	unitFollowNative5158C0(nil, nil)
	unitFollowNative5158C0(&Object{ObjClass: object.ClassMonster}, nil)
	target := new(Object)
	unitFollowNative5158C0(&Object{ObjClass: object.ClassPlayer}, target)
	unitFollowNative5158C0(&Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagDead}, target)
	unit := &Object{ObjClass: object.ClassMonster}
	unitFollowNative5158C0(unit, unit)
}

func TestUnitFollowNative5158C0DoesNotHideMissingUpdateData(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("eligible monster without UpdateData did not preserve the original fault contract")
		}
	}()
	unitFollowNative5158C0(&Object{ObjClass: object.ClassMonster}, new(Object))
}

func TestUnitFollow5158C0ServerMethodUsesNativeAdapter(t *testing.T) {
	s, unit, target := unitFollowNativeFixture5158C0(t)
	s.UnitSetFollow5158C0(unit, target)
	if head := unit.UpdateDataMonster().AIStackHead(); head.Type() != ai.ACTION_ESCORT || head.ArgObj(2) != target {
		t.Fatalf("head = %#v target %p, want ACTION_ESCORT target %p", head, head.ArgObj(2), target)
	}
	runtime.KeepAlive(target)
}

func TestUnitFollowNativeLayout5158C0(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantPosition := uintptr(56)
	wantActionSize := uintptr(24)
	wantActionArgs := uintptr(4)
	wantActionField5 := uintptr(20)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantFlags = 20
		wantPosition = 60
		wantActionSize = 48
		wantActionArgs = 8
		wantActionField5 = 40
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.PosVec.Y", unsafe.Offsetof(Object{}.PosVec) + unsafe.Offsetof(types.Pointf{}.Y), wantPosition + 4},
		{"AIStackItem size", unsafe.Sizeof(AIStackItem{}), wantActionSize},
		{"AIStackItem.Action", unsafe.Offsetof(AIStackItem{}.Action), 0},
		{"AIStackItem.Args", unsafe.Offsetof(AIStackItem{}.Args), wantActionArgs},
		{"AIStackItem.Field5", unsafe.Offsetof(AIStackItem{}.Field5), wantActionField5},
		{"AIStackItem.Args element", unsafe.Sizeof(AIStackItem{}.Args[0]), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
