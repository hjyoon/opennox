package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type moverXferTestData4F5730 struct {
	name string
}

type moverXferTestObject4F5730 struct {
	field34    uint32
	updateData *moverXferTestData4F5730
}

type moverXferMapCall4F5730 struct {
	object  *moverXferTestObject4F5730
	version int32
}

type moverXferWaypointCall4F5730 struct {
	object *moverXferTestObject4F5730
	data   *moverXferTestData4F5730
	slot   int
}

type moverXferInventoryCall4F5730 struct {
	version uint16
	object  *moverXferTestObject4F5730
	count   int32
}

type moverXferTestWorld4F5730 struct {
	version         uint16
	mapResult       int32
	readOnlyValues  []int32
	waypointIndexes map[int]uint32
	inventoryResult int32

	updateLoads    int
	field34Loads   int
	versionValues  []uint16
	mapCalls       []moverXferMapCall4F5730
	transfers      []string
	transferData   []*moverXferTestData4F5730
	readOnlyCalls  int
	waypointCalls  []moverXferWaypointCall4F5730
	waypointWire   []uint32
	inventoryCalls []moverXferInventoryCall4F5730
	field34Stores  int
	events         []string
	after          map[string]func()
}

func newMoverXferTestWorld4F5730() *moverXferTestWorld4F5730 {
	return &moverXferTestWorld4F5730{
		version:         moverXferCurrentVersion4F5730,
		mapResult:       1,
		readOnlyValues:  []int32{0, 1},
		waypointIndexes: map[int]uint32{3: 0x11223344, 5: 0x55667788},
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *moverXferTestWorld4F5730) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *moverXferTestWorld4F5730) transfer(name string, data *moverXferTestData4F5730) {
	w.transfers = append(w.transfers, name)
	w.transferData = append(w.transferData, data)
	w.event("rw-" + name)
}

