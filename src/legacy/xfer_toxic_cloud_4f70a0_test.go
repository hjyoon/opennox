package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type toxicCloudXferTestData4F70A0 struct {
	duration int32
}

type toxicCloudXferTestObject4F70A0 struct {
	data    *toxicCloudXferTestData4F70A0
	field34 uint32
}

type toxicCloudXferTestInventoryCall4F70A0 struct {
	version uint16
	object  *toxicCloudXferTestObject4F70A0
	count   int32
}

type toxicCloudXferTestWorld4F70A0 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	inventoryResult int32

	events            []string
	after             map[string]func()
	updateDataLoads   int
	field34Loads      int
	field34Stores     int
	mapVersions       []int32
	durationTransfers []*toxicCloudXferTestData4F70A0
	modeCalls         int
	inventoryCalls    []toxicCloudXferTestInventoryCall4F70A0
	panicDuration     bool
}

func newToxicCloudXferTestWorld4F70A0() *toxicCloudXferTestWorld4F70A0 {
	return &toxicCloudXferTestWorld4F70A0{
		version:         toxicCloudXferCurrentVersion4F70A0,
		mapResult:       1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *toxicCloudXferTestWorld4F70A0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *toxicCloudXferTestWorld4F70A0) deps() toxicCloudXferDeps4F70A0[
	*toxicCloudXferTestObject4F70A0,
	*toxicCloudXferTestData4F70A0,
] {
	return toxicCloudXferDeps4F70A0[
		*toxicCloudXferTestObject4F70A0,
		*toxicCloudXferTestData4F70A0,
	]{
		loadUpdateData: func(object *toxicCloudXferTestObject4F70A0) *toxicCloudXferTestData4F70A0 {
			value := object.data
			w.updateDataLoads++
			w.event("load-update-data")
			return value
		},
		loadField34: func(object *toxicCloudXferTestObject4F70A0) uint32 {
			value := object.field34
			w.field34Loads++
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		storeField34: func(object *toxicCloudXferTestObject4F70A0, value uint32) {
			object.field34 = value
			w.field34Stores++
			w.event("store-field34")
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *toxicCloudXferTestObject4F70A0, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
		},
		rwDuration: func(data *toxicCloudXferTestData4F70A0) {
			w.durationTransfers = append(w.durationTransfers, data)
			w.event("rw-duration")
			if w.panicDuration {
				panic("ToxicCloudXfer duration stream fault")
			}
		},
		readMode: func() int32 {
			index := w.modeCalls
			w.modeCalls++
			value := int32(0)
			if index < len(w.modes) {
				value = w.modes[index]
			}
			w.event(fmt.Sprintf("read-mode:%d=%d", index+1, value))
			return value
		},
		transferInventory: func(version uint16, object *toxicCloudXferTestObject4F70A0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, toxicCloudXferTestInventoryCall4F70A0{
				version: version,
				object:  object,
				count:   count,
			})
			w.event(fmt.Sprintf("transfer-inventory:%d:%d", version, count))
			return w.inventoryResult
		},
	}
}

func toxicCloudXferAssertPanics4F70A0(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("call did not panic")
		}
	}()
	call()
}

func TestToxicCloudXfer4F70A0CachesEntryDataAndPreservesOrder(t *testing.T) {
	entry := &toxicCloudXferTestData4F70A0{duration: 7}
	replacement := &toxicCloudXferTestData4F70A0{duration: 99}
	object := &toxicCloudXferTestObject4F70A0{data: entry, field34: 0x11223344}
	w := newToxicCloudXferTestWorld4F70A0()
	w.mapResult = -7
	w.modes = []int32{1}
	w.after["map-read-write:61"] = func() { object.data = replacement }
	w.after["rw-duration"] = func() {
		entry.duration = -1430532899
		object.field34 = 0x80000003
	}
	w.after["read-mode:1=1"] = func() { object.field34 = 5 }

	if got := toxicCloudXfer4F70A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if uint32(entry.duration) != 0xaabbccdd || replacement.duration != 99 || object.data != replacement {
		t.Fatalf("cached/replacement duration and pointer = %#x/%d/%p",
			uint32(entry.duration), replacement.duration, object.data)
	}
	if object.field34 != 0x11223344 {
		t.Fatalf("Field34 = %#x, want entry value", object.field34)
	}
	if !reflect.DeepEqual(w.durationTransfers, []*toxicCloudXferTestData4F70A0{entry}) {
		t.Fatalf("duration transfer pointers = %v, want cached entry", w.durationTransfers)
	}
	wantInventory := []toxicCloudXferTestInventoryCall4F70A0{{
		version: 61,
		object:  object,
		count:   -2147483645,
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %+v, want %+v", w.inventoryCalls, wantInventory)
	}
	wantEvents := []string{
		"load-update-data",
		"load-field34:1",
		"rw-version:61",
		"map-read-write:61",
		"rw-duration",
		"load-field34:2",
		"read-mode:1=1",
		"transfer-inventory:61:-2147483645",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, wantEvents)
	}
}

