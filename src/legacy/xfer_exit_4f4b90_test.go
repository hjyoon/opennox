package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type exitXferTestData4F4B90 struct {
	mapNameSize uint32
	bytes       map[uint32]byte
}

type exitXferTestObject4F4B90 struct {
	field34     uint32
	collideData *exitXferTestData4F4B90
}

type exitXferMapCall4F4B90 struct {
	object  *exitXferTestObject4F4B90
	version int32
}

type exitXferDataCall4F4B90 struct {
	data *exitXferTestData4F4B90
	size uint32
}

type exitXferByteCall4F4B90 struct {
	data   *exitXferTestData4F4B90
	offset uint32
}

type exitXferInventoryCall4F4B90 struct {
	version uint16
	object  *exitXferTestObject4F4B90
	count   int32
}

type exitXferTestWorld4F4B90 struct {
	version         uint16
	transferredSize uint32
	mapResult       int32
	readOnlyDefault int32
	readOnlyValues  []int32
	inventoryResult int32

	collideDataLoads  int
	mapNameSizeLoads  int
	field34Loads      int
	versionTransfers  []uint16
	mapCalls          []exitXferMapCall4F4B90
	readOnlyCalls     int
	mapNameSizeCalls  []uint32
	mapNameCalls      []exitXferDataCall4F4B90
	legacyByteCalls   []exitXferByteCall4F4B90
	mapNameByteLoads  []exitXferByteCall4F4B90
	destinationXCalls []*exitXferTestData4F4B90
	destinationYCalls []*exitXferTestData4F4B90
	inventoryCalls    []exitXferInventoryCall4F4B90
	field34Stores     int
	events            []string
	after             map[string]func()
	faultAt           int
}

