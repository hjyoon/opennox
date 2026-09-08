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

func monsterWanderTestServer512930(t *testing.T) *Server {
	t.Helper()
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	return s
}

func monsterWanderNativeFixture512930(t *testing.T) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	s := monsterWanderTestServer512930(t)
	update := new(MonsterUpdateData)
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{
		Action: uint32(ai.ACTION_WAIT),
		Args:   [4]uintptr{1, 2, 3, 4},
	}
	update.Field333 = 0xdeadbea5
	unit := &Object{
		ObjClass:     object.ClassMonster,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
		}
		if uintptr(unsafe.Pointer(update)) <= math.MaxUint32 {
			t.Fatalf("UpdateData pointer = %p, want native address above 4 GiB", update)
		}
	}
	return s, unit, update
}

func TestMonsterWanderNative512930BuildsExactActionStack(t *testing.T) {
	s, unit, update := monsterWanderNativeFixture512930(t)
	monsterWanderNative512930(unit)

	if update.AIStackInd != 1 {
		t.Fatalf("AIStackInd = %d, want 1", update.AIStackInd)
	}
	report := &update.AIStack[0]
	if report.Type() != ai.ACTION_REPORT {
		t.Fatalf("first action = %s, want ACTION_REPORT", report.Type())
	}
	if report.Args != [4]uintptr{10, 0, 0, 0} || report.Field5 != 0 {
		t.Fatalf("REPORT payload = %#v/%d, want [10 0 0 0]/0", report.Args, report.Field5)
	}
	roam := &update.AIStack[1]
	if roam.Type() != ai.ACTION_ROAM {
		t.Fatalf("head action = %s, want ACTION_ROAM", roam.Type())
	}
	if roam.Args != [4]uintptr{0, 0, 0xa5, 0} || roam.Field5 != 0 {
		t.Fatalf("ROAM payload = %#v/%d, want [0 0 0xa5 0]/0", roam.Args, roam.Field5)
	}
	if update.Field333 != 0xdeadbea5 {
		t.Fatalf("Field333 = %#08x, want unchanged 0xdeadbea5", update.Field333)
	}
	if !s.AI.StackChanged {
		t.Fatal("native action-stack callbacks did not mark the stack changed")
	}
	runtime.KeepAlive(unit)
}

func TestMonsterWanderNative512930PartialStoresPreserveUpperBits(t *testing.T) {
	item := new(AIStackItem)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		highArg0 := uint64(0xfedcba9800000000)
		highArg2 := uint64(0x0123456789abcd00)
		item.Args[0] = uintptr(highArg0)
		item.Args[2] = uintptr(highArg2)
	}
	monsterWanderStoreArgU32512930(item, 0, 0x76543210)
	monsterWanderStoreArgLow512930(item, 2, 0xa5)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantArg0 := uint64(0xfedcba9876543210)
		wantArg2 := uint64(0x0123456789abcda5)
		if item.Args[0] != uintptr(wantArg0) {
			t.Fatalf("dword store = %#x, want upper native bits preserved", item.Args[0])
		}
		if item.Args[2] != uintptr(wantArg2) {
			t.Fatalf("byte store = %#x, want upper native bits preserved", item.Args[2])
		}
	} else {
		if item.Args[0] != uintptr(0x76543210) || item.Args[2] != uintptr(0xa5) {
			t.Fatalf("32-bit stores = %#x/%#x, want 0x76543210/0xa5", item.Args[0], item.Args[2])
		}
	}
}

func TestMonsterWanderNative512930PreservesOriginalGates(t *testing.T) {
	monsterWanderNative512930(&Object{ObjClass: object.ClassPlayer})
	monsterWanderNative512930(&Object{
		ObjClass: object.ClassMonster,
		ObjFlags: object.FlagDead,
	})
}

func TestMonsterWanderNative512930DoesNotHideNullUnit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("null unit did not preserve the original class-load fault")
		}
	}()
	monsterWanderNative512930(nil)
}

func TestMonsterWanderNative512930DoesNotHideMissingUpdateData(t *testing.T) {
	s := monsterWanderTestServer512930(t)
	unit := &Object{ObjClass: object.ClassMonster, serverHandle: s.handle}
	defer func() {
		if recover() == nil {
			t.Fatal("eligible monster without UpdateData did not preserve the original fault contract")
		}
	}()
	monsterWanderNative512930(unit)
}

func TestMonsterWander512930ServerMethodUsesNativeAdapter(t *testing.T) {
	s, unit, update := monsterWanderNativeFixture512930(t)
	s.ScriptMonsterRoam512930(unit)
	if update.AIStackInd != 1 || update.AIStack[0].Type() != ai.ACTION_REPORT || update.AIStack[1].Type() != ai.ACTION_ROAM {
		t.Fatalf("stack = %#v, want REPORT then ROAM", update.GetAIStack())
	}
}

func TestMonsterWanderNativeLayout512930(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantUpdate := uintptr(748)
	wantMonsterSize := uintptr(2200)
	wantAIStackInd := uintptr(544)
	wantAIStack := uintptr(552)
	wantField333 := uintptr(1332)
	wantActionSize := uintptr(24)
	wantActionArgs := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantFlags = 20
		wantUpdate = 872
		wantMonsterSize = 2960
		wantAIStackInd = 628
		wantAIStack = 640
		wantField333 = 2072
		wantActionSize = 48
		wantActionArgs = 8
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"MonsterUpdateData size", unsafe.Sizeof(MonsterUpdateData{}), wantMonsterSize},
		{"MonsterUpdateData.AIStackInd", unsafe.Offsetof(MonsterUpdateData{}.AIStackInd), wantAIStackInd},
		{"MonsterUpdateData.AIStack", unsafe.Offsetof(MonsterUpdateData{}.AIStack), wantAIStack},
		{"MonsterUpdateData.Field333", unsafe.Offsetof(MonsterUpdateData{}.Field333), wantField333},
		{"AIStackItem size", unsafe.Sizeof(AIStackItem{}), wantActionSize},
		{"AIStackItem.Args", unsafe.Offsetof(AIStackItem{}.Args), wantActionArgs},
		{"AIStackItem.Args[2]", unsafe.Offsetof(AIStackItem{}.Args) + 2*unsafe.Sizeof(uintptr(0)), wantActionArgs + 2*unsafe.Sizeof(uintptr(0))},
		{"AIStackItem.Args element", unsafe.Sizeof(AIStackItem{}.Args[0]), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
