package legacy

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

type teamXferTestModifier4F6D20 struct{ name string }

type teamXferTestModifierData4F6D20 struct {
	slots [4]*teamXferTestModifier4F6D20
}

type teamXferTestUpdate4F6D20 struct {
	x float32
	y float32
}

type teamXferTestObject4F6D20 struct {
	field34   uint32
	modifiers *teamXferTestModifierData4F6D20
	objClass  uint32
	update    *teamXferTestUpdate4F6D20
	positionX float32
	positionY float32
	applied   [4]int32
	tail      uint32
}

type teamXferTestNameWrite4F6D20 struct {
	slot *teamXferTestModifier4F6D20
	size uint8
}

type teamXferTestInventoryCall4F6D20 struct {
	version uint16
	object  *teamXferTestObject4F6D20
	count   int32
}

type teamXferTestWorld4F6D20 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	byteOutputs     []uint8
	readNames       []string
	nameIDs         map[string]int32
	inventoryResult int32

	events                []string
	after                 map[string]func()
	field34Loads          int
	field34Stores         int
	mapVersions           []int32
	modeCalls             int
	modifierDataLoads     int
	modifierSlotLoads     [4]int
	byteInputs            []uint8
	modifierNameWrites    []teamXferTestNameWrite4F6D20
	readNameSizes         []uint8
	modifierApplications  int
	objClassLoads         int
	updateDataLoads       int
	positionXLoads        int
	positionYLoads        int
	updatePositionXStores int
	updatePositionYStores int
	inventoryCalls        []teamXferTestInventoryCall4F6D20
}

func newTeamXferTestObject4F6D20() *teamXferTestObject4F6D20 {
	return &teamXferTestObject4F6D20{
		modifiers: &teamXferTestModifierData4F6D20{},
		update:    &teamXferTestUpdate4F6D20{},
	}
}

func newTeamXferTestWorld4F6D20() *teamXferTestWorld4F6D20 {
	return &teamXferTestWorld4F6D20{
		version:         teamXferCurrentVersion4F6D20,
		mapResult:       1,
		nameIDs:         make(map[string]int32),
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *teamXferTestWorld4F6D20) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *teamXferTestWorld4F6D20) deps() teamXferDeps4F6D20[
	*teamXferTestObject4F6D20,
	*teamXferTestModifierData4F6D20,
	*teamXferTestModifier4F6D20,
	int32,
	*teamXferTestUpdate4F6D20,
] {
	return teamXferDeps4F6D20[
		*teamXferTestObject4F6D20,
		*teamXferTestModifierData4F6D20,
		*teamXferTestModifier4F6D20,
		int32,
		*teamXferTestUpdate4F6D20,
	]{
		loadField34: func(object *teamXferTestObject4F6D20) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		storeField34: func(object *teamXferTestObject4F6D20, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *teamXferTestObject4F6D20, version int32) int32 {
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
		loadModifierData: func(object *teamXferTestObject4F6D20) *teamXferTestModifierData4F6D20 {
			w.modifierDataLoads++
			value := object.modifiers
			w.event("load-modifier-data")
			return value
		},
		loadModifierSlot: func(data *teamXferTestModifierData4F6D20, index int) *teamXferTestModifier4F6D20 {
			w.modifierSlotLoads[index]++
			value := data.slots[index]
			w.event(fmt.Sprintf("load-modifier-slot:%d:%d", index, w.modifierSlotLoads[index]))
			return value
		},
		modifierNameLength: func(slot *teamXferTestModifier4F6D20) uint32 {
			value := uint32(len(slot.name))
			w.event("modifier-name-length:" + slot.name)
			return value
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
		rwModifierName: func(slot *teamXferTestModifier4F6D20, size uint8) {
			name := slot.name
			w.modifierNameWrites = append(w.modifierNameWrites, teamXferTestNameWrite4F6D20{slot: slot, size: size})
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
			return -1
		},
		modifierByID: func(id int32) int32 {
			w.event(fmt.Sprintf("modifier-by-id:%d", id))
			return id
		},
		applyModifiers: func(object *teamXferTestObject4F6D20, modifiers [4]int32, tail uint32) {
			w.modifierApplications++
			object.applied = modifiers
			object.tail = tail
			w.event("apply-modifiers")
		},
		loadObjClass: func(object *teamXferTestObject4F6D20) uint32 {
			w.objClassLoads++
			value := object.objClass
			w.event("load-obj-class")
			return value
		},
		loadUpdateData: func(object *teamXferTestObject4F6D20) *teamXferTestUpdate4F6D20 {
			w.updateDataLoads++
			value := object.update
			w.event("load-update-data")
			return value
		},
		loadPositionX: func(object *teamXferTestObject4F6D20) float32 {
			w.positionXLoads++
			value := object.positionX
			w.event("load-position-x")
			return value
		},
		storeUpdatePositionX: func(update *teamXferTestUpdate4F6D20, value float32) {
			w.updatePositionXStores++
			update.x = value
			w.event("store-update-position-x")
		},
		loadPositionY: func(object *teamXferTestObject4F6D20) float32 {
			w.positionYLoads++
			value := object.positionY
			w.event("load-position-y")
			return value
		},
		storeUpdatePositionY: func(update *teamXferTestUpdate4F6D20, value float32) {
			w.updatePositionYStores++
			update.y = value
			w.event("store-update-position-y")
		},
		transferInventory: func(version uint16, object *teamXferTestObject4F6D20, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, teamXferTestInventoryCall4F6D20{
				version: version,
				object:  object,
				count:   count,
			})
			w.event(fmt.Sprintf("transfer-inventory:%d:%d", version, count))
			return w.inventoryResult
		},
	}
}

func teamXferAssertPanics4F6D20(t *testing.T, call func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("call did not panic")
		}
	}()
	call()
}

