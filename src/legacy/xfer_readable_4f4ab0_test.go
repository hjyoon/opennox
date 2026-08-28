package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type readableXferTestData4F4AB0 struct {
	textSize  uint32
	transient uint32
}

type readableXferTestObject4F4AB0 struct {
	field34 uint32
	useData *readableXferTestData4F4AB0
}

type readableXferMapCall4F4AB0 struct {
	object  *readableXferTestObject4F4AB0
	version int32
}

type readableXferTextCall4F4AB0 struct {
	data *readableXferTestData4F4AB0
	size uint32
}

type readableXferInventoryCall4F4AB0 struct {
	version uint16
	object  *readableXferTestObject4F4AB0
	count   int32
}

type readableXferTestWorld4F4AB0 struct {
	version         uint16
	transferredSize uint32
	mapResult       int32
	readOnlyValue   int32
	inventoryResult int32

	useDataLoads      int
	textSizeLoads     int
	field34Loads      int
	versionTransfers  []uint16
	mapCalls          []readableXferMapCall4F4AB0
	textSizeTransfers []uint32
	textCalls         []readableXferTextCall4F4AB0
	readOnlyCalls     int
	transientClears   []*readableXferTestData4F4AB0
	inventoryCalls    []readableXferInventoryCall4F4AB0
	field34Stores     int
	events            []string
	after             map[string]func()
	faultAt           int
}

