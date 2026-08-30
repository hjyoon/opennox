package legacy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type ammoXferTestModifier4F6B20 struct{ name string }

type ammoXferTestModifierData4F6B20 struct {
	slots [4]*ammoXferTestModifier4F6B20
}

type ammoXferTestUse4F6B20 struct {
	bytes [3]uint8
}

type ammoXferTestObject4F6B20 struct {
	use       *ammoXferTestUse4F6B20
	field34   uint32
	modifiers *ammoXferTestModifierData4F6B20
	applied   [4]int32
	tail      uint32
}

type ammoXferTestNameWrite4F6B20 struct {
	slot *ammoXferTestModifier4F6B20
	size uint8
}

type ammoXferTestInventoryCall4F6B20 struct {
	version uint16
	object  *ammoXferTestObject4F6B20
	count   int32
}

type ammoXferTestWorld4F6B20 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	byteOutputs     []uint8
	readNames       []string
	nameIDs         map[string]int32
	quest           bool
	inventoryResult int32

	events               []string
	after                map[string]func()
	useDataLoads         int
	field34Loads         int
	field34Stores        int
	mapVersions          []int32
	modeCalls            int
	modifierDataLoads    int
	modifierSlotLoads    [4]int
	byteInputs           []uint8
	modifierNameWrites   []ammoXferTestNameWrite4F6B20
	readNameSizes        []uint8
	modifierApplications int
	useByteLoads         []int
	useByteStores        []int
	questCalls           int
	inventoryCalls       []ammoXferTestInventoryCall4F6B20
}

func newAmmoXferTestObject4F6B20() *ammoXferTestObject4F6B20 {
	return &ammoXferTestObject4F6B20{
		use:       &ammoXferTestUse4F6B20{},
		modifiers: &ammoXferTestModifierData4F6B20{},
	}
}

func newAmmoXferTestWorld4F6B20() *ammoXferTestWorld4F6B20 {
	return &ammoXferTestWorld4F6B20{
		version:         ammoXferCurrentVersion4F6B20,
		mapResult:       1,
		nameIDs:         make(map[string]int32),
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *ammoXferTestWorld4F6B20) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *ammoXferTestWorld4F6B20) deps() ammoXferDeps4F6B20[
	*ammoXferTestObject4F6B20,
	*ammoXferTestModifierData4F6B20,
	*ammoXferTestModifier4F6B20,
	int32,
	*ammoXferTestUse4F6B20,
] {
	return ammoXferDeps4F6B20[
		*ammoXferTestObject4F6B20,
		*ammoXferTestModifierData4F6B20,
		*ammoXferTestModifier4F6B20,
		int32,
		*ammoXferTestUse4F6B20,
	]{
		loadUseData: func(object *ammoXferTestObject4F6B20) *ammoXferTestUse4F6B20 {
			w.useDataLoads++
			w.event("load-use-data")
			return object.use
		},
		loadField34: func(object *ammoXferTestObject4F6B20) uint32 {
			w.field34Loads++
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return object.field34
		},
		storeField34: func(object *ammoXferTestObject4F6B20, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *ammoXferTestObject4F6B20, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
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
		loadModifierData: func(object *ammoXferTestObject4F6B20) *ammoXferTestModifierData4F6B20 {
			w.modifierDataLoads++
			w.event("load-modifier-data")
			return object.modifiers
		},
		loadModifierSlot: func(data *ammoXferTestModifierData4F6B20, index int) *ammoXferTestModifier4F6B20 {
			w.modifierSlotLoads[index]++
			w.event(fmt.Sprintf("load-modifier-slot:%d:%d", index, w.modifierSlotLoads[index]))
			return data.slots[index]
		},
		modifierNameLength: func(slot *ammoXferTestModifier4F6B20) uint32 {
			w.event("modifier-name-length:" + slot.name)
			return uint32(len(slot.name))
		},
		rwByte: func(value uint8) uint8 {
			index := len(w.byteInputs)
			w.byteInputs = append(w.byteInputs, value)
			w.event(fmt.Sprintf("rw-byte:%d:%d", index+1, value))
			if index < len(w.byteOutputs) {
				return w.byteOutputs[index]
			}
			return value
		},
		rwModifierName: func(slot *ammoXferTestModifier4F6B20, size uint8) {
			name := slot.name
			w.modifierNameWrites = append(w.modifierNameWrites, ammoXferTestNameWrite4F6B20{slot: slot, size: size})
			w.event(fmt.Sprintf("rw-modifier-name:%s:%d", name, size))
		},
		readModifierName: func(size uint8) string {
			index := len(w.readNameSizes)
			w.readNameSizes = append(w.readNameSizes, size)
			w.event(fmt.Sprintf("read-modifier-name:%d", size))
			if index < len(w.readNames) {
				return w.readNames[index]
			}
			return ""
		},
		modifierIDByName: func(name string) int32 {
			w.event("modifier-id:" + name)
			if value, ok := w.nameIDs[name]; ok {
				return value
			}
			return 255
		},
		modifierByID: func(id int32) int32 {
			w.event(fmt.Sprintf("modifier-by-id:%d", id))
			return id
		},
		applyModifiers: func(object *ammoXferTestObject4F6B20, modifiers [4]int32, tail uint32) {
			w.modifierApplications++
			object.applied = modifiers
			object.tail = tail
			w.event("apply-modifiers")
		},
		loadUseByte: func(data *ammoXferTestUse4F6B20, index int) uint8 {
			w.useByteLoads = append(w.useByteLoads, index)
			w.event(fmt.Sprintf("load-use-byte:%d", index))
			return data.bytes[index]
		},
		storeUseByte: func(data *ammoXferTestUse4F6B20, index int, value uint8) {
			w.useByteStores = append(w.useByteStores, index)
			w.event(fmt.Sprintf("store-use-byte:%d:%d", index, value))
			data.bytes[index] = value
		},
		gameFlag4096: func() bool {
			w.questCalls++
			w.event(fmt.Sprintf("game-flag-4096:%t", w.quest))
			return w.quest
		},
		transferInventory: func(version uint16, object *ammoXferTestObject4F6B20, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, ammoXferTestInventoryCall4F6B20{
				version: version,
				object:  object,
				count:   count,
			})
			w.event(fmt.Sprintf("transfer-inventory:%d:%d", version, count))
			return w.inventoryResult
		},
	}
}