func TestTeamXfer4F6D20VersionAndCommonFailurePrefixes(t *testing.T) {
	t.Run("nil object faults at entry Field34", func(t *testing.T) {
		w := newTeamXferTestWorld4F6D20()
		teamXferAssertPanics4F6D20(t, func() { teamXfer4F6D20(nil, w.deps()) })
		if len(w.events) != 0 || w.field34Stores != 0 {
			t.Fatalf("events/stores = %v/%d, want none", w.events, w.field34Stores)
		}
	})

	t.Run("unsupported positive version keeps callback mutation", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 0x11223344
		w := newTeamXferTestWorld4F6D20()
		w.version = 61
		w.after["rw-version:60"] = func() { object.field34 = 0x55667788 }

		if got := teamXfer4F6D20(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "rw-version:60"}
		if !reflect.DeepEqual(w.events, want) || object.field34 != 0x55667788 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/stores = %v/%#x/%d, want %v/live/0", w.events, object.field34, w.field34Stores, want)
		}
	})

	t.Run("negative version is sign extended", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		w := newTeamXferTestWorld4F6D20()
		w.version = math.MaxUint16
		w.modes = []int32{0}

		if got := teamXfer4F6D20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.mapVersions, []int32{-1}) {
			t.Fatalf("map versions = %v, want signed -1", w.mapVersions)
		}
	})

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 7
		w := newTeamXferTestWorld4F6D20()
		w.mapResult = 0
		w.after["map-read-write:60"] = func() { object.field34 = 9 }

		if got := teamXfer4F6D20(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 9 || w.field34Loads != 1 || w.field34Stores != 0 || w.modeCalls != 0 {
			t.Fatalf("Field34/loads/stores/modes = %d/%d/%d/%d, want 9/1/0/0",
				object.field34, w.field34Loads, w.field34Stores, w.modeCalls)
		}
	})
}

func TestTeamXfer4F6D20WritesCachedModifierDataAndLiveSlots(t *testing.T) {
	first := &teamXferTestModifier4F6D20{name: "Old"}
	replacement := &teamXferTestModifier4F6D20{name: "New"}
	long := &teamXferTestModifier4F6D20{name: strings.Repeat("L", 257)}
	empty := &teamXferTestModifier4F6D20{}
	entryData := &teamXferTestModifierData4F6D20{
		slots: [4]*teamXferTestModifier4F6D20{first, nil, long, empty},
	}
	object := newTeamXferTestObject4F6D20()
	object.modifiers = entryData
	w := newTeamXferTestWorld4F6D20()
	w.modes = []int32{0}
	w.byteOutputs = []uint8{2, 0, 1, 0}
	w.after["load-modifier-data"] = func() {
		object.modifiers = &teamXferTestModifierData4F6D20{}
	}
	w.after["rw-byte:1:3"] = func() { entryData.slots[0] = replacement }

	if got := teamXfer4F6D20(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.byteInputs, []uint8{3, 0, 1, 0}) {
		t.Fatalf("byte inputs = %v, want low-byte modifier lengths", w.byteInputs)
	}
	wantWrites := []teamXferTestNameWrite4F6D20{
		{slot: replacement, size: 2},
		{slot: long, size: 1},
		{slot: empty, size: 0},
	}
	if !reflect.DeepEqual(w.modifierNameWrites, wantWrites) ||
		!reflect.DeepEqual(w.modifierSlotLoads, [4]int{2, 1, 2, 2}) || w.modifierDataLoads != 1 {
		t.Fatalf("writes/slot loads/data loads = %#v/%v/%d, want live slots/%v/1",
			w.modifierNameWrites, w.modifierSlotLoads, w.modifierDataLoads, [4]int{2, 1, 2, 2})
	}
	if w.modeCalls != 1 || w.field34Stores != 1 {
		t.Fatalf("mode calls/Field34 stores = %d/%d, want 1/1 for zero live count", w.modeCalls, w.field34Stores)
	}
}

