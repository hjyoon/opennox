package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type holeXferTestData4F51D0 struct {
	scriptFlags uint32
	scriptFunc  int32
	destination [2]int32
	extent      uint32
	netCode     uint16
	reserved22  uint16
	field24     uint32
}

type holeXferTestObject4F51D0 struct {
	field34     uint32
	scriptData  uintptr
	collideData *holeXferTestData4F51D0
}

type holeXferMapCall4F51D0 struct {
	object  *holeXferTestObject4F51D0
	version int32
}

type holeXferScriptCall4F51D0 struct {
	data   *holeXferTestData4F51D0
	script uintptr
	offset uintptr
}

type holeXferInventoryCall4F51D0 struct {
	version uint16
	object  *holeXferTestObject4F51D0
	count   int32
}

type holeXferTestWorld4F51D0 struct {
	version         uint16
	mapResult       int32
	readOnlyDefault int32
	inventoryResult int32

	field34Loads     int
	scriptDataLoads  int
	collideDataLoads int
	versionTransfers []uint16
	mapCalls         []holeXferMapCall4F51D0
	field24Transfers []*holeXferTestData4F51D0
	scriptCalls      []holeXferScriptCall4F51D0
	destinationCalls []*holeXferTestData4F51D0
	extentTransfers  []*holeXferTestData4F51D0
	netCodeTransfers []*holeXferTestData4F51D0
	readOnlyCalls    int
	inventoryCalls   []holeXferInventoryCall4F51D0
	field34Stores    int
	events           []string
	after            map[string]func()
	faultAt          int
}

