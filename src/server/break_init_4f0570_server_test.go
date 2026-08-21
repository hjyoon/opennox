package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
)

func TestBreakInit4F0570NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantField5 := uintptr(20)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantField5 = 24
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Field5", unsafe.Offsetof(Object{}.Field5), wantField5},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestBreakInit4F0570NativeUsesLowByteAndExactObject(t *testing.T) {
	unit := &Object{Field5: 0x0e000001}
	calls := 0
	breakInitNative4F0570(unit, breakInitNativeDeps4F0570{
		setXStatus: func(got *Object, bit uint32) {
			calls++
			if got != unit || bit != 2 {
				t.Fatalf("set args = %p/%#x, want %p/2", got, bit, unit)
			}
			got.Field5 |= bit
		},
	})
	if calls != 1 || unit.Field5 != 0x0e000003 {
		t.Fatalf("calls/status = %d/%#x, want 1/0x0e000003", calls, unit.Field5)
	}
}

func TestBreakInit4F0570NativeMaskedBitsSkipSet(t *testing.T) {
	for _, status := range []uint32{2, 4, 8, 0xffffff0e} {
		unit := &Object{Field5: status}
		calls := 0
		breakInitNative4F0570(unit, breakInitNativeDeps4F0570{
			setXStatus: func(*Object, uint32) {
				calls++
			},
		})
		if calls != 0 || unit.Field5 != status {
			t.Fatalf("status %#x: calls/result = %d/%#x", status, calls, unit.Field5)
		}
	}
}

func TestBreakInit4F0570NativeNilUnitFaultsBeforeSet(t *testing.T) {
	calls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object did not preserve the original Field5-load fault")
		}
		if calls != 0 {
			t.Fatalf("set calls = %d, want 0", calls)
		}
	}()
	breakInitNative4F0570(nil, breakInitNativeDeps4F0570{
		setXStatus: func(*Object, uint32) {
			calls++
		},
	})
}

func TestBreakInit4F0570NativeBindsSetXStatus(t *testing.T) {
	s := new(Server)
	unit := &Object{ObjClass: object.ClassImmobile, Field5: 1, Field38: 0x12345678}
	for index := range unit.Field140 {
		unit.Field140[index] = uint32(index)
	}
	s.BreakInit4F0570(unit)
	if unit.Field5 != 3 || unit.Field38 != math.MaxUint32 {
		t.Fatalf("status/sync = %#x/%#x, want 3/max", unit.Field5, unit.Field38)
	}
	for index, value := range unit.Field140 {
		if value != 0x80000 {
			t.Fatalf("Field140[%d] = %#x, want 0x80000", index, value)
		}
	}

	masked := &Object{ObjClass: object.ClassImmobile, Field5: 2, Field38: 0x89abcdef}
	masked.Field140[0] = 0x12345678
	s.BreakInit4F0570(masked)
	if masked.Field5 != 2 || masked.Field38 != 0x89abcdef || masked.Field140[0] != 0x12345678 {
		t.Fatalf("masked object changed: status/sync/data = %#x/%#x/%#x", masked.Field5, masked.Field38, masked.Field140[0])
	}
}

func TestBreakInitParse536910NativeBinding(t *testing.T) {
	for _, args := range [][]string{nil, {}, {"ignored"}, {"2", "3"}} {
		objt := &ObjectType{Field9: math.MaxUint32}
		if err := objectBreakInitParse536910(objt, args); err != nil {
			t.Fatalf("args %q: parse error: %v", args, err)
		}
		if objt.Field9 != 2 {
			t.Fatalf("args %q: Field9 = %#x, want 2", args, objt.Field9)
		}
	}
}

func TestBreakInitParse536910NilObjectTypeFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil ObjectType did not preserve the original offset-36 store fault")
		}
	}()
	_ = objectBreakInitParse536910(nil, nil)
}

func TestObjectTypeParseInitRunsBreakParserWithZeroDataSize(t *testing.T) {
	const name = "BreakInit"
	oldDef, hadDef := initFuncs[name]
	initFuncs[name] = objectDefFunc{DataSize: 0}
	t.Cleanup(func() {
		if hadDef {
			initFuncs[name] = oldDef
		} else {
			delete(initFuncs, name)
		}
	})

	stale := byte(0xa5)
	objt := &ObjectType{
		Field9:       math.MaxUint32,
		InitData:     unsafe.Pointer(&stale),
		InitDataSize: 99,
	}
	if err := objt.parseInit(&things.ProcFunc{Name: name, Args: []string{"ignored"}}); err != nil {
		t.Fatal(err)
	}
	if objt.InitData != nil || objt.InitDataSize != 0 {
		t.Fatalf("init data/size = %p/%d, want nil/0", objt.InitData, objt.InitDataSize)
	}
	if objt.Field9 != 2 {
		t.Fatalf("Field9 = %#x, want 2", objt.Field9)
	}
}
