package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestSummonedCreatureLimitNative500D70PreservesNativeObjects(t *testing.T) {
	first := &Object{ObjSubClass: object.SubClass(object.MonsterSmall)}
	second := &Object{ObjSubClass: object.SubClass(object.MonsterMedium)}
	owner := &Object{Field129: first}
	first.Field128 = second
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 {
		t.Fatalf("owner pointer = %p, want native address above 4 GiB", owner)
	}

	var events []string
	var visited []*Object
	got := summonedCreatureLimitNative500D70(
		owner,
		math.MinInt32+0x500d70,
		func(index int32) int32 {
			events = append(events, "guide")
			if index != math.MinInt32+0x500d70 {
				t.Fatalf("guide index = %d, want %d", index, int32(math.MinInt32+0x500d70))
			}
			return 0x12345601
		},
		func(gotOwner, unit *Object) bool {
			events = append(events, "monitor")
			if gotOwner != owner {
				t.Fatalf("monitor owner = %p, want %p", gotOwner, owner)
			}
			visited = append(visited, unit)
			return true
		},
	)
	if !got {
		t.Fatal("one guide slot plus three controlled slots was rejected")
	}
	if want := []string{"guide", "monitor", "monitor"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if want := []*Object{first, second}; !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited = %v, want %v", visited, want)
	}
	runtime.KeepAlive(owner)
}

func TestSummonedCreatureLimitNative500D70GuidePrecedesNilOwnerFault(t *testing.T) {
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		summonedCreatureLimitNative500D70(
			nil,
			7,
			func(int32) int32 {
				events = append(events, "guide")
				return 1
			},
			func(*Object, *Object) bool {
				events = append(events, "monitor")
				return true
			},
		)
	}()
	if recovered == nil {
		t.Fatal("nil owner returned before the original first-owned load")
	}
	if want := []string{"guide"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want exact fault prefix %q", events, want)
	}
}

func TestCheckSummonedCreaturesLimit500D70PublicBinding(t *testing.T) {
	firstUpdate := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	secondUpdate := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	owner := new(Object)
	first := &Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterSmall),
		ObjOwner: owner, UpdateData: unsafe.Pointer(firstUpdate),
	}
	second := &Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterMedium),
		ObjOwner: owner, UpdateData: unsafe.Pointer(secondUpdate),
	}
	owner.Field129 = first
	first.Field128 = second

	if !CheckSummonedCreaturesLimit500D70(owner, 11, func(index int32) int32 {
		if index != 11 {
			t.Fatalf("guide index = %d, want 11", index)
		}
		return 1
	}) {
		t.Fatal("exact four-slot public binding result was rejected")
	}
	if CheckSummonedCreaturesLimit500D70(owner, 11, func(int32) int32 { return 2 }) {
		t.Fatal("five-slot public binding result was accepted")
	}
	runtime.KeepAlive(firstUpdate)
	runtime.KeepAlive(secondUpdate)
}
