package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type obeliskXferTestData4F6F60 struct {
	mana int32
}

type obeliskXferTestObject4F6F60 struct {
	data    *obeliskXferTestData4F6F60
	field34 uint32
	extent  uint32
}

type obeliskXferTestDrawable4F6F60 struct {
	name string
	next *obeliskXferTestDrawable4F6F60
}

type obeliskXferTestSync4F6F60 struct {
	object *obeliskXferTestObject4F6F60
	bits   uint32
}

type obeliskXferTestInventory4F6F60 struct {
	version uint16
	object  *obeliskXferTestObject4F6F60
	count   int32
}

type obeliskXferTestWorld4F6F60 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	gameResult      int32
	manaRead        *int32
	minimapRead     uint8
	static          *obeliskXferTestDrawable4F6F60
	first           *obeliskXferTestDrawable4F6F60
	inventoryResult int32

	events          []string
	after           map[string]func()
	updateDataLoads int
	field34Loads    int
	field34Stores   int
	mapVersions     []int32
	manaTransfers   []*obeliskXferTestData4F6F60
	modeCalls       int
	manaLoads       int
	syncCalls       []obeliskXferTestSync4F6F60
	gameMasks       []uint32
	extentLoads     int
	staticCodes     []uint32
	firstCalls      int
	nextCalls       int
	minimapInputs   []uint8
	inventoryCalls  []obeliskXferTestInventory4F6F60
	panicMana       bool
}

func newObeliskXferTestWorld4F6F60() *obeliskXferTestWorld4F6F60 {
	return &obeliskXferTestWorld4F6F60{
		version:         obeliskXferCurrentVersion4F6F60,
		mapResult:       1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *obeliskXferTestWorld4F6F60) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *obeliskXferTestWorld4F6F60) deps() obeliskXferDeps4F6F60[
	*obeliskXferTestObject4F6F60,
	*obeliskXferTestData4F6F60,
	*obeliskXferTestDrawable4F6F60,
] {
	return obeliskXferDeps4F6F60[
		*obeliskXferTestObject4F6F60,
		*obeliskXferTestData4F6F60,
		*obeliskXferTestDrawable4F6F60,
	]{
		loadUpdateData: func(object *obeliskXferTestObject4F6F60) *obeliskXferTestData4F6F60 {
			value := object.data
			w.updateDataLoads++
			w.event("load-update-data")
			return value
		},
		loadField34: func(object *obeliskXferTestObject4F6F60) uint32 {
			value := object.field34
			w.field34Loads++
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		storeField34: func(object *obeliskXferTestObject4F6F60, value uint32) {
			object.field34 = value
			w.field34Stores++
			w.event("store-field34")
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *obeliskXferTestObject4F6F60, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
		},
		rwMana: func(data *obeliskXferTestData4F6F60) {
			w.manaTransfers = append(w.manaTransfers, data)
			if w.manaRead != nil {
				data.mana = *w.manaRead
			}
			w.event("rw-mana")
			if w.panicMana {
				panic("ObeliskXfer mana stream fault")
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
		loadMana: func(data *obeliskXferTestData4F6F60) int32 {
			value := data.mana
			w.manaLoads++
			w.event(fmt.Sprintf("load-mana:%d", value))
			return value
		},
		syncManaLevel: func(object *obeliskXferTestObject4F6F60, level float32) {
			bits := math.Float32bits(level)
			w.syncCalls = append(w.syncCalls, obeliskXferTestSync4F6F60{object: object, bits: bits})
			w.event(fmt.Sprintf("sync-mana:%08x", bits))
		},
		gameFlags: func(mask uint32) int32 {
			w.gameMasks = append(w.gameMasks, mask)
			w.event(fmt.Sprintf("game-flags:%#x=%d", mask, w.gameResult))
			return w.gameResult
		},
		loadExtent: func(object *obeliskXferTestObject4F6F60) uint32 {
			value := object.extent
			w.extentLoads++
			w.event(fmt.Sprintf("load-extent:%#x", value))
			return value
		},
		staticDrawable: func(code uint32) *obeliskXferTestDrawable4F6F60 {
			w.staticCodes = append(w.staticCodes, code)
			w.event(fmt.Sprintf("static-drawable:%#x", code))
			return w.static
		},
		firstMinimap: func() *obeliskXferTestDrawable4F6F60 {
			w.firstCalls++
			w.event("first-minimap")
			return w.first
		},
		nextMinimap: func(drawable *obeliskXferTestDrawable4F6F60) *obeliskXferTestDrawable4F6F60 {
			w.nextCalls++
			w.event("next-minimap:" + drawable.name)
			return drawable.next
		},
		rwMinimapPresent: func(value uint8) uint8 {
			w.minimapInputs = append(w.minimapInputs, value)
			w.event(fmt.Sprintf("rw-minimap:%d", value))
			return w.minimapRead
		},
		transferInventory: func(version uint16, object *obeliskXferTestObject4F6F60, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, obeliskXferTestInventory4F6F60{
				version: version,
				object:  object,
				count:   count,
			})
			w.event(fmt.Sprintf("transfer-inventory:%d:%d", version, count))
			return w.inventoryResult
		},
	}
}

func obeliskXferAssertPanics4F6F60(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("call did not panic")
		}
	}()
	call()
}

