package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type transporterXferTestData4F5300 struct {
	target       uintptr
	targetExtent uint32
}

type transporterXferTestObject4F5300 struct {
	field34    uint32
	updateData *transporterXferTestData4F5300
}

type transporterXferMapCall4F5300 struct {
	object  *transporterXferTestObject4F5300
	version int32
}

type transporterXferInventoryCall4F5300 struct {
	version uint16
	object  *transporterXferTestObject4F5300
	count   int32
}

type transporterXferTestWorld4F5300 struct {
	version         uint16
	mapResult       int32
	readOnlyValues  []int32
	inventoryResult int32
	readExtentValue uint32

	updateLoads      int
	field34Loads     int
	versionTransfers []uint16
	mapCalls         []transporterXferMapCall4F5300
	readOnlyCalls    int
	inPlaceTransfers []*transporterXferTestData4F5300
	hasTargetCalls   []*transporterXferTestData4F5300
	extentLoads      []*transporterXferTestData4F5300
	localTransfers   []uint32
	inventoryCalls   []transporterXferInventoryCall4F5300
	field34Stores    int
	events           []string
	after            map[string]func()
	faultAt          int
}

func newTransporterXferTestWorld4F5300() *transporterXferTestWorld4F5300 {
	return &transporterXferTestWorld4F5300{
		version:         transporterXferCurrentVersion4F5300,
		mapResult:       1,
		readOnlyValues:  []int32{1, 1},
		inventoryResult: 1,
		readExtentValue: 0x55667788,
		after:           make(map[string]func()),
	}
}