func TestAmmoXfer4F6B20VersionAndCommonFailurePrefixes(t *testing.T) {
	t.Run("unsupported positive version", func(t *testing.T) {
		object := newAmmoXferTestObject4F6B20()
		object.field34 = 0x11223344
		w := newAmmoXferTestWorld4F6B20()
		w.version = 61
		w.after["rw-version:60"] = func() { object.field34 = 0x55667788 }

		if got := ammoXfer4F6B20(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-use-data", "load-field34:1", "rw-version:60"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 0x55667788 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/stores = %v/%#x/%d, want %v/live/0", w.events, object.field34, w.field34Stores, want)
		}
	})

	t.Run("negative version is sign extended", func(t *testing.T) {
		object := newAmmoXferTestObject4F6B20()
		w := newAmmoXferTestWorld4F6B20()
		w.version = 0xffff
		w.modes = []int32{0}

		if got := ammoXfer4F6B20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.mapVersions, []int32{-1}) {
			t.Fatalf("map versions = %v, want signed -1", w.mapVersions)
		}
	})

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := newAmmoXferTestObject4F6B20()
		object.field34 = 7
		w := newAmmoXferTestWorld4F6B20()
		w.mapResult = 0
		w.after["map-read-write:60"] = func() { object.field34 = 9 }

		if got := ammoXfer4F6B20(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 9 || w.field34Loads != 1 || w.field34Stores != 0 || w.modeCalls != 0 {
			t.Fatalf("Field34/loads/stores/modes = %d/%d/%d/%d, want 9/1/0/0",
				object.field34, w.field34Loads, w.field34Stores, w.modeCalls)
		}
	})
}

