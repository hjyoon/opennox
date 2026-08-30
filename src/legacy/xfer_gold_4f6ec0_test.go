package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type goldXferTestData4F6EC0 struct {
	amount uint32
}

type goldXferTestObject4F6EC0 struct {
	data    *goldXferTestData4F6EC0
	field34 uint32
}

type goldXferTestInventoryCall4F6EC0 struct {
	version uint16
	object  *goldXferTestObject4F6EC0
	count   int32
}

type goldXferTestWorld4F6EC0 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	inventoryResult int32

	events          []string
	after           map[string]func()
	initDataLoads   int
	field34Loads    int
	field34Stores   int
	mapVersions     []int32
	amountTransfers []*goldXferTestData4F6EC0
	modeCalls       int
	inventoryCalls  []goldXferTestInventoryCall4F6EC0
	panicAmount     bool
}

func newGoldXferTestWorld4F6EC0() *goldXferTestWorld4F6EC0 {
	return &goldXferTestWorld4F6EC0{
		version:         goldXferCurrentVersion4F6EC0,
		mapResult:       1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *goldXferTestWorld4F6EC0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *goldXferTestWorld4F6EC0) deps() goldXferDeps4F6EC0[
	*goldXferTestObject4F6EC0,
	*goldXferTestData4F6EC0,
] {
	return goldXferDeps4F6EC0[
		*goldXferTestObject4F6EC0,
		*goldXferTestData4F6EC0,
	]{
		loadInitData: func(object *goldXferTestObject4F6EC0) *goldXferTestData4F6EC0 {
			value := object.data
			w.initDataLoads++
			w.event("load-init-data")
			return value
		},
		loadField34: func(object *goldXferTestObject4F6EC0) uint32 {
			value := object.field34
			w.field34Loads++
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		storeField34: func(object *goldXferTestObject4F6EC0, value uint32) {
			object.field34 = value
			w.field34Stores++
			w.event("store-field34")
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *goldXferTestObject4F6EC0, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
		},
		rwGoldAmount: func(data *goldXferTestData4F6EC0) {
			w.amountTransfers = append(w.amountTransfers, data)
			w.event("rw-gold-amount")
			if w.panicAmount {
				panic("GoldXfer amount stream fault")
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
		transferInventory: func(version uint16, object *goldXferTestObject4F6EC0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, goldXferTestInventoryCall4F6EC0{
				version: version,
				object:  object,
				count:   count,
			})
			w.event(fmt.Sprintf("transfer-inventory:%d:%d", version, count))
			return w.inventoryResult
		},
	}
}

func goldXferAssertPanics4F6EC0(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("call did not panic")
		}
	}()
	call()
}

func TestGoldXfer4F6EC0CachesEntryPointersAndPreservesOrder(t *testing.T) {
	entry := &goldXferTestData4F6EC0{amount: 7}
	replacement := &goldXferTestData4F6EC0{amount: 99}
	object := &goldXferTestObject4F6EC0{data: entry, field34: 0x11223344}
	w := newGoldXferTestWorld4F6EC0()
	w.mapResult = -7
	w.modes = []int32{1}
	w.after["map-read-write:60"] = func() {
		object.data = replacement
	}
	w.after["rw-gold-amount"] = func() {
		entry.amount = 0xaabbccdd
		object.field34 = 0x80000003
	}
	w.after["read-mode:1=1"] = func() {
		object.field34 = 5
	}

	if got := goldXfer4F6EC0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if entry.amount != 0xaabbccdd || replacement.amount != 99 || object.data != replacement {
		t.Fatalf("cached/replacement amount and pointer = %#x/%#x/%p",
			entry.amount, replacement.amount, object.data)
	}
	if object.field34 != 0x11223344 {
		t.Fatalf("Field34 = %#x, want entry value", object.field34)
	}
	if !reflect.DeepEqual(w.amountTransfers, []*goldXferTestData4F6EC0{entry}) {
		t.Fatalf("amount transfer pointers = %v, want cached entry", w.amountTransfers)
	}
	wantInventory := []goldXferTestInventoryCall4F6EC0{{
		version: 60,
		object:  object,
		count:   int32(-2147483645),
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %+v, want %+v", w.inventoryCalls, wantInventory)
	}
	wantEvents := []string{
		"load-init-data",
		"load-field34:1",
		"rw-version:60",
		"map-read-write:60",
		"rw-gold-amount",
		"load-field34:2",
		"read-mode:1=1",
		"transfer-inventory:60:-2147483645",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, wantEvents)
	}
}

func TestGoldXfer4F6EC0SignedVersionGateAndCommonPrefix(t *testing.T) {
	tests := []struct {
		name       string
		version    uint16
		mapResult  int32
		wantResult int32
		wantMap    []int32
		wantAmount int
		wantStores int
	}{
		{name: "current", version: 60, mapResult: 1, wantResult: 1, wantMap: []int32{60}, wantAmount: 1, wantStores: 1},
		{name: "positive too new", version: 61, mapResult: 1, wantResult: 0},
		{name: "largest positive", version: 0x7fff, mapResult: 1, wantResult: 0},
		{name: "most negative", version: 0x8000, mapResult: 1, wantResult: 1, wantMap: []int32{-32768}, wantAmount: 1, wantStores: 1},
		{name: "minus one", version: 0xffff, mapResult: 1, wantResult: 1, wantMap: []int32{-1}, wantAmount: 1, wantStores: 1},
		{name: "common failure", version: 60, mapResult: 0, wantResult: 0, wantMap: []int32{60}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := &goldXferTestObject4F6EC0{data: &goldXferTestData4F6EC0{}}
			w := newGoldXferTestWorld4F6EC0()
			w.version = test.version
			w.mapResult = test.mapResult
			if got := goldXfer4F6EC0(object, w.deps()); got != test.wantResult {
				t.Fatalf("result = %d, want %d", got, test.wantResult)
			}
			if !reflect.DeepEqual(w.mapVersions, test.wantMap) {
				t.Fatalf("map versions = %v, want %v", w.mapVersions, test.wantMap)
			}
			if len(w.amountTransfers) != test.wantAmount || w.field34Stores != test.wantStores {
				t.Fatalf("amount transfers/stores = %d/%d, want %d/%d",
					len(w.amountTransfers), w.field34Stores, test.wantAmount, test.wantStores)
			}
		})
	}
}