func TestTeamXfer4F6D20ReadsModifiersThenResetsLiveFlagHome(t *testing.T) {
	entryUpdate := &teamXferTestUpdate4F6D20{x: -1, y: -2}
	replacementUpdate := &teamXferTestUpdate4F6D20{x: 100, y: 200}
	object := newTeamXferTestObject4F6D20()
	object.update = entryUpdate
	w := newTeamXferTestWorld4F6D20()
	w.modes = []int32{-7}
	w.byteOutputs = []uint8{0, 1, 255, 3}
	w.readNames = []string{"", "A", "Huge", "End"}
	w.nameIDs = map[string]int32{"": -1, "A": 10, "Huge": 20, "End": 30}
	w.after["apply-modifiers"] = func() {
		object.objClass = 0x10000000
		object.positionX = 12.5
		object.positionY = 25.5
	}
	w.after["load-update-data"] = func() { object.update = replacementUpdate }
	w.after["store-update-position-x"] = func() { object.positionY = 77.25 }

	if got := teamXfer4F6D20(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.readNameSizes, []uint8{0, 1, 255, 3}) ||
		object.applied != ([4]int32{-1, 10, 20, 30}) || object.tail != math.MaxUint32 {
		t.Fatalf("read sizes/applied/tail = %v/%v/%#x", w.readNameSizes, object.applied, object.tail)
	}
	if *entryUpdate != (teamXferTestUpdate4F6D20{x: 12.5, y: 77.25}) ||
		*replacementUpdate != (teamXferTestUpdate4F6D20{x: 100, y: 200}) {
		t.Fatalf("cached/replacement updates = %+v/%+v", *entryUpdate, *replacementUpdate)
	}
	if w.objClassLoads != 1 || w.updateDataLoads != 1 || w.positionXLoads != 1 || w.positionYLoads != 1 ||
		w.updatePositionXStores != 1 || w.updatePositionYStores != 1 {
		t.Fatalf("flag access counts = class=%d update=%d x=%d/%d y=%d/%d, want all one",
			w.objClassLoads, w.updateDataLoads, w.positionXLoads, w.updatePositionXStores,
			w.positionYLoads, w.updatePositionYStores)
	}

	applyIndex := slicesIndexTeam4F6D20(w.events, "apply-modifiers")
	classIndex := slicesIndexTeam4F6D20(w.events, "load-obj-class")
	updateIndex := slicesIndexTeam4F6D20(w.events, "load-update-data")
	xLoadIndex := slicesIndexTeam4F6D20(w.events, "load-position-x")
	xStoreIndex := slicesIndexTeam4F6D20(w.events, "store-update-position-x")
	yLoadIndex := slicesIndexTeam4F6D20(w.events, "load-position-y")
	yStoreIndex := slicesIndexTeam4F6D20(w.events, "store-update-position-y")
	if !(applyIndex < classIndex && classIndex < updateIndex && updateIndex < xLoadIndex &&
		xLoadIndex < xStoreIndex && xStoreIndex < yLoadIndex && yLoadIndex < yStoreIndex) {
		t.Fatalf("flag access order is wrong: %v", w.events)
	}

	t.Run("non-flag skips update data", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		w := newTeamXferTestWorld4F6D20()
		w.modes = []int32{1}
		if got := teamXfer4F6D20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.objClassLoads != 1 || w.updateDataLoads != 0 || w.positionXLoads != 0 || w.positionYLoads != 0 {
			t.Fatalf("class/update/x/y loads = %d/%d/%d/%d, want 1/0/0/0",
				w.objClassLoads, w.updateDataLoads, w.positionXLoads, w.positionYLoads)
		}
	})
}

