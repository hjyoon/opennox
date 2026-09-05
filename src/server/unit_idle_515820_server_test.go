package server

import (
	"math"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func unitIdleTestServer515820(t *testing.T) *Server {
	t.Helper()
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	return s
}

func TestUnitIdleNative515820PreservesHighPointerAndRestoresStack(t *testing.T) {
	s := unitIdleTestServer515820(t)
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{
		Action: uint32(ai.ACTION_WAIT),
		Args:   [4]uintptr{1, 2, 3, 4},
		Field5: 0,
	}

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %#x, want native address above 4 GiB", uintptr(unsafe.Pointer(unit)))
	}
	unitIdleNative515820(unit)

	if update.AIStackInd != 0 {
		t.Fatalf("AIStackInd = %d, want 0", update.AIStackInd)
	}
	if head := update.AIStackHead(); head == nil || head.Type() != ai.ACTION_IDLE {
		t.Fatalf("head = %#v, want ACTION_IDLE", head)
	} else if head.Args != ([4]uintptr{}) || head.Field5 != 0 {
		t.Fatalf("idle payload = args %#v field5 %d, want zero", head.Args, head.Field5)
	}
	if !s.AI.StackChanged {
		t.Fatal("native action-stack callbacks did not mark the stack changed")
	}
	runtime.KeepAlive(unit)
}

func TestUnitIdleNative515820PreservesOriginalGates(t *testing.T) {
	unitIdleNative515820(nil)
	unitIdleNative515820(&Object{ObjClass: object.ClassPlayer})
	unitIdleNative515820(&Object{
		ObjClass: object.ClassMonster,
		ObjFlags: object.FlagDead,
	})
}

func TestUnitIdleNative515820DoesNotHideMissingUpdateData(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("eligible monster without UpdateData did not preserve the original fault contract")
		}
	}()
	unitIdleNative515820(&Object{ObjClass: object.ClassMonster})
}

func TestUnitIdle515820ServerMethodUsesNativeAdapter(t *testing.T) {
	s := unitIdleTestServer515820(t)
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_WAIT)

	s.UnitIdle515820(unit)
	if got := unit.UpdateDataMonster().AIStackHead().Type(); got != ai.ACTION_IDLE {
		t.Fatalf("head = %v, want ACTION_IDLE", got)
	}
}

func TestUnitIdleNativeLayout515820(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantFlags = 20
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