func TestObeliskXfer4F6F60CachesEntryDataAndPreservesOrder(t *testing.T) {
	entry := &obeliskXferTestData4F6F60{mana: 7}
	replacement := &obeliskXferTestData4F6F60{mana: 99}
	target := &obeliskXferTestDrawable4F6F60{name: "target"}
	first := &obeliskXferTestDrawable4F6F60{name: "first", next: target}
	object := &obeliskXferTestObject4F6F60{
		data: entry, field34: 0x11223344, extent: 0x10,
	}
	readMana := int32(0x02000000)
	w := newObeliskXferTestWorld4F6F60()
	w.mapResult = -7
	w.modes = []int32{1, 1}
	w.gameResult = 1
	w.manaRead = &readMana
	w.static = target
	w.first = first
	w.after["map-read-write:61"] = func() { object.data = replacement }
	w.after["rw-mana"] = func() {
		object.extent = 0x20
		object.field34 = 0x80000003
	}
	w.after["transfer-inventory:61:-2147483645"] = func() { object.field34 = 5 }

	if got := obeliskXfer4F6F60(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if entry.mana != readMana || replacement.mana != 99 || object.data != replacement {
		t.Fatalf("cached/replacement mana and pointer = %d/%d/%p", entry.mana, replacement.mana, object.data)
	}
	if object.field34 != 0x11223344 {
		t.Fatalf("Field34 = %#x, want entry value", object.field34)
	}
	if !reflect.DeepEqual(w.manaTransfers, []*obeliskXferTestData4F6F60{entry}) {
		t.Fatalf("mana transfer pointers = %v, want cached entry", w.manaTransfers)
	}
	wantLevelBits := math.Float32bits(float32(-32212254))
	if !reflect.DeepEqual(w.syncCalls, []obeliskXferTestSync4F6F60{{object: object, bits: wantLevelBits}}) {
		t.Fatalf("sync calls = %+v, want object and bits %#x", w.syncCalls, wantLevelBits)
	}
	wantInventory := []obeliskXferTestInventory4F6F60{{
		version: 61, object: object, count: -2147483645,
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %+v, want %+v", w.inventoryCalls, wantInventory)
	}
	wantEvents := []string{
		"load-update-data",
		"load-field34:1",
		"rw-version:61",
		"map-read-write:61",
		"rw-mana",
		"read-mode:1=1",
		fmt.Sprintf("load-mana:%d", readMana),
		fmt.Sprintf("sync-mana:%08x", wantLevelBits),
		"game-flags:0x800=1",
		"load-extent:0x20",
		"static-drawable:0x20",
		"first-minimap",
		"next-minimap:first",
		"rw-minimap:1",
		"load-field34:2",
		"read-mode:2=1",
		"transfer-inventory:61:-2147483645",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, wantEvents)
	}
}

func TestObeliskXfer4F6F60SignedVersionGateAndVersion61Suffix(t *testing.T) {
	tests := []struct {
		name        string
		version     uint16
		mapResult   int32
		wantResult  int32
		wantMap     []int32
		wantMana    int
		wantMinimap []uint8
		wantModes   int
		wantStores  int
	}{
		{name: "current", version: 61, mapResult: 1, wantResult: 1, wantMap: []int32{61}, wantMana: 1, wantMinimap: []uint8{0}, wantModes: 1, wantStores: 1},
		{name: "previous", version: 60, mapResult: 1, wantResult: 1, wantMap: []int32{60}, wantStores: 1},
		{name: "zero", version: 0, mapResult: 1, wantResult: 1, wantMap: []int32{0}, wantStores: 1},
		{name: "positive too new", version: 62, mapResult: 1, wantResult: 0},
		{name: "largest positive", version: 0x7fff, mapResult: 1, wantResult: 0},
		{name: "most negative", version: 0x8000, mapResult: 1, wantResult: 1, wantMap: []int32{-32768}, wantStores: 1},
		{name: "minus one", version: 0xffff, mapResult: 1, wantResult: 1, wantMap: []int32{-1}, wantStores: 1},
		{name: "common failure", version: 61, mapResult: 0, wantResult: 0, wantMap: []int32{61}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := &obeliskXferTestObject4F6F60{data: &obeliskXferTestData4F6F60{}}
			w := newObeliskXferTestWorld4F6F60()
			w.version = test.version
			w.mapResult = test.mapResult

			if got := obeliskXfer4F6F60(object, w.deps()); got != test.wantResult {
				t.Fatalf("result = %d, want %d", got, test.wantResult)
			}
			if !reflect.DeepEqual(w.mapVersions, test.wantMap) ||
				len(w.manaTransfers) != test.wantMana ||
				!reflect.DeepEqual(w.minimapInputs, test.wantMinimap) ||
				w.modeCalls != test.wantModes || w.field34Stores != test.wantStores {
				t.Fatalf("map/mana/minimap/modes/stores = %v/%d/%v/%d/%d",
					w.mapVersions, len(w.manaTransfers), w.minimapInputs, w.modeCalls, w.field34Stores)
			}
		})
	}

	t.Run("unsupported version keeps callback mutation", func(t *testing.T) {
		object := &obeliskXferTestObject4F6F60{data: &obeliskXferTestData4F6F60{}, field34: 7}
		w := newObeliskXferTestWorld4F6F60()
		w.version = 62
		w.after["rw-version:61"] = func() { object.field34 = 9 }

		if got := obeliskXfer4F6F60(object, w.deps()); got != 0 || object.field34 != 9 || w.field34Stores != 0 {
			t.Fatalf("result/Field34/stores = %d/%d/%d, want 0/9/0", got, object.field34, w.field34Stores)
		}
	})

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := &obeliskXferTestObject4F6F60{data: &obeliskXferTestData4F6F60{}, field34: 7}
		w := newObeliskXferTestWorld4F6F60()
		w.mapResult = 0
		w.after["map-read-write:61"] = func() { object.field34 = 11 }

		if got := obeliskXfer4F6F60(object, w.deps()); got != 0 || object.field34 != 11 || w.field34Stores != 0 {
			t.Fatalf("result/Field34/stores = %d/%d/%d, want 0/11/0", got, object.field34, w.field34Stores)
		}
	})
}

