package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type elevatorShaftXferTestData4F54A0 struct {
	elevatorExtent uint32
}

type elevatorShaftXferTestObject4F54A0 struct {
	field34    uint32
	updateData *elevatorShaftXferTestData4F54A0
}

type elevatorShaftXferMapCall4F54A0 struct {
	object  *elevatorShaftXferTestObject4F54A0
	version int32
}

type elevatorShaftXferInventoryCall4F54A0 struct {
	version uint16
	object  *elevatorShaftXferTestObject4F54A0
	count   int32
}

type elevatorShaftXferTestWorld4F54A0 struct {
	version         uint16
	mapResult       int32
	readOnlyValue   int32
	inventoryResult int32

	updateLoads     int
	field34Loads    int
	versionValues   []uint16
	mapCalls        []elevatorShaftXferMapCall4F54A0
	extentTransfers []*elevatorShaftXferTestData4F54A0
	readOnlyCalls   int
	inventoryCalls  []elevatorShaftXferInventoryCall4F54A0
	field34Stores   int
	events          []string
	after           map[string]func()
	faultAt         int
}

func newElevatorShaftXferTestWorld4F54A0() *elevatorShaftXferTestWorld4F54A0 {
	return &elevatorShaftXferTestWorld4F54A0{
		version:         elevatorShaftXferCurrentVersion4F54A0,
		mapResult:       1,
		readOnlyValue:   1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *elevatorShaftXferTestWorld4F54A0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *elevatorShaftXferTestWorld4F54A0) deps() elevatorShaftXferDeps4F54A0[
	*elevatorShaftXferTestObject4F54A0,
	*elevatorShaftXferTestData4F54A0,
] {
	return elevatorShaftXferDeps4F54A0[
		*elevatorShaftXferTestObject4F54A0,
		*elevatorShaftXferTestData4F54A0,
	]{
		loadUpdateData: func(object *elevatorShaftXferTestObject4F54A0) *elevatorShaftXferTestData4F54A0 {
			w.updateLoads++
			value := object.updateData
			w.event("load-update-data")
			return value
		},
		loadField34: func(object *elevatorShaftXferTestObject4F54A0) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.versionValues = append(w.versionValues, value)
			w.event("rw-version")
			return w.version
		},
		mapReadWrite: func(object *elevatorShaftXferTestObject4F54A0, version int32) int32 {
			w.mapCalls = append(w.mapCalls, elevatorShaftXferMapCall4F54A0{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		rwElevatorExtent: func(data *elevatorShaftXferTestData4F54A0) {
			w.extentTransfers = append(w.extentTransfers, data)
			_ = data.elevatorExtent
			w.event("rw-elevator-extent")
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			w.event("read-only")
			return w.readOnlyValue
		},
		transferInventory: func(version uint16, object *elevatorShaftXferTestObject4F54A0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, elevatorShaftXferInventoryCall4F54A0{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *elevatorShaftXferTestObject4F54A0, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestElevatorShaftXfer4F54A0CachesEntryDataAndRestoresField34(t *testing.T) {
	const original = uint32(0x11223344)
	entryData := &elevatorShaftXferTestData4F54A0{elevatorExtent: 7}
	liveData := &elevatorShaftXferTestData4F54A0{elevatorExtent: 17}
	object := &elevatorShaftXferTestObject4F54A0{field34: original, updateData: entryData}
	w := newElevatorShaftXferTestWorld4F54A0()
	w.after["load-update-data"] = func() { object.updateData = liveData }
	w.after["rw-elevator-extent"] = func() { object.field34 = 0x80000003 }
	w.after["load-field34:2"] = func() { object.field34 = 5 }

	if got := elevatorShaftXfer4F54A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != original || object.updateData != liveData {
		t.Fatalf("object = Field34 %#x UpdateData %p, want %#x/%p",
			object.field34, object.updateData, original, liveData)
	}
	if !reflect.DeepEqual(w.extentTransfers, []*elevatorShaftXferTestData4F54A0{entryData}) {
		t.Fatalf("cached extent transfers = %p, want entry %p", w.extentTransfers, entryData)
	}
	if !reflect.DeepEqual(w.versionValues, []uint16{elevatorShaftXferCurrentVersion4F54A0}) {
		t.Fatalf("version values = %v, want [60]", w.versionValues)
	}
	if !reflect.DeepEqual(w.mapCalls, []elevatorShaftXferMapCall4F54A0{{object: object, version: 60}}) {
		t.Fatalf("map calls = %#v, want object/version 60", w.mapCalls)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []elevatorShaftXferInventoryCall4F54A0{{
		version: 60, object: object, count: -2147483645,
	}}) {
		t.Fatalf("inventory calls = %#v, want zero-extended version and live count bits", w.inventoryCalls)
	}
	wantEvents := []string{
		"load-update-data", "load-field34:1", "rw-version", "map-read-write",
		"rw-elevator-extent", "load-field34:2", "read-only",
		"transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestElevatorShaftXfer4F54A0SignedVersionThresholds(t *testing.T) {
	tests := []struct {
		name       string
		version    uint16
		wantResult int32
		wantMap    []int32
		wantExtent bool
	}{
		{name: "negative thirty two thousand", version: 0x8000, wantResult: 1, wantMap: []int32{-32768}, wantExtent: true},
		{name: "negative one", version: 0xffff, wantResult: 1, wantMap: []int32{-1}, wantExtent: true},
		{name: "sixty", version: 60, wantResult: 1, wantMap: []int32{60}, wantExtent: true},
		{name: "sixty one", version: 61, wantResult: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &elevatorShaftXferTestData4F54A0{}
			object := &elevatorShaftXferTestObject4F54A0{updateData: data}
			w := newElevatorShaftXferTestWorld4F54A0()
			w.version = tc.version

			if got := elevatorShaftXfer4F54A0(object, w.deps()); got != tc.wantResult {
				t.Fatalf("result = %d, want %d", got, tc.wantResult)
			}
			var gotMap []int32
			for _, call := range w.mapCalls {
				gotMap = append(gotMap, call.version)
			}
			if !reflect.DeepEqual(gotMap, tc.wantMap) || (len(w.extentTransfers) != 0) != tc.wantExtent {
				t.Fatalf("map/extent = %v/%d, want %v/%t",
					gotMap, len(w.extentTransfers), tc.wantMap, tc.wantExtent)
			}
		})
	}
}

func TestElevatorShaftXfer4F54A0InventoryShortCircuitAndExactReadGate(t *testing.T) {
	t.Run("zero live count skips mode read", func(t *testing.T) {
		object := &elevatorShaftXferTestObject4F54A0{field34: 7, updateData: &elevatorShaftXferTestData4F54A0{}}
		w := newElevatorShaftXferTestWorld4F54A0()
		w.after["rw-elevator-extent"] = func() { object.field34 = 0 }

		if got := elevatorShaftXfer4F54A0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 0 || len(w.inventoryCalls) != 0 || object.field34 != 7 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 0/0/7",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("noncanonical read value skips inventory", func(t *testing.T) {
		object := &elevatorShaftXferTestObject4F54A0{field34: 9, updateData: &elevatorShaftXferTestData4F54A0{}}
		w := newElevatorShaftXferTestWorld4F54A0()
		w.readOnlyValue = 2

		if got := elevatorShaftXfer4F54A0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 9 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 1/0/9",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})
}

func TestElevatorShaftXfer4F54A0FailurePrefixesDoNotRollback(t *testing.T) {
	tests := []struct {
		name        string
		configure   func(*elevatorShaftXferTestWorld4F54A0, *elevatorShaftXferTestObject4F54A0)
		wantField34 uint32
		wantEvents  []string
	}{
		{
			name: "version greater than sixty",
			configure: func(w *elevatorShaftXferTestWorld4F54A0, object *elevatorShaftXferTestObject4F54A0) {
				w.version = 61
				w.after["rw-version"] = func() { object.field34 = 19 }
			},
			wantField34: 19,
			wantEvents:  []string{"load-update-data", "load-field34:1", "rw-version"},
		},
		{
			name: "map serializer failure",
			configure: func(w *elevatorShaftXferTestWorld4F54A0, object *elevatorShaftXferTestObject4F54A0) {
				w.mapResult = 0
				w.after["map-read-write"] = func() { object.field34 = 23 }
			},
			wantField34: 23,
			wantEvents:  []string{"load-update-data", "load-field34:1", "rw-version", "map-read-write"},
		},
		{
			name: "inventory failure",
			configure: func(w *elevatorShaftXferTestWorld4F54A0, object *elevatorShaftXferTestObject4F54A0) {
				w.inventoryResult = 0
				w.after["rw-elevator-extent"] = func() { object.field34 = 3 }
				w.after["transfer-inventory"] = func() { object.field34 = 29 }
			},
			wantField34: 29,
			wantEvents: []string{
				"load-update-data", "load-field34:1", "rw-version", "map-read-write",
				"rw-elevator-extent", "load-field34:2", "read-only", "transfer-inventory",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := &elevatorShaftXferTestObject4F54A0{
				field34: 11, updateData: &elevatorShaftXferTestData4F54A0{},
			}
			w := newElevatorShaftXferTestWorld4F54A0()
			tc.configure(w, object)
			if got := elevatorShaftXfer4F54A0(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) || w.field34Stores != 0 || object.field34 != tc.wantField34 {
				t.Fatalf("events/stores/Field34 = %v/%d/%d, want %v/0/%d",
					w.events, w.field34Stores, object.field34, tc.wantEvents, tc.wantField34)
			}
		})
	}
}

func TestElevatorShaftXfer4F54A0AnyNonzeroResultsAndFaultPrefixes(t *testing.T) {
	object := &elevatorShaftXferTestObject4F54A0{
		field34: 5, updateData: &elevatorShaftXferTestData4F54A0{},
	}
	w := newElevatorShaftXferTestWorld4F54A0()
	w.mapResult = -7
	w.inventoryResult = -9
	if got := elevatorShaftXfer4F54A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}

	wantEvents := []string{
		"load-update-data", "load-field34:1", "rw-version", "map-read-write",
		"rw-elevator-extent", "load-field34:2", "read-only",
		"transfer-inventory", "store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			object := &elevatorShaftXferTestObject4F54A0{
				field34: 13, updateData: &elevatorShaftXferTestData4F54A0{},
			}
			w := newElevatorShaftXferTestWorld4F54A0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = elevatorShaftXfer4F54A0(object, w.deps())
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

func TestElevatorShaftXfer4F54A0NilFaultOrder(t *testing.T) {
	t.Run("nil object faults on update data before Field34 and version", func(t *testing.T) {
		w := newElevatorShaftXferTestWorld4F54A0()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = elevatorShaftXfer4F54A0((*elevatorShaftXferTestObject4F54A0)(nil), w.deps())
		}()
		if recovered == nil {
			t.Fatal("nil object did not fault")
		}
		if w.updateLoads != 1 || w.field34Loads != 0 || len(w.versionValues) != 0 || len(w.events) != 0 {
			t.Fatalf("update/field/version/events = %d/%d/%d/%v, want 1/0/0/[]",
				w.updateLoads, w.field34Loads, len(w.versionValues), w.events)
		}
	})

	t.Run("nil update data faults after common serializer", func(t *testing.T) {
		object := &elevatorShaftXferTestObject4F54A0{field34: 7}
		w := newElevatorShaftXferTestWorld4F54A0()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = elevatorShaftXfer4F54A0(object, w.deps())
		}()
		if recovered == nil {
			t.Fatal("nil update data did not fault")
		}
		want := []string{"load-update-data", "load-field34:1", "rw-version", "map-read-write"}
		if !reflect.DeepEqual(w.events, want) || w.field34Loads != 1 || w.field34Stores != 0 {
			t.Fatalf("events/loads/stores = %v/%d/%d, want %v/1/0",
				w.events, w.field34Loads, w.field34Stores, want)
		}
	})
}
