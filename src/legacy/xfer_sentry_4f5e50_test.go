package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type sentryXferTestData4F5E50 struct {
	words [3]uint32
}

type sentryXferTestObject4F5E50 struct {
	data    *sentryXferTestData4F5E50
	field34 uint32
}

type sentryXferTestInventoryCall4F5E50 struct {
	version uint16
	object  *sentryXferTestObject4F5E50
	count   int32
}

type sentryXferTestWorld4F5E50 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	flags           []int32
	inventoryResult int32

	dataLoads      int
	field34Loads   int
	mapVersions    []int32
	updateRWs      []int
	modeCalls      int
	flagMasks      []uint32
	updateLoads    []int
	updateStores   []int
	inventoryCalls []sentryXferTestInventoryCall4F5E50
	field34Stores  int
	events         []string
	after          map[string]func()
}

func newSentryXferTestWorld4F5E50() *sentryXferTestWorld4F5E50 {
	return &sentryXferTestWorld4F5E50{
		version:         sentryXferCurrentVersion4F5E50,
		mapResult:       1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *sentryXferTestWorld4F5E50) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *sentryXferTestWorld4F5E50) deps() sentryXferDeps4F5E50[
	*sentryXferTestObject4F5E50,
	*sentryXferTestData4F5E50,
] {
	return sentryXferDeps4F5E50[
		*sentryXferTestObject4F5E50,
		*sentryXferTestData4F5E50,
	]{
		loadUpdateData: func(object *sentryXferTestObject4F5E50) *sentryXferTestData4F5E50 {
			w.dataLoads++
			value := object.data
			w.event("load-update-data")
			return value
		},
		loadField34: func(object *sentryXferTestObject4F5E50) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *sentryXferTestObject4F5E50, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
		},
		rwUpdateData: func(_ *sentryXferTestData4F5E50, offset int) {
			w.updateRWs = append(w.updateRWs, offset)
			w.event(fmt.Sprintf("rw-update:%d", offset))
		},
		readMode: func() int32 {
			call := w.modeCalls
			w.modeCalls++
			value := int32(0)
			if call < len(w.modes) {
				value = w.modes[call]
			}
			w.event(fmt.Sprintf("read-mode:%d", call+1))
			return value
		},
		gameFlags: func(mask uint32) int32 {
			call := len(w.flagMasks)
			w.flagMasks = append(w.flagMasks, mask)
			value := int32(0)
			if call < len(w.flags) {
				value = w.flags[call]
			}
			w.event(fmt.Sprintf("game-flags:%#x", mask))
			return value
		},
		loadUpdateU32: func(data *sentryXferTestData4F5E50, offset int) uint32 {
			w.updateLoads = append(w.updateLoads, offset)
			value := data.words[offset/4]
			w.event(fmt.Sprintf("load-update:%d", offset))
			return value
		},
		storeUpdateU32: func(data *sentryXferTestData4F5E50, offset int, value uint32) {
			w.updateStores = append(w.updateStores, offset)
			data.words[offset/4] = value
			w.event(fmt.Sprintf("store-update:%d", offset))
		},
		transferInventory: func(version uint16, object *sentryXferTestObject4F5E50, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, sentryXferTestInventoryCall4F5E50{
				version: version,
				object:  object,
				count:   count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *sentryXferTestObject4F5E50, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestSentryXfer4F5E50CachesUpdateDataAndPreservesInstructionOrder(t *testing.T) {
	entry := &sentryXferTestData4F5E50{words: [3]uint32{0x11111111, 0x22222222, 0x33333333}}
	replacement := &sentryXferTestData4F5E50{words: [3]uint32{0xaaaaaaaa, 0xbbbbbbbb, 0xcccccccc}}
	object := &sentryXferTestObject4F5E50{data: entry, field34: 0x10203040}
	w := newSentryXferTestWorld4F5E50()
	w.mapResult = -7
	w.modes = []int32{0, 0}
	w.flags = []int32{1}
	w.after["map-read-write:61"] = func() {
		object.data = replacement
		object.field34 = 0x80000003
	}
	w.after["rw-update:8"] = func() {
		entry.words[1] = 0x55667788
	}

	if got := sentryXfer4F5E50(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.data != replacement || object.field34 != 0x10203040 {
		t.Fatalf("object data/Field34 = %p/%#x, want replacement/%#x",
			object.data, object.field34, uint32(0x10203040))
	}
	if entry.words[0] != 0x55667788 || replacement.words[0] != 0xaaaaaaaa {
		t.Fatalf("entry/replacement word zero = %#x/%#x, want cached copy/unchanged",
			entry.words[0], replacement.words[0])
	}
	if !reflect.DeepEqual(w.updateRWs, []int{4, 8, 0}) ||
		!reflect.DeepEqual(w.updateLoads, []int{4}) ||
		!reflect.DeepEqual(w.updateStores, []int{0}) {
		t.Fatalf("update rw/load/store = %v/%v/%v, want [4 8 0]/[4]/[0]",
			w.updateRWs, w.updateLoads, w.updateStores)
	}
	wantEvents := []string{
		"load-update-data",
		"load-field34:1",
		"rw-version:61",
		"map-read-write:61",
		"rw-update:4",
		"rw-update:8",
		"read-mode:1",
		"game-flags:0x200000",
		"load-update:4",
		"store-update:0",
		"rw-update:0",
		"load-field34:2",
		"read-mode:2",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, wantEvents)
	}
}

func TestSentryXfer4F5E50SignedVersionAndInventoryABI(t *testing.T) {
	data := &sentryXferTestData4F5E50{words: [3]uint32{1, 0x76543210, 3}}
	object := &sentryXferTestObject4F5E50{data: data, field34: 0x80000002}
	w := newSentryXferTestWorld4F5E50()
	w.version = 0xffff
	w.mapResult = -1
	w.modes = []int32{1, 1}
	w.inventoryResult = -9

	if got := sentryXfer4F5E50(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if !reflect.DeepEqual(w.mapVersions, []int32{-1}) {
		t.Fatalf("map versions = %v, want signed -1", w.mapVersions)
	}
	if !reflect.DeepEqual(w.updateRWs, []int{4, 8}) || data.words[0] != data.words[1] {
		t.Fatalf("update rw/copy = %v/%#x, want [4 8]/%#x", w.updateRWs, data.words[0], data.words[1])
	}
	if len(w.flagMasks) != 0 {
		t.Fatalf("read mode called game flags: %v", w.flagMasks)
	}
	wantInventory := []sentryXferTestInventoryCall4F5E50{{
		version: 0xffff,
		object:  object,
		count:   -2147483646,
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want zero-extended version and signed count", w.inventoryCalls)
	}
	if object.field34 != 0x80000002 || w.field34Stores != 1 {
		t.Fatalf("Field34/stores = %#x/%d, want restored entry value/1", object.field34, w.field34Stores)
	}
}

func TestSentryXfer4F5E50VersionThresholds(t *testing.T) {
	for _, version := range []uint16{62, 0x7fff} {
		t.Run(fmt.Sprintf("reject-%#x", version), func(t *testing.T) {
			object := &sentryXferTestObject4F5E50{data: &sentryXferTestData4F5E50{}, field34: 7}
			w := newSentryXferTestWorld4F5E50()
			w.version = version

			if got := sentryXfer4F5E50(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if w.dataLoads != 1 || w.field34Loads != 1 || len(w.mapVersions) != 0 || w.field34Stores != 0 {
				t.Fatalf("entry loads/map/store = %d/%d/%v/%d, want 1/1/none/0",
					w.dataLoads, w.field34Loads, w.mapVersions, w.field34Stores)
			}
		})
	}

	for _, tc := range []struct {
		version uint16
		wantRWs []int
	}{
		{version: 60, wantRWs: []int{4, 8}},
		{version: 61, wantRWs: []int{4, 8, 0}},
		{version: 0x8000, wantRWs: []int{4, 8}},
	} {
		t.Run(fmt.Sprintf("accept-%#x", tc.version), func(t *testing.T) {
			object := &sentryXferTestObject4F5E50{data: &sentryXferTestData4F5E50{}}
			w := newSentryXferTestWorld4F5E50()
			w.version = tc.version
			w.modes = []int32{2}
			w.flags = []int32{0}

			if got := sentryXfer4F5E50(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.mapVersions, []int32{int32(int16(tc.version))}) ||
				!reflect.DeepEqual(w.updateRWs, tc.wantRWs) {
				t.Fatalf("map versions/update RWs = %v/%v, want [%d]/%v",
					w.mapVersions, w.updateRWs, int16(tc.version), tc.wantRWs)
			}
		})
	}
}

func TestSentryXfer4F5E50ExactOneModeAndGameFlagGates(t *testing.T) {
	tests := []struct {
		name      string
		mode      int32
		flag      int32
		wantCopy  bool
		wantFlags int
	}{
		{name: "read mode", mode: 1, flag: 0, wantCopy: true, wantFlags: 0},
		{name: "write flagged", mode: 0, flag: 1, wantCopy: true, wantFlags: 1},
		{name: "other mode flagged", mode: -7, flag: 1, wantCopy: true, wantFlags: 1},
		{name: "flag two is false", mode: 2, flag: 2, wantCopy: false, wantFlags: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &sentryXferTestData4F5E50{words: [3]uint32{0x11, 0x22, 0x33}}
			object := &sentryXferTestObject4F5E50{data: data}
			w := newSentryXferTestWorld4F5E50()
			w.modes = []int32{tc.mode}
			w.flags = []int32{tc.flag}

			if got := sentryXfer4F5E50(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			wantWord := uint32(0x11)
			if tc.wantCopy {
				wantWord = 0x22
			}
			if data.words[0] != wantWord || len(w.flagMasks) != tc.wantFlags {
				t.Fatalf("word zero/flag calls = %#x/%d, want %#x/%d",
					data.words[0], len(w.flagMasks), wantWord, tc.wantFlags)
			}
		})
	}
}

func TestSentryXfer4F5E50FailurePrefixesAndField34Restore(t *testing.T) {
	t.Run("map failure leaves callback state and never dereferences cached update", func(t *testing.T) {
		object := &sentryXferTestObject4F5E50{field34: 0x11223344}
		w := newSentryXferTestWorld4F5E50()
		w.mapResult = 0
		w.after["map-read-write:61"] = func() { object.field34 = 0x55667788 }
		deps := w.deps()
		deps.rwUpdateData = func(*sentryXferTestData4F5E50, int) {
			t.Fatal("map failure dereferenced cached nil UpdateData")
		}

		if got := sentryXfer4F5E50(object, deps); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 0x55667788 || w.field34Stores != 0 {
			t.Fatalf("Field34/stores = %#x/%d, want callback state/no restore", object.field34, w.field34Stores)
		}
	})

	t.Run("zero live count skips second mode read and restores", func(t *testing.T) {
		object := &sentryXferTestObject4F5E50{
			data: &sentryXferTestData4F5E50{}, field34: 0x10203040,
		}
		w := newSentryXferTestWorld4F5E50()
		w.modes = []int32{0, 1}
		w.flags = []int32{0}
		w.after["map-read-write:61"] = func() { object.field34 = 0 }

		if got := sentryXfer4F5E50(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 0x10203040 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%#x, want 1/0/restored",
				w.modeCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("inventory failure leaves its live Field34 mutation", func(t *testing.T) {
		object := &sentryXferTestObject4F5E50{
			data: &sentryXferTestData4F5E50{}, field34: 0x11223344,
		}
		w := newSentryXferTestWorld4F5E50()
		w.modes = []int32{1, 1}
		w.inventoryResult = 0
		w.after["map-read-write:61"] = func() { object.field34 = 9 }
		w.after["transfer-inventory"] = func() { object.field34 = 0xaabbccdd }

		if got := sentryXfer4F5E50(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 0xaabbccdd || w.field34Stores != 0 {
			t.Fatalf("Field34/stores = %#x/%d, want inventory mutation/no restore",
				object.field34, w.field34Stores)
		}
		if len(w.inventoryCalls) != 1 || w.inventoryCalls[0].count != 9 {
			t.Fatalf("inventory calls = %#v, want live count 9", w.inventoryCalls)
		}
	})
}

func TestSentryXfer4F5E50UpdateFaultBeginsAfterCommonTransfer(t *testing.T) {
	object := &sentryXferTestObject4F5E50{field34: 7}
	w := newSentryXferTestWorld4F5E50()
	deps := w.deps()
	deps.rwUpdateData = func(_ *sentryXferTestData4F5E50, offset int) {
		w.event(fmt.Sprintf("rw-update:%d", offset))
		panic("original nil UpdateData dereference boundary")
	}

	deferred := false
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("missing update-data fault")
			}
			deferred = true
		}()
		_ = sentryXfer4F5E50(object, deps)
	}()
	if !deferred {
		t.Fatal("fault recovery did not run")
	}
	want := []string{
		"load-update-data",
		"load-field34:1",
		"rw-version:61",
		"map-read-write:61",
		"rw-update:4",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("fault prefix = %v, want %v", w.events, want)
	}
}