func TestToxicCloudXfer4F70A0SignedVersionRangeAndCommonPrefix(t *testing.T) {
	tests := []struct {
		name         string
		version      uint16
		mapResult    int32
		wantResult   int32
		wantMap      []int32
		wantDuration int
		wantStores   int
	}{
		{name: "lowest supported", version: 1, mapResult: 1, wantResult: 1, wantMap: []int32{1}, wantDuration: 1, wantStores: 1},
		{name: "current", version: 61, mapResult: -7, wantResult: 1, wantMap: []int32{61}, wantDuration: 1, wantStores: 1},
		{name: "zero", version: 0, mapResult: 1, wantResult: 0},
		{name: "positive too new", version: 62, mapResult: 1, wantResult: 0},
		{name: "largest positive", version: 0x7fff, mapResult: 1, wantResult: 0},
		{name: "most negative", version: 0x8000, mapResult: 1, wantResult: 0},
		{name: "minus one", version: 0xffff, mapResult: 1, wantResult: 0},
		{name: "common failure", version: 61, mapResult: 0, wantResult: 0, wantMap: []int32{61}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := &toxicCloudXferTestObject4F70A0{data: &toxicCloudXferTestData4F70A0{}}
			w := newToxicCloudXferTestWorld4F70A0()
			w.version = test.version
			w.mapResult = test.mapResult

			if got := toxicCloudXfer4F70A0(object, w.deps()); got != test.wantResult {
				t.Fatalf("result = %d, want %d", got, test.wantResult)
			}
			if !reflect.DeepEqual(w.mapVersions, test.wantMap) ||
				len(w.durationTransfers) != test.wantDuration || w.field34Stores != test.wantStores {
				t.Fatalf("map/duration/stores = %v/%d/%d, want %v/%d/%d",
					w.mapVersions, len(w.durationTransfers), w.field34Stores,
					test.wantMap, test.wantDuration, test.wantStores)
			}
		})
	}

	t.Run("invalid version keeps callback mutation", func(t *testing.T) {
		object := &toxicCloudXferTestObject4F70A0{data: &toxicCloudXferTestData4F70A0{}, field34: 7}
		w := newToxicCloudXferTestWorld4F70A0()
		w.version = 0
		w.after["rw-version:61"] = func() { object.field34 = 9 }

		if got := toxicCloudXfer4F70A0(object, w.deps()); got != 0 || object.field34 != 9 || w.field34Stores != 0 {
			t.Fatalf("result/Field34/stores = %d/%d/%d, want 0/9/0", got, object.field34, w.field34Stores)
		}
	})

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := &toxicCloudXferTestObject4F70A0{data: &toxicCloudXferTestData4F70A0{}, field34: 7}
		w := newToxicCloudXferTestWorld4F70A0()
		w.mapResult = 0
		w.after["map-read-write:61"] = func() { object.field34 = 11 }

		if got := toxicCloudXfer4F70A0(object, w.deps()); got != 0 || object.field34 != 11 || w.field34Stores != 0 {
			t.Fatalf("result/Field34/stores = %d/%d/%d, want 0/11/0", got, object.field34, w.field34Stores)
		}
	})
}