func TestAmmoXfer4F6B20CachesUseDataAndWritesLowByteNamesAndCharges(t *testing.T) {
	first := &ammoXferTestModifier4F6B20{name: "Old"}
	replacementSlot := &ammoXferTestModifier4F6B20{name: "New"}
	long := &ammoXferTestModifier4F6B20{name: strings.Repeat("L", 257)}
	last := &ammoXferTestModifier4F6B20{name: "Z"}
	entryUse := &ammoXferTestUse4F6B20{bytes: [3]uint8{10, 20, 30}}
	replacementUse := &ammoXferTestUse4F6B20{bytes: [3]uint8{40, 50, 60}}
	object := newAmmoXferTestObject4F6B20()
	object.use = entryUse
	object.modifiers.slots = [4]*ammoXferTestModifier4F6B20{first, nil, long, last}
	w := newAmmoXferTestWorld4F6B20()
	w.modes = []int32{0}
	w.byteOutputs = []uint8{2, 0, 1, 1, 99, 88}
	w.after["map-read-write:60"] = func() { object.use = replacementUse }
	w.after["rw-byte:1:3"] = func() { object.modifiers.slots[0] = replacementSlot }

	if got := ammoXfer4F6B20(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if !reflect.DeepEqual(w.byteInputs, []uint8{3, 0, 1, 1, 20, 10}) {
		t.Fatalf("byte inputs = %v, want modifier lengths then cached charge1/charge0", w.byteInputs)
	}
	wantWrites := []ammoXferTestNameWrite4F6B20{
		{slot: replacementSlot, size: 2},
		{slot: long, size: 1},
		{slot: last, size: 1},
	}
	if !reflect.DeepEqual(w.modifierNameWrites, wantWrites) ||
		!reflect.DeepEqual(w.modifierSlotLoads, [4]int{2, 1, 2, 2}) {
		t.Fatalf("writes/slot loads = %#v/%v, want live reloads/%v",
			w.modifierNameWrites, w.modifierSlotLoads, [4]int{2, 1, 2, 2})
	}
	if !reflect.DeepEqual(w.useByteLoads, []int{1, 0}) || entryUse.bytes != [3]uint8{10, 20, 30} ||
		replacementUse.bytes != [3]uint8{40, 50, 60} || w.useDataLoads != 1 {
		t.Fatalf("use loads/entry/replacement/cache loads = %v/%v/%v/%d, want [1 0]/unchanged/unchanged/1",
			w.useByteLoads, entryUse.bytes, replacementUse.bytes, w.useDataLoads)
	}
}

func TestAmmoXfer4F6B20ReadModifiersAndQuestChargeOrder(t *testing.T) {
	entryUse := &ammoXferTestUse4F6B20{bytes: [3]uint8{1, 2, 3}}
	replacementUse := &ammoXferTestUse4F6B20{bytes: [3]uint8{4, 5, 6}}
	object := newAmmoXferTestObject4F6B20()
	object.use = entryUse
	w := newAmmoXferTestWorld4F6B20()
	w.modes = []int32{-7}
	w.byteOutputs = []uint8{0, 1, 255, 3, 0xa1, 0xb2}
	w.readNames = []string{"", "A", "Huge", "End"}
	w.nameIDs = map[string]int32{"A": 10, "Huge": 20, "End": 30}
	w.quest = true
	w.after["map-read-write:60"] = func() { object.use = replacementUse }

	if got := ammoXfer4F6B20(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.readNameSizes, []uint8{0, 1, 255, 3}) ||
		!reflect.DeepEqual(object.applied, [4]int32{255, 10, 20, 30}) || object.tail != ^uint32(0) {
		t.Fatalf("sizes/modifiers/tail = %v/%v/%#x, want all byte sizes/[255 10 20 30]/max",
			w.readNameSizes, object.applied, object.tail)
	}
	if entryUse.bytes != [3]uint8{0xb2, 0xa1, 0} || replacementUse.bytes != [3]uint8{4, 5, 6} {
		t.Fatalf("entry/replacement use = %v/%v, want [178 161 0]/unchanged", entryUse.bytes, replacementUse.bytes)
	}
	if !reflect.DeepEqual(w.useByteStores, []int{2, 1, 0}) || w.questCalls != 1 ||
		w.modifierApplications != 1 || w.modifierDataLoads != 0 {
		t.Fatalf("stores/quest/applications/data loads = %v/%d/%d/%d, want [2 1 0]/1/1/0",
			w.useByteStores, w.questCalls, w.modifierApplications, w.modifierDataLoads)
	}
	wantTail := []string{
		"apply-modifiers",
		"rw-byte:5:0",
		"rw-byte:6:0",
		"game-flag-4096:true",
		"store-use-byte:2:0",
		"store-use-byte:1:161",
		"store-use-byte:0:178",
		"load-field34:2",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events[len(w.events)-len(wantTail):], wantTail) {
		t.Fatalf("tail events = %v, want %v", w.events[len(w.events)-len(wantTail):], wantTail)
	}
}

func TestAmmoXfer4F6B20InventoryGateAndFailureRollback(t *testing.T) {
	for _, result := range []int32{-3, 0} {
		t.Run(fmt.Sprintf("inventory-result-%d", result), func(t *testing.T) {
			object := newAmmoXferTestObject4F6B20()
			object.field34 = 7
			w := newAmmoXferTestWorld4F6B20()
			w.version = 0xffff
			w.modes = []int32{0, 1}
			w.inventoryResult = result
			w.after["map-read-write:-1"] = func() { object.field34 = 0x80000004 }

			wantResult := int32(1)
			if result == 0 {
				wantResult = 0
			}
			if got := ammoXfer4F6B20(object, w.deps()); got != wantResult {
				t.Fatalf("result = %d, want %d", got, wantResult)
			}
			wantCall := []ammoXferTestInventoryCall4F6B20{{
				version: 0xffff,
				object:  object,
				count:   -2147483644,
			}}
			if !reflect.DeepEqual(w.inventoryCalls, wantCall) {
				t.Fatalf("inventory calls = %#v, want %#v", w.inventoryCalls, wantCall)
			}
			if result != 0 {
				if object.field34 != 7 || w.field34Stores != 1 {
					t.Fatalf("success Field34/stores = %#x/%d, want 7/1", object.field34, w.field34Stores)
				}
			} else if object.field34 != 0x80000004 || w.field34Stores != 0 {
				t.Fatalf("failure Field34/stores = %#x/%d, want live/0", object.field34, w.field34Stores)
			}
		})
	}

	t.Run("non-exact mode skips inventory and restores", func(t *testing.T) {
		object := newAmmoXferTestObject4F6B20()
		object.field34 = 9
		w := newAmmoXferTestWorld4F6B20()
		w.modes = []int32{0, 2}
		w.after["map-read-write:60"] = func() { object.field34 = 4 }

		if got := ammoXfer4F6B20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.inventoryCalls) != 0 || object.field34 != 9 || w.field34Stores != 1 {
			t.Fatalf("inventory/Field34/stores = %d/%d/%d, want 0/9/1", len(w.inventoryCalls), object.field34, w.field34Stores)
		}
	})
}

