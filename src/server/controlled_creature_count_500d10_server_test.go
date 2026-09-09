package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestControlledCreatureCountNative500D10Layout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantSubClass := uintptr(12)
	wantNextOwned := uintptr(512)
	wantFirstOwned := uintptr(516)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantSubClass = 16
		wantNextOwned = 560
		wantFirstOwned = 568
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.ObjSubClass width", unsafe.Sizeof(Object{}.ObjSubClass), 4},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantNextOwned},
		{"Object.Field128 width", unsafe.Sizeof(Object{}.Field128), unsafe.Sizeof(uintptr(0))},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantFirstOwned},
		{"Object.Field129 width", unsafe.Sizeof(Object{}.Field129), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestControlledCreatureCountNative500D10BindsLiveObjectLinks(t *testing.T) {
	first := &Object{ObjSubClass: object.SubClass(object.MonsterSmall)}
	second := &Object{ObjSubClass: object.SubClass(object.MonsterMedium)}
	third := &Object{ObjSubClass: object.SubClass(0x80000000)}
	owner := &Object{Field129: first}
	first.Field128 = second
	second.Field128 = third

	var visited []*Object
	got := controlledCreatureCountNative500D10(owner, func(gotOwner, unit *Object) bool {
		if gotOwner != owner {
			t.Fatalf("monitor owner = %p, want %p", gotOwner, owner)
		}
		visited = append(visited, unit)
		return unit != second
	})
	if got != 5 {
		t.Fatalf("count = %d, want 5", got)
	}
	if want := []*Object{first, second, third}; !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited = %v, want %v", visited, want)
	}
}

func TestControlledCreatureCountNative500D10DoesNotGuardNilOwner(t *testing.T) {
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		controlledCreatureCountNative500D10(nil, func(*Object, *Object) bool {
			t.Fatal("nil owner reached monitor callback")
			return false
		})
	}()
	if recovered == nil {
		t.Fatal("nil owner returned before the original first-owned load")
	}
}

func TestControlledCreatureCount500D10PublicBinding(t *testing.T) {
	firstUpdate := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	secondUpdate := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	thirdUpdate := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	owner := &Object{}
	first := &Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterSmall),
		ObjOwner: owner, UpdateData: unsafe.Pointer(firstUpdate),
	}
	second := &Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterMedium),
		ObjOwner: owner, UpdateData: unsafe.Pointer(secondUpdate),
	}
	third := &Object{
		ObjClass: object.ClassMonster,
		ObjOwner: owner, UpdateData: unsafe.Pointer(thirdUpdate),
	}
	owner.Field129 = first
	first.Field128 = second
	second.Field128 = third

	if got := owner.Nox_xxx_countControlledCreatures_500D10(); got != 7 {
		t.Fatalf("count = %d, want 7", got)
	}
	runtime.KeepAlive(firstUpdate)
	runtime.KeepAlive(secondUpdate)
	runtime.KeepAlive(thirdUpdate)
}
