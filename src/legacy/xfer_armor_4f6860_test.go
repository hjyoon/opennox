package legacy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type armorXferTestModifier4F6860 struct {
	name string
}

type armorXferTestModifierData4F6860 struct {
	slots [4]*armorXferTestModifier4F6860
}

type armorXferTestHealth4F6860 struct {
	current uint16
	field2  uint16
	maximum uint16
}

type armorXferTestArmor4F6860 struct {
	hp uint16
}

type armorXferTestUpdate4F6860 struct {
	field4 uint32
}

type armorXferTestObject4F6860 struct {
	field34   uint32
	typeIndex uint16
	modifiers *armorXferTestModifierData4F6860
	health    *armorXferTestHealth4F6860
	update    *armorXferTestUpdate4F6860

	legacyEmpty bool
	applied     [4]int32
	appliedTail uint32
	setHP       []uint16
}

type armorXferTestNameWrite4F6860 struct {
	slot *armorXferTestModifier4F6860
	size uint8
}

type armorXferTestInventoryCall4F6860 struct {
	version uint16
	object  *armorXferTestObject4F6860
	count   int32
}

type armorXferTestWorld4F6860 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	byteOutputs     []uint8
	wordOutput      *uint16
	readNames       []string
	nameIDs         map[string]int32
	flagOutputs     []bool
	switchToSolo    int32
	notMultiplayer  int32
	anyPlayers      int32
	armor           *armorXferTestArmor4F6860
	inventoryResult int32

	events               []string
	after                map[string]func()
	field34Loads         int
	field34Stores        int
	modeCalls            int
	byteInputs           []uint8
	wordInputs           []uint16
	modifierDataLoads    int
	modifierSlotLoads    [4]int
	modifierNameWrites   []armorXferTestNameWrite4F6860
	readNameSizes        []uint8
	modifierApplications int
	legacyApplications   int
	healthLoads          int
	healthMaximumLoads   int
	switchCalls          int
	notMultiplayerCalls  int
	flagCalls            int
	anyPlayersCalls      int
	typeIndexLoads       int
	armorLookups         int
	armorHPLoads         int
	dummyTransfers       int
	updateLoads          int
	updateTransfers      int
	inventoryCalls       []armorXferTestInventoryCall4F6860
}

func newArmorXferTestObject4F6860() *armorXferTestObject4F6860 {
	return &armorXferTestObject4F6860{
		modifiers: &armorXferTestModifierData4F6860{},
		health:    &armorXferTestHealth4F6860{maximum: 100},
		update:    &armorXferTestUpdate4F6860{},
	}
}