func newHoleXferTestWorld4F51D0() *holeXferTestWorld4F51D0 {
	return &holeXferTestWorld4F51D0{
		version:         holeXferCurrentVersion4F51D0,
		mapResult:       1,
		readOnlyDefault: 1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *holeXferTestWorld4F51D0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *holeXferTestWorld4F51D0) deps() holeXferDeps4F51D0[
	*holeXferTestObject4F51D0,
	*holeXferTestData4F51D0,
	uintptr,
] {
	return holeXferDeps4F51D0[
		*holeXferTestObject4F51D0,
		*holeXferTestData4F51D0,
		uintptr,
	]{
		loadField34: func(object *holeXferTestObject4F51D0) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		loadScriptData: func(object *holeXferTestObject4F51D0) uintptr {
			w.scriptDataLoads++
			value := object.scriptData
			w.event("load-script-data")
			return value
		},
		loadCollideData: func(object *holeXferTestObject4F51D0) *holeXferTestData4F51D0 {
			w.collideDataLoads++
			value := object.collideData
			w.event("load-collide-data")
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.versionTransfers = append(w.versionTransfers, value)
			w.event("rw-version")
			return w.version
		},
		mapReadWrite: func(object *holeXferTestObject4F51D0, version int32) int32 {
			w.mapCalls = append(w.mapCalls, holeXferMapCall4F51D0{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		rwField24: func(data *holeXferTestData4F51D0) {
			data.field24 = 0x24242424
			w.field24Transfers = append(w.field24Transfers, data)
			w.event("rw-field24")
		},
		storeField24: func(data *holeXferTestData4F51D0, value uint32) {
			data.field24 = value
			w.event("store-field24")
		},
		transferScript: func(data *holeXferTestData4F51D0, script uintptr, offset uintptr) {
			data.scriptFlags = 0x11223344
			data.scriptFunc = 0x55667788
			w.scriptCalls = append(w.scriptCalls, holeXferScriptCall4F51D0{
				data: data, script: script, offset: offset,
			})
			w.event("transfer-script")
		},
		rwDestinationXY: func(data *holeXferTestData4F51D0) {
			data.destination = [2]int32{101, -202}
			w.destinationCalls = append(w.destinationCalls, data)
			w.event("rw-destination-xy")
		},
		rwDestinationExtent: func(data *holeXferTestData4F51D0) {
			data.extent = 0x33445566
			w.extentTransfers = append(w.extentTransfers, data)
			w.event("rw-destination-extent")
		},
		rwDestinationNetCode: func(data *holeXferTestData4F51D0) {
			data.netCode = 0x7788
			w.netCodeTransfers = append(w.netCodeTransfers, data)
			w.event("rw-destination-net-code")
		},
		storeScriptFunc: func(data *holeXferTestData4F51D0, value int32) {
			data.scriptFunc = value
			w.event("store-script-func")
		},
		storeScriptFlags: func(data *holeXferTestData4F51D0, value uint32) {
			data.scriptFlags = value
			w.event("store-script-flags")
		},
		storeDestinationExtent: func(data *holeXferTestData4F51D0, value uint32) {
			data.extent = value
			w.event("store-destination-extent")
		},
		storeDestinationNetCode: func(data *holeXferTestData4F51D0, value uint16) {
			data.netCode = value
			w.event("store-destination-net-code")
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			w.event("read-only")
			return w.readOnlyDefault
		},
		transferInventory: func(version uint16, object *holeXferTestObject4F51D0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, holeXferInventoryCall4F51D0{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *holeXferTestObject4F51D0, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestHoleXfer4F51D0PreservesEntryCachesAndReadOrder(t *testing.T) {
	liveCount := uint32(0x80000003)
	entryData := &holeXferTestData4F51D0{reserved22: 0xbeef}
	liveData := &holeXferTestData4F51D0{}
	object := &holeXferTestObject4F51D0{
		field34: 0x11223344, scriptData: 0x123, collideData: entryData,
	}
	w := newHoleXferTestWorld4F51D0()
	w.after["load-field34:1"] = func() { object.field34 = 7 }
	w.after["load-script-data"] = func() { object.scriptData = 0x456 }
	w.after["load-collide-data"] = func() { object.collideData = liveData }
	w.after["rw-destination-net-code"] = func() { object.field34 = liveCount }
	w.after["load-field34:2"] = func() { object.field34 = 9 }

	if got := holeXfer4F51D0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != 0x11223344 {
		t.Fatalf("Field34 = %#08x, want entry cache %#08x", object.field34, uint32(0x11223344))
	}
	if object.collideData != liveData || object.scriptData != 0x456 {
		t.Fatalf("live pointers = %p/%#x, want %p/%#x", object.collideData, object.scriptData, liveData, uintptr(0x456))
	}
	if !reflect.DeepEqual(w.versionTransfers, []uint16{60}) ||
		!reflect.DeepEqual(w.mapCalls, []holeXferMapCall4F51D0{{object: object, version: 60}}) {
		t.Fatalf("version/map calls = %v/%#v, want 60", w.versionTransfers, w.mapCalls)
	}
	if !reflect.DeepEqual(w.field24Transfers, []*holeXferTestData4F51D0{entryData}) ||
		!reflect.DeepEqual(w.destinationCalls, []*holeXferTestData4F51D0{entryData}) ||
		!reflect.DeepEqual(w.extentTransfers, []*holeXferTestData4F51D0{entryData}) ||
		!reflect.DeepEqual(w.netCodeTransfers, []*holeXferTestData4F51D0{entryData}) {
		t.Fatalf("data calls did not use entry CollideData: %#v/%#v/%#v/%#v",
			w.field24Transfers, w.destinationCalls, w.extentTransfers, w.netCodeTransfers)
	}
	if !reflect.DeepEqual(w.scriptCalls, []holeXferScriptCall4F51D0{{
		data: entryData, script: 0x123, offset: 128,
	}}) {
		t.Fatalf("script calls = %#v, want entry ScriptData +128 context", w.scriptCalls)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []holeXferInventoryCall4F51D0{{
		version: 60, object: object, count: int32(liveCount),
	}}) {
		t.Fatalf("inventory calls = %#v, want live count bits", w.inventoryCalls)
	}
	if entryData.reserved22 != 0xbeef {
		t.Fatalf("reserved22 = %#x, want untouched 0xbeef", entryData.reserved22)
	}
	wantEvents := []string{
		"load-field34:1", "load-script-data", "load-collide-data", "rw-version", "map-read-write",
		"rw-field24", "transfer-script", "rw-destination-xy", "rw-destination-extent",
		"rw-destination-net-code", "load-field34:2", "read-only", "transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestHoleXfer4F51D0SignedVersionAndWireBoundaries(t *testing.T) {
	t.Run("signed version above 60 rejects after entry caches", func(t *testing.T) {
		object := &holeXferTestObject4F51D0{field34: 8, scriptData: 9, collideData: &holeXferTestData4F51D0{}}
		w := newHoleXferTestWorld4F51D0()
		w.version = 61
		if got := holeXfer4F51D0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "load-script-data", "load-collide-data", "rw-version"}
		if !reflect.DeepEqual(w.events, want) || w.field34Stores != 0 {
			t.Fatalf("events/stores = %v/%d, want %v/0", w.events, w.field34Stores, want)
		}
	})

	t.Run("negative version is signed for branches and unsigned for inventory", func(t *testing.T) {
		liveCount := uint32(0x80000002)
		data := &holeXferTestData4F51D0{
			scriptFlags: 7, scriptFunc: 8, extent: 9, netCode: 10, reserved22: 0xcafe, field24: 11,
		}
		object := &holeXferTestObject4F51D0{field34: liveCount, collideData: data}
		w := newHoleXferTestWorld4F51D0()
		w.version = 0xffff
		if got := holeXfer4F51D0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.mapCalls, []holeXferMapCall4F51D0{{object: object, version: -1}}) {
			t.Fatalf("map calls = %#v, want signed -1", w.mapCalls)
		}
		if !reflect.DeepEqual(w.inventoryCalls, []holeXferInventoryCall4F51D0{{
			version: 0xffff, object: object, count: int32(liveCount),
		}}) {
			t.Fatalf("inventory calls = %#v, want zero-extended version word", w.inventoryCalls)
		}
		if data.field24 != 0 || data.scriptFunc != -1 || data.scriptFlags != 0 ||
			data.extent != 0 || data.netCode != 0 || data.reserved22 != 0xcafe {
			t.Fatalf("legacy data = %#v, want zero/-1 defaults with reserved22 untouched", data)
		}
		if len(w.scriptCalls) != 0 || len(w.extentTransfers) != 0 || len(w.netCodeTransfers) != 0 {
			t.Fatalf("new wire calls = %d/%d/%d, want zero", len(w.scriptCalls), len(w.extentTransfers), len(w.netCodeTransfers))
		}
		wantLegacy := []string{
			"store-field24", "rw-destination-xy", "store-script-func", "store-script-flags",
			"store-destination-extent", "store-destination-net-code",
		}
		if !reflect.DeepEqual(w.events[5:11], wantLegacy) {
			t.Fatalf("legacy events = %v, want %v", w.events[5:11], wantLegacy)
		}
	})

	for _, tc := range []struct {
		name             string
		version          uint16
		wantField24Wire  int
		wantField24Value uint32
		wantScriptWire   int
		wantExtentWire   int
	}{
		{name: "version 40 uses both legacy branches", version: 40, wantField24Value: 0},
		{name: "version 41 zeros field24 but transfers script state", version: 41, wantField24Value: 0, wantScriptWire: 1, wantExtentWire: 1},
		{name: "version 42 transfers field24 and script state", version: 42, wantField24Wire: 1, wantField24Value: 0x24242424, wantScriptWire: 1, wantExtentWire: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &holeXferTestData4F51D0{field24: 0xaaaaaaaa, reserved22: 0xface}
			object := &holeXferTestObject4F51D0{scriptData: 1, collideData: data}
			w := newHoleXferTestWorld4F51D0()
			w.version = tc.version
			if got := holeXfer4F51D0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if len(w.field24Transfers) != tc.wantField24Wire || data.field24 != tc.wantField24Value {
				t.Fatalf("field24 wire/value = %d/%#x, want %d/%#x",
					len(w.field24Transfers), data.field24, tc.wantField24Wire, tc.wantField24Value)
			}
			if len(w.scriptCalls) != tc.wantScriptWire || len(w.extentTransfers) != tc.wantExtentWire ||
				len(w.netCodeTransfers) != tc.wantExtentWire {
				t.Fatalf("script/extent/net wire = %d/%d/%d, want %d/%d/%d",
					len(w.scriptCalls), len(w.extentTransfers), len(w.netCodeTransfers),
					tc.wantScriptWire, tc.wantExtentWire, tc.wantExtentWire)
			}
			if data.reserved22 != 0xface {
				t.Fatalf("reserved22 = %#x, want untouched 0xface", data.reserved22)
			}
		})
	}
}

func TestHoleXfer4F51D0FailureAndInventoryPrefixes(t *testing.T) {
	t.Run("common failure does not restore entry Field34", func(t *testing.T) {
		object := &holeXferTestObject4F51D0{field34: 17, collideData: &holeXferTestData4F51D0{}}
		w := newHoleXferTestWorld4F51D0()
		w.mapResult = 0
		w.after["map-read-write"] = func() { object.field34 = 23 }
		if got := holeXfer4F51D0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 23 || w.field34Stores != 0 {
			t.Fatalf("Field34/stores = %d/%d, want 23/0", object.field34, w.field34Stores)
		}
	})

	t.Run("inventory failure does not restore entry Field34", func(t *testing.T) {
		object := &holeXferTestObject4F51D0{field34: 17, collideData: &holeXferTestData4F51D0{}}
		w := newHoleXferTestWorld4F51D0()
		w.inventoryResult = 0
		w.after["rw-destination-net-code"] = func() { object.field34 = 23 }
		w.after["transfer-inventory"] = func() { object.field34 = 29 }
		if got := holeXfer4F51D0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 29 || w.field34Stores != 0 {
			t.Fatalf("Field34/stores = %d/%d, want 29/0", object.field34, w.field34Stores)
		}
	})

	t.Run("zero live count skips mode read and inventory", func(t *testing.T) {
		object := &holeXferTestObject4F51D0{field34: 17, collideData: &holeXferTestData4F51D0{}}
		w := newHoleXferTestWorld4F51D0()
		w.after["rw-destination-net-code"] = func() { object.field34 = 0 }
		if got := holeXfer4F51D0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.readOnlyCalls != 0 || len(w.inventoryCalls) != 0 || object.field34 != 17 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 0/0/17", w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	for _, mode := range []int32{0, 2, -1} {
		t.Run(fmt.Sprintf("mode-%d-is-not-exact-one", mode), func(t *testing.T) {
			object := &holeXferTestObject4F51D0{field34: 17, collideData: &holeXferTestData4F51D0{}}
			w := newHoleXferTestWorld4F51D0()
			w.readOnlyDefault = mode
			w.mapResult = -9
			if got := holeXfer4F51D0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want normalized 1", got)
			}
			if w.readOnlyCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 17 {
				t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 1/0/17", w.readOnlyCalls, len(w.inventoryCalls), object.field34)
			}
		})
	}
}

func TestHoleXfer4F51D0FaultPrefixes(t *testing.T) {
	wantEvents := []string{
		"load-field34:1", "load-script-data", "load-collide-data", "rw-version", "map-read-write",
		"rw-field24", "transfer-script", "rw-destination-xy", "rw-destination-extent",
		"rw-destination-net-code", "load-field34:2", "read-only", "transfer-inventory", "store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			object := &holeXferTestObject4F51D0{
				field34: 3, scriptData: 1, collideData: &holeXferTestData4F51D0{},
			}
			w := newHoleXferTestWorld4F51D0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = holeXfer4F51D0(object, w.deps())
			}()
			if recovered == nil {
				t.Fatalf("fault %d did not panic", faultAt)
			}
			if !reflect.DeepEqual(w.events, wantEvents[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", w.events, wantEvents[:faultAt])
			}
			if faultAt < len(wantEvents) && w.field34Stores != 0 {
				t.Fatalf("fault prefix restored Field34 early at event %d", faultAt)
			}
		})
	}
}

func TestHoleXfer4F51D0NilFaultOrder(t *testing.T) {
	t.Run("nil object faults on entry Field34 before other caches", func(t *testing.T) {
		w := newHoleXferTestWorld4F51D0()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = holeXfer4F51D0((*holeXferTestObject4F51D0)(nil), w.deps())
		}()
		if recovered == nil {
			t.Fatal("nil object did not fault")
		}
		if w.field34Loads != 1 || w.scriptDataLoads != 0 || w.collideDataLoads != 0 || len(w.events) != 0 {
			t.Fatalf("loads/events = %d/%d/%d/%v, want 1/0/0/[]",
				w.field34Loads, w.scriptDataLoads, w.collideDataLoads, w.events)
		}
	})

	t.Run("nil CollideData reaches version and common before field24 fault", func(t *testing.T) {
		object := &holeXferTestObject4F51D0{field34: 3, scriptData: 1}
		w := newHoleXferTestWorld4F51D0()
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			_ = holeXfer4F51D0(object, w.deps())
		}()
		if recovered == nil {
			t.Fatal("nil CollideData did not fault")
		}
		want := []string{"load-field34:1", "load-script-data", "load-collide-data", "rw-version", "map-read-write"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("nil ScriptData is forwarded as null context", func(t *testing.T) {
		data := &holeXferTestData4F51D0{}
		object := &holeXferTestObject4F51D0{collideData: data}
		w := newHoleXferTestWorld4F51D0()
		if got := holeXfer4F51D0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.scriptCalls, []holeXferScriptCall4F51D0{{data: data, script: 0, offset: 128}}) {
			t.Fatalf("script calls = %#v, want null context", w.scriptCalls)
		}
	})
}
