package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type elevatorXferTestData4F53D0 struct {
	shaftExtent uint32
	field3      byte
	field4      uint32
}

type elevatorXferTestObject4F53D0 struct {
	field34    uint32
	updateData *elevatorXferTestData4F53D0
}

type elevatorXferMapCall4F53D0 struct {
	object  *elevatorXferTestObject4F53D0
	version int32
}

type elevatorXferInventoryCall4F53D0 struct {
	version uint16
	object  *elevatorXferTestObject4F53D0
	count   int32
}

type elevatorXferTestWorld4F53D0 struct {
	version         uint16
	mapResult       int32
	readOnlyValue   int32
	inventoryResult int32

	updateLoads     int
	field34Loads    int
	versionValues   []uint16
	mapCalls        []elevatorXferMapCall4F53D0
	shaftTransfers  []*elevatorXferTestData4F53D0
	field4Transfers []*elevatorXferTestData4F53D0
	field3Transfers []*elevatorXferTestData4F53D0
	readOnlyCalls   int
	inventoryCalls  []elevatorXferInventoryCall4F53D0
	field34Stores   int
	events          []string
	after           map[string]func()
	faultAt         int
}

func newElevatorXferTestWorld4F53D0() *elevatorXferTestWorld4F53D0 {
	return &elevatorXferTestWorld4F53D0{
		version:         elevatorXferCurrentVersion4F53D0,
		mapResult:       1,
		readOnlyValue:   1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *elevatorXferTestWorld4F53D0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *elevatorXferTestWorld4F53D0) deps() elevatorXferDeps4F53D0[
	*elevatorXferTestObject4F53D0,
	*elevatorXferTestData4F53D0,
] {
	return elevatorXferDeps4F53D0[
		*elevatorXferTestObject4F53D0,
		*elevatorXferTestData4F53D0,
	]{
		loadUpdateData: func(object *elevatorXferTestObject4F53D0) *elevatorXferTestData4F53D0 {
			w.updateLoads++
			value := object.updateData
			w.event("load-update-data")
			return value
		},
		loadField34: func(object *elevatorXferTestObject4F53D0) uint32 {
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
		mapReadWrite: func(object *elevatorXferTestObject4F53D0, version int32) int32 {
			w.mapCalls = append(w.mapCalls, elevatorXferMapCall4F53D0{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		rwShaftExtent: func(data *elevatorXferTestData4F53D0) {
			w.shaftTransfers = append(w.shaftTransfers, data)
			_ = data.shaftExtent
			w.event("rw-shaft-extent")
		},
		rwField4: func(data *elevatorXferTestData4F53D0) {
			w.field4Transfers = append(w.field4Transfers, data)
			_ = data.field4
			w.event("rw-field4")
		},
		rwField3: func(data *elevatorXferTestData4F53D0) {
			w.field3Transfers = append(w.field3Transfers, data)
			_ = data.field3
			w.event("rw-field3")
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			w.event("read-only")
			return w.readOnlyValue
		},
		transferInventory: func(version uint16, object *elevatorXferTestObject4F53D0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, elevatorXferInventoryCall4F53D0{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *elevatorXferTestObject4F53D0, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestElevatorXfer4F53D0CachesEntryDataAndRestoresField34(t *testing.T) {
	const original = uint32(0x11223344)
	entryData := &elevatorXferTestData4F53D0{shaftExtent: 7, field3: 8, field4: 9}
	liveData := &elevatorXferTestData4F53D0{shaftExtent: 17, field3: 18, field4: 19}
	object := &elevatorXferTestObject4F53D0{field34: original, updateData: entryData}
	w := newElevatorXferTestWorld4F53D0()
	w.after["load-update-data"] = func() { object.updateData = liveData }
	w.after["rw-field3"] = func() { object.field34 = 0x80000003 }
	w.after["load-field34:2"] = func() { object.field34 = 5 }

	if got := elevatorXfer4F53D0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != original || object.updateData != liveData {
		t.Fatalf("object = Field34 %#x UpdateData %p, want %#x/%p",
			object.field34, object.updateData, original, liveData)
	}
	if !reflect.DeepEqual(w.shaftTransfers, []*elevatorXferTestData4F53D0{entryData}) ||
		!reflect.DeepEqual(w.field4Transfers, []*elevatorXferTestData4F53D0{entryData}) ||
		!reflect.DeepEqual(w.field3Transfers, []*elevatorXferTestData4F53D0{entryData}) {
		t.Fatalf("cached transfers = %p/%p/%p, want entry %p",
			w.shaftTransfers, w.field4Transfers, w.field3Transfers, entryData)
	}
	if !reflect.DeepEqual(w.versionValues, []uint16{elevatorXferCurrentVersion4F53D0}) {
		t.Fatalf("version values = %v, want [61]", w.versionValues)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []elevatorXferInventoryCall4F53D0{{
		version: 61, object: object, count: -2147483645,
	}}) {
		t.Fatalf("inventory calls = %#v, want zero-extended version and live count bits", w.inventoryCalls)
	}
	wantEvents := []string{
		"load-update-data", "load-field34:1", "rw-version", "map-read-write",
		"rw-shaft-extent", "rw-field4", "rw-field3", "load-field34:2",
		"read-only", "transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestElevatorXfer4F53D0SignedVersionThresholds(t *testing.T) {
	tests := []struct {
		name       string
		version    uint16
		wantResult int32
		wantMap    []int32
		wantData   []string
	}{
		{name: "negative one", version: 0xffff, wantResult: 1, wantMap: []int32{-1}, wantData: []string{"shaft"}},
		{name: "forty", version: 40, wantResult: 1, wantMap: []int32{40}, wantData: []string{"shaft"}},
		{name: "forty one", version: 41, wantResult: 1, wantMap: []int32{41}, wantData: []string{"shaft", "field4"}},
		{name: "sixty", version: 60, wantResult: 1, wantMap: []int32{60}, wantData: []string{"shaft", "field4"}},
		{name: "sixty one", version: 61, wantResult: 1, wantMap: []int32{61}, wantData: []string{"shaft", "field4", "field3"}},
		{name: "sixty two", version: 62, wantResult: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &elevatorXferTestData4F53D0{}
			object := &elevatorXferTestObject4F53D0{updateData: data}
			w := newElevatorXferTestWorld4F53D0()
			w.version = tc.version

			if got := elevatorXfer4F53D0(object, w.deps()); got != tc.wantResult {
				t.Fatalf("result = %d, want %d", got, tc.wantResult)
			}
			var gotMap []int32
			for _, call := range w.mapCalls {
				gotMap = append(gotMap, call.version)
			}
			var gotData []string
			if len(w.shaftTransfers) != 0 {
				gotData = append(gotData, "shaft")
			}
			if len(w.field4Transfers) != 0 {
				gotData = append(gotData, "field4")
			}
			if len(w.field3Transfers) != 0 {
				gotData = append(gotData, "field3")
			}
			if !reflect.DeepEqual(gotMap, tc.wantMap) || !reflect.DeepEqual(gotData, tc.wantData) {
				t.Fatalf("map/data = %v/%v, want %v/%v", gotMap, gotData, tc.wantMap, tc.wantData)
			}
		})
	}
}

func TestElevatorXfer4F53D0InventoryShortCircuitAndExactReadGate(t *testing.T) {
	t.Run("zero live count skips mode read", func(t *testing.T) {
		object := &elevatorXferTestObject4F53D0{field34: 7, updateData: &elevatorXferTestData4F53D0{}}
		w := newElevatorXferTestWorld4F53D0()
		w.after["rw-field3"] = func() { object.field34 = 0 }

		if got := elevatorXfer4F53D0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 0 || len(w.inventoryCalls) != 0 || object.field34 != 7 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 0/0/7",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("noncanonical read value skips inventory", func(t *testing.T) {
		object := &elevatorXferTestObject4F53D0{field34: 9, updateData: &elevatorXferTestData4F53D0{}}
		w := newElevatorXferTestWorld4F53D0()
		w.readOnlyValue = 2

		if got := elevatorXfer4F53D0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 9 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 1/0/9",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})
}

func TestElevatorXfer4F53D0FailurePrefixesDoNotRollback(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(*elevatorXferTestWorld4F53D0, *elevatorXferTestObject4F53D0)
		wantEvents []string
	}{
		{
			name: "version greater than sixty one",
			configure: func(w *elevatorXferTestWorld4F53D0, object *elevatorXferTestObject4F53D0) {
				w.version = 62
				w.after["rw-version"] = func() { object.field34 = 19 }
			},
			wantEvents: []string{"load-update-data", "load-field34:1", "rw-version"},
		},
		{
			name: "map serializer failure",
			configure: func(w *elevatorXferTestWorld4F53D0, object *elevatorXferTestObject4F53D0) {
				w.mapResult = 0
				w.after["map-read-write"] = func() { object.field34 = 23 }
			},
			wantEvents: []string{"load-update-data", "load-field34:1", "rw-version", "map-read-write"},
		},
		{
			name: "inventory failure",
			configure: func(w *elevatorXferTestWorld4F53D0, object *elevatorXferTestObject4F53D0) {
				w.inventoryResult = 0
				w.after["rw-field3"] = func() { object.field34 = 3 }
				w.after["transfer-inventory"] = func() { object.field34 = 29 }
			},
			wantEvents: []string{
				"load-update-data", "load-field34:1", "rw-version", "map-read-write",
				"rw-shaft-extent", "rw-field4", "rw-field3", "load-field34:2",
				"read-only", "transfer-inventory",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := &elevatorXferTestObject4F53D0{
				field34: 11, updateData: &elevatorXferTestData4F53D0{},
			}
			w := newElevatorXferTestWorld4F53D0()
			tc.configure(w, object)
			if got := elevatorXfer4F53D0(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) || w.field34Stores != 0 {
				t.Fatalf("events/stores = %v/%d, want %v/0", w.events, w.field34Stores, tc.wantEvents)
			}
		})
	}
}

func TestElevatorXfer4F53D0AnyNonzeroResultsAndFaultPrefixes(t *testing.T) {
	object := &elevatorXferTestObject4F53D0{
		field34: 5, updateData: &elevatorXferTestData4F53D0{},
	}
	w := newElevatorXferTestWorld4F53D0()
	w.mapResult = -7
	w.inventoryResult = -9
	if got := elevatorXfer4F53D0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}

	wantEvents := []string{
		"load-update-data", "load-field34:1", "rw-version", "map-read-write",
		"rw-shaft-extent", "rw-field4", "rw-field3", "load-field34:2",
		"read-only", "transfer-inventory", "store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			object := &elevatorXferTestObject4F53D0{
				field34: 13, updateData: &elevatorXferTestData4F53D0{},
			}
			w := newElevatorXferTestWorld4F53D0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = elevatorXfer4F53D0(object, w.deps())
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

func TestElevatorXfer4F53D0NilFaultOrder(t *testing.T) {
	t.Run("nil object faults on update data before Field34 and version", func(t *testing.T) {
		w := newElevatorXferTestWorld4F53D0()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = elevatorXfer4F53D0((*elevatorXferTestObject4F53D0)(nil), w.deps())
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
		object := &elevatorXferTestObject4F53D0{field34: 7}
		w := newElevatorXferTestWorld4F53D0()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = elevatorXfer4F53D0(object, w.deps())
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
