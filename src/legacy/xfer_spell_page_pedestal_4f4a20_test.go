package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type spellPagePedestalXferTestData4F4A20 struct {
	spell uint32
}

type spellPagePedestalXferTestObject4F4A20 struct {
	field34     uint32
	collideData *spellPagePedestalXferTestData4F4A20
}

type spellPagePedestalXferMapCall4F4A20 struct {
	object  *spellPagePedestalXferTestObject4F4A20
	version int32
}

type spellPagePedestalXferInventoryCall4F4A20 struct {
	version uint16
	object  *spellPagePedestalXferTestObject4F4A20
	count   int32
}

type spellPagePedestalXferTestWorld4F4A20 struct {
	version         uint16
	mapResult       int32
	readOnlyValue   int32
	inventoryResult int32

	field34Loads      int
	versionTransfers  []uint16
	mapCalls          []spellPagePedestalXferMapCall4F4A20
	collideDataLoads  int
	spellPayloadCalls []*spellPagePedestalXferTestData4F4A20
	readOnlyCalls     int
	inventoryCalls    []spellPagePedestalXferInventoryCall4F4A20
	field34Stores     int
	events            []string
	after             map[string]func()
	faultAt           int
}

func newSpellPagePedestalXferTestWorld4F4A20() *spellPagePedestalXferTestWorld4F4A20 {
	return &spellPagePedestalXferTestWorld4F4A20{
		version:         spellPagePedestalXferCurrentVersion4F4A20,
		mapResult:       1,
		readOnlyValue:   1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *spellPagePedestalXferTestWorld4F4A20) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *spellPagePedestalXferTestWorld4F4A20) deps() spellPagePedestalXferDeps4F4A20[
	*spellPagePedestalXferTestObject4F4A20,
	*spellPagePedestalXferTestData4F4A20,
] {
	return spellPagePedestalXferDeps4F4A20[
		*spellPagePedestalXferTestObject4F4A20,
		*spellPagePedestalXferTestData4F4A20,
	]{
		loadField34: func(object *spellPagePedestalXferTestObject4F4A20) uint32 {
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
		mapReadWrite: func(object *spellPagePedestalXferTestObject4F4A20, version int32) int32 {
			w.mapCalls = append(w.mapCalls, spellPagePedestalXferMapCall4F4A20{
				object: object, version: version,
			})
			w.event("map-read-write")
			return w.mapResult
		},
		loadCollideData: func(object *spellPagePedestalXferTestObject4F4A20) *spellPagePedestalXferTestData4F4A20 {
			w.collideDataLoads++
			value := object.collideData
			w.event("load-collide-data")
			return value
		},
		rwSpellPayload: func(data *spellPagePedestalXferTestData4F4A20) {
			w.spellPayloadCalls = append(w.spellPayloadCalls, data)
			w.event("rw-spell-payload")
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			w.event("read-only")
			return w.readOnlyValue
		},
		transferInventory: func(version uint16, object *spellPagePedestalXferTestObject4F4A20, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, spellPagePedestalXferInventoryCall4F4A20{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *spellPagePedestalXferTestObject4F4A20, value uint32) {
			w.field34Stores++
			w.event("store-field34")
			object.field34 = value
		},
	}
}

func TestSpellPagePedestalXfer4F4A20PreservesEntryCacheAndLiveLoads(t *testing.T) {
	const original = uint32(0x11223344)
	liveCount := uint32(0x80000005)
	entryData := &spellPagePedestalXferTestData4F4A20{spell: 1}
	liveData := &spellPagePedestalXferTestData4F4A20{spell: 2}
	afterLoadData := &spellPagePedestalXferTestData4F4A20{spell: 3}
	object := &spellPagePedestalXferTestObject4F4A20{
		field34: original, collideData: entryData,
	}
	w := newSpellPagePedestalXferTestWorld4F4A20()
	w.version = 0xffff
	w.after["rw-version"] = func() { object.field34 = 0x55667788 }
	w.after["map-read-write"] = func() {
		object.collideData = liveData
		object.field34 = 7
	}
	w.after["load-collide-data"] = func() { object.collideData = afterLoadData }
	w.after["rw-spell-payload"] = func() { object.field34 = liveCount }
	w.after["load-field34:2"] = func() { object.field34 = 9 }

	if got := spellPagePedestalXfer4F4A20(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != original {
		t.Fatalf("Field34 = %#08x, want entry value %#08x", object.field34, original)
	}
	if !reflect.DeepEqual(w.versionTransfers, []uint16{spellPagePedestalXferCurrentVersion4F4A20}) {
		t.Fatalf("version transfer arguments = %v, want [60]", w.versionTransfers)
	}
	wantMap := []spellPagePedestalXferMapCall4F4A20{{object: object, version: -1}}
	if !reflect.DeepEqual(w.mapCalls, wantMap) {
		t.Fatalf("map calls = %#v, want %#v", w.mapCalls, wantMap)
	}
	if !reflect.DeepEqual(w.spellPayloadCalls, []*spellPagePedestalXferTestData4F4A20{liveData}) {
		t.Fatalf("spell payload calls = %p, want cached live pointer %p", w.spellPayloadCalls, liveData)
	}
	wantInventory := []spellPagePedestalXferInventoryCall4F4A20{{
		version: 0xffff, object: object, count: int32(liveCount),
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want %#v", w.inventoryCalls, wantInventory)
	}
	wantEvents := []string{
		"load-field34:1", "rw-version", "map-read-write",
		"load-collide-data", "rw-spell-payload", "load-field34:2",
		"read-only", "transfer-inventory", "store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func TestSpellPagePedestalXfer4F4A20PayloadAndInventoryGates(t *testing.T) {
	t.Run("non-one read skips inventory after payload", func(t *testing.T) {
		data := &spellPagePedestalXferTestData4F4A20{spell: 11}
		object := &spellPagePedestalXferTestObject4F4A20{field34: 17, collideData: data}
		w := newSpellPagePedestalXferTestWorld4F4A20()
		w.readOnlyValue = 2
		w.after["rw-spell-payload"] = func() { object.field34 = 3 }

		if got := spellPagePedestalXfer4F4A20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.spellPayloadCalls) != 1 || w.spellPayloadCalls[0] != data {
			t.Fatalf("spell payload calls = %v, want [%p]", w.spellPayloadCalls, data)
		}
		if w.readOnlyCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 17 {
			t.Fatalf("read-only/inventory/Field34 = %d/%d/%d, want 1/0/17",
				w.readOnlyCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("zero live count skips read-only after payload", func(t *testing.T) {
		data := &spellPagePedestalXferTestData4F4A20{spell: 12}
		object := &spellPagePedestalXferTestObject4F4A20{field34: 23, collideData: data}
		w := newSpellPagePedestalXferTestWorld4F4A20()
		w.after["rw-spell-payload"] = func() { object.field34 = 0 }

		if got := spellPagePedestalXfer4F4A20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.spellPayloadCalls) != 1 || w.readOnlyCalls != 0 || len(w.inventoryCalls) != 0 {
			t.Fatalf("payload/read-only/inventory calls = %d/%d/%d, want 1/0/0",
				len(w.spellPayloadCalls), w.readOnlyCalls, len(w.inventoryCalls))
		}
		if object.field34 != 23 {
			t.Fatalf("Field34 = %d, want 23", object.field34)
		}
	})

	t.Run("nil collide data is forwarded without a guard", func(t *testing.T) {
		object := &spellPagePedestalXferTestObject4F4A20{}
		w := newSpellPagePedestalXferTestWorld4F4A20()

		if got := spellPagePedestalXfer4F4A20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.spellPayloadCalls) != 1 || w.spellPayloadCalls[0] != nil {
			t.Fatalf("spell payload calls = %v, want one nil pointer", w.spellPayloadCalls)
		}
	})
}

func TestSpellPagePedestalXfer4F4A20FailurePrefixesDoNotRollback(t *testing.T) {
	t.Run("version greater than sixty", func(t *testing.T) {
		object := &spellPagePedestalXferTestObject4F4A20{field34: 7}
		w := newSpellPagePedestalXferTestWorld4F4A20()
		w.version = 61
		w.after["rw-version"] = func() { object.field34 = 19 }

		if got := spellPagePedestalXfer4F4A20(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "rw-version"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 19 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/stores = %v/%d/%d, want %v/19/0",
				w.events, object.field34, w.field34Stores, want)
		}
	})

	t.Run("map serializer failure", func(t *testing.T) {
		object := &spellPagePedestalXferTestObject4F4A20{field34: 11}
		w := newSpellPagePedestalXferTestWorld4F4A20()
		w.mapResult = 0
		w.after["map-read-write"] = func() { object.field34 = 29 }

		if got := spellPagePedestalXfer4F4A20(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "rw-version", "map-read-write"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 29 || w.collideDataLoads != 0 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/collide-loads/stores = %v/%d/%d/%d, want %v/29/0/0",
				w.events, object.field34, w.collideDataLoads, w.field34Stores, want)
		}
	})

	t.Run("inventory transfer failure", func(t *testing.T) {
		data := &spellPagePedestalXferTestData4F4A20{spell: 13}
		object := &spellPagePedestalXferTestObject4F4A20{field34: 13, collideData: data}
		w := newSpellPagePedestalXferTestWorld4F4A20()
		w.inventoryResult = 0
		w.after["rw-spell-payload"] = func() { object.field34 = 5 }
		w.after["transfer-inventory"] = func() { object.field34 = 31 }

		if got := spellPagePedestalXfer4F4A20(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{
			"load-field34:1", "rw-version", "map-read-write",
			"load-collide-data", "rw-spell-payload", "load-field34:2",
			"read-only", "transfer-inventory",
		}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 31 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/stores = %v/%d/%d, want %v/31/0",
				w.events, object.field34, w.field34Stores, want)
		}
	})
}

func TestSpellPagePedestalXfer4F4A20TreatsAnyNonzeroCallbackResultAsSuccess(t *testing.T) {
	object := &spellPagePedestalXferTestObject4F4A20{field34: 37}
	w := newSpellPagePedestalXferTestWorld4F4A20()
	w.mapResult = -7
	w.inventoryResult = -9

	if got := spellPagePedestalXfer4F4A20(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if len(w.inventoryCalls) != 1 {
		t.Fatalf("inventory calls = %v, want one", w.inventoryCalls)
	}
}

func TestSpellPagePedestalXfer4F4A20FaultPrefixes(t *testing.T) {
	wantEvents := []string{
		"load-field34:1", "rw-version", "map-read-write",
		"load-collide-data", "rw-spell-payload", "load-field34:2",
		"read-only", "transfer-inventory", "store-field34",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			object := &spellPagePedestalXferTestObject4F4A20{
				field34:     41,
				collideData: &spellPagePedestalXferTestData4F4A20{spell: 14},
			}
			w := newSpellPagePedestalXferTestWorld4F4A20()
			w.faultAt = faultAt
			w.after["rw-spell-payload"] = func() { object.field34 = 2 }

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				_ = spellPagePedestalXfer4F4A20(object, w.deps())
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

func TestSpellPagePedestalXfer4F4A20NilObjectFaultsBeforeVersionTransfer(t *testing.T) {
	w := newSpellPagePedestalXferTestWorld4F4A20()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = spellPagePedestalXfer4F4A20((*spellPagePedestalXferTestObject4F4A20)(nil), w.deps())
	}()
	if recovered == nil {
		t.Fatal("nil object did not fault")
	}
	if w.field34Loads != 1 || len(w.versionTransfers) != 0 || len(w.events) != 0 {
		t.Fatalf("field loads/version transfers/events = %d/%d/%v, want 1/0/[]",
			w.field34Loads, len(w.versionTransfers), w.events)
	}
}
