package legacy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type spellRewardXferTestData4F5F30 struct {
	spell uint8
}

type spellRewardXferTestObject4F5F30 struct {
	data    *spellRewardXferTestData4F5F30
	field34 uint32
}

type spellRewardXferTestInventoryCall4F5F30 struct {
	version uint16
	object  *spellRewardXferTestObject4F5F30
	count   int32
}

type spellRewardXferTestWorld4F5F30 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	streamRead      bool
	wire            []byte
	position        int
	output          []byte
	nameByID        map[uint8]string
	idByName        map[string]uint8
	inventoryResult int32

	field34Loads   int
	useDataLoads   int
	mapVersions    []int32
	modeCalls      int
	loadedSpells   []uint8
	storedSpells   []uint8
	nameLookups    []uint8
	idLookups      []string
	inventoryCalls []spellRewardXferTestInventoryCall4F5F30
	field34Stores  int
	events         []string
	after          map[string]func()
}

func newSpellRewardXferTestWorld4F5F30() *spellRewardXferTestWorld4F5F30 {
	return &spellRewardXferTestWorld4F5F30{
		version:         spellRewardXferCurrentVersion4F5F30,
		mapResult:       1,
		nameByID:        make(map[uint8]string),
		idByName:        make(map[string]uint8),
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *spellRewardXferTestWorld4F5F30) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *spellRewardXferTestWorld4F5F30) deps() spellRewardXferDeps4F5F30[
	*spellRewardXferTestObject4F5F30,
	*spellRewardXferTestData4F5F30,
] {
	return spellRewardXferDeps4F5F30[
		*spellRewardXferTestObject4F5F30,
		*spellRewardXferTestData4F5F30,
	]{
		loadField34: func(object *spellRewardXferTestObject4F5F30) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		loadUseData: func(object *spellRewardXferTestObject4F5F30) *spellRewardXferTestData4F5F30 {
			w.useDataLoads++
			value := object.data
			w.event("load-use-data")
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *spellRewardXferTestObject4F5F30, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
		},
		readMode: func() int32 {
			call := w.modeCalls
			w.modeCalls++
			value := int32(0)
			if call < len(w.modes) {
				value = w.modes[call]
			}
			w.event(fmt.Sprintf("read-mode:%d=%d", call+1, value))
			return value
		},
		rwByte: func(value uint8) uint8 {
			if w.streamRead {
				if w.position >= len(w.wire) {
					panic("SpellRewardXfer test byte-stream underrun")
				}
				value = w.wire[w.position]
				w.position++
				w.event(fmt.Sprintf("rw-byte:read:%d", value))
				return value
			}
			w.output = append(w.output, value)
			w.event(fmt.Sprintf("rw-byte:write:%d", value))
			return value
		},
		rwBytes: func(value []byte) {
			if w.streamRead {
				if len(w.wire)-w.position < len(value) {
					panic("SpellRewardXfer test byte-stream underrun")
				}
				copy(value, w.wire[w.position:w.position+len(value)])
				w.position += len(value)
				w.event(fmt.Sprintf("rw-bytes:read:%d", len(value)))
				return
			}
			w.output = append(w.output, value...)
			w.event(fmt.Sprintf("rw-bytes:write:%d", len(value)))
		},
		loadSpell: func(data *spellRewardXferTestData4F5F30) uint8 {
			w.event("load-spell")
			value := data.spell
			w.loadedSpells = append(w.loadedSpells, value)
			return value
		},
		storeSpell: func(data *spellRewardXferTestData4F5F30, value uint8) {
			w.event(fmt.Sprintf("store-spell:%d", value))
			data.spell = value
			w.storedSpells = append(w.storedSpells, value)
		},
		spellName: func(id uint8) string {
			w.nameLookups = append(w.nameLookups, id)
			w.event(fmt.Sprintf("spell-name:%d", id))
			return w.nameByID[id]
		},
		spellID: func(name string) uint8 {
			w.idLookups = append(w.idLookups, name)
			w.event("spell-id:" + name)
			return w.idByName[name]
		},
		transferInventory: func(version uint16, object *spellRewardXferTestObject4F5F30, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, spellRewardXferTestInventoryCall4F5F30{
				version: version,
				object:  object,
				count:   count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *spellRewardXferTestObject4F5F30, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func spellRewardNameTestWire4F5F30(name string) []byte {
	return append([]byte{uint8(len(name))}, []byte(name)...)
}

func TestSpellRewardXfer4F5F30CachesEntryUseDataAndPreservesOrder(t *testing.T) {
	entry := &spellRewardXferTestData4F5F30{spell: 3}
	replacement := &spellRewardXferTestData4F5F30{spell: 4}
	object := &spellRewardXferTestObject4F5F30{data: entry, field34: 0x10203040}
	w := newSpellRewardXferTestWorld4F5F30()
	w.streamRead = true
	w.wire = spellRewardNameTestWire4F5F30("SPELL_BLINK")
	w.idByName["SPELL_BLINK"] = 22
	w.modes = []int32{1, 0}
	w.mapResult = -7
	w.after["map-read-write:60"] = func() {
		object.data = replacement
		object.field34 = 0x80000003
	}

	if got := spellRewardXfer4F5F30(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if entry.spell != 22 || replacement.spell != 4 || object.data != replacement {
		t.Fatalf("cached/replacement spell and live pointer = %d/%d/%p, want 22/4/replacement",
			entry.spell, replacement.spell, object.data)
	}
	if object.field34 != 0x10203040 || len(w.inventoryCalls) != 0 {
		t.Fatalf("Field34/inventory = %#x/%d, want restored/none", object.field34, len(w.inventoryCalls))
	}
	wantEvents := []string{
		"load-field34:1",
		"load-use-data",
		"rw-version:60",
		"map-read-write:60",
		"read-mode:1=1",
		"rw-byte:read:11",
		"rw-bytes:read:11",
		"spell-id:SPELL_BLINK",
		"store-spell:22",
		"load-field34:2",
		"read-mode:2=0",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, wantEvents)
	}
}

func TestSpellRewardXfer4F5F30SignedVersionAndInventoryABI(t *testing.T) {
	data := &spellRewardXferTestData4F5F30{spell: 1}
	object := &spellRewardXferTestObject4F5F30{data: data, field34: 0x11223344}
	w := newSpellRewardXferTestWorld4F5F30()
	w.version = 0xffff
	w.streamRead = true
	w.wire = []byte{0xaa, 0x88, 0}
	w.modes = []int32{1, 1}
	w.mapResult = -9
	w.inventoryResult = -3
	w.after["map-read-write:-1"] = func() { object.field34 = 0x80000002 }

	if got := spellRewardXfer4F5F30(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if !reflect.DeepEqual(w.mapVersions, []int32{-1}) {
		t.Fatalf("map versions = %v, want signed -1", w.mapVersions)
	}
	wantInventory := []spellRewardXferTestInventoryCall4F5F30{{
		version: 0xffff,
		object:  object,
		count:   -2147483646,
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want zero-extended version and signed count", w.inventoryCalls)
	}
	if data.spell != 0x88 || object.field34 != 0x11223344 || w.field34Stores != 1 {
		t.Fatalf("spell/Field34/stores = %#x/%#x/%d, want %#x/restored/1",
			data.spell, object.field34, w.field34Stores, uint8(0x88))
	}
}

func TestSpellRewardXfer4F5F30VersionAndEarlyFailurePrefixes(t *testing.T) {
	for _, version := range []uint16{61, 0x7fff} {
		t.Run(fmt.Sprintf("reject-%#x", version), func(t *testing.T) {
			object := &spellRewardXferTestObject4F5F30{field34: 7}
			w := newSpellRewardXferTestWorld4F5F30()
			w.version = version

			if got := spellRewardXfer4F5F30(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if w.field34Loads != 1 || w.useDataLoads != 1 || len(w.mapVersions) != 0 || w.field34Stores != 0 {
				t.Fatalf("entry loads/map/store = %d/%d/%v/%d, want 1/1/none/0",
					w.field34Loads, w.useDataLoads, w.mapVersions, w.field34Stores)
			}
		})
	}

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := &spellRewardXferTestObject4F5F30{field34: 0x11223344}
		w := newSpellRewardXferTestWorld4F5F30()
		w.mapResult = 0
		w.after["map-read-write:60"] = func() { object.field34 = 0x55667788 }

		if got := spellRewardXfer4F5F30(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 0x55667788 || w.modeCalls != 0 || w.field34Stores != 0 {
			t.Fatalf("Field34/mode/store = %#x/%d/%d, want callback state/0/0",
				object.field34, w.modeCalls, w.field34Stores)
		}
	})

	t.Run("modern name length 128 returns before bytes and rollback", func(t *testing.T) {
		object := &spellRewardXferTestObject4F5F30{
			data: &spellRewardXferTestData4F5F30{spell: 17}, field34: 0x11223344,
		}
		w := newSpellRewardXferTestWorld4F5F30()
		w.streamRead = true
		w.wire = []byte{spellRewardXferNameLimit4F5F30}
		w.modes = []int32{1}
		w.after["map-read-write:60"] = func() { object.field34 = 0xaabbccdd }

		if got := spellRewardXfer4F5F30(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.data.spell != 17 || object.field34 != 0xaabbccdd ||
			len(w.idLookups) != 0 || len(w.storedSpells) != 0 || w.field34Loads != 1 || w.field34Stores != 0 {
			t.Fatalf("spell/Field34/lookups/stores/loads/restore = %d/%#x/%v/%v/%d/%d",
				object.data.spell, object.field34, w.idLookups, w.storedSpells,
				w.field34Loads, w.field34Stores)
		}
	})
}

func TestSpellRewardXfer4F5F30HistoricalRawIDs(t *testing.T) {
	tests := []struct {
		name         string
		version      uint16
		wire         []byte
		want         uint8
		wantStores   []uint8
		wantMap      int32
		wantConsumed int
	}{
		{name: "version 10 consumes trailer and clamps second", version: 10, wire: []byte{0xaa, 0x89, 0x88, 0xcc}, want: 0x88, wantStores: []uint8{0, 0x88}, wantMap: 10, wantConsumed: 4},
		{name: "third threshold preserves second", version: 30, wire: []byte{0xaa, 0x44, 0x89}, want: 0x44, wantStores: []uint8{0x44}, wantMap: 30, wantConsumed: 3},
		{name: "negative version takes raw branch", version: 0x8000, wire: []byte{0xaa, 0x22, 0}, want: 0x22, wantStores: []uint8{0x22}, wantMap: -32768, wantConsumed: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &spellRewardXferTestData4F5F30{spell: 0xff}
			object := &spellRewardXferTestObject4F5F30{data: data}
			w := newSpellRewardXferTestWorld4F5F30()
			w.version = tc.version
			w.streamRead = true
			w.wire = tc.wire
			w.modes = []int32{1}

			if got := spellRewardXfer4F5F30(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if data.spell != tc.want || !reflect.DeepEqual(w.storedSpells, tc.wantStores) ||
				!reflect.DeepEqual(w.mapVersions, []int32{tc.wantMap}) || w.position != tc.wantConsumed {
				t.Fatalf("spell/stores/map/consumed = %#x/%v/%v/%d, want %#x/%v/[%d]/%d",
					data.spell, w.storedSpells, w.mapVersions, w.position,
					tc.want, tc.wantStores, tc.wantMap, tc.wantConsumed)
			}
		})
	}
}

func TestSpellRewardXfer4F5F30HistoricalNamesDelayStoreAndApplyPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name       string
		third      string
		thirdID    uint8
		want       uint8
		wantStores []uint8
	}{
		{name: "third overrides second", third: "SPELL_THIRD", thirdID: 33, want: 33, wantStores: []uint8{22, 33}},
		{name: "unknown third preserves second", third: "SPELL_UNKNOWN", thirdID: 0, want: 22, wantStores: []uint8{22}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &spellRewardXferTestData4F5F30{spell: 99}
			object := &spellRewardXferTestObject4F5F30{data: data}
			w := newSpellRewardXferTestWorld4F5F30()
			w.version = 40
			w.streamRead = true
			w.wire = append(w.wire, spellRewardNameTestWire4F5F30("SPELL_IGNORED")...)
			w.wire = append(w.wire, spellRewardNameTestWire4F5F30("SPELL_SECOND")...)
			w.wire = append(w.wire, spellRewardNameTestWire4F5F30(tc.third)...)
			w.idByName["SPELL_IGNORED"] = 11
			w.idByName["SPELL_SECOND"] = 22
			w.idByName[tc.third] = tc.thirdID
			w.modes = []int32{1}

			if got := spellRewardXfer4F5F30(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if data.spell != tc.want || !reflect.DeepEqual(w.storedSpells, tc.wantStores) ||
				!reflect.DeepEqual(w.idLookups, []string{"SPELL_IGNORED", "SPELL_SECOND", tc.third}) {
				t.Fatalf("spell/stores/lookups = %d/%v/%v, want %d/%v/all names",
					data.spell, w.storedSpells, w.idLookups, tc.want, tc.wantStores)
			}
		})
	}

	t.Run("third length failure leaves entry spell and live count", func(t *testing.T) {
		data := &spellRewardXferTestData4F5F30{spell: 99}
		object := &spellRewardXferTestObject4F5F30{data: data, field34: 7}
		w := newSpellRewardXferTestWorld4F5F30()
		w.version = 31
		w.streamRead = true
		w.wire = append(w.wire, spellRewardNameTestWire4F5F30("SPELL_IGNORED")...)
		w.wire = append(w.wire, spellRewardNameTestWire4F5F30("SPELL_SECOND")...)
		w.wire = append(w.wire, spellRewardXferNameLimit4F5F30)
		w.idByName["SPELL_IGNORED"] = 11
		w.idByName["SPELL_SECOND"] = 22
		w.modes = []int32{1}
		w.after["map-read-write:31"] = func() { object.field34 = 9 }

		if got := spellRewardXfer4F5F30(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if data.spell != 99 || len(w.storedSpells) != 0 ||
			!reflect.DeepEqual(w.idLookups, []string{"SPELL_IGNORED", "SPELL_SECOND"}) ||
			object.field34 != 9 || w.field34Stores != 0 {
			t.Fatalf("spell/stores/lookups/Field34/restore = %d/%v/%v/%d/%d",
				data.spell, w.storedSpells, w.idLookups, object.field34, w.field34Stores)
		}
	})
}

func TestSpellRewardXfer4F5F30AllNonOneModesWriteOneName(t *testing.T) {
	for index, tc := range []struct {
		version uint16
		mode    int32
	}{
		{version: 0, mode: 0},
		{version: 30, mode: 2},
		{version: 31, mode: -1},
		{version: 41, mode: 0},
		{version: 60, mode: 3},
		{version: 0x8000, mode: -7},
	} {
		t.Run(fmt.Sprintf("case-%d", index), func(t *testing.T) {
			data := &spellRewardXferTestData4F5F30{spell: 7}
			object := &spellRewardXferTestObject4F5F30{data: data}
			w := newSpellRewardXferTestWorld4F5F30()
			w.version = tc.version
			w.modes = []int32{tc.mode}
			w.nameByID[7] = "SPELL_BLINK"

			if got := spellRewardXfer4F5F30(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			wantWire := spellRewardNameTestWire4F5F30("SPELL_BLINK")
			if !reflect.DeepEqual(w.output, wantWire) || !reflect.DeepEqual(w.loadedSpells, []uint8{7}) ||
				!reflect.DeepEqual(w.nameLookups, []uint8{7}) || w.modeCalls != 1 ||
				!reflect.DeepEqual(w.mapVersions, []int32{int32(int16(tc.version))}) {
				t.Fatalf("wire/load/name/mode/map = %x/%v/%v/%d/%v, want %x/[7]/[7]/1/signed",
					w.output, w.loadedSpells, w.nameLookups, w.modeCalls, w.mapVersions, wantWire)
			}
		})
	}
}

func TestSpellRewardXfer4F5F30FinalGateAndFailureDoNotInventRollback(t *testing.T) {
	t.Run("zero live count skips second mode read and restores", func(t *testing.T) {
		object := &spellRewardXferTestObject4F5F30{
			data: &spellRewardXferTestData4F5F30{}, field34: 0x10203040,
		}
		w := newSpellRewardXferTestWorld4F5F30()
		w.streamRead = true
		w.wire = spellRewardNameTestWire4F5F30("SPELL_ZERO")
		w.idByName["SPELL_ZERO"] = 0
		w.modes = []int32{1, 1}
		w.after["map-read-write:60"] = func() { object.field34 = 0 }

		if got := spellRewardXfer4F5F30(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 0x10203040 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%#x, want 1/0/restored",
				w.modeCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("inventory failure keeps inventory mutation", func(t *testing.T) {
		object := &spellRewardXferTestObject4F5F30{
			data: &spellRewardXferTestData4F5F30{}, field34: 0x11223344,
		}
		w := newSpellRewardXferTestWorld4F5F30()
		w.streamRead = true
		w.wire = spellRewardNameTestWire4F5F30("SPELL_BLINK")
		w.idByName["SPELL_BLINK"] = 22
		w.modes = []int32{1, 1}
		w.inventoryResult = 0
		w.after["map-read-write:60"] = func() { object.field34 = 9 }
		w.after["transfer-inventory"] = func() { object.field34 = 0xaabbccdd }

		if got := spellRewardXfer4F5F30(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 0xaabbccdd || w.field34Stores != 0 ||
			len(w.inventoryCalls) != 1 || w.inventoryCalls[0].count != 9 {
			t.Fatalf("Field34/stores/inventory = %#x/%d/%#v, want mutation/0/live count 9",
				object.field34, w.field34Stores, w.inventoryCalls)
		}
	})
}

func TestSpellRewardXfer4F5F30UseDataAndWriteOverflowFaultBoundaries(t *testing.T) {
	t.Run("nil cached UseData faults after common transfer and first mode", func(t *testing.T) {
		object := &spellRewardXferTestObject4F5F30{field34: 7}
		w := newSpellRewardXferTestWorld4F5F30()
		w.modes = []int32{0}

		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("missing cached UseData fault")
				}
			}()
			_ = spellRewardXfer4F5F30(object, w.deps())
		}()
		wantPrefix := []string{
			"load-field34:1",
			"load-use-data",
			"rw-version:60",
			"map-read-write:60",
			"read-mode:1=0",
			"load-spell",
		}
		if !reflect.DeepEqual(w.events, wantPrefix) || w.field34Stores != 0 {
			t.Fatalf("fault prefix/restore = %v/%d, want %v/0", w.events, w.field34Stores, wantPrefix)
		}
	})

	t.Run("oversized runtime name faults without clean rollback", func(t *testing.T) {
		object := &spellRewardXferTestObject4F5F30{
			data: &spellRewardXferTestData4F5F30{spell: 7}, field34: 0x11223344,
		}
		w := newSpellRewardXferTestWorld4F5F30()
		w.modes = []int32{2}
		w.nameByID[7] = strings.Repeat("x", spellRewardXferNameLimit4F5F30)
		w.after["map-read-write:60"] = func() { object.field34 = 0x55667788 }

		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("missing PE32 stack-buffer boundary fault")
				}
			}()
			_ = spellRewardXfer4F5F30(object, w.deps())
		}()
		if object.field34 != 0x55667788 || w.field34Loads != 1 || w.field34Stores != 0 || len(w.output) != 0 {
			t.Fatalf("Field34/loads/stores/wire = %#x/%d/%d/%x, want live/1/0/empty",
				object.field34, w.field34Loads, w.field34Stores, w.output)
		}
	})
}