func TestGoldXfer4F6EC0LiveInventoryGateAndUnsignedVersion(t *testing.T) {
	t.Run("zero live count skips mode", func(t *testing.T) {
		object := &goldXferTestObject4F6EC0{data: &goldXferTestData4F6EC0{}, field34: 9}
		w := newGoldXferTestWorld4F6EC0()
		w.after["rw-gold-amount"] = func() { object.field34 = 0 }
		if got := goldXfer4F6EC0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 0 || len(w.inventoryCalls) != 0 || object.field34 != 9 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d", w.modeCalls, len(w.inventoryCalls), object.field34)
		}
	})

	for _, mode := range []int32{-1, 0, 2} {
		t.Run(fmt.Sprintf("mode_%d_skips", mode), func(t *testing.T) {
			object := &goldXferTestObject4F6EC0{data: &goldXferTestData4F6EC0{}, field34: 3}
			w := newGoldXferTestWorld4F6EC0()
			w.modes = []int32{mode}
			if got := goldXfer4F6EC0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if len(w.inventoryCalls) != 0 || object.field34 != 3 {
				t.Fatalf("inventory/Field34 = %d/%d", len(w.inventoryCalls), object.field34)
			}
		})
	}

	t.Run("exact one passes zero-extended version and signed count bits", func(t *testing.T) {
		object := &goldXferTestObject4F6EC0{data: &goldXferTestData4F6EC0{}, field34: 0x80000001}
		w := newGoldXferTestWorld4F6EC0()
		w.version = 0xffff
		w.modes = []int32{1}
		if got := goldXfer4F6EC0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		want := []goldXferTestInventoryCall4F6EC0{{version: 0xffff, object: object, count: -2147483647}}
		if !reflect.DeepEqual(w.inventoryCalls, want) {
			t.Fatalf("inventory calls = %+v, want %+v", w.inventoryCalls, want)
		}
	})
}

func TestGoldXfer4F6EC0InventoryFailureKeepsAllPrefixSideEffects(t *testing.T) {
	entry := &goldXferTestData4F6EC0{amount: 1}
	object := &goldXferTestObject4F6EC0{data: entry, field34: 7}
	w := newGoldXferTestWorld4F6EC0()
	w.modes = []int32{1}
	w.inventoryResult = 0
	w.after["rw-gold-amount"] = func() {
		entry.amount = 0x12345678
		object.field34 = 11
	}
	w.after["transfer-inventory:60:11"] = func() {
		object.field34 = 13
	}

	if got := goldXfer4F6EC0(object, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if entry.amount != 0x12345678 || object.field34 != 13 || w.field34Stores != 0 {
		t.Fatalf("amount/Field34/stores = %#x/%d/%d, want transferred/13/0",
			entry.amount, object.field34, w.field34Stores)
	}
}

func TestGoldXfer4F6EC0FaultBoundaries(t *testing.T) {
	t.Run("nil object faults before the first event", func(t *testing.T) {
		w := newGoldXferTestWorld4F6EC0()
		goldXferAssertPanics4F6EC0(t, func() { goldXfer4F6EC0(nil, w.deps()) })
		if len(w.events) != 0 {
			t.Fatalf("events = %v, want none", w.events)
		}
	})

	t.Run("nil InitData reaches amount transfer after common prefix", func(t *testing.T) {
		object := &goldXferTestObject4F6EC0{field34: 5}
		w := newGoldXferTestWorld4F6EC0()
		w.panicAmount = true
		goldXferAssertPanics4F6EC0(t, func() { goldXfer4F6EC0(object, w.deps()) })
		want := []string{
			"load-init-data",
			"load-field34:1",
			"rw-version:60",
			"map-read-write:60",
			"rw-gold-amount",
		}
		if !reflect.DeepEqual(w.events, want) || w.field34Stores != 0 {
			t.Fatalf("events/stores = %v/%d, want %v/0", w.events, w.field34Stores, want)
		}
	})
}