func slicesIndexTeam4F6D20(values []string, target string) int {
	for index, value := range values {
		if value == target {
			return index
		}
	}
	return -1
}

func TestTeamXfer4F6D20InventoryGateAndRestoration(t *testing.T) {
	t.Run("live count exact mode and nonzero result restore entry", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 7
		w := newTeamXferTestWorld4F6D20()
		w.modes = []int32{0, 1}
		w.after["map-read-write:60"] = func() { object.field34 = 0x80000004 }
		w.after["transfer-inventory:60:-2147483644"] = func() { object.field34 = 99 }
		w.inventoryResult = -7

		if got := teamXfer4F6D20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want canonical 1", got)
		}
		want := []teamXferTestInventoryCall4F6D20{{
			version: 60, object: object, count: -2147483644,
		}}
		if !reflect.DeepEqual(w.inventoryCalls, want) || object.field34 != 7 || w.field34Stores != 1 {
			t.Fatalf("calls/Field34/stores = %+v/%d/%d", w.inventoryCalls, object.field34, w.field34Stores)
		}
	})

	t.Run("inventory zero result preserves live callback state", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 5
		w := newTeamXferTestWorld4F6D20()
		w.modes = []int32{0, 1}
		w.inventoryResult = 0
		w.after["transfer-inventory:60:5"] = func() { object.field34 = 88 }

		if got := teamXfer4F6D20(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 88 || w.field34Stores != 0 {
			t.Fatalf("Field34/stores = %d/%d, want live 88/0", object.field34, w.field34Stores)
		}
	})

	t.Run("zero live count skips second mode", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 5
		w := newTeamXferTestWorld4F6D20()
		w.modes = []int32{0}
		w.after["map-read-write:60"] = func() { object.field34 = 0 }

		if got := teamXfer4F6D20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 5 {
			t.Fatalf("modes/inventory/Field34 = %d/%d/%d, want 1/0/5", w.modeCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("non-exact second mode skips inventory", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 5
		w := newTeamXferTestWorld4F6D20()
		w.modes = []int32{0, -1}

		if got := teamXfer4F6D20(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 2 || len(w.inventoryCalls) != 0 || w.field34Stores != 1 {
			t.Fatalf("modes/inventory/stores = %d/%d/%d, want 2/0/1", w.modeCalls, len(w.inventoryCalls), w.field34Stores)
		}
	})
}

func TestTeamXfer4F6D20FaultBoundariesDoNotRestoreField34(t *testing.T) {
	t.Run("nil write modifier data", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 4
		object.modifiers = nil
		w := newTeamXferTestWorld4F6D20()
		w.modes = []int32{0}

		teamXferAssertPanics4F6D20(t, func() { teamXfer4F6D20(object, w.deps()) })
		if w.modifierDataLoads != 1 || w.field34Stores != 0 || object.field34 != 4 {
			t.Fatalf("data loads/stores/Field34 = %d/%d/%d, want 1/0/4", w.modifierDataLoads, w.field34Stores, object.field34)
		}
	})

	t.Run("live write slot becomes nil", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 4
		object.modifiers.slots[0] = &teamXferTestModifier4F6D20{name: "A"}
		w := newTeamXferTestWorld4F6D20()
		w.modes = []int32{0}
		w.after["rw-byte:1:1"] = func() { object.modifiers.slots[0] = nil }

		teamXferAssertPanics4F6D20(t, func() { teamXfer4F6D20(object, w.deps()) })
		if w.modifierSlotLoads[0] != 2 || w.field34Stores != 0 {
			t.Fatalf("slot loads/stores = %d/%d, want 2/0", w.modifierSlotLoads[0], w.field34Stores)
		}
	})

	t.Run("nil flag update faults after live X load", func(t *testing.T) {
		object := newTeamXferTestObject4F6D20()
		object.field34 = 4
		object.objClass = 0x10000000
		object.update = nil
		w := newTeamXferTestWorld4F6D20()
		w.modes = []int32{1}

		teamXferAssertPanics4F6D20(t, func() { teamXfer4F6D20(object, w.deps()) })
		if w.updateDataLoads != 1 || w.positionXLoads != 1 || w.updatePositionXStores != 1 ||
			w.positionYLoads != 0 || w.field34Stores != 0 {
			t.Fatalf("update/x/store/y/restores = %d/%d/%d/%d/%d, want 1/1/1/0/0",
				w.updateDataLoads, w.positionXLoads, w.updatePositionXStores, w.positionYLoads, w.field34Stores)
		}
	})
}