func (w *moverXferTestWorld4F5730) deps() moverXferDeps4F5730[
	*moverXferTestObject4F5730,
	*moverXferTestData4F5730,
] {
	return moverXferDeps4F5730[
		*moverXferTestObject4F5730,
		*moverXferTestData4F5730,
	]{
		loadUpdateData: func(object *moverXferTestObject4F5730) *moverXferTestData4F5730 {
			w.updateLoads++
			value := object.updateData
			w.event("load-update-data")
			return value
		},
		loadField34: func(object *moverXferTestObject4F5730) uint32 {
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
		mapReadWrite: func(object *moverXferTestObject4F5730, version int32) int32 {
			w.mapCalls = append(w.mapCalls, moverXferMapCall4F5730{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		rwField1: func(data *moverXferTestData4F5730) { w.transfer("field1", data) },
		rwField2: func(data *moverXferTestData4F5730) { w.transfer("field2", data) },
		rwField8: func(data *moverXferTestData4F5730) { w.transfer("field8", data) },
		rwField0: func(data *moverXferTestData4F5730) { w.transfer("field0", data) },
		readOnly: func() int32 {
			call := w.readOnlyCalls
			w.readOnlyCalls++
			w.event(fmt.Sprintf("read-only:%d", call+1))
			if call >= len(w.readOnlyValues) {
				return 0
			}
			return w.readOnlyValues[call]
		},
		rwField4: func(data *moverXferTestData4F5730) { w.transfer("field4", data) },
		rwField6: func(data *moverXferTestData4F5730) { w.transfer("field6", data) },
		waypointIndex: func(object *moverXferTestObject4F5730, data *moverXferTestData4F5730, slot int) uint32 {
			w.waypointCalls = append(w.waypointCalls, moverXferWaypointCall4F5730{object: object, data: data, slot: slot})
			w.event(fmt.Sprintf("waypoint-index:%d", slot))
			return w.waypointIndexes[slot]
		},
		rwWaypointIndex: func(value uint32) {
			w.waypointWire = append(w.waypointWire, value)
			w.event(fmt.Sprintf("rw-waypoint:%#x", value))
		},
		rwSpeedBase: func(*moverXferTestObject4F5730) { w.event("rw-speed-base") },
		rwSpeedCur:  func(*moverXferTestObject4F5730) { w.event("rw-speed-cur") },
		transferInventory: func(version uint16, object *moverXferTestObject4F5730, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, moverXferInventoryCall4F5730{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *moverXferTestObject4F5730, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestMoverXfer4F5730CachesEntryDataAndPreservesWriteOrder(t *testing.T) {
	const original = uint32(0xa1b2c3d4)
	entryData := &moverXferTestData4F5730{name: "entry"}
	liveData := &moverXferTestData4F5730{name: "live"}
	object := &moverXferTestObject4F5730{field34: original, updateData: entryData}
	w := newMoverXferTestWorld4F5730()
	w.after["load-update-data"] = func() { object.updateData = liveData }
	w.after["rw-speed-cur"] = func() { object.field34 = 0x80000003 }

	if got := moverXfer4F5730(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.updateData != liveData || object.field34 != original {
		t.Fatalf("object = UpdateData %p Field34 %#x, want %p/%#x", object.updateData, object.field34, liveData, original)
	}
	if !reflect.DeepEqual(w.versionValues, []uint16{60}) {
		t.Fatalf("version values = %v, want [60]", w.versionValues)
	}
	if !reflect.DeepEqual(w.mapCalls, []moverXferMapCall4F5730{{object: object, version: 60}}) {
		t.Fatalf("map calls = %#v, want object/version 60", w.mapCalls)
	}
	for i, data := range w.transferData {
		if data != entryData {
			t.Fatalf("transfer data %d = %p, want cached entry %p", i, data, entryData)
		}
	}
	if !reflect.DeepEqual(w.transfers, []string{"field1", "field2", "field8", "field0"}) {
		t.Fatalf("transfers = %v, want write-side fixed fields", w.transfers)
	}
	if !reflect.DeepEqual(w.waypointWire, []uint32{0x11223344, 0x55667788}) {
		t.Fatalf("waypoint wire values = %#v, want ordered indexes", w.waypointWire)
	}
	if len(w.waypointCalls) != 2 || w.waypointCalls[0].data != entryData || w.waypointCalls[1].data != entryData ||
		w.waypointCalls[0].slot != 3 || w.waypointCalls[1].slot != 5 {
		t.Fatalf("waypoint calls = %#v, want cached slots 3 then 5", w.waypointCalls)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []moverXferInventoryCall4F5730{{
		version: 60, object: object, count: -2147483645,
	}}) {
		t.Fatalf("inventory calls = %#v, want zero-extended version and live count bits", w.inventoryCalls)
	}
	wantEvents := []string{
		"load-update-data", "load-field34:1", "rw-version", "map-read-write",
		"rw-field1", "rw-field2", "rw-field8", "rw-field0", "read-only:1",
		"waypoint-index:3", "rw-waypoint:0x11223344",
		"waypoint-index:5", "rw-waypoint:0x55667788",
		"rw-speed-base", "rw-speed-cur", "load-field34:2", "read-only:2",
		"transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestMoverXfer4F5730ReadModeUsesSerializedWaypointIDs(t *testing.T) {
	data := &moverXferTestData4F5730{name: "entry"}
	object := &moverXferTestObject4F5730{field34: 7, updateData: data}
	w := newMoverXferTestWorld4F5730()
	// Any nonzero value selects the direct Field4/Field6 read path, while the
	// later inventory gate independently requires exact one.
	w.readOnlyValues = []int32{2, 1}

	if got := moverXfer4F5730(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.transfers, []string{"field1", "field2", "field8", "field0", "field4", "field6"}) {
		t.Fatalf("transfers = %v, want direct waypoint ID fields", w.transfers)
	}
	if len(w.waypointCalls) != 0 || len(w.waypointWire) != 0 {
		t.Fatalf("write-side waypoint calls = %d/%d, want none", len(w.waypointCalls), len(w.waypointWire))
	}
	if len(w.inventoryCalls) != 1 || w.readOnlyCalls != 2 {
		t.Fatalf("inventory/mode calls = %d/%d, want 1/2", len(w.inventoryCalls), w.readOnlyCalls)
	}
}

func TestMoverXfer4F5730SignedVersionThresholds(t *testing.T) {
	tests := []struct {
		name          string
		version       uint16
		wantResult    int32
		wantMap       []int32
		wantTransfers []string
		wantSpeeds    bool
	}{
		{name: "negative thirty two thousand", version: 0x8000, wantResult: 1, wantMap: []int32{-32768}, wantTransfers: []string{"field1", "field2", "field8"}},
		{name: "negative one", version: 0xffff, wantResult: 1, wantMap: []int32{-1}, wantTransfers: []string{"field1", "field2", "field8"}},
		{name: "forty", version: 40, wantResult: 1, wantMap: []int32{40}, wantTransfers: []string{"field1", "field2", "field8"}},
		{name: "forty one", version: 41, wantResult: 1, wantMap: []int32{41}, wantTransfers: []string{"field1", "field2", "field8", "field0"}},
		{name: "forty two", version: 42, wantResult: 1, wantMap: []int32{42}, wantTransfers: []string{"field1", "field2", "field8", "field0"}, wantSpeeds: true},
		{name: "sixty", version: 60, wantResult: 1, wantMap: []int32{60}, wantTransfers: []string{"field1", "field2", "field8", "field0"}, wantSpeeds: true},
		{name: "sixty one", version: 61, wantResult: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := &moverXferTestObject4F5730{updateData: &moverXferTestData4F5730{}}
			w := newMoverXferTestWorld4F5730()
			w.version = tc.version
			w.readOnlyValues = []int32{0}

			if got := moverXfer4F5730(object, w.deps()); got != tc.wantResult {
				t.Fatalf("result = %d, want %d", got, tc.wantResult)
			}
			var gotMap []int32
			for _, call := range w.mapCalls {
				gotMap = append(gotMap, call.version)
			}
			if !reflect.DeepEqual(gotMap, tc.wantMap) || !reflect.DeepEqual(w.transfers, tc.wantTransfers) {
				t.Fatalf("map/transfers = %v/%v, want %v/%v", gotMap, w.transfers, tc.wantMap, tc.wantTransfers)
			}
			gotSpeeds := false
			for _, event := range w.events {
				if event == "rw-speed-base" {
					gotSpeeds = true
				}
			}
			if gotSpeeds != tc.wantSpeeds {
				t.Fatalf("speed transfers = %t, want %t", gotSpeeds, tc.wantSpeeds)
			}
		})
	}
}

func TestMoverXfer4F5730InventoryShortCircuitAndExactReadGate(t *testing.T) {
	t.Run("zero live count skips second mode read", func(t *testing.T) {
		object := &moverXferTestObject4F5730{field34: 7, updateData: &moverXferTestData4F5730{}}
		w := newMoverXferTestWorld4F5730()
		w.after["rw-speed-cur"] = func() { object.field34 = 0 }

		if got := moverXfer4F5730(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 7 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 1/0/7", w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("noncanonical inventory mode skips inventory", func(t *testing.T) {
		object := &moverXferTestObject4F5730{field34: 9, updateData: &moverXferTestData4F5730{}}
		w := newMoverXferTestWorld4F5730()
		w.readOnlyValues = []int32{0, 2}

		if got := moverXfer4F5730(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 2 || len(w.inventoryCalls) != 0 || object.field34 != 9 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 2/0/9", w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})
}

func TestMoverXfer4F5730ZeroExtendsNegativeVersionForInventory(t *testing.T) {
	object := &moverXferTestObject4F5730{field34: 3, updateData: &moverXferTestData4F5730{}}
	w := newMoverXferTestWorld4F5730()
	w.version = 0xffff
	// Version -1 skips the version-41 mode branch, so this is the inventory
	// gate's first and only mode read.
	w.readOnlyValues = []int32{1}

	if got := moverXfer4F5730(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []moverXferInventoryCall4F5730{{
		version: 0xffff, object: object, count: 3,
	}}) {
		t.Fatalf("inventory calls = %#v, want version word 0xffff", w.inventoryCalls)
	}
	if len(w.mapCalls) != 1 || w.mapCalls[0].version != -1 {
		t.Fatalf("map calls = %#v, want signed version -1", w.mapCalls)
	}
}

func TestMoverXfer4F5730FailurePrefixesDoNotRollback(t *testing.T) {
	tests := []struct {
		name        string
		configure   func(*moverXferTestWorld4F5730, *moverXferTestObject4F5730)
		wantField34 uint32
		wantEvents  []string
	}{
		{
			name: "version greater than sixty",
			configure: func(w *moverXferTestWorld4F5730, object *moverXferTestObject4F5730) {
				w.version = 61
				w.after["rw-version"] = func() { object.field34 = 19 }
			},
			wantField34: 19,
			wantEvents:  []string{"load-update-data", "load-field34:1", "rw-version"},
		},
		{
			name: "map serializer failure",
			configure: func(w *moverXferTestWorld4F5730, object *moverXferTestObject4F5730) {
				w.mapResult = 0
				w.after["map-read-write"] = func() { object.field34 = 23 }
			},
			wantField34: 23,
			wantEvents:  []string{"load-update-data", "load-field34:1", "rw-version", "map-read-write"},
		},
		{
			name: "inventory failure",
			configure: func(w *moverXferTestWorld4F5730, object *moverXferTestObject4F5730) {
				w.inventoryResult = 0
				w.after["rw-speed-cur"] = func() { object.field34 = 3 }
				w.after["transfer-inventory"] = func() { object.field34 = 29 }
			},
			wantField34: 29,
			wantEvents: []string{
				"load-update-data", "load-field34:1", "rw-version", "map-read-write",
				"rw-field1", "rw-field2", "rw-field8", "rw-field0", "read-only:1",
				"waypoint-index:3", "rw-waypoint:0x11223344",
				"waypoint-index:5", "rw-waypoint:0x55667788",
				"rw-speed-base", "rw-speed-cur", "load-field34:2", "read-only:2",
				"transfer-inventory",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := &moverXferTestObject4F5730{field34: 11, updateData: &moverXferTestData4F5730{}}
			w := newMoverXferTestWorld4F5730()
			tc.configure(w, object)

			if got := moverXfer4F5730(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if object.field34 != tc.wantField34 || w.field34Stores != 0 {
				t.Fatalf("Field34/stores = %d/%d, want %d/0", object.field34, w.field34Stores, tc.wantField34)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", w.events, tc.wantEvents)
			}
		})
	}
}