func TestObeliskXferManaLevel4F6F60WrapsBeforeSignedDivision(t *testing.T) {
	tests := []struct {
		mana     int32
		quotient int32
	}{
		{mana: 0, quotient: 0},
		{mana: 1, quotient: 1},
		{mana: 50, quotient: 80},
		{mana: -1, quotient: -1},
		{mana: math.MaxInt32, quotient: -1},
		{mana: math.MinInt32, quotient: 0},
		{mana: 0x40000001, quotient: 1},
		{mana: 0x02000000, quotient: -32212254},
	}
	for _, test := range tests {
		got := obeliskXferManaLevel4F6F60(test.mana)
		want := float32(test.quotient)
		if math.Float32bits(got) != math.Float32bits(want) {
			t.Fatalf("mana %#x: level bits = %#x, want %#x", uint32(test.mana), math.Float32bits(got), math.Float32bits(want))
		}
	}
}

func TestObeliskXfer4F6F60QuestMinimapUsesIdentityAndIgnoresReadByte(t *testing.T) {
	target := &obeliskXferTestDrawable4F6F60{name: "target"}
	equalContent := &obeliskXferTestDrawable4F6F60{name: "target"}
	second := &obeliskXferTestDrawable4F6F60{name: "second"}
	first := &obeliskXferTestDrawable4F6F60{name: "first", next: second}
	tests := []struct {
		name       string
		flags      int32
		static     *obeliskXferTestDrawable4F6F60
		first      *obeliskXferTestDrawable4F6F60
		wantByte   uint8
		wantExtent int
		wantFirst  int
		wantNext   int
	}{
		{name: "quest disabled", flags: 0},
		{name: "missing static drawable", flags: 1, wantExtent: 1},
		{name: "empty list", flags: 1, static: target, wantExtent: 1, wantFirst: 1},
		{name: "head identity", flags: 1, static: target, first: target, wantByte: 1, wantExtent: 1, wantFirst: 1},
		{name: "later identity", flags: 1, static: second, first: first, wantByte: 1, wantExtent: 1, wantFirst: 1, wantNext: 1},
		{name: "equal content is not identity", flags: 1, static: target, first: equalContent, wantExtent: 1, wantFirst: 1, wantNext: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := &obeliskXferTestObject4F6F60{
				data: &obeliskXferTestData4F6F60{}, extent: 0xaabbccdd,
			}
			w := newObeliskXferTestWorld4F6F60()
			w.gameResult = test.flags
			w.static = test.static
			w.first = test.first
			w.minimapRead = 1

			if got := obeliskXfer4F6F60(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.minimapInputs, []uint8{test.wantByte}) ||
				w.extentLoads != test.wantExtent || w.firstCalls != test.wantFirst || w.nextCalls != test.wantNext {
				t.Fatalf("byte/extent/first/next = %v/%d/%d/%d, want %d/%d/%d/%d",
					w.minimapInputs, w.extentLoads, w.firstCalls, w.nextCalls,
					test.wantByte, test.wantExtent, test.wantFirst, test.wantNext)
			}
		})
	}
}