func (w *transporterXferTestWorld4F5300) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *transporterXferTestWorld4F5300) deps() transporterXferDeps4F5300[
	*transporterXferTestObject4F5300,
	*transporterXferTestData4F5300,
] {
	return transporterXferDeps4F5300[
		*transporterXferTestObject4F5300,
		*transporterXferTestData4F5300,
	]{
		loadUpdateData: func(object *transporterXferTestObject4F5300) *transporterXferTestData4F5300 {
			w.updateLoads++
			value := object.updateData
			w.event("load-update-data")
			return value
		},
		loadField34: func(object *transporterXferTestObject4F5300) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.versionTransfers = append(w.versionTransfers, value)
			w.event("rw-version")
			return w.version
		},
		mapReadWrite: func(object *transporterXferTestObject4F5300, version int32) int32 {
			w.mapCalls = append(w.mapCalls, transporterXferMapCall4F5300{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		readOnly: func() int32 {
			index := w.readOnlyCalls
			w.readOnlyCalls++
			value := int32(0)
			if index < len(w.readOnlyValues) {
				value = w.readOnlyValues[index]
			}
			w.event(fmt.Sprintf("read-only:%d", w.readOnlyCalls))
			return value
		},
		rwTargetExtent: func(data *transporterXferTestData4F5300) {
			w.inPlaceTransfers = append(w.inPlaceTransfers, data)
			data.targetExtent = w.readExtentValue
			w.event("rw-target-extent")
		},
		hasTarget: func(data *transporterXferTestData4F5300) bool {
			w.hasTargetCalls = append(w.hasTargetCalls, data)
			value := data.target != 0
			w.event("has-target")
			return value
		},
		loadTargetExtent: func(data *transporterXferTestData4F5300) uint32 {
			w.extentLoads = append(w.extentLoads, data)
			value := data.targetExtent
			w.event("load-target-extent")
			return value
		},
		rwLocalTargetExtent: func(value uint32) {
			w.localTransfers = append(w.localTransfers, value)
			w.event("rw-local-target-extent")
		},
		transferInventory: func(version uint16, object *transporterXferTestObject4F5300, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, transporterXferInventoryCall4F5300{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *transporterXferTestObject4F5300, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestTransporterXfer4F5300ReadPathCachesUpdateAndRestoresField34(t *testing.T) {
	const original = uint32(0x11223344)
	liveCount := uint32(0x80000003)
	entryData := &transporterXferTestData4F5300{target: 0x123, targetExtent: 7}
	replacementData := &transporterXferTestData4F5300{target: 0x456, targetExtent: 9}
	object := &transporterXferTestObject4F5300{field34: original, updateData: entryData}
	w := newTransporterXferTestWorld4F5300()
	w.version = 0xffff
	w.readOnlyValues = []int32{2, 1}
	w.after["load-update-data"] = func() { object.updateData = replacementData }
	w.after["rw-target-extent"] = func() { object.field34 = liveCount }
	w.after["load-field34:2"] = func() { object.field34 = 5 }

	if got := transporterXfer4F5300(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != original || object.updateData != replacementData {
		t.Fatalf("object = Field34 %#x UpdateData %p, want %#x/%p", object.field34, object.updateData, original, replacementData)
	}
	if entryData.targetExtent != w.readExtentValue || replacementData.targetExtent != 9 {
		t.Fatalf("extents = %#x/%#x, want %#x/9", entryData.targetExtent, replacementData.targetExtent, w.readExtentValue)
	}
	if !reflect.DeepEqual(w.mapCalls, []transporterXferMapCall4F5300{{object: object, version: -1}}) {
		t.Fatalf("map calls = %#v, want signed -1", w.mapCalls)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []transporterXferInventoryCall4F5300{{
		version: 0xffff, object: object, count: int32(liveCount),
	}}) {
		t.Fatalf("inventory calls = %#v, want zero-extended version and live count bits", w.inventoryCalls)
	}
	wantEvents := []string{
		"load-update-data", "load-field34:1", "rw-version", "map-read-write",
		"read-only:1", "rw-target-extent", "load-field34:2", "read-only:2",
		"transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestTransporterXfer4F5300WritePathUsesLocalExtent(t *testing.T) {
	t.Run("target uses cached extent without mutating update data", func(t *testing.T) {
		data := &transporterXferTestData4F5300{target: 0x12345678, targetExtent: 0xaabbccdd}
		object := &transporterXferTestObject4F5300{field34: 7, updateData: data}
		w := newTransporterXferTestWorld4F5300()
		w.readOnlyValues = []int32{0, 0}

		if got := transporterXfer4F5300(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.localTransfers, []uint32{0xaabbccdd}) ||
			len(w.hasTargetCalls) != 1 || len(w.extentLoads) != 1 || len(w.inPlaceTransfers) != 0 {
			t.Fatalf("local/target/load/in-place = %v/%d/%d/%d", w.localTransfers,
				len(w.hasTargetCalls), len(w.extentLoads), len(w.inPlaceTransfers))
		}
		if data.targetExtent != 0xaabbccdd || len(w.inventoryCalls) != 0 || object.field34 != 7 {
			t.Fatalf("extent/inventory/Field34 = %#x/%d/%d", data.targetExtent, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("missing target serializes local zero without loading extent", func(t *testing.T) {
		data := &transporterXferTestData4F5300{targetExtent: 0xdeadbeef}
		object := &transporterXferTestObject4F5300{field34: 0, updateData: data}
		w := newTransporterXferTestWorld4F5300()
		w.readOnlyValues = []int32{0}

		if got := transporterXfer4F5300(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.localTransfers, []uint32{0}) || len(w.extentLoads) != 0 ||
			w.readOnlyCalls != 1 || data.targetExtent != 0xdeadbeef {
			t.Fatalf("local/load/mode/extent = %v/%d/%d/%#x, want [0]/0/1/0xdeadbeef",
				w.localTransfers, len(w.extentLoads), w.readOnlyCalls, data.targetExtent)
		}
	})
}

func TestTransporterXfer4F5300FailurePrefixesDoNotRollback(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(*transporterXferTestWorld4F5300, *transporterXferTestObject4F5300)
		wantEvents []string
	}{
		{
			name: "version greater than sixty",
			configure: func(w *transporterXferTestWorld4F5300, object *transporterXferTestObject4F5300) {
				w.version = 61
				w.after["rw-version"] = func() { object.field34 = 19 }
			},
			wantEvents: []string{"load-update-data", "load-field34:1", "rw-version"},
		},
		{
			name: "map serializer failure",
			configure: func(w *transporterXferTestWorld4F5300, object *transporterXferTestObject4F5300) {
				w.mapResult = 0
				w.after["map-read-write"] = func() { object.field34 = 23 }
			},
			wantEvents: []string{"load-update-data", "load-field34:1", "rw-version", "map-read-write"},
		},
		{
			name: "inventory failure",
			configure: func(w *transporterXferTestWorld4F5300, object *transporterXferTestObject4F5300) {
				w.inventoryResult = 0
				w.after["rw-target-extent"] = func() { object.field34 = 3 }
				w.after["transfer-inventory"] = func() { object.field34 = 29 }
			},
			wantEvents: []string{
				"load-update-data", "load-field34:1", "rw-version", "map-read-write",
				"read-only:1", "rw-target-extent", "load-field34:2", "read-only:2",
				"transfer-inventory",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := &transporterXferTestObject4F5300{
				field34: 11, updateData: &transporterXferTestData4F5300{},
			}
			w := newTransporterXferTestWorld4F5300()
			tc.configure(w, object)
			if got := transporterXfer4F5300(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) || w.field34Stores != 0 {
				t.Fatalf("events/stores = %v/%d, want %v/0", w.events, w.field34Stores, tc.wantEvents)
			}
		})
	}
}

func TestTransporterXfer4F5300AnyNonzeroResultsAndFaultPrefixes(t *testing.T) {
	object := &transporterXferTestObject4F5300{
		field34: 5, updateData: &transporterXferTestData4F5300{},
	}
	w := newTransporterXferTestWorld4F5300()
	w.mapResult = -7
	w.inventoryResult = -9
	if got := transporterXfer4F5300(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}

	wantEvents := []string{
		"load-update-data", "load-field34:1", "rw-version", "map-read-write",
		"read-only:1", "rw-target-extent", "load-field34:2", "read-only:2",
		"transfer-inventory", "store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			object := &transporterXferTestObject4F5300{
				field34: 13, updateData: &transporterXferTestData4F5300{},
			}
			w := newTransporterXferTestWorld4F5300()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = transporterXfer4F5300(object, w.deps())
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

func TestTransporterXfer4F5300NilObjectFaultsBeforeField34AndVersion(t *testing.T) {
	w := newTransporterXferTestWorld4F5300()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = transporterXfer4F5300((*transporterXferTestObject4F5300)(nil), w.deps())
	}()
	if recovered == nil {
		t.Fatal("nil object did not fault")
	}
	if w.updateLoads != 1 || w.field34Loads != 0 || len(w.versionTransfers) != 0 || len(w.events) != 0 {
		t.Fatalf("update/field/version/events = %d/%d/%d/%v, want 1/0/0/[]",
			w.updateLoads, w.field34Loads, len(w.versionTransfers), w.events)
	}
}
