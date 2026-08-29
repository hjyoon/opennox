package legacy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type abilityRewardXferTestData4F6240 struct {
	ability uint8
}

type abilityRewardXferTestObject4F6240 struct {
	data    *abilityRewardXferTestData4F6240
	field34 uint32
}

type abilityRewardXferTestInventoryCall4F6240 struct {
	version uint16
	object  *abilityRewardXferTestObject4F6240
	count   int32
}

type abilityRewardXferTestWorld4F6240 struct {
	version         uint16
	mapResult       int32
	byteResult      *uint8
	payload         []byte
	nameByID        map[uint8]string
	idByName        map[string]int32
	modes           []int32
	inventoryResult int32

	field34Loads    int
	useDataLoads    int
	mapVersions     []int32
	byteInputs      []uint8
	bytePayloads    [][]byte
	loadedAbilities []uint8
	nameLookups     []uint8
	idLookups       []string
	storedAbilities []uint8
	modeCalls       int
	inventoryCalls  []abilityRewardXferTestInventoryCall4F6240
	field34Stores   int
	events          []string
	after           map[string]func()
}

func newAbilityRewardXferTestWorld4F6240() *abilityRewardXferTestWorld4F6240 {
	return &abilityRewardXferTestWorld4F6240{
		version:         abilityRewardXferCurrentVersion4F6240,
		mapResult:       1,
		nameByID:        make(map[uint8]string),
		idByName:        make(map[string]int32),
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func abilityRewardXferTestByte4F6240(value uint8) *uint8 {
	return &value
}

func (w *abilityRewardXferTestWorld4F6240) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *abilityRewardXferTestWorld4F6240) deps() abilityRewardXferDeps4F6240[
	*abilityRewardXferTestObject4F6240,
	*abilityRewardXferTestData4F6240,
] {
	return abilityRewardXferDeps4F6240[
		*abilityRewardXferTestObject4F6240,
		*abilityRewardXferTestData4F6240,
	]{
		loadField34: func(object *abilityRewardXferTestObject4F6240) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		loadUseData: func(object *abilityRewardXferTestObject4F6240) *abilityRewardXferTestData4F6240 {
			w.useDataLoads++
			value := object.data
			w.event("load-use-data")
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *abilityRewardXferTestObject4F6240, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
		},
		loadAbility: func(data *abilityRewardXferTestData4F6240) uint8 {
			value := data.ability
			w.loadedAbilities = append(w.loadedAbilities, value)
			w.event(fmt.Sprintf("load-ability:%d", value))
			return value
		},
		abilityName: func(id uint8) string {
			w.nameLookups = append(w.nameLookups, id)
			w.event(fmt.Sprintf("ability-name:%d", id))
			return w.nameByID[id]
		},
		rwByte: func(value uint8) uint8 {
			w.byteInputs = append(w.byteInputs, value)
			w.event(fmt.Sprintf("rw-byte:%d", value))
			if w.byteResult != nil {
				return *w.byteResult
			}
			return value
		},
		rwBytes: func(value []byte) {
			if len(w.payload) < len(value) {
				panic("AbilityRewardXfer test byte-stream underrun")
			}
			if w.payload != nil {
				copy(value, w.payload[:len(value)])
			}
			w.bytePayloads = append(w.bytePayloads, append([]byte(nil), value...))
			w.event(fmt.Sprintf("rw-bytes:%d", len(value)))
		},
		abilityID: func(name string) int32 {
			w.idLookups = append(w.idLookups, name)
			w.event("ability-id:" + name)
			return w.idByName[name]
		},
		storeAbility: func(data *abilityRewardXferTestData4F6240, value uint8) {
			data.ability = value
			w.storedAbilities = append(w.storedAbilities, value)
			w.event(fmt.Sprintf("store-ability:%d", value))
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
		transferInventory: func(version uint16, object *abilityRewardXferTestObject4F6240, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, abilityRewardXferTestInventoryCall4F6240{
				version: version,
				object:  object,
				count:   count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *abilityRewardXferTestObject4F6240, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestAbilityRewardXfer4F6240CachesEntryUseDataAndOrdersNameRoundTrip(t *testing.T) {
	entry := &abilityRewardXferTestData4F6240{ability: 1}
	replacement := &abilityRewardXferTestData4F6240{ability: 4}
	object := &abilityRewardXferTestObject4F6240{data: entry, field34: 0x10203040}
	w := newAbilityRewardXferTestWorld4F6240()
	w.nameByID[2] = "ABILITY_WARCRY"
	w.byteResult = abilityRewardXferTestByte4F6240(uint8(len("ABILITY_HARPOON")))
	w.payload = []byte("ABILITY_HARPOON")
	w.idByName["ABILITY_HARPOON"] = 3
	w.modes = []int32{0}
	w.mapResult = -7
	w.after["map-read-write:61"] = func() {
		entry.ability = 2
		object.data = replacement
		object.field34 = 0x80000003
	}

	if got := abilityRewardXfer4F6240(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if entry.ability != 3 || replacement.ability != 4 || object.data != replacement {
		t.Fatalf("cached/replacement ability and live pointer = %d/%d/%p, want 3/4/replacement",
			entry.ability, replacement.ability, object.data)
	}
	if object.field34 != 0x10203040 || len(w.inventoryCalls) != 0 {
		t.Fatalf("Field34/inventory = %#x/%d, want restored/none", object.field34, len(w.inventoryCalls))
	}
	wantEvents := []string{
		"load-field34:1",
		"load-use-data",
		"rw-version:61",
		"map-read-write:61",
		"load-ability:2",
		"ability-name:2",
		"rw-byte:14",
		"rw-bytes:15",
		"ability-id:ABILITY_HARPOON",
		"store-ability:3",
		"load-field34:2",
		"read-mode:1=0",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, wantEvents)
	}
}

func TestAbilityRewardXfer4F6240SignedVersionInventoryAndLowByteABI(t *testing.T) {
	data := &abilityRewardXferTestData4F6240{ability: 1}
	object := &abilityRewardXferTestObject4F6240{data: data, field34: 0x11223344}
	w := newAbilityRewardXferTestWorld4F6240()
	w.version = 0xffff
	w.nameByID[1] = "A"
	w.byteResult = abilityRewardXferTestByte4F6240(1)
	w.payload = []byte("X")
	w.idByName["X"] = 0x123
	w.modes = []int32{1}
	w.mapResult = -9
	w.inventoryResult = -3
	w.after["map-read-write:-1"] = func() { object.field34 = 0x80000002 }

	if got := abilityRewardXfer4F6240(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if !reflect.DeepEqual(w.mapVersions, []int32{-1}) {
		t.Fatalf("map versions = %v, want signed -1", w.mapVersions)
	}
	wantInventory := []abilityRewardXferTestInventoryCall4F6240{{
		version: 0xffff,
		object:  object,
		count:   -2147483646,
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want zero-extended version and signed count", w.inventoryCalls)
	}
	if data.ability != 0x23 || object.field34 != 0x11223344 || w.field34Stores != 1 {
		t.Fatalf("ability/Field34/stores = %#x/%#x/%d, want 0x23/restored/1",
			data.ability, object.field34, w.field34Stores)
	}
}

func TestAbilityRewardXfer4F6240EmptyAndEmbeddedNULNames(t *testing.T) {
	for _, tc := range []struct {
		name       string
		payload    []byte
		parsedName string
	}{
		{name: "empty still transfers and parses", payload: []byte{}, parsedName: ""},
		{name: "embedded NUL uses C string prefix", payload: []byte{'A', 0, 'B'}, parsedName: "A"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &abilityRewardXferTestData4F6240{ability: 5}
			object := &abilityRewardXferTestObject4F6240{data: data}
			w := newAbilityRewardXferTestWorld4F6240()
			w.nameByID[5] = "ABILITY_EYE_OF_THE_WOLF"
			w.byteResult = abilityRewardXferTestByte4F6240(uint8(len(tc.payload)))
			w.payload = tc.payload
			w.idByName[tc.parsedName] = 4

			if got := abilityRewardXfer4F6240(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if data.ability != 4 || !reflect.DeepEqual(w.idLookups, []string{tc.parsedName}) {
				t.Fatalf("ability/lookups = %d/%q, want 4/%q", data.ability, w.idLookups, tc.parsedName)
			}
			if len(w.bytePayloads) != 1 || len(w.bytePayloads[0]) != len(tc.payload) {
				t.Fatalf("payload callbacks = %#v, want one callback of length %d", w.bytePayloads, len(tc.payload))
			}
			if w.modeCalls != 0 {
				t.Fatalf("mode calls = %d, want zero for live Field34 zero", w.modeCalls)
			}
		})
	}
}

func TestAbilityRewardXfer4F6240FailurePrefixes(t *testing.T) {
	for _, version := range []uint16{62, 0x7fff} {
		t.Run(fmt.Sprintf("reject-%#x", version), func(t *testing.T) {
			object := &abilityRewardXferTestObject4F6240{field34: 7}
			w := newAbilityRewardXferTestWorld4F6240()
			w.version = version

			if got := abilityRewardXfer4F6240(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if w.field34Loads != 1 || w.useDataLoads != 1 || len(w.mapVersions) != 0 || w.field34Stores != 0 {
				t.Fatalf("entry loads/map/store = %d/%d/%v/%d, want 1/1/none/0",
					w.field34Loads, w.useDataLoads, w.mapVersions, w.field34Stores)
			}
		})
	}

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := &abilityRewardXferTestObject4F6240{
			data: &abilityRewardXferTestData4F6240{ability: 1}, field34: 0x11223344,
		}
		w := newAbilityRewardXferTestWorld4F6240()
		w.mapResult = 0
		w.after["map-read-write:61"] = func() { object.field34 = 0x55667788 }

		if got := abilityRewardXfer4F6240(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 0x55667788 || len(w.nameLookups) != 0 || w.field34Stores != 0 {
			t.Fatalf("Field34/name/store = %#x/%v/%d, want callback state/none/0",
				object.field34, w.nameLookups, w.field34Stores)
		}
	})

	t.Run("length 128 stops before payload parse and rollback", func(t *testing.T) {
		data := &abilityRewardXferTestData4F6240{ability: 2}
		object := &abilityRewardXferTestObject4F6240{data: data, field34: 0x11223344}
		w := newAbilityRewardXferTestWorld4F6240()
		w.nameByID[2] = "ABILITY_WARCRY"
		w.byteResult = abilityRewardXferTestByte4F6240(abilityRewardXferNameLimit4F6240)
		w.after["map-read-write:61"] = func() { object.field34 = 0xaabbccdd }

		if got := abilityRewardXfer4F6240(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if data.ability != 2 || object.field34 != 0xaabbccdd || len(w.bytePayloads) != 0 ||
			len(w.idLookups) != 0 || len(w.storedAbilities) != 0 || w.field34Loads != 1 || w.field34Stores != 0 {
			t.Fatalf("ability/Field34/payload/lookups/stores/loads/restore = %d/%#x/%v/%v/%v/%d/%d",
				data.ability, object.field34, w.bytePayloads, w.idLookups, w.storedAbilities,
				w.field34Loads, w.field34Stores)
		}
	})

	t.Run("inventory failure keeps parsed ability and live count", func(t *testing.T) {
		data := &abilityRewardXferTestData4F6240{ability: 1}
		object := &abilityRewardXferTestObject4F6240{data: data, field34: 7}
		w := newAbilityRewardXferTestWorld4F6240()
		w.nameByID[1] = "A"
		w.byteResult = abilityRewardXferTestByte4F6240(1)
		w.payload = []byte("B")
		w.idByName["B"] = 2
		w.modes = []int32{1}
		w.inventoryResult = 0
		w.after["store-ability:2"] = func() { object.field34 = 0x80000004 }

		if got := abilityRewardXfer4F6240(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if data.ability != 2 || object.field34 != 0x80000004 || len(w.inventoryCalls) != 1 || w.field34Stores != 0 {
			t.Fatalf("ability/Field34/inventory/restore = %d/%#x/%d/%d, want 2/live/1/0",
				data.ability, object.field34, len(w.inventoryCalls), w.field34Stores)
		}
	})
}

func TestAbilityRewardXfer4F6240FaultBoundaries(t *testing.T) {
	t.Run("nil cached use data faults after common serialization", func(t *testing.T) {
		object := &abilityRewardXferTestObject4F6240{field34: 3}
		w := newAbilityRewardXferTestWorld4F6240()

		defer func() {
			if recover() == nil {
				t.Fatal("expected nil UseData fault")
			}
			want := []string{
				"load-field34:1", "load-use-data", "rw-version:61",
				"map-read-write:61",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		}()
		_ = abilityRewardXfer4F6240(object, w.deps())
	})

	t.Run("overlong runtime name faults before length IO", func(t *testing.T) {
		data := &abilityRewardXferTestData4F6240{ability: 1}
		object := &abilityRewardXferTestObject4F6240{data: data, field34: 3}
		w := newAbilityRewardXferTestWorld4F6240()
		w.nameByID[1] = strings.Repeat("A", abilityRewardXferNameLimit4F6240)

		defer func() {
			if recover() == nil {
				t.Fatal("expected PE32 stack-buffer contract fault")
			}
			if !reflect.DeepEqual(w.byteInputs, []uint8(nil)) || w.field34Stores != 0 {
				t.Fatalf("length IO/restore = %v/%d, want none/0", w.byteInputs, w.field34Stores)
			}
			wantSuffix := []string{"load-ability:1", "ability-name:1"}
			if len(w.events) < len(wantSuffix) || !reflect.DeepEqual(w.events[len(w.events)-len(wantSuffix):], wantSuffix) {
				t.Fatalf("event suffix = %v, want %v", w.events, wantSuffix)
			}
		}()
		_ = abilityRewardXfer4F6240(object, w.deps())
	})
}
