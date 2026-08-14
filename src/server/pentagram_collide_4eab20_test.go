package server

import (
	"reflect"
	"testing"
)

type pentagramCollideTestObject4EAB20 struct {
	data  *pentagramCollideTestData4EAB20
	guard uint32
}

type pentagramCollideTestData4EAB20 struct {
	before    uint32
	triggered uint32
	after     uint32
}

func defaultPentagramCollideHooks4EAB20() pentagramCollideHooks4EAB20[*pentagramCollideTestObject4EAB20, *pentagramCollideTestData4EAB20] {
	return pentagramCollideHooks4EAB20[*pentagramCollideTestObject4EAB20, *pentagramCollideTestData4EAB20]{
		loadUpdateData: func(obj *pentagramCollideTestObject4EAB20) *pentagramCollideTestData4EAB20 {
			return obj.data
		},
		storeTriggered: func(data *pentagramCollideTestData4EAB20, value uint32) {
			data.triggered = value
		},
	}
}

func TestPentagramCollide4EAB20LoadsOnceStoresOneAndReturnsSource(t *testing.T) {
	first := &pentagramCollideTestData4EAB20{
		before:    0x11223344,
		triggered: 0xaabbccdd,
		after:     0x55667788,
	}
	second := &pentagramCollideTestData4EAB20{
		before:    0x01020304,
		triggered: 0xdeadbeef,
		after:     0x05060708,
	}
	source := &pentagramCollideTestObject4EAB20{data: first, guard: 0x89abcdef}
	target := &pentagramCollideTestObject4EAB20{data: second, guard: 0x76543210}
	collision := &struct{ guard uint32 }{guard: 0x31415926}
	events := make([]string, 0, 2)
	hooks := defaultPentagramCollideHooks4EAB20()
	hooks.loadUpdateData = func(obj *pentagramCollideTestObject4EAB20) *pentagramCollideTestData4EAB20 {
		events = append(events, "data")
		if obj != source {
			t.Fatalf("source = %p", obj)
		}
		source.data = second
		return first
	}
	hooks.storeTriggered = func(data *pentagramCollideTestData4EAB20, value uint32) {
		events = append(events, "store")
		if data != first || value != pentagramCollideTriggered4EAB20 {
			t.Fatalf("store = %p/%#x", data, value)
		}
		data.triggered = value
	}

	got := pentagramCollide4EAB20(source, target, collision, hooks)
	if got != source {
		t.Fatalf("return = %p, want %p", got, source)
	}
	if want := []string{"data", "store"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if first.before != 0x11223344 || first.triggered != 1 || first.after != 0x55667788 {
		t.Fatalf("first data = %+v", first)
	}
	if second.before != 0x01020304 || second.triggered != 0xdeadbeef || second.after != 0x05060708 {
		t.Fatalf("second data = %+v", second)
	}
	if source.guard != 0x89abcdef || target.guard != 0x76543210 || collision.guard != 0x31415926 {
		t.Fatalf("state = source %+v target %+v collision %+v", source, target, collision)
	}
}

func TestPentagramCollide4EAB20IgnoresTargetAndCollision(t *testing.T) {
	data := &pentagramCollideTestData4EAB20{triggered: 7}
	source := &pentagramCollideTestObject4EAB20{data: data}
	hooks := defaultPentagramCollideHooks4EAB20()
	got := pentagramCollide4EAB20(source, func() { t.Fatal("target observed") }, func() { t.Fatal("collision observed") }, hooks)
	if got != source || data.triggered != 1 {
		t.Fatalf("return/data = %p/%#x", got, data.triggered)
	}
}

func TestPentagramCollide4EAB20NilSourceFaultsBeforeStore(t *testing.T) {
	events := make([]string, 0, 1)
	hooks := defaultPentagramCollideHooks4EAB20()
	hooks.loadUpdateData = func(obj *pentagramCollideTestObject4EAB20) *pentagramCollideTestData4EAB20 {
		events = append(events, "data")
		return obj.data
	}
	hooks.storeTriggered = func(*pentagramCollideTestData4EAB20, uint32) {
		t.Fatal("store reached")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
		if want := []string{"data"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	pentagramCollide4EAB20[*pentagramCollideTestObject4EAB20](nil, 1, 2, hooks)
}

func TestPentagramCollide4EAB20NilUpdateDataFaultsAtStore(t *testing.T) {
	source := &pentagramCollideTestObject4EAB20{}
	events := make([]string, 0, 2)
	hooks := defaultPentagramCollideHooks4EAB20()
	hooks.loadUpdateData = func(obj *pentagramCollideTestObject4EAB20) *pentagramCollideTestData4EAB20 {
		events = append(events, "data")
		return obj.data
	}
	hooks.storeTriggered = func(data *pentagramCollideTestData4EAB20, value uint32) {
		events = append(events, "store")
		data.triggered = value
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update data did not fault")
		}
		if want := []string{"data", "store"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	pentagramCollide4EAB20(source, 1, 2, hooks)
}