func TestObeliskXfer4F6F60ReloadsExactOneModeIndependently(t *testing.T) {
	tests := []struct {
		name          string
		modes         []int32
		wantSync      int
		wantInventory int
	}{
		{name: "read then write", modes: []int32{1, 0}, wantSync: 1},
		{name: "write then read", modes: []int32{0, 1}, wantInventory: 1},
		{name: "minus one is neither", modes: []int32{-1, -1}},
		{name: "two is neither", modes: []int32{2, 2}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := &obeliskXferTestObject4F6F60{
				data: &obeliskXferTestData4F6F60{mana: 50}, field34: 3,
			}
			w := newObeliskXferTestWorld4F6F60()
			w.modes = test.modes

			if got := obeliskXfer4F6F60(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if len(w.syncCalls) != test.wantSync || len(w.inventoryCalls) != test.wantInventory || w.modeCalls != 2 {
				t.Fatalf("sync/inventory/modes = %d/%d/%d, want %d/%d/2",
					len(w.syncCalls), len(w.inventoryCalls), w.modeCalls, test.wantSync, test.wantInventory)
			}
		})
	}
}

func TestObeliskXfer4F6F60InventoryGateAndFailurePrefix(t *testing.T) {
	t.Run("legacy negative version passes unsigned word and signed count bits", func(t *testing.T) {
		object := &obeliskXferTestObject4F6F60{
			data: &obeliskXferTestData4F6F60{}, field34: 0x80000001,
		}
		w := newObeliskXferTestWorld4F6F60()
		w.version = 0xffff
		w.modes = []int32{1}

		if got := obeliskXfer4F6F60(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		want := []obeliskXferTestInventory4F6F60{{version: 0xffff, object: object, count: -2147483647}}
		if !reflect.DeepEqual(w.inventoryCalls, want) || len(w.manaTransfers) != 0 || w.modeCalls != 1 {
			t.Fatalf("inventory/mana/modes = %+v/%d/%d", w.inventoryCalls, len(w.manaTransfers), w.modeCalls)
		}
	})

	t.Run("zero live count skips second mode", func(t *testing.T) {
		object := &obeliskXferTestObject4F6F60{data: &obeliskXferTestData4F6F60{}, field34: 9}
		w := newObeliskXferTestWorld4F6F60()
		w.modes = []int32{0}
		w.after["rw-minimap:0"] = func() { object.field34 = 0 }

		if got := obeliskXfer4F6F60(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 9 {
			t.Fatalf("modes/inventory/Field34 = %d/%d/%d", w.modeCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("inventory failure keeps all prefix side effects", func(t *testing.T) {
		data := &obeliskXferTestData4F6F60{mana: 1}
		object := &obeliskXferTestObject4F6F60{data: data, field34: 7}
		readMana := int32(123)
		w := newObeliskXferTestWorld4F6F60()
		w.modes = []int32{1, 1}
		w.manaRead = &readMana
		w.minimapRead = 1
		w.inventoryResult = 0
		w.after["rw-minimap:0"] = func() { object.field34 = 11 }
		w.after["transfer-inventory:61:11"] = func() { object.field34 = 13 }

		if got := obeliskXfer4F6F60(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if data.mana != readMana || len(w.syncCalls) != 1 || !reflect.DeepEqual(w.minimapInputs, []uint8{0}) ||
			object.field34 != 13 || w.field34Stores != 0 {
			t.Fatalf("mana/sync/minimap/Field34/stores = %d/%d/%v/%d/%d",
				data.mana, len(w.syncCalls), w.minimapInputs, object.field34, w.field34Stores)
		}
	})

	t.Run("nonzero inventory result is canonicalized", func(t *testing.T) {
		object := &obeliskXferTestObject4F6F60{data: &obeliskXferTestData4F6F60{}, field34: 5}
		w := newObeliskXferTestWorld4F6F60()
		w.modes = []int32{0, 1}
		w.inventoryResult = -7
		if got := obeliskXfer4F6F60(object, w.deps()); got != 1 || object.field34 != 5 || w.field34Stores != 1 {
			t.Fatalf("result/Field34/stores = %d/%d/%d, want 1/5/1", got, object.field34, w.field34Stores)
		}
	})
}

func TestObeliskXfer4F6F60FaultBoundaries(t *testing.T) {
	t.Run("nil object faults before the first event", func(t *testing.T) {
		w := newObeliskXferTestWorld4F6F60()
		obeliskXferAssertPanics4F6F60(t, func() { obeliskXfer4F6F60(nil, w.deps()) })
		if len(w.events) != 0 {
			t.Fatalf("events = %v, want none", w.events)
		}
	})

	t.Run("nil UpdateData reaches mana transfer after common prefix", func(t *testing.T) {
		object := &obeliskXferTestObject4F6F60{field34: 5}
		w := newObeliskXferTestWorld4F6F60()
		w.panicMana = true
		obeliskXferAssertPanics4F6F60(t, func() { obeliskXfer4F6F60(object, w.deps()) })
		want := []string{
			"load-update-data", "load-field34:1", "rw-version:61", "map-read-write:61", "rw-mana",
		}
		if !reflect.DeepEqual(w.events, want) || w.field34Stores != 0 {
			t.Fatalf("events/stores = %v/%d, want %v/0", w.events, w.field34Stores, want)
		}
	})
}