func TestToxicCloudXfer4F70A0LiveInventoryGate(t *testing.T) {
	t.Run("zero live count skips mode", func(t *testing.T) {
		object := &toxicCloudXferTestObject4F70A0{data: &toxicCloudXferTestData4F70A0{}, field34: 9}
		w := newToxicCloudXferTestWorld4F70A0()
		w.after["rw-duration"] = func() { object.field34 = 0 }

		if got := toxicCloudXfer4F70A0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 0 || len(w.inventoryCalls) != 0 || object.field34 != 9 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d", w.modeCalls, len(w.inventoryCalls), object.field34)
		}
	})

	for _, mode := range []int32{-1, 0, 2} {
		t.Run(fmt.Sprintf("mode_%d_skips", mode), func(t *testing.T) {
			object := &toxicCloudXferTestObject4F70A0{data: &toxicCloudXferTestData4F70A0{}, field34: 3}
			w := newToxicCloudXferTestWorld4F70A0()
			w.modes = []int32{mode}

			if got := toxicCloudXfer4F70A0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if len(w.inventoryCalls) != 0 || object.field34 != 3 {
				t.Fatalf("inventory/Field34 = %d/%d", len(w.inventoryCalls), object.field34)
			}
		})
	}

	t.Run("exact one passes zero-extended version and signed count bits", func(t *testing.T) {
		object := &toxicCloudXferTestObject4F70A0{data: &toxicCloudXferTestData4F70A0{}, field34: 0x80000001}
		w := newToxicCloudXferTestWorld4F70A0()
		w.version = 1
		w.modes = []int32{1}

		if got := toxicCloudXfer4F70A0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		want := []toxicCloudXferTestInventoryCall4F70A0{{version: 1, object: object, count: -2147483647}}
		if !reflect.DeepEqual(w.inventoryCalls, want) {
			t.Fatalf("inventory calls = %+v, want %+v", w.inventoryCalls, want)
		}
	})
}

func TestToxicCloudXfer4F70A0InventoryFailureKeepsAllPrefixSideEffects(t *testing.T) {
	entry := &toxicCloudXferTestData4F70A0{duration: 1}
	object := &toxicCloudXferTestObject4F70A0{data: entry, field34: 7}
	w := newToxicCloudXferTestWorld4F70A0()
	w.modes = []int32{1}
	w.inventoryResult = 0
	w.after["rw-duration"] = func() {
		entry.duration = 0x12345678
		object.field34 = 11
	}
	w.after["transfer-inventory:61:11"] = func() { object.field34 = 13 }

	if got := toxicCloudXfer4F70A0(object, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if entry.duration != 0x12345678 || object.field34 != 13 || w.field34Stores != 0 {
		t.Fatalf("duration/Field34/stores = %#x/%d/%d, want transferred/13/0",
			entry.duration, object.field34, w.field34Stores)
	}
}

func TestToxicCloudXfer4F70A0NonzeroInventoryResultIsCanonicalized(t *testing.T) {
	object := &toxicCloudXferTestObject4F70A0{data: &toxicCloudXferTestData4F70A0{}, field34: 5}
	w := newToxicCloudXferTestWorld4F70A0()
	w.modes = []int32{1}
	w.inventoryResult = -7
	if got := toxicCloudXfer4F70A0(object, w.deps()); got != 1 || object.field34 != 5 || w.field34Stores != 1 {
		t.Fatalf("result/Field34/stores = %d/%d/%d, want 1/5/1", got, object.field34, w.field34Stores)
	}
}

func TestToxicCloudXfer4F70A0FaultBoundaries(t *testing.T) {
	t.Run("nil object faults before the first event", func(t *testing.T) {
		w := newToxicCloudXferTestWorld4F70A0()
		toxicCloudXferAssertPanics4F70A0(t, func() { toxicCloudXfer4F70A0(nil, w.deps()) })
		if len(w.events) != 0 {
			t.Fatalf("events = %v, want none", w.events)
		}
	})

	t.Run("nil UpdateData reaches duration transfer after common prefix", func(t *testing.T) {
		object := &toxicCloudXferTestObject4F70A0{field34: 5}
		w := newToxicCloudXferTestWorld4F70A0()
		w.panicDuration = true
		toxicCloudXferAssertPanics4F70A0(t, func() { toxicCloudXfer4F70A0(object, w.deps()) })
		want := []string{
			"load-update-data", "load-field34:1", "rw-version:61", "map-read-write:61", "rw-duration",
		}
		if !reflect.DeepEqual(w.events, want) || w.field34Stores != 0 {
			t.Fatalf("events/stores = %v/%d, want %v/0", w.events, w.field34Stores, want)
		}
	})
}
