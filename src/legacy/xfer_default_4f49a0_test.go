package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type defaultXferTestObject4F49A0 struct {
	field34 uint32
}

type defaultXferMapCall4F49A0 struct {
	object  *defaultXferTestObject4F49A0
	version int32
}

type defaultXferInventoryCall4F49A0 struct {
	version uint16
	object  *defaultXferTestObject4F49A0
	count   int32
}

type defaultXferTestWorld4F49A0 struct {
	version         uint16
	mapResult       int32
	readOnlyValue   int32
	inventoryResult int32

	loadCalls      int
	rwCalls        int
	readOnlyCalls  int
	storeCalls     int
	rwArguments    []uint16
	mapCalls       []defaultXferMapCall4F49A0
	inventoryCalls []defaultXferInventoryCall4F49A0
	events         []string
	after          map[string]func()
	faultAt        int
}

func newDefaultXferTestWorld4F49A0() *defaultXferTestWorld4F49A0 {
	return &defaultXferTestWorld4F49A0{
		version:         defaultXferCurrentVersion4F49A0,
		mapResult:       1,
		readOnlyValue:   1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *defaultXferTestWorld4F49A0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *defaultXferTestWorld4F49A0) deps() defaultXferDeps4F49A0[*defaultXferTestObject4F49A0] {
	return defaultXferDeps4F49A0[*defaultXferTestObject4F49A0]{
		loadField34: func(object *defaultXferTestObject4F49A0) uint32 {
			w.loadCalls++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.loadCalls))
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.rwCalls++
			w.rwArguments = append(w.rwArguments, value)
			w.event("rw-version")
			return w.version
		},
		mapReadWrite: func(object *defaultXferTestObject4F49A0, version int32) int32 {
			w.mapCalls = append(w.mapCalls, defaultXferMapCall4F49A0{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			w.event("read-only")
			return w.readOnlyValue
		},
		transferInventory: func(version uint16, object *defaultXferTestObject4F49A0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, defaultXferInventoryCall4F49A0{
				version: version,
				object:  object,
				count:   count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *defaultXferTestObject4F49A0, value uint32) {
			w.storeCalls++
			w.event("store-field34")
			object.field34 = value
		},
	}
}

func TestDefaultXfer4F49A0PreservesEntryCacheSignedVersionAndLiveCount(t *testing.T) {
	const original = uint32(0x11223344)
	liveCount := uint32(0x80000005)
	object := &defaultXferTestObject4F49A0{field34: original}
	w := newDefaultXferTestWorld4F49A0()
	w.version = 0xffff
	w.after["rw-version"] = func() { object.field34 = 0x55667788 }
	w.after["map-read-write"] = func() { object.field34 = liveCount }
	w.after["load-field34:2"] = func() { object.field34 = 9 }

	if got := defaultXfer4F49A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != original {
		t.Fatalf("Field34 = %#08x, want entry value %#08x", object.field34, original)
	}
	if !reflect.DeepEqual(w.rwArguments, []uint16{defaultXferCurrentVersion4F49A0}) {
		t.Fatalf("version transfer arguments = %v, want [60]", w.rwArguments)
	}
	wantMap := []defaultXferMapCall4F49A0{{object: object, version: -1}}
	if !reflect.DeepEqual(w.mapCalls, wantMap) {
		t.Fatalf("map calls = %#v, want %#v", w.mapCalls, wantMap)
	}
	wantInventory := []defaultXferInventoryCall4F49A0{{
		version: 0xffff,
		object:  object,
		count:   int32(liveCount),
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want %#v", w.inventoryCalls, wantInventory)
	}
	wantEvents := []string{
		"load-field34:1",
		"rw-version",
		"map-read-write",
		"load-field34:2",
		"read-only",
		"transfer-inventory",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestDefaultXfer4F49A0ReadGateAndZeroCount(t *testing.T) {
	t.Run("non-one read value skips inventory", func(t *testing.T) {
		object := &defaultXferTestObject4F49A0{field34: 17}
		w := newDefaultXferTestWorld4F49A0()
		w.readOnlyValue = 2
		w.after["map-read-write"] = func() { object.field34 = 3 }

		if got := defaultXfer4F49A0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.inventoryCalls) != 0 {
			t.Fatalf("inventory calls = %v, want none", w.inventoryCalls)
		}
		if w.readOnlyCalls != 1 || object.field34 != 17 {
			t.Fatalf("read-only calls/Field34 = %d/%d, want 1/17", w.readOnlyCalls, object.field34)
		}
	})

	t.Run("zero live count skips read-only global", func(t *testing.T) {
		object := &defaultXferTestObject4F49A0{field34: 23}
		w := newDefaultXferTestWorld4F49A0()
		w.after["map-read-write"] = func() { object.field34 = 0 }

		if got := defaultXfer4F49A0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 0 || len(w.inventoryCalls) != 0 {
			t.Fatalf("read-only/inventory calls = %d/%d, want 0/0", w.readOnlyCalls, len(w.inventoryCalls))
		}
		if object.field34 != 23 {
			t.Fatalf("Field34 = %d, want 23", object.field34)
		}
	})
}

func TestDefaultXfer4F49A0FailurePrefixesDoNotRollback(t *testing.T) {
	t.Run("version greater than sixty", func(t *testing.T) {
		object := &defaultXferTestObject4F49A0{field34: 7}
		w := newDefaultXferTestWorld4F49A0()
		w.version = 61

		if got := defaultXfer4F49A0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "rw-version"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 7 {
			t.Fatalf("events/Field34 = %v/%d, want %v/7", w.events, object.field34, want)
		}
	})

	t.Run("map serializer failure", func(t *testing.T) {
		object := &defaultXferTestObject4F49A0{field34: 11}
		w := newDefaultXferTestWorld4F49A0()
		w.mapResult = 0
		w.after["map-read-write"] = func() { object.field34 = 29 }

		if got := defaultXfer4F49A0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "rw-version", "map-read-write"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 29 || w.storeCalls != 0 {
			t.Fatalf("events/Field34/stores = %v/%d/%d, want %v/29/0", w.events, object.field34, w.storeCalls, want)
		}
	})

	t.Run("inventory transfer failure", func(t *testing.T) {
		object := &defaultXferTestObject4F49A0{field34: 13}
		w := newDefaultXferTestWorld4F49A0()
		w.inventoryResult = 0
		w.after["map-read-write"] = func() { object.field34 = 5 }
		w.after["transfer-inventory"] = func() { object.field34 = 31 }

		if got := defaultXfer4F49A0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{
			"load-field34:1", "rw-version", "map-read-write",
			"load-field34:2", "read-only", "transfer-inventory",
		}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 31 || w.storeCalls != 0 {
			t.Fatalf("events/Field34/stores = %v/%d/%d, want %v/31/0", w.events, object.field34, w.storeCalls, want)
		}
	})
}

func TestDefaultXfer4F49A0FaultPrefixes(t *testing.T) {
	wantEvents := []string{
		"load-field34:1",
		"rw-version",
		"map-read-write",
		"load-field34:2",
		"read-only",
		"transfer-inventory",
		"store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			object := &defaultXferTestObject4F49A0{field34: 41}
			w := newDefaultXferTestWorld4F49A0()
			w.faultAt = faultAt
			w.after["map-read-write"] = func() { object.field34 = 2 }

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = defaultXfer4F49A0(object, w.deps())
			}()
			if recovered == nil {
				t.Fatal("call did not preserve the injected fault")
			}
			if !reflect.DeepEqual(w.events, wantEvents[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", w.events, wantEvents[:faultAt])
			}
		})
	}
}

func TestDefaultXfer4F49A0NilObjectFaultsBeforeVersionTransfer(t *testing.T) {
	w := newDefaultXferTestWorld4F49A0()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = defaultXfer4F49A0((*defaultXferTestObject4F49A0)(nil), w.deps())
	}()
	if recovered == nil {
		t.Fatal("nil object did not fault")
	}
	if w.loadCalls != 1 || w.rwCalls != 0 || len(w.events) != 0 {
		t.Fatalf("load/version/events = %d/%d/%v, want 1/0/[]", w.loadCalls, w.rwCalls, w.events)
	}
}