func newArmorXferTestWorld4F6860() *armorXferTestWorld4F6860 {
	return &armorXferTestWorld4F6860{
		version:         armorXferCurrentVersion4F6860,
		mapResult:       1,
		nameIDs:         make(map[string]int32),
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func armorXferTestUint16Ptr4F6860(value uint16) *uint16 { return &value }

func (w *armorXferTestWorld4F6860) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *armorXferTestWorld4F6860) deps() armorXferDeps4F6860[
	*armorXferTestObject4F6860,
	*armorXferTestModifierData4F6860,
	*armorXferTestModifier4F6860,
	int32,
	*armorXferTestHealth4F6860,
	*armorXferTestArmor4F6860,
	*armorXferTestUpdate4F6860,
] {
	return armorXferDeps4F6860[
		*armorXferTestObject4F6860,
		*armorXferTestModifierData4F6860,
		*armorXferTestModifier4F6860,
		int32,
		*armorXferTestHealth4F6860,
		*armorXferTestArmor4F6860,
		*armorXferTestUpdate4F6860,
	]{
		loadField34: func(object *armorXferTestObject4F6860) uint32 {
			w.field34Loads++
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return object.field34
		},
		storeField34: func(object *armorXferTestObject4F6860, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *armorXferTestObject4F6860, version int32) int32 {
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
		applyLegacyEmptyModifiers: func(object *armorXferTestObject4F6860) {
			w.legacyApplications++
			object.legacyEmpty = true
			w.event("apply-legacy-empty-modifiers")
		},
		loadModifierData: func(object *armorXferTestObject4F6860) *armorXferTestModifierData4F6860 {
			w.modifierDataLoads++
			data := object.modifiers
			w.event("load-modifier-data")
			return data
		},
		loadModifierSlot: func(data *armorXferTestModifierData4F6860, index int) *armorXferTestModifier4F6860 {
			w.modifierSlotLoads[index]++
			w.event(fmt.Sprintf("load-modifier-slot:%d:%d", index, w.modifierSlotLoads[index]))
			return data.slots[index]
		},
		modifierNameLength: func(slot *armorXferTestModifier4F6860) uint32 {
			w.event(fmt.Sprintf("modifier-name-length:%s", slot.name))
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
		rwModifierName: func(slot *armorXferTestModifier4F6860, size uint8) {
			_ = slot.name
			w.modifierNameWrites = append(w.modifierNameWrites, armorXferTestNameWrite4F6860{slot: slot, size: size})
			w.event(fmt.Sprintf("rw-modifier-name:%s:%d", slot.name, size))
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
			w.event(fmt.Sprintf("modifier-id:%s", name))
			if id, ok := w.nameIDs[name]; ok {
				return id
			}
			return 255
		},
		modifierByID: func(id int32) int32 {
			w.event(fmt.Sprintf("modifier-by-id:%d", id))
			return id
		},
		applyModifiers: func(object *armorXferTestObject4F6860, modifiers [4]int32, tail uint32) {
			w.modifierApplications++
			object.applied = modifiers
			object.appliedTail = tail
			w.event("apply-modifiers")
		},
		unitGetHP: func(object *armorXferTestObject4F6860) uint16 {
			w.event("unit-get-hp")
			if object.health == nil {
				return 0
			}
			return object.health.current
		},
		rwWord: func(value uint16) uint16 {
			w.wordInputs = append(w.wordInputs, value)
			w.event(fmt.Sprintf("rw-word:%d", value))
			if w.wordOutput != nil {
				return *w.wordOutput
			}
			return value
		},
		loadHealthData: func(object *armorXferTestObject4F6860) *armorXferTestHealth4F6860 {
			w.healthLoads++
			health := object.health
			w.event(fmt.Sprintf("load-health-data:%d", w.healthLoads))
			return health
		},
		loadHealthMaximum: func(health *armorXferTestHealth4F6860) uint16 {
			w.healthMaximumLoads++
			w.event("load-health-maximum")
			return health.maximum
		},
		switchToSolo: func() int32 {
			w.switchCalls++
			w.event("switch-to-solo")
			return w.switchToSolo
		},
		notMultiplayer: func() int32 {
			w.notMultiplayerCalls++
			w.event("not-multiplayer")
			return w.notMultiplayer
		},
		gameFlag4096: func() bool {
			index := w.flagCalls
			w.flagCalls++
			value := false
			if index < len(w.flagOutputs) {
				value = w.flagOutputs[index]
			}
			w.event(fmt.Sprintf("game-flag-4096:%t", value))
			return value
		},
		anyTrackedPlayers: func() int32 {
			w.anyPlayersCalls++
			w.event("any-tracked-players")
			return w.anyPlayers
		},
		unitSetHP: func(object *armorXferTestObject4F6860, value uint16) {
			object.setHP = append(object.setHP, value)
			w.event(fmt.Sprintf("unit-set-hp:%d", value))
		},
		loadTypeIndex: func(object *armorXferTestObject4F6860) uint16 {
			w.typeIndexLoads++
			value := object.typeIndex
			w.event(fmt.Sprintf("load-type-index:%d", value))
			return value
		},
		armorClass: func(typeIndex uint16) *armorXferTestArmor4F6860 {
			w.armorLookups++
			w.event(fmt.Sprintf("armor-class:%d", typeIndex))
			return w.armor
		},
		loadArmorHP: func(armor *armorXferTestArmor4F6860) uint16 {
			w.armorHPLoads++
			value := armor.hp
			w.event(fmt.Sprintf("load-armor-hp:%d", w.armorHPLoads))
			return value
		},
		storeHealthMaximum: func(health *armorXferTestHealth4F6860, value uint16) {
			health.maximum = value
			w.event(fmt.Sprintf("store-health-maximum:%d", value))
		},
		storeHealthField2: func(health *armorXferTestHealth4F6860, value uint16) {
			health.field2 = value
			w.event(fmt.Sprintf("store-health-field2:%d", value))
		},
		rwDummyByte: func() {
			w.dummyTransfers++
			w.event("rw-dummy-byte")
		},
		loadUpdateData: func(object *armorXferTestObject4F6860) *armorXferTestUpdate4F6860 {
			w.updateLoads++
			update := object.update
			w.event("load-update-data")
			return update
		},
		rwUpdateField4: func(update *armorXferTestUpdate4F6860) {
			w.updateTransfers++
			w.event("rw-update-field4")
			_ = update.field4
		},
		transferInventory: func(version uint16, object *armorXferTestObject4F6860, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, armorXferTestInventoryCall4F6860{
				version: version,
				object:  object,
				count:   count,
			})
			w.event(fmt.Sprintf("transfer-inventory:%d:%d", version, count))
			return w.inventoryResult
		},
	}
}

func TestArmorXfer4F6860VersionAndCommonFailurePrefixes(t *testing.T) {
	t.Run("unsupported positive version", func(t *testing.T) {
		object := newArmorXferTestObject4F6860()
		object.field34 = 0x11223344
		w := newArmorXferTestWorld4F6860()
		w.version = 63
		w.after["rw-version:62"] = func() { object.field34 = 0x55667788 }

		if got := armorXfer4F6860(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if !reflect.DeepEqual(w.events, []string{"load-field34:1", "rw-version:62"}) ||
			object.field34 != 0x55667788 || w.field34Stores != 0 {
			t.Fatalf("events/Field34/stores = %v/%#x/%d, want exact rejection prefix/live/0",
				w.events, object.field34, w.field34Stores)
		}
	})

	t.Run("negative version is sign extended", func(t *testing.T) {
		object := newArmorXferTestObject4F6860()
		w := newArmorXferTestWorld4F6860()
		w.version = 0xffff
		w.modes = []int32{0}

		if got := armorXfer4F6860(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.events[:3], []string{"load-field34:1", "rw-version:62", "map-read-write:-1"}) {
			t.Fatalf("prefix = %v, want signed common version -1", w.events[:3])
		}
	})

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := newArmorXferTestObject4F6860()
		object.field34 = 7
		w := newArmorXferTestWorld4F6860()
		w.mapResult = 0
		w.after["map-read-write:62"] = func() { object.field34 = 9 }

		if got := armorXfer4F6860(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 9 || w.field34Loads != 1 || w.field34Stores != 0 || w.modeCalls != 0 {
			t.Fatalf("Field34/loads/stores/modes = %d/%d/%d/%d, want 9/1/0/0",
				object.field34, w.field34Loads, w.field34Stores, w.modeCalls)
		}
	})
}

func TestArmorXfer4F6860LegacyModifierShortcutDoesNotRestoreField34(t *testing.T) {
	object := newArmorXferTestObject4F6860()
	object.field34 = 5
	w := newArmorXferTestWorld4F6860()
	w.version = 10
	w.modes = []int32{1}
	w.after["map-read-write:10"] = func() { object.field34 = 9 }

	if got := armorXfer4F6860(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !object.legacyEmpty || object.field34 != 9 || w.legacyApplications != 1 ||
		w.modifierApplications != 0 || w.field34Loads != 1 || w.field34Stores != 0 || w.modeCalls != 1 {
		t.Fatalf("legacy/Field34/apply/loads/stores/modes = %t/%d/%d/%d/%d/%d, want true/9/1/1/0/1",
			object.legacyEmpty, object.field34, w.legacyApplications, w.field34Loads, w.field34Stores, w.modeCalls)
	}
}

func TestArmorXfer4F6860WriteModifiersUseLowByteAndLiveSlot(t *testing.T) {
	first := &armorXferTestModifier4F6860{name: "Old"}
	replacement := &armorXferTestModifier4F6860{name: "New"}
	long := &armorXferTestModifier4F6860{name: strings.Repeat("L", 257)}
	last := &armorXferTestModifier4F6860{name: "Z"}
	object := newArmorXferTestObject4F6860()
	object.modifiers.slots = [4]*armorXferTestModifier4F6860{first, nil, long, last}
	w := newArmorXferTestWorld4F6860()
	w.version = 40
	w.modes = []int32{0}
	w.byteOutputs = []uint8{2, 0, 1, 1}
	w.after["rw-byte:1:3"] = func() { object.modifiers.slots[0] = replacement }

	if got := armorXfer4F6860(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.byteInputs, []uint8{3, 0, 1, 1}) {
		t.Fatalf("length inputs = %v, want [3 0 1 1]", w.byteInputs)
	}
	wantWrites := []armorXferTestNameWrite4F6860{
		{slot: replacement, size: 2},
		{slot: long, size: 1},
		{slot: last, size: 1},
	}
	if !reflect.DeepEqual(w.modifierNameWrites, wantWrites) ||
		!reflect.DeepEqual(w.modifierSlotLoads, [4]int{2, 1, 2, 2}) {
		t.Fatalf("name writes/slot loads = %#v/%v, want live replacement/%v",
			w.modifierNameWrites, w.modifierSlotLoads, [4]int{2, 1, 2, 2})
	}
	if w.modifierDataLoads != 1 || w.modifierApplications != 0 || w.modeCalls != 1 {
		t.Fatalf("data loads/applications/modes = %d/%d/%d, want 1/0/1",
			w.modifierDataLoads, w.modifierApplications, w.modeCalls)
	}
}

func TestArmorXfer4F6860ReadModifiersAcceptsEveryByteLength(t *testing.T) {
	object := newArmorXferTestObject4F6860()
	w := newArmorXferTestWorld4F6860()
	w.version = 40
	w.modes = []int32{-7}
	w.byteOutputs = []uint8{0, 1, 255, 3}
	w.readNames = []string{"", "A", "Huge", "End"}
	w.nameIDs = map[string]int32{"A": 10, "Huge": 20, "End": 30}

	if got := armorXfer4F6860(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.readNameSizes, []uint8{0, 1, 255, 3}) ||
		!reflect.DeepEqual(object.applied, [4]int32{255, 10, 20, 30}) ||
		object.appliedTail != ^uint32(0) {
		t.Fatalf("sizes/modifiers/tail = %v/%v/%#x, want all byte sizes/[255 10 20 30]/max",
			w.readNameSizes, object.applied, object.appliedTail)
	}
	if w.modifierApplications != 1 || w.modifierDataLoads != 0 || w.modeCalls != 1 {
		t.Fatalf("applications/data loads/modes = %d/%d/%d, want 1/0/1",
			w.modifierApplications, w.modifierDataLoads, w.modeCalls)
	}
}

func TestArmorXfer4F6860HPClampAndShortCircuit(t *testing.T) {
	object := newArmorXferTestObject4F6860()
	object.health = &armorXferTestHealth4F6860{current: 17, maximum: 50}
	w := newArmorXferTestWorld4F6860()
	w.version = 41
	w.modes = []int32{0, 1}
	w.wordOutput = armorXferTestUint16Ptr4F6860(99)
	w.switchToSolo = 1

	if got := armorXfer4F6860(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(object.setHP, []uint16{50}) || !reflect.DeepEqual(w.wordInputs, []uint16{17}) {
		t.Fatalf("set HP/word input = %v/%v, want [50]/[17]", object.setHP, w.wordInputs)
	}
	if w.switchCalls != 1 || w.notMultiplayerCalls != 0 || w.flagCalls != 0 || w.armorLookups != 0 {
		t.Fatalf("switch/notMP/flag/armor = %d/%d/%d/%d, want 1/0/0/0",
			w.switchCalls, w.notMultiplayerCalls, w.flagCalls, w.armorLookups)
	}
}

func TestArmorXfer4F6860HPFallbackUsesLiveHealthAndArmorLoads(t *testing.T) {
	h0 := &armorXferTestHealth4F6860{current: 17, field2: 2, maximum: 50}
	h1 := &armorXferTestHealth4F6860{field2: 11, maximum: 12}
	h2 := &armorXferTestHealth4F6860{field2: 21, maximum: 22}
	armor := &armorXferTestArmor4F6860{hp: 70}
	object := newArmorXferTestObject4F6860()
	object.health = h0
	object.typeIndex = 0xabcd
	w := newArmorXferTestWorld4F6860()
	w.version = 41
	w.modes = []int32{0, 1}
	w.wordOutput = armorXferTestUint16Ptr4F6860(30)
	w.armor = armor
	w.flagOutputs = []bool{false}
	w.after["armor-class:43981"] = func() { object.health = h1 }
	w.after["load-health-data:2"] = func() { object.health = h2 }
	w.after["load-armor-hp:1"] = func() { armor.hp = 71 }
	w.after["load-armor-hp:2"] = func() { armor.hp = 72 }

	if got := armorXfer4F6860(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if h0.maximum != 50 || h0.field2 != 2 || h1.maximum != 70 || h1.field2 != 11 ||
		h2.maximum != 22 || h2.field2 != 71 || !reflect.DeepEqual(object.setHP, []uint16{72}) {
		t.Fatalf("healths/setHP = h0:%+v h1:%+v h2:%+v set:%v, want live 70/71/72 stores",
			*h0, *h1, *h2, object.setHP)
	}
	if w.healthLoads != 3 || w.armorHPLoads != 3 || w.typeIndexLoads != 1 ||
		w.armorLookups != 1 || w.anyPlayersCalls != 0 {
		t.Fatalf("health/armorHP/type/lookup/any = %d/%d/%d/%d/%d, want 3/3/1/1/0",
			w.healthLoads, w.armorHPLoads, w.typeIndexLoads, w.armorLookups, w.anyPlayersCalls)
	}
	wantOrder := []string{
		"armor-class:43981",
		"load-health-data:2",
		"load-armor-hp:1",
		"store-health-maximum:70",
		"load-health-data:3",
		"load-armor-hp:2",
		"store-health-field2:71",
		"load-armor-hp:3",
		"unit-set-hp:72",
	}
	found := false
	for i := 0; i+len(wantOrder) <= len(w.events); i++ {
		if reflect.DeepEqual(w.events[i:i+len(wantOrder)], wantOrder) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("fallback event order = %v, want contiguous %v", w.events, wantOrder)
	}
}

func TestArmorXfer4F6860HPSpecialModeUsesExactPredicates(t *testing.T) {
	object := newArmorXferTestObject4F6860()
	object.health.current = 25
	w := newArmorXferTestWorld4F6860()
	w.version = 41
	w.modes = []int32{0, 1}
	w.switchToSolo = 2
	w.notMultiplayer = -1
	w.flagOutputs = []bool{true}
	w.anyPlayers = -7

	if got := armorXfer4F6860(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(object.setHP, []uint16{25}) || w.switchCalls != 1 ||
		w.notMultiplayerCalls != 1 || w.flagCalls != 1 || w.anyPlayersCalls != 1 || w.armorLookups != 0 {
		t.Fatalf("set/switch/notMP/flag/any/armor = %v/%d/%d/%d/%d/%d, want [25]/1/1/1/1/0",
			object.setHP, w.switchCalls, w.notMultiplayerCalls, w.flagCalls, w.anyPlayersCalls, w.armorLookups)
	}
}

func TestArmorXfer4F6860VersionedSuffixAndInventoryRestore(t *testing.T) {
	t.Run("version 61 has only dummy byte", func(t *testing.T) {
		object := newArmorXferTestObject4F6860()
		w := newArmorXferTestWorld4F6860()
		w.version = 61
		w.modes = []int32{0, 0}

		if got := armorXfer4F6860(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.dummyTransfers != 1 || w.updateLoads != 0 || w.updateTransfers != 0 {
			t.Fatalf("dummy/update loads/transfers = %d/%d/%d, want 1/0/0",
				w.dummyTransfers, w.updateLoads, w.updateTransfers)
		}
	})

	for _, result := range []int32{1, 0} {
		t.Run(fmt.Sprintf("signed-version-inventory-result-%d", result), func(t *testing.T) {
			object := newArmorXferTestObject4F6860()
			object.field34 = 7
			w := newArmorXferTestWorld4F6860()
			w.version = 0xffff
			w.modes = []int32{0, 1}
			w.inventoryResult = result
			w.after["map-read-write:-1"] = func() { object.field34 = 0x80000004 }

			if got := armorXfer4F6860(object, w.deps()); got != result {
				t.Fatalf("result = %d, want %d", got, result)
			}
			wantCall := []armorXferTestInventoryCall4F6860{{
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

	t.Run("version 62 transfers update before live inventory", func(t *testing.T) {
		object := newArmorXferTestObject4F6860()
		object.field34 = 3
		w := newArmorXferTestWorld4F6860()
		w.modes = []int32{0, 0, 1}

		if got := armorXfer4F6860(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.dummyTransfers != 0 || w.updateLoads != 1 || w.updateTransfers != 1 || len(w.inventoryCalls) != 1 {
			t.Fatalf("dummy/update/inventory = %d/%d/%d/%d, want 0/1/1/1",
				w.dummyTransfers, w.updateLoads, w.updateTransfers, len(w.inventoryCalls))
		}
	})
}

func TestArmorXfer4F6860FaultBoundaries(t *testing.T) {
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

	t.Run("nil object faults at entry Field34", func(t *testing.T) {
		w := newArmorXferTestWorld4F6860()
		expectPanic(t, func() { armorXfer4F6860((*armorXferTestObject4F6860)(nil), w.deps()) })
		if !reflect.DeepEqual(w.events, []string{"load-field34:1"}) {
			t.Fatalf("events = %v, want entry Field34 load only", w.events)
		}
	})

	t.Run("nil modifier data faults after mode", func(t *testing.T) {
		object := newArmorXferTestObject4F6860()
		object.modifiers = nil
		w := newArmorXferTestWorld4F6860()
		w.version = 40
		w.modes = []int32{0}
		expectPanic(t, func() { armorXfer4F6860(object, w.deps()) })
		if w.modifierDataLoads != 1 || w.modifierSlotLoads[0] != 1 || w.field34Stores != 0 {
			t.Fatalf("modifier/slot/stores = %d/%d/%d, want 1/1/0",
				w.modifierDataLoads, w.modifierSlotLoads[0], w.field34Stores)
		}
	})

	t.Run("nil health faults after HP transfer", func(t *testing.T) {
		object := newArmorXferTestObject4F6860()
		object.health = nil
		w := newArmorXferTestWorld4F6860()
		w.version = 41
		w.modes = []int32{0}
		expectPanic(t, func() { armorXfer4F6860(object, w.deps()) })
		if !reflect.DeepEqual(w.wordInputs, []uint16{0}) || w.healthLoads != 1 ||
			w.healthMaximumLoads != 1 || w.field34Stores != 0 {
			t.Fatalf("word/health/max/stores = %v/%d/%d/%d, want [0]/1/1/0",
				w.wordInputs, w.healthLoads, w.healthMaximumLoads, w.field34Stores)
		}
	})

	t.Run("nil v62 update faults before inventory", func(t *testing.T) {
		object := newArmorXferTestObject4F6860()
		object.update = nil
		w := newArmorXferTestWorld4F6860()
		w.modes = []int32{0, 0}
		expectPanic(t, func() { armorXfer4F6860(object, w.deps()) })
		if w.updateLoads != 1 || w.updateTransfers != 1 || len(w.inventoryCalls) != 0 || w.field34Stores != 0 {
			t.Fatalf("update loads/transfers/inventory/stores = %d/%d/%d/%d, want 1/1/0/0",
				w.updateLoads, w.updateTransfers, len(w.inventoryCalls), w.field34Stores)
		}
	})
}
