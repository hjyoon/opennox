package legacy

import (
	"reflect"
	"testing"
)

type objectInventoryFixture4E7980 struct {
	first *objectInventoryFixture4E7980
	next  *objectInventoryFixture4E7980
}

func TestObjectInventoryAccessors4E7980(t *testing.T) {
	owner := &objectInventoryFixture4E7980{}
	first := &objectInventoryFixture4E7980{}
	next := &objectInventoryFixture4E7980{}
	owner.first = first
	first.next = next
	var events []string

	gotFirst := objectInventoryFirst4E7980(owner, func(obj *objectInventoryFixture4E7980) *objectInventoryFixture4E7980 {
		events = append(events, "first")
		return obj.first
	})
	gotNext := objectInventoryNext4E7990(gotFirst, func(obj *objectInventoryFixture4E7980) *objectInventoryFixture4E7980 {
		events = append(events, "next")
		return obj.next
	})
	if gotFirst != first || gotNext != next {
		t.Fatalf("accessors = (%p, %p), want (%p, %p)", gotFirst, gotNext, first, next)
	}
	if want := []string{"first", "next"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectInventoryNext4E7990NilDoesNotLoad(t *testing.T) {
	loads := 0
	got := objectInventoryNext4E7990[*objectInventoryFixture4E7980](nil, func(*objectInventoryFixture4E7980) *objectInventoryFixture4E7980 {
		loads++
		return &objectInventoryFixture4E7980{}
	})
	if got != nil || loads != 0 {
		t.Fatalf("next(nil) = %p with %d loads, want nil with 0 loads", got, loads)
	}
}

func TestObjectInventoryFirst4E7980NilFaults(t *testing.T) {
	loads := 0
	defer func() {
		if recover() == nil {
			t.Fatal("first(nil) did not fault")
		}
		if loads != 1 {
			t.Fatalf("first loads = %d, want 1", loads)
		}
	}()
	_ = objectInventoryFirst4E7980[*objectInventoryFixture4E7980](nil, func(obj *objectInventoryFixture4E7980) *objectInventoryFixture4E7980 {
		loads++
		return obj.first
	})
}