func newExitXferTestWorld4F4B90() *exitXferTestWorld4F4B90 {
	return &exitXferTestWorld4F4B90{
		version:         exitXferCurrentVersion4F4B90,
		transferredSize: 5,
		mapResult:       1,
		readOnlyDefault: 1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *exitXferTestWorld4F4B90) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *exitXferTestWorld4F4B90) deps() exitXferDeps4F4B90[
	*exitXferTestObject4F4B90,
	*exitXferTestData4F4B90,
] {
	return exitXferDeps4F4B90[
		*exitXferTestObject4F4B90,
		*exitXferTestData4F4B90,
	]{
		loadCollideData: func(object *exitXferTestObject4F4B90) *exitXferTestData4F4B90 {
			w.collideDataLoads++
			value := object.collideData
			w.event("load-collide-data")
			return value
		},
		mapNameSizeWithNUL: func(data *exitXferTestData4F4B90) uint32 {
			w.mapNameSizeLoads++
			value := data.mapNameSize
			w.event("map-name-size")
			return value
		},
		loadField34: func(object *exitXferTestObject4F4B90) uint32 {
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
		mapReadWrite: func(object *exitXferTestObject4F4B90, version int32) int32 {
			w.mapCalls = append(w.mapCalls, exitXferMapCall4F4B90{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			value := w.readOnlyDefault
			if w.readOnlyCalls <= len(w.readOnlyValues) {
				value = w.readOnlyValues[w.readOnlyCalls-1]
			}
			w.event(fmt.Sprintf("read-only:%d", w.readOnlyCalls))
			return value
		},
		rwMapNameSize: func(value uint32) uint32 {
			w.mapNameSizeCalls = append(w.mapNameSizeCalls, value)
			w.event("rw-map-name-size")
			return w.transferredSize
		},
		rwMapName: func(data *exitXferTestData4F4B90, size uint32) {
			w.mapNameCalls = append(w.mapNameCalls, exitXferDataCall4F4B90{data: data, size: size})
			w.event("rw-map-name")
		},
		rwLegacyMapNameByte: func(data *exitXferTestData4F4B90, offset uint32) {
			w.legacyByteCalls = append(w.legacyByteCalls, exitXferByteCall4F4B90{data: data, offset: offset})
			w.event(fmt.Sprintf("rw-legacy-byte:%d", offset))
		},
		loadMapNameByte: func(data *exitXferTestData4F4B90, offset uint32) byte {
			w.mapNameByteLoads = append(w.mapNameByteLoads, exitXferByteCall4F4B90{data: data, offset: offset})
			value := data.bytes[offset]
			w.event(fmt.Sprintf("load-map-byte:%d", offset))
			return value
		},
		rwDestinationX: func(data *exitXferTestData4F4B90) {
			w.destinationXCalls = append(w.destinationXCalls, data)
			w.event("rw-destination-x")
		},
		rwDestinationY: func(data *exitXferTestData4F4B90) {
			w.destinationYCalls = append(w.destinationYCalls, data)
			w.event("rw-destination-y")
		},
		transferInventory: func(version uint16, object *exitXferTestObject4F4B90, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, exitXferInventoryCall4F4B90{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *exitXferTestObject4F4B90, value uint32) {
			w.field34Stores++
			w.event("store-field34")
			object.field34 = value
		},
	}
}

func TestExitXfer4F4B90PreservesEntryCachesAndLiveCount(t *testing.T) {
	const originalAfterLength = uint32(0x55667788)
	liveCount := uint32(0x80000003)
	entryData := &exitXferTestData4F4B90{mapNameSize: 17, bytes: map[uint32]byte{0: 'A', 1: 0}}
	liveData := &exitXferTestData4F4B90{mapNameSize: 23, bytes: map[uint32]byte{0: 'B', 1: 0}}
	afterMapData := &exitXferTestData4F4B90{mapNameSize: 29, bytes: map[uint32]byte{0: 'C', 1: 0}}
	object := &exitXferTestObject4F4B90{field34: 0x11223344, collideData: entryData}
	w := newExitXferTestWorld4F4B90()
	w.transferredSize = 0x10001
	w.after["map-name-size"] = func() {
		object.collideData = liveData
		object.field34 = originalAfterLength
	}
	w.after["map-read-write"] = func() {
		object.collideData = afterMapData
		object.field34 = 7
	}
	w.after["rw-map-name"] = func() { object.collideData = liveData }
	w.after["rw-destination-y"] = func() { object.field34 = liveCount }
	w.after["load-field34:2"] = func() { object.field34 = 9 }

	if got := exitXfer4F4B90(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != originalAfterLength {
		t.Fatalf("Field34 = %#08x, want cache %#08x", object.field34, originalAfterLength)
	}
	if !reflect.DeepEqual(w.versionTransfers, []uint16{exitXferCurrentVersion4F4B90}) {
		t.Fatalf("version args = %v, want [60]", w.versionTransfers)
	}
	if !reflect.DeepEqual(w.mapCalls, []exitXferMapCall4F4B90{{object: object, version: 60}}) {
		t.Fatalf("map calls = %#v, want signed version 60", w.mapCalls)
	}
	if !reflect.DeepEqual(w.mapNameSizeCalls, []uint32{17}) ||
		!reflect.DeepEqual(w.mapNameCalls, []exitXferDataCall4F4B90{{data: entryData, size: 0x10001}}) {
		t.Fatalf("name size/data calls = %v/%#v, want cached data and unbounded wire size", w.mapNameSizeCalls, w.mapNameCalls)
	}
	if !reflect.DeepEqual(w.destinationXCalls, []*exitXferTestData4F4B90{entryData}) ||
		!reflect.DeepEqual(w.destinationYCalls, []*exitXferTestData4F4B90{entryData}) {
		t.Fatalf("destination calls = %v/%v, want cached entry data %p", w.destinationXCalls, w.destinationYCalls, entryData)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []exitXferInventoryCall4F4B90{{
		version: 60, object: object, count: int32(liveCount),
	}}) {
		t.Fatalf("inventory calls = %#v, want signed live count", w.inventoryCalls)
	}
	wantEvents := []string{
		"load-collide-data", "map-name-size", "load-field34:1", "rw-version", "map-read-write",
		"rw-map-name-size", "rw-map-name", "rw-destination-x", "rw-destination-y",
		"load-field34:2", "read-only:1", "transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestExitXfer4F4B90VersionedMapNameAndDestinationWire(t *testing.T) {
	t.Run("legacy exact-one read transfers bytes until live NUL", func(t *testing.T) {
		data := &exitXferTestData4F4B90{mapNameSize: 91, bytes: make(map[uint32]byte)}
		object := &exitXferTestObject4F4B90{field34: 3, collideData: data}
		w := newExitXferTestWorld4F4B90()
		w.version = 1
		w.readOnlyValues = []int32{1, 1}
		w.after["rw-legacy-byte:0"] = func() { data.bytes[0] = 'Q' }
		w.after["rw-legacy-byte:1"] = func() { data.bytes[1] = 0 }

		if got := exitXfer4F4B90(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		wantBytes := []exitXferByteCall4F4B90{{data: data, offset: 0}, {data: data, offset: 1}}
		if !reflect.DeepEqual(w.legacyByteCalls, wantBytes) || !reflect.DeepEqual(w.mapNameByteLoads, wantBytes) {
			t.Fatalf("byte transfer/loads = %#v/%#v, want %#v", w.legacyByteCalls, w.mapNameByteLoads, wantBytes)
		}
		if len(w.mapNameSizeCalls) != 0 || len(w.mapNameCalls) != 0 ||
			len(w.destinationXCalls) != 0 || len(w.destinationYCalls) != 0 {
			t.Fatalf("length/block/destination calls = %d/%d/%d/%d, want all zero",
				len(w.mapNameSizeCalls), len(w.mapNameCalls), len(w.destinationXCalls), len(w.destinationYCalls))
		}
		if w.readOnlyCalls != 2 || len(w.inventoryCalls) != 1 {
			t.Fatalf("read-only/inventory calls = %d/%d, want 2/1", w.readOnlyCalls, len(w.inventoryCalls))
		}
	})

	t.Run("negative version uses entry length in non-one mode", func(t *testing.T) {
		data := &exitXferTestData4F4B90{mapNameSize: 257, bytes: map[uint32]byte{0: 0}}
		object := &exitXferTestObject4F4B90{field34: 0, collideData: data}
		w := newExitXferTestWorld4F4B90()
		w.version = 0xffff
		w.readOnlyValues = []int32{2}

		if got := exitXfer4F4B90(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.mapCalls, []exitXferMapCall4F4B90{{object: object, version: -1}}) {
			t.Fatalf("map calls = %#v, want signed version -1", w.mapCalls)
		}
		if !reflect.DeepEqual(w.mapNameCalls, []exitXferDataCall4F4B90{{data: data, size: 257}}) ||
			len(w.mapNameSizeCalls) != 0 || len(w.legacyByteCalls) != 0 {
			t.Fatalf("block/size/byte calls = %#v/%v/%v, want unbounded 257/none/none",
				w.mapNameCalls, w.mapNameSizeCalls, w.legacyByteCalls)
		}
		if w.readOnlyCalls != 1 || len(w.destinationXCalls) != 0 || len(w.inventoryCalls) != 0 {
			t.Fatalf("read-only/destination/inventory = %d/%d/%d, want 1/0/0",
				w.readOnlyCalls, len(w.destinationXCalls), len(w.inventoryCalls))
		}
	})

	t.Run("version two accepts wire length above eighty and omits coordinates", func(t *testing.T) {
		data := &exitXferTestData4F4B90{mapNameSize: 4, bytes: map[uint32]byte{0: 0}}
		object := &exitXferTestObject4F4B90{field34: 0, collideData: data}
		w := newExitXferTestWorld4F4B90()
		w.version = 2
		w.transferredSize = 0x10001

		if got := exitXfer4F4B90(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.mapNameSizeCalls, []uint32{4}) ||
			!reflect.DeepEqual(w.mapNameCalls, []exitXferDataCall4F4B90{{data: data, size: 0x10001}}) {
			t.Fatalf("size/name calls = %v/%#v, want 4 and unbounded wire size", w.mapNameSizeCalls, w.mapNameCalls)
		}
		if len(w.destinationXCalls) != 0 || len(w.destinationYCalls) != 0 || w.readOnlyCalls != 0 {
			t.Fatalf("destination/read-only calls = %d/%d/%d, want 0/0/0",
				len(w.destinationXCalls), len(w.destinationYCalls), w.readOnlyCalls)
		}
	})
}

func TestExitXfer4F4B90InventoryGatesUseLiveState(t *testing.T) {
	t.Run("zero live count skips read-only", func(t *testing.T) {
		data := &exitXferTestData4F4B90{mapNameSize: 1, bytes: map[uint32]byte{0: 0}}
		object := &exitXferTestObject4F4B90{field34: 23, collideData: data}
		w := newExitXferTestWorld4F4B90()
		w.after["rw-destination-y"] = func() { object.field34 = 0 }

		if got := exitXfer4F4B90(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.field34Loads != 2 || w.readOnlyCalls != 0 || len(w.inventoryCalls) != 0 || object.field34 != 23 {
			t.Fatalf("field loads/read-only/inventory/Field34 = %d/%d/%d/%d, want 2/0/0/23",
				w.field34Loads, w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("non-one read skips inventory", func(t *testing.T) {
		data := &exitXferTestData4F4B90{mapNameSize: 1, bytes: map[uint32]byte{0: 0}}
		object := &exitXferTestObject4F4B90{field34: 29, collideData: data}
		w := newExitXferTestWorld4F4B90()
		w.readOnlyDefault = 2

		if got := exitXfer4F4B90(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 29 {
			t.Fatalf("read-only/inventory/Field34 = %d/%d/%d, want 1/0/29",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})
}

func TestExitXfer4F4B90FailurePrefixesDoNotRollback(t *testing.T) {
	t.Run("version greater than sixty", func(t *testing.T) {
		data := &exitXferTestData4F4B90{mapNameSize: 2, bytes: map[uint32]byte{0: 0}}
		object := &exitXferTestObject4F4B90{field34: 7, collideData: data}
		w := newExitXferTestWorld4F4B90()
		w.version = 61
		w.after["rw-version"] = func() { object.field34 = 19 }

		if got := exitXfer4F4B90(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-collide-data", "map-name-size", "load-field34:1", "rw-version"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 19 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/stores = %v/%d/%d, want %v/19/0", w.events, object.field34, w.field34Stores, want)
		}
	})

	t.Run("map serializer failure", func(t *testing.T) {
		data := &exitXferTestData4F4B90{mapNameSize: 3, bytes: map[uint32]byte{0: 0}}
		object := &exitXferTestObject4F4B90{field34: 11, collideData: data}
		w := newExitXferTestWorld4F4B90()
		w.mapResult = 0
		w.after["map-read-write"] = func() { object.field34 = 29 }

		if got := exitXfer4F4B90(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-collide-data", "map-name-size", "load-field34:1", "rw-version", "map-read-write"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 29 || len(w.mapNameCalls) != 0 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/name/stores = %v/%d/%d/%d, want %v/29/0/0",
				w.events, object.field34, len(w.mapNameCalls), w.field34Stores, want)
		}
	})

	t.Run("inventory transfer failure", func(t *testing.T) {
		data := &exitXferTestData4F4B90{mapNameSize: 4, bytes: map[uint32]byte{0: 0}}
		object := &exitXferTestObject4F4B90{field34: 13, collideData: data}
		w := newExitXferTestWorld4F4B90()
		w.inventoryResult = 0
		w.after["rw-destination-y"] = func() { object.field34 = 5 }
		w.after["transfer-inventory"] = func() { object.field34 = 31 }

		if got := exitXfer4F4B90(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 31 || w.field34Stores != 0 {
			t.Fatalf("Field34/stores = %d/%d, want 31/0", object.field34, w.field34Stores)
		}
	})
}

func TestExitXfer4F4B90TreatsAnyNonzeroCallbackResultAsSuccess(t *testing.T) {
	data := &exitXferTestData4F4B90{mapNameSize: 1, bytes: map[uint32]byte{0: 0}}
	object := &exitXferTestObject4F4B90{field34: 37, collideData: data}
	w := newExitXferTestWorld4F4B90()
	w.mapResult = -7
	w.inventoryResult = -9

	if got := exitXfer4F4B90(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if len(w.inventoryCalls) != 1 || object.field34 != 37 {
		t.Fatalf("inventory/Field34 = %d/%d, want 1/37", len(w.inventoryCalls), object.field34)
	}
}

func TestExitXfer4F4B90FaultPrefixes(t *testing.T) {
	wantEvents := []string{
		"load-collide-data", "map-name-size", "load-field34:1", "rw-version", "map-read-write",
		"rw-map-name-size", "rw-map-name", "rw-destination-x", "rw-destination-y",
		"load-field34:2", "read-only:1", "transfer-inventory", "store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			data := &exitXferTestData4F4B90{mapNameSize: 6, bytes: map[uint32]byte{0: 0}}
			object := &exitXferTestObject4F4B90{field34: 41, collideData: data}
			w := newExitXferTestWorld4F4B90()
			w.faultAt = faultAt
			w.after["rw-destination-y"] = func() { object.field34 = 2 }

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = exitXfer4F4B90(object, w.deps())
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

func TestExitXfer4F4B90NilFaultOrder(t *testing.T) {
	t.Run("nil object faults on CollideData before map-name length or Field34", func(t *testing.T) {
		w := newExitXferTestWorld4F4B90()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = exitXfer4F4B90((*exitXferTestObject4F4B90)(nil), w.deps())
		}()
		if recovered == nil {
			t.Fatal("nil object did not fault")
		}
		if w.collideDataLoads != 1 || w.mapNameSizeLoads != 0 || w.field34Loads != 0 || len(w.events) != 0 {
			t.Fatalf("loads/events = %d/%d/%d/%v, want 1/0/0/[]",
				w.collideDataLoads, w.mapNameSizeLoads, w.field34Loads, w.events)
		}
	})

	t.Run("nil CollideData faults during map-name length before Field34", func(t *testing.T) {
		object := &exitXferTestObject4F4B90{field34: 9}
		w := newExitXferTestWorld4F4B90()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = exitXfer4F4B90(object, w.deps())
		}()
		if recovered == nil {
			t.Fatal("nil CollideData did not fault")
		}
		if w.mapNameSizeLoads != 1 || w.field34Loads != 0 || !reflect.DeepEqual(w.events, []string{"load-collide-data"}) {
			t.Fatalf("name/field/events = %d/%d/%v, want 1/0/[load-collide-data]",
				w.mapNameSizeLoads, w.field34Loads, w.events)
		}
	})
}