func newReadableXferTestWorld4F4AB0() *readableXferTestWorld4F4AB0 {
	return &readableXferTestWorld4F4AB0{
		version:         readableXferCurrentVersion4F4AB0,
		transferredSize: 5,
		mapResult:       1,
		readOnlyValue:   1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *readableXferTestWorld4F4AB0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *readableXferTestWorld4F4AB0) deps() readableXferDeps4F4AB0[
	*readableXferTestObject4F4AB0,
	*readableXferTestData4F4AB0,
] {
	return readableXferDeps4F4AB0[
		*readableXferTestObject4F4AB0,
		*readableXferTestData4F4AB0,
	]{
		loadUseData: func(object *readableXferTestObject4F4AB0) *readableXferTestData4F4AB0 {
			w.useDataLoads++
			value := object.useData
			w.event("load-use-data")
			return value
		},
		textSizeWithNUL: func(data *readableXferTestData4F4AB0) uint32 {
			w.textSizeLoads++
			value := data.textSize
			w.event("text-size")
			return value
		},
		loadField34: func(object *readableXferTestObject4F4AB0) uint32 {
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
		mapReadWrite: func(object *readableXferTestObject4F4AB0, version int32) int32 {
			w.mapCalls = append(w.mapCalls, readableXferMapCall4F4AB0{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		rwTextSize: func(value uint32) uint32 {
			w.textSizeTransfers = append(w.textSizeTransfers, value)
			w.event("rw-text-size")
			return w.transferredSize
		},
		rwText: func(data *readableXferTestData4F4AB0, size uint32) {
			w.textCalls = append(w.textCalls, readableXferTextCall4F4AB0{data: data, size: size})
			w.event("rw-text")
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			w.event("read-only")
			return w.readOnlyValue
		},
		clearTransientRead: func(data *readableXferTestData4F4AB0) {
			w.transientClears = append(w.transientClears, data)
			w.event("clear-transient")
			data.transient = 0
		},
		transferInventory: func(version uint16, object *readableXferTestObject4F4AB0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, readableXferInventoryCall4F4AB0{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *readableXferTestObject4F4AB0, value uint32) {
			w.field34Stores++
			w.event("store-field34")
			object.field34 = value
		},
	}
}

func TestReadableXfer4F4AB0PreservesEntryCachesAndLiveCount(t *testing.T) {
	entryData := &readableXferTestData4F4AB0{textSize: 17, transient: 0x11111111}
	liveData := &readableXferTestData4F4AB0{textSize: 23, transient: 0x22222222}
	afterSizeData := &readableXferTestData4F4AB0{textSize: 29, transient: 0x33333333}
	object := &readableXferTestObject4F4AB0{field34: 0x11223344, useData: entryData}
	w := newReadableXferTestWorld4F4AB0()
	w.version = 2
	w.transferredSize = 0x80000005
	w.after["text-size"] = func() {
		object.useData = liveData
		object.field34 = 0x55667788
	}
	w.after["map-read-write"] = func() {
		object.useData = afterSizeData
		object.field34 = 7
	}
	w.after["rw-text"] = func() { object.field34 = 0x80000003 }
	w.after["load-field34:2"] = func() { object.field34 = 9 }

	if got := readableXfer4F4AB0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != 0x55667788 {
		t.Fatalf("Field34 = %#08x, want value cached after strlen", object.field34)
	}
	if entryData.transient != 0 || liveData.transient != 0x22222222 || afterSizeData.transient != 0x33333333 {
		t.Fatalf("transient state = %#x/%#x/%#x, want cached data only cleared",
			entryData.transient, liveData.transient, afterSizeData.transient)
	}
	if !reflect.DeepEqual(w.versionTransfers, []uint16{readableXferCurrentVersion4F4AB0}) {
		t.Fatalf("version args = %v, want [60]", w.versionTransfers)
	}
	if !reflect.DeepEqual(w.mapCalls, []readableXferMapCall4F4AB0{{object: object, version: 2}}) {
		t.Fatalf("map calls = %#v, want signed version 2", w.mapCalls)
	}
	if !reflect.DeepEqual(w.textSizeTransfers, []uint32{17}) {
		t.Fatalf("text-size args = %v, want entry size 17", w.textSizeTransfers)
	}
	if !reflect.DeepEqual(w.textCalls, []readableXferTextCall4F4AB0{{data: entryData, size: 0x80000005}}) {
		t.Fatalf("text calls = %#v, want cached pointer and unbounded wire size", w.textCalls)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []readableXferInventoryCall4F4AB0{{
		version: 2, object: object, count: -2147483645,
	}}) {
		t.Fatalf("inventory calls = %#v, want signed live count", w.inventoryCalls)
	}
	wantEvents := []string{
		"load-use-data", "text-size", "load-field34:1", "rw-version", "map-read-write",
		"rw-text-size", "rw-text", "read-only", "clear-transient", "load-field34:2",
		"transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestReadableXfer4F4AB0VersionedTextWireAndReadGate(t *testing.T) {
	t.Run("negative signed version omits length and reaches serializer", func(t *testing.T) {
		data := &readableXferTestData4F4AB0{textSize: 257, transient: 9}
		object := &readableXferTestObject4F4AB0{field34: 3, useData: data}
		w := newReadableXferTestWorld4F4AB0()
		w.version = 0xffff
		w.readOnlyValue = 0

		if got := readableXfer4F4AB0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.textSizeTransfers) != 0 || !reflect.DeepEqual(w.textCalls, []readableXferTextCall4F4AB0{{data: data, size: 257}}) {
			t.Fatalf("size/text calls = %v/%#v, want no length record and unbounded 257", w.textSizeTransfers, w.textCalls)
		}
		if !reflect.DeepEqual(w.mapCalls, []readableXferMapCall4F4AB0{{object: object, version: -1}}) {
			t.Fatalf("map calls = %#v, want signed version -1", w.mapCalls)
		}
		if data.transient != 9 || w.field34Loads != 1 || len(w.inventoryCalls) != 0 {
			t.Fatalf("transient/field loads/inventory = %d/%d/%d, want 9/1/0",
				data.transient, w.field34Loads, len(w.inventoryCalls))
		}
	})

	t.Run("version two accepts wire length above use-data text capacity", func(t *testing.T) {
		data := &readableXferTestData4F4AB0{textSize: 4, transient: 7}
		object := &readableXferTestObject4F4AB0{field34: 5, useData: data}
		w := newReadableXferTestWorld4F4AB0()
		w.version = 2
		w.transferredSize = 0x10001
		w.readOnlyValue = 2

		if got := readableXfer4F4AB0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.textCalls, []readableXferTextCall4F4AB0{{data: data, size: 0x10001}}) {
			t.Fatalf("text calls = %#v, want original unbounded wire size", w.textCalls)
		}
		if data.transient != 7 || w.readOnlyCalls != 1 || w.field34Loads != 1 {
			t.Fatalf("transient/read-only/field loads = %d/%d/%d, want 7/1/1", data.transient, w.readOnlyCalls, w.field34Loads)
		}
	})

	t.Run("exact-one read clears transient before zero live count", func(t *testing.T) {
		data := &readableXferTestData4F4AB0{textSize: 2, transient: 11}
		object := &readableXferTestObject4F4AB0{field34: 13, useData: data}
		w := newReadableXferTestWorld4F4AB0()
		w.after["rw-text"] = func() { object.field34 = 0 }

		if got := readableXfer4F4AB0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if data.transient != 0 || w.field34Loads != 2 || len(w.inventoryCalls) != 0 || object.field34 != 13 {
			t.Fatalf("transient/loads/inventory/Field34 = %d/%d/%d/%d, want 0/2/0/13",
				data.transient, w.field34Loads, len(w.inventoryCalls), object.field34)
		}
	})
}

func TestReadableXfer4F4AB0FailurePrefixesDoNotRollback(t *testing.T) {
	t.Run("version greater than sixty", func(t *testing.T) {
		data := &readableXferTestData4F4AB0{textSize: 3, transient: 1}
		object := &readableXferTestObject4F4AB0{field34: 7, useData: data}
		w := newReadableXferTestWorld4F4AB0()
		w.version = 61
		w.after["rw-version"] = func() { object.field34 = 19 }

		if got := readableXfer4F4AB0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-use-data", "text-size", "load-field34:1", "rw-version"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 19 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/stores = %v/%d/%d, want %v/19/0", w.events, object.field34, w.field34Stores, want)
		}
	})

	t.Run("map serializer failure", func(t *testing.T) {
		data := &readableXferTestData4F4AB0{textSize: 4, transient: 2}
		object := &readableXferTestObject4F4AB0{field34: 11, useData: data}
		w := newReadableXferTestWorld4F4AB0()
		w.mapResult = 0
		w.after["map-read-write"] = func() { object.field34 = 29 }

		if got := readableXfer4F4AB0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-use-data", "text-size", "load-field34:1", "rw-version", "map-read-write"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 29 || len(w.textCalls) != 0 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/text/stores = %v/%d/%d/%d, want %v/29/0/0",
				w.events, object.field34, len(w.textCalls), w.field34Stores, want)
		}
	})

	t.Run("inventory transfer failure leaves cleared transient and live field", func(t *testing.T) {
		data := &readableXferTestData4F4AB0{textSize: 5, transient: 3}
		object := &readableXferTestObject4F4AB0{field34: 13, useData: data}
		w := newReadableXferTestWorld4F4AB0()
		w.inventoryResult = 0
		w.after["rw-text"] = func() { object.field34 = 5 }
		w.after["transfer-inventory"] = func() { object.field34 = 31 }

		if got := readableXfer4F4AB0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if data.transient != 0 || object.field34 != 31 || w.field34Stores != 0 {
			t.Fatalf("transient/Field34/stores = %d/%d/%d, want 0/31/0", data.transient, object.field34, w.field34Stores)
		}
	})
}

func TestReadableXfer4F4AB0TreatsAnyNonzeroCallbackResultAsSuccess(t *testing.T) {
	data := &readableXferTestData4F4AB0{textSize: 2, transient: 4}
	object := &readableXferTestObject4F4AB0{field34: 37, useData: data}
	w := newReadableXferTestWorld4F4AB0()
	w.mapResult = -7
	w.inventoryResult = -9

	if got := readableXfer4F4AB0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if len(w.inventoryCalls) != 1 || data.transient != 0 {
		t.Fatalf("inventory/transient = %d/%d, want 1/0", len(w.inventoryCalls), data.transient)
	}
}

func TestReadableXfer4F4AB0FaultPrefixes(t *testing.T) {
	wantEvents := []string{
		"load-use-data", "text-size", "load-field34:1", "rw-version", "map-read-write",
		"rw-text-size", "rw-text", "read-only", "clear-transient", "load-field34:2",
		"transfer-inventory", "store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			data := &readableXferTestData4F4AB0{textSize: 6, transient: 5}
			object := &readableXferTestObject4F4AB0{field34: 41, useData: data}
			w := newReadableXferTestWorld4F4AB0()
			w.faultAt = faultAt
			w.after["rw-text"] = func() { object.field34 = 2 }

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = readableXfer4F4AB0(object, w.deps())
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

func TestReadableXfer4F4AB0NilFaultOrder(t *testing.T) {
	t.Run("nil object faults on UseData before strlen or Field34", func(t *testing.T) {
		w := newReadableXferTestWorld4F4AB0()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = readableXfer4F4AB0((*readableXferTestObject4F4AB0)(nil), w.deps())
		}()
		if recovered == nil {
			t.Fatal("nil object did not fault")
		}
		if w.useDataLoads != 1 || w.textSizeLoads != 0 || w.field34Loads != 0 || len(w.events) != 0 {
			t.Fatalf("loads/events = %d/%d/%d/%v, want 1/0/0/[]",
				w.useDataLoads, w.textSizeLoads, w.field34Loads, w.events)
		}
	})

	t.Run("nil UseData faults during strlen before Field34", func(t *testing.T) {
		object := &readableXferTestObject4F4AB0{field34: 9}
		w := newReadableXferTestWorld4F4AB0()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = readableXfer4F4AB0(object, w.deps())
		}()
		if recovered == nil {
			t.Fatal("nil UseData did not fault")
		}
		if w.textSizeLoads != 1 || w.field34Loads != 0 || !reflect.DeepEqual(w.events, []string{"load-use-data"}) {
			t.Fatalf("text/field/events = %d/%d/%v, want 1/0/[load-use-data]",
				w.textSizeLoads, w.field34Loads, w.events)
		}
	})
}