func TestAmmoXfer4F6B20FaultBoundaries(t *testing.T) {
	expectPanic := func(t *testing.T, run func()) {
		t.Helper()
		panicked := false
		func() {
			defer func() { panicked = recover() != nil }()
			run()
		}()
		if !panicked {
			t.Fatal("call did not panic")
		}
	}

	t.Run("nil object faults at entry UseData", func(t *testing.T) {
		w := newAmmoXferTestWorld4F6B20()
		expectPanic(t, func() { ammoXfer4F6B20((*ammoXferTestObject4F6B20)(nil), w.deps()) })
		if !reflect.DeepEqual(w.events, []string{"load-use-data"}) {
			t.Fatalf("events = %v, want entry UseData load only", w.events)
		}
	})

	t.Run("nil modifier data faults before cached UseData", func(t *testing.T) {
		object := newAmmoXferTestObject4F6B20()
		object.modifiers = nil
		w := newAmmoXferTestWorld4F6B20()
		w.modes = []int32{0}
		expectPanic(t, func() { ammoXfer4F6B20(object, w.deps()) })
		if w.modifierDataLoads != 1 || w.modifierSlotLoads[0] != 1 || len(w.useByteLoads) != 0 || w.field34Stores != 0 {
			t.Fatalf("modifier/slot/use/stores = %d/%d/%v/%d, want 1/1/none/0",
				w.modifierDataLoads, w.modifierSlotLoads[0], w.useByteLoads, w.field34Stores)
		}
	})

	t.Run("nil cached UseData write faults after four modifiers", func(t *testing.T) {
		object := newAmmoXferTestObject4F6B20()
		object.use = nil
		w := newAmmoXferTestWorld4F6B20()
		w.modes = []int32{0}
		expectPanic(t, func() { ammoXfer4F6B20(object, w.deps()) })
		if !reflect.DeepEqual(w.byteInputs, []uint8{0, 0, 0, 0}) || !reflect.DeepEqual(w.useByteLoads, []int{1}) || w.field34Stores != 0 {
			t.Fatalf("bytes/use loads/stores = %v/%v/%d, want four zero lengths/[1]/0",
				w.byteInputs, w.useByteLoads, w.field34Stores)
		}
	})

	t.Run("nil cached UseData read faults after quest predicate", func(t *testing.T) {
		object := newAmmoXferTestObject4F6B20()
		object.use = nil
		w := newAmmoXferTestWorld4F6B20()
		w.modes = []int32{1}
		w.quest = false
		expectPanic(t, func() { ammoXfer4F6B20(object, w.deps()) })
		if w.modifierApplications != 1 || w.questCalls != 1 || !reflect.DeepEqual(w.useByteStores, []int{1}) || w.field34Stores != 0 {
			t.Fatalf("applications/quest/stores/Field34 stores = %d/%d/%v/%d, want 1/1/[1]/0",
				w.modifierApplications, w.questCalls, w.useByteStores, w.field34Stores)
		}
	})
}
