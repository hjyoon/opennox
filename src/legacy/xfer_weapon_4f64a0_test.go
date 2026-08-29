package legacy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type weaponXferTestModifier4F64A0 struct {
	name string
}

type weaponXferTestModifierData4F64A0 struct {
	slots [4]*weaponXferTestModifier4F64A0
}

type weaponXferTestUse4F64A0 struct {
	current uint8
	maximum uint8
	value   int32
}

type weaponXferTestHealth4F64A0 struct {
	current uint16
	field2  uint16
	maximum uint16
}

type weaponXferTestProjectile4F64A0 struct {
	hp uint16
}

type weaponXferTestUpdate4F64A0 struct {
	field4 int32
}

type weaponXferTestObject4F64A0 struct {
	field34   uint32
	class     uint32
	subclass  uint32
	typeIndex uint16

	modifiers *weaponXferTestModifierData4F64A0
	use       *weaponXferTestUse4F64A0
	health    *weaponXferTestHealth4F64A0
	update    *weaponXferTestUpdate4F64A0

	legacyEmpty bool
	applied     [4]int32
	appliedTail uint32
	setHP       []uint16
}

type weaponXferTestNameWrite4F64A0 struct {
	slot *weaponXferTestModifier4F64A0
	size uint8
}

type weaponXferTestInventoryCall4F64A0 struct {
	version uint16
	object  *weaponXferTestObject4F64A0
	count   int32
}

type weaponXferTestWorld4F64A0 struct {
	version         uint16
	mapResult       int32
	modes           []int32
	byteOutputs     []uint8
	dwordOutput     *int32
	wordOutput      *uint16
	readNames       []string
	nameIDs         map[string]int32
	flagOutputs     []bool
	switchToSolo    int32
	notMultiplayer  int32
	anyPlayers      int32
	projectile      *weaponXferTestProjectile4F64A0
	inventoryResult int32

	events               []string
	after                map[string]func()
	field34Loads         int
	field34Stores        int
	modeCalls            int
	byteInputs           []uint8
	dwordInputs          []int32
	wordInputs           []uint16
	modifierDataLoads    int
	modifierSlotLoads    [4]int
	modifierNameWrites   []weaponXferTestNameWrite4F64A0
	readNameSizes        []uint8
	modifierApplications int
	legacyApplications   int
	classLoads           int
	subclassLoads        int
	useLoads             int
	chargeCurrentLoads   int
	chargeMaximumLoads   int
	chargeValueLoads     int
	flagCalls            int
	healthLoads          int
	healthMaximumLoads   int
	switchCalls          int
	notMultiplayerCalls  int
	anyPlayersCalls      int
	typeIndexLoads       int
	projectileLookups    int
	projectileHPLoads    int
	updateLoads          int
	updateTransfers      int
	inventoryCalls       []weaponXferTestInventoryCall4F64A0
}

func newWeaponXferTestObject4F64A0() *weaponXferTestObject4F64A0 {
	return &weaponXferTestObject4F64A0{
		modifiers: &weaponXferTestModifierData4F64A0{},
		use:       &weaponXferTestUse4F64A0{},
		health:    &weaponXferTestHealth4F64A0{maximum: 100},
		update:    &weaponXferTestUpdate4F64A0{},
	}
}

func newWeaponXferTestWorld4F64A0() *weaponXferTestWorld4F64A0 {
	return &weaponXferTestWorld4F64A0{
		version:         weaponXferCurrentVersion4F64A0,
		mapResult:       1,
		nameIDs:         make(map[string]int32),
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func weaponXferTestInt32Ptr4F64A0(value int32) *int32    { return &value }
func weaponXferTestUint16Ptr4F64A0(value uint16) *uint16 { return &value }

func (w *weaponXferTestWorld4F64A0) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *weaponXferTestWorld4F64A0) deps() weaponXferDeps4F64A0[
	*weaponXferTestObject4F64A0,
	*weaponXferTestModifierData4F64A0,
	*weaponXferTestModifier4F64A0,
	int32,
	*weaponXferTestUse4F64A0,
	*weaponXferTestHealth4F64A0,
	*weaponXferTestProjectile4F64A0,
	*weaponXferTestUpdate4F64A0,
] {
	return weaponXferDeps4F64A0[
		*weaponXferTestObject4F64A0,
		*weaponXferTestModifierData4F64A0,
		*weaponXferTestModifier4F64A0,
		int32,
		*weaponXferTestUse4F64A0,
		*weaponXferTestHealth4F64A0,
		*weaponXferTestProjectile4F64A0,
		*weaponXferTestUpdate4F64A0,
	]{
		loadField34: func(object *weaponXferTestObject4F64A0) uint32 {
			w.field34Loads++
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return object.field34
		},
		storeField34: func(object *weaponXferTestObject4F64A0, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *weaponXferTestObject4F64A0, version int32) int32 {
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
		applyLegacyEmptyModifiers: func(object *weaponXferTestObject4F64A0) {
			w.legacyApplications++
			object.legacyEmpty = true
			w.event("apply-legacy-empty-modifiers")
		},
		loadModifierData: func(object *weaponXferTestObject4F64A0) *weaponXferTestModifierData4F64A0 {
			w.modifierDataLoads++
			data := object.modifiers
			w.event("load-modifier-data")
			return data
		},
		loadModifierSlot: func(data *weaponXferTestModifierData4F64A0, index int) *weaponXferTestModifier4F64A0 {
			w.modifierSlotLoads[index]++
			w.event(fmt.Sprintf("load-modifier-slot:%d:%d", index, w.modifierSlotLoads[index]))
			return data.slots[index]
		},
		modifierNameLength: func(slot *weaponXferTestModifier4F64A0) uint32 {
			name := slot.name
			w.event(fmt.Sprintf("modifier-name-length:%s", name))
			return uint32(len(name))
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
		rwModifierName: func(slot *weaponXferTestModifier4F64A0, size uint8) {
			_ = slot.name
			w.modifierNameWrites = append(w.modifierNameWrites, weaponXferTestNameWrite4F64A0{slot: slot, size: size})
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
		applyModifiers: func(object *weaponXferTestObject4F64A0, modifiers [4]int32, tail uint32) {
			w.modifierApplications++
			object.applied = modifiers
			object.appliedTail = tail
			w.event("apply-modifiers")
		},
		loadClass: func(object *weaponXferTestObject4F64A0) uint32 {
			w.classLoads++
			value := object.class
			w.event("load-class")
			return value
		},
		loadSubclass: func(object *weaponXferTestObject4F64A0) uint32 {
			w.subclassLoads++
			value := object.subclass
			w.event(fmt.Sprintf("load-subclass:%d", w.subclassLoads))
			return value
		},
		loadUseData: func(object *weaponXferTestObject4F64A0) *weaponXferTestUse4F64A0 {
			w.useLoads++
			data := object.use
			w.event("load-use-data")
			return data
		},
		loadChargeCurrent: func(data *weaponXferTestUse4F64A0) uint8 {
			w.chargeCurrentLoads++
			value := data.current
			w.event(fmt.Sprintf("load-charge-current:%d", w.chargeCurrentLoads))
			return value
		},
		loadChargeMaximum: func(data *weaponXferTestUse4F64A0) uint8 {
			w.chargeMaximumLoads++
			value := data.maximum
			w.event(fmt.Sprintf("load-charge-maximum:%d", w.chargeMaximumLoads))
			return value
		},
		loadChargeValue: func(data *weaponXferTestUse4F64A0) int32 {
			w.chargeValueLoads++
			value := data.value
			w.event(fmt.Sprintf("load-charge-value:%d", w.chargeValueLoads))
			return value
		},
		rwDword: func(value int32) int32 {
			w.dwordInputs = append(w.dwordInputs, value)
			w.event(fmt.Sprintf("rw-dword:%d", value))
			if w.dwordOutput != nil {
				return *w.dwordOutput
			}
			return value
		},
		storeChargeCurrent: func(data *weaponXferTestUse4F64A0, value uint8) {
			data.current = value
			w.event(fmt.Sprintf("store-charge-current:%d", value))
		},
		storeChargeMaximum: func(data *weaponXferTestUse4F64A0, value uint8) {
			data.maximum = value
			w.event(fmt.Sprintf("store-charge-maximum:%d", value))
		},
		storeChargeValue: func(data *weaponXferTestUse4F64A0, value int32) {
			data.value = value
			w.event(fmt.Sprintf("store-charge-value:%d", value))
		},
		gameFlag4096: func() bool {
			index := w.flagCalls
			w.flagCalls++
			value := false
			if index < len(w.flagOutputs) {
				value = w.flagOutputs[index]
			}
			w.event(fmt.Sprintf("game-flag-4096:%d=%t", index+1, value))
			return value
		},
		unitGetHP: func(object *weaponXferTestObject4F64A0) uint16 {
			value := uint16(0)
			if object.health != nil {
				value = object.health.current
			}
			w.event(fmt.Sprintf("unit-get-hp:%d", value))
			return value
		},
		rwWord: func(value uint16) uint16 {
			w.wordInputs = append(w.wordInputs, value)
			w.event(fmt.Sprintf("rw-word:%d", value))
			if w.wordOutput != nil {
				return *w.wordOutput
			}
			return value
		},
		loadHealthData: func(object *weaponXferTestObject4F64A0) *weaponXferTestHealth4F64A0 {
			w.healthLoads++
			health := object.health
			w.event(fmt.Sprintf("load-health-data:%d", w.healthLoads))
			return health
		},
		loadHealthMaximum: func(health *weaponXferTestHealth4F64A0) uint16 {
			w.healthMaximumLoads++
			w.event("load-health-maximum")
			return health.maximum
		},
		switchToSolo: func() int32 {
			w.switchCalls++
			w.event(fmt.Sprintf("switch-to-solo:%d", w.switchToSolo))
			return w.switchToSolo
		},
		notMultiplayer: func() int32 {
			w.notMultiplayerCalls++
			w.event(fmt.Sprintf("not-multiplayer:%d", w.notMultiplayer))
			return w.notMultiplayer
		},
		anyTrackedPlayers: func() int32 {
			w.anyPlayersCalls++
			w.event(fmt.Sprintf("any-tracked-players:%d", w.anyPlayers))
			return w.anyPlayers
		},
		unitSetHP: func(object *weaponXferTestObject4F64A0, value uint16) {
			object.setHP = append(object.setHP, value)
			w.event(fmt.Sprintf("unit-set-hp:%d", value))
		},
		loadTypeIndex: func(object *weaponXferTestObject4F64A0) uint16 {
			w.typeIndexLoads++
			value := object.typeIndex
			w.event(fmt.Sprintf("load-type-index:%d", value))
			return value
		},
		projectileClass: func(index uint16) *weaponXferTestProjectile4F64A0 {
			w.projectileLookups++
			w.event(fmt.Sprintf("projectile-class:%d", index))
			return w.projectile
		},
		loadProjectileHP: func(projectile *weaponXferTestProjectile4F64A0) uint16 {
			w.projectileHPLoads++
			value := projectile.hp
			w.event(fmt.Sprintf("load-projectile-hp:%d", w.projectileHPLoads))
			return value
		},
		storeHealthMaximum: func(health *weaponXferTestHealth4F64A0, value uint16) {
			health.maximum = value
			w.event(fmt.Sprintf("store-health-maximum:%d", value))
		},
		storeHealthField2: func(health *weaponXferTestHealth4F64A0, value uint16) {
			health.field2 = value
			w.event(fmt.Sprintf("store-health-field2:%d", value))
		},
		loadUpdateData: func(object *weaponXferTestObject4F64A0) *weaponXferTestUpdate4F64A0 {
			w.updateLoads++
			update := object.update
			w.event("load-update-data")
			return update
		},
		rwUpdateField4: func(update *weaponXferTestUpdate4F64A0) {
			w.updateTransfers++
			w.event("rw-update-field4")
			update.field4++
		},
		transferInventory: func(version uint16, object *weaponXferTestObject4F64A0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, weaponXferTestInventoryCall4F64A0{
				version: version,
				object:  object,
				count:   count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
	}
}

func TestWeaponXfer4F64A0VersionAndCommonFailurePrefixes(t *testing.T) {
	t.Run("signed maximum reject", func(t *testing.T) {
		object := newWeaponXferTestObject4F64A0()
		object.field34 = 7
		w := newWeaponXferTestWorld4F64A0()
		w.version = 65

		if got := weaponXfer4F64A0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"load-field34:1", "rw-version:64"}
		if !reflect.DeepEqual(w.events, want) || w.field34Stores != 0 || object.field34 != 7 {
			t.Fatalf("events/stores/Field34 = %v/%d/%d, want %v/0/7", w.events, w.field34Stores, object.field34, want)
		}
	})

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := newWeaponXferTestObject4F64A0()
		object.field34 = 0x11223344
		w := newWeaponXferTestWorld4F64A0()
		w.mapResult = 0
		w.after["map-read-write:64"] = func() { object.field34 = 0x55667788 }

		if got := weaponXfer4F64A0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 0x55667788 || w.field34Loads != 1 || w.field34Stores != 0 || w.modeCalls != 0 {
			t.Fatalf("Field34/loads/stores/modes = %#x/%d/%d/%d, want callback state/1/0/0",
				object.field34, w.field34Loads, w.field34Stores, w.modeCalls)
		}
	})
}

func TestWeaponXfer4F64A0LegacyModifierShortcutDoesNotRestoreField34(t *testing.T) {
	object := newWeaponXferTestObject4F64A0()
	object.field34 = 5
	w := newWeaponXferTestWorld4F64A0()
	w.version = 10
	w.modes = []int32{1}
	w.after["map-read-write:10"] = func() { object.field34 = 9 }

	if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !object.legacyEmpty || object.field34 != 9 || w.legacyApplications != 1 ||
		w.modifierApplications != 0 || w.field34Loads != 1 || w.field34Stores != 0 || w.modeCalls != 1 {
		t.Fatalf("legacy/Field34/apply/loads/stores/modes = %t/%d/%d/%d/%d/%d, want true/9/1/1/0/1",
			object.legacyEmpty, object.field34, w.legacyApplications, w.field34Loads, w.field34Stores, w.modeCalls)
	}
}

func TestWeaponXfer4F64A0WriteModifiersUseLowByteAndLiveSlot(t *testing.T) {
	first := &weaponXferTestModifier4F64A0{name: "Old"}
	replacement := &weaponXferTestModifier4F64A0{name: "New"}
	long := &weaponXferTestModifier4F64A0{name: strings.Repeat("L", 257)}
	last := &weaponXferTestModifier4F64A0{name: "Z"}
	object := newWeaponXferTestObject4F64A0()
	object.modifiers.slots = [4]*weaponXferTestModifier4F64A0{first, nil, long, last}
	w := newWeaponXferTestWorld4F64A0()
	w.version = 40
	w.modes = []int32{0}
	w.byteOutputs = []uint8{2, 0, 1, 1}
	w.after["rw-byte:1:3"] = func() { object.modifiers.slots[0] = replacement }

	if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.byteInputs, []uint8{3, 0, 1, 1}) {
		t.Fatalf("length inputs = %v, want [3 0 1 1]", w.byteInputs)
	}
	wantWrites := []weaponXferTestNameWrite4F64A0{
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

func TestWeaponXfer4F64A0ReadModifiersAcceptsEveryByteLength(t *testing.T) {
	object := newWeaponXferTestObject4F64A0()
	w := newWeaponXferTestWorld4F64A0()
	w.version = 40
	w.modes = []int32{2}
	w.byteOutputs = []uint8{0, 1, 255, 3}
	w.readNames = []string{"", "A", "Huge", "End"}
	w.nameIDs = map[string]int32{"A": 10, "Huge": 20, "End": 30}

	if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
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

func TestWeaponXfer4F64A0ChargeUsesCachedValidationAndLiveTransferLoads(t *testing.T) {
	object := newWeaponXferTestObject4F64A0()
	object.class = weaponXferClassMask4F64A0
	object.subclass = 0x00010000
	object.use = &weaponXferTestUse4F64A0{current: 3, maximum: 9, value: 100}
	w := newWeaponXferTestWorld4F64A0()
	w.modes = []int32{0, 0}
	w.byteOutputs = []uint8{0, 0, 0, 0, 8, 9}
	w.dwordOutput = weaponXferTestInt32Ptr4F64A0(50)
	w.flagOutputs = []bool{true}
	w.after["load-charge-current:1"] = func() { object.use.current = 7 }
	w.after["load-charge-maximum:1"] = func() { object.use.maximum = 12 }
	w.after["load-charge-value:1"] = func() { object.use.value = 80 }

	if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if got := *object.use; got != (weaponXferTestUse4F64A0{current: 8, maximum: 9, value: 50}) {
		t.Fatalf("charge = %+v, want current/max/value 8/9/50", got)
	}
	if !reflect.DeepEqual(w.byteInputs, []uint8{0, 0, 0, 0, 7, 12}) ||
		!reflect.DeepEqual(w.dwordInputs, []int32{80}) {
		t.Fatalf("byte/dword inputs = %v/%v, want live 7/12 and live 80", w.byteInputs, w.dwordInputs)
	}
	if w.chargeCurrentLoads != 2 || w.chargeMaximumLoads != 2 || w.chargeValueLoads != 2 ||
		w.flagCalls != 1 || w.updateTransfers != 1 {
		t.Fatalf("charge loads/flags/update = %d/%d/%d/%d/%d, want 2/2/2/1/1",
			w.chargeCurrentLoads, w.chargeMaximumLoads, w.chargeValueLoads, w.flagCalls, w.updateTransfers)
	}
}

func TestWeaponXfer4F64A0ChargeValidationResetAndFlagBypass(t *testing.T) {
	tests := []struct {
		name    string
		current uint8
		maximum uint8
		value   int32
		flag    bool
		want    weaponXferTestUse4F64A0
	}{
		{"current above cached max", 10, 9, 50, true, weaponXferTestUse4F64A0{maximum: 9}},
		{"transferred max mismatch", 8, 8, 50, true, weaponXferTestUse4F64A0{maximum: 9}},
		{"negative value", 8, 9, -1, true, weaponXferTestUse4F64A0{maximum: 9}},
		{"value above cached value", 8, 9, 101, true, weaponXferTestUse4F64A0{maximum: 9}},
		{"flag disabled bypasses validation", 250, 8, 200, false, weaponXferTestUse4F64A0{current: 250, maximum: 8, value: 200}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := newWeaponXferTestObject4F64A0()
			object.class = weaponXferClassMask4F64A0
			object.subclass = 0x00010000
			object.use = &weaponXferTestUse4F64A0{current: 3, maximum: 9, value: 100}
			w := newWeaponXferTestWorld4F64A0()
			w.version = 61
			w.modes = []int32{0, 0}
			w.byteOutputs = []uint8{0, 0, 0, 0, tc.current, tc.maximum}
			w.dwordOutput = weaponXferTestInt32Ptr4F64A0(tc.value)
			w.flagOutputs = []bool{tc.flag}

			if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if got := *object.use; got != tc.want {
				t.Fatalf("charge = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestWeaponXfer4F64A0ChargeVersion62SubclassBoundary(t *testing.T) {
	tests := []struct {
		name         string
		version      uint16
		class        uint32
		subclass     uint32
		wantUseLoads int
		wantSubLoads int
	}{
		{"v60 forbidden subclass skips", 60, weaponXferClassMask4F64A0, weaponXferLegacySkipMask4F64A0, 0, 2},
		{"v62 forbidden subclass transfers", 62, weaponXferClassMask4F64A0, weaponXferLegacySkipMask4F64A0, 1, 1},
		{"class miss skips every subclass load", 62, 0, weaponXferLegacySkipMask4F64A0, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := newWeaponXferTestObject4F64A0()
			object.class = tc.class
			object.subclass = tc.subclass
			object.use = &weaponXferTestUse4F64A0{current: 1, maximum: 2, value: 3}
			w := newWeaponXferTestWorld4F64A0()
			w.version = tc.version
			w.modes = []int32{0, 0}

			if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if w.useLoads != tc.wantUseLoads || w.subclassLoads != tc.wantSubLoads {
				t.Fatalf("use/subclass loads = %d/%d, want %d/%d",
					w.useLoads, w.subclassLoads, tc.wantUseLoads, tc.wantSubLoads)
			}
		})
	}
}

func TestWeaponXfer4F64A0HPClampAndShortCircuit(t *testing.T) {
	object := newWeaponXferTestObject4F64A0()
	object.health = &weaponXferTestHealth4F64A0{current: 17, maximum: 50}
	w := newWeaponXferTestWorld4F64A0()
	w.version = 42
	w.modes = []int32{0, 1}
	w.wordOutput = weaponXferTestUint16Ptr4F64A0(99)
	w.switchToSolo = 1

	if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(object.setHP, []uint16{50}) || !reflect.DeepEqual(w.wordInputs, []uint16{17}) {
		t.Fatalf("set HP/word input = %v/%v, want [50]/[17]", object.setHP, w.wordInputs)
	}
	if w.switchCalls != 1 || w.notMultiplayerCalls != 0 || w.flagCalls != 0 ||
		w.projectileLookups != 0 {
		t.Fatalf("switch/notMP/flag/projectile = %d/%d/%d/%d, want 1/0/0/0",
			w.switchCalls, w.notMultiplayerCalls, w.flagCalls, w.projectileLookups)
	}
}

func TestWeaponXfer4F64A0HPFallbackUsesLiveHealthAndProjectileLoads(t *testing.T) {
	h0 := &weaponXferTestHealth4F64A0{current: 17, field2: 2, maximum: 50}
	h1 := &weaponXferTestHealth4F64A0{field2: 11, maximum: 12}
	h2 := &weaponXferTestHealth4F64A0{field2: 21, maximum: 22}
	projectile := &weaponXferTestProjectile4F64A0{hp: 70}
	object := newWeaponXferTestObject4F64A0()
	object.health = h0
	object.typeIndex = 0xabcd
	w := newWeaponXferTestWorld4F64A0()
	w.version = 42
	w.modes = []int32{0, 1}
	w.wordOutput = weaponXferTestUint16Ptr4F64A0(30)
	w.projectile = projectile
	w.flagOutputs = []bool{false}
	w.after["projectile-class:43981"] = func() { object.health = h1 }
	w.after["load-health-data:2"] = func() { object.health = h2 }
	w.after["load-projectile-hp:1"] = func() {
		projectile.hp = 71
	}
	w.after["load-projectile-hp:2"] = func() {
		projectile.hp = 72
	}

	if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if h0.maximum != 50 || h0.field2 != 2 || h1.maximum != 70 || h1.field2 != 11 ||
		h2.maximum != 22 || h2.field2 != 71 || !reflect.DeepEqual(object.setHP, []uint16{72}) {
		t.Fatalf("healths/setHP = h0:%+v h1:%+v h2:%+v set:%v, want live 70/71/72 stores",
			*h0, *h1, *h2, object.setHP)
	}
	if w.healthLoads != 3 || w.projectileHPLoads != 3 || w.typeIndexLoads != 1 ||
		w.projectileLookups != 1 || w.anyPlayersCalls != 0 {
		t.Fatalf("health/projectileHP/type/lookup/any = %d/%d/%d/%d/%d, want 3/3/1/1/0",
			w.healthLoads, w.projectileHPLoads, w.typeIndexLoads, w.projectileLookups, w.anyPlayersCalls)
	}
	wantOrder := []string{
		"projectile-class:43981",
		"load-health-data:2",
		"load-projectile-hp:1",
		"store-health-maximum:70",
		"load-health-data:3",
		"load-projectile-hp:2",
		"store-health-field2:71",
		"load-projectile-hp:3",
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

func TestWeaponXfer4F64A0HPSpecialModeUsesExactPredicates(t *testing.T) {
	object := newWeaponXferTestObject4F64A0()
	object.health.current = 25
	w := newWeaponXferTestWorld4F64A0()
	w.version = 42
	w.modes = []int32{0, 1}
	w.switchToSolo = 2
	w.notMultiplayer = -1
	w.flagOutputs = []bool{true}
	w.anyPlayers = -7

	if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(object.setHP, []uint16{25}) || w.switchCalls != 1 ||
		w.notMultiplayerCalls != 1 || w.flagCalls != 1 || w.anyPlayersCalls != 1 || w.projectileLookups != 0 {
		t.Fatalf("set/switch/notMP/flag/any/projectile = %v/%d/%d/%d/%d/%d, want [25]/1/1/1/1/0",
			object.setHP, w.switchCalls, w.notMultiplayerCalls, w.flagCalls, w.anyPlayersCalls, w.projectileLookups)
	}
}

func TestWeaponXfer4F64A0Version63DummyUsesTransferredChargeByte(t *testing.T) {
	object := newWeaponXferTestObject4F64A0()
	object.class = weaponXferClassMask4F64A0
	object.subclass = 0x00010000
	object.use = &weaponXferTestUse4F64A0{current: 4, maximum: 5, value: 6}
	w := newWeaponXferTestWorld4F64A0()
	w.version = 63
	w.modes = []int32{0, 0}
	w.byteOutputs = []uint8{0, 0, 0, 0, 7, 5, 8}

	if got := weaponXfer4F64A0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.byteInputs, []uint8{0, 0, 0, 0, 4, 5, 7}) {
		t.Fatalf("byte inputs = %v, want modifier bytes, charge 4/5, dummy transferred-current 7", w.byteInputs)
	}
	if w.updateLoads != 0 || w.updateTransfers != 0 {
		t.Fatalf("update loads/transfers = %d/%d, want 0/0 at v63", w.updateLoads, w.updateTransfers)
	}
}

func TestWeaponXfer4F64A0SignedVersionInventoryAndRestore(t *testing.T) {
	for _, result := range []int32{1, 0} {
		t.Run(fmt.Sprintf("inventory-result-%d", result), func(t *testing.T) {
			object := newWeaponXferTestObject4F64A0()
			object.field34 = 7
			w := newWeaponXferTestWorld4F64A0()
			w.version = 0xffff
			w.modes = []int32{0, 1}
			w.inventoryResult = result
			w.after["map-read-write:-1"] = func() { object.field34 = 0x80000004 }

			if got := weaponXfer4F64A0(object, w.deps()); got != result {
				t.Fatalf("result = %d, want %d", got, result)
			}
			wantCall := []weaponXferTestInventoryCall4F64A0{{
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
}

func TestWeaponXfer4F64A0FaultBoundaries(t *testing.T) {
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
		w := newWeaponXferTestWorld4F64A0()
		expectPanic(t, func() { weaponXfer4F64A0((*weaponXferTestObject4F64A0)(nil), w.deps()) })
		if !reflect.DeepEqual(w.events, []string{"load-field34:1"}) {
			t.Fatalf("events = %v, want entry Field34 load only", w.events)
		}
	})

	t.Run("nil modifier data faults after mode", func(t *testing.T) {
		object := newWeaponXferTestObject4F64A0()
		object.modifiers = nil
		w := newWeaponXferTestWorld4F64A0()
		w.version = 40
		w.modes = []int32{0}
		expectPanic(t, func() { weaponXfer4F64A0(object, w.deps()) })
		if w.modifierDataLoads != 1 || w.modifierSlotLoads[0] != 1 || w.field34Stores != 0 {
			t.Fatalf("modifier/slot/stores = %d/%d/%d, want 1/1/0",
				w.modifierDataLoads, w.modifierSlotLoads[0], w.field34Stores)
		}
	})

	t.Run("nil eligible use data faults at first charge load", func(t *testing.T) {
		object := newWeaponXferTestObject4F64A0()
		object.class = weaponXferClassMask4F64A0
		object.subclass = 0x00010000
		object.use = nil
		w := newWeaponXferTestWorld4F64A0()
		w.version = 41
		w.modes = []int32{0}
		expectPanic(t, func() { weaponXfer4F64A0(object, w.deps()) })
		if w.useLoads != 1 || w.chargeCurrentLoads != 1 || w.field34Stores != 0 {
			t.Fatalf("use/current/stores = %d/%d/%d, want 1/1/0", w.useLoads, w.chargeCurrentLoads, w.field34Stores)
		}
	})

	t.Run("nil health faults after HP transfer", func(t *testing.T) {
		object := newWeaponXferTestObject4F64A0()
		object.health = nil
		w := newWeaponXferTestWorld4F64A0()
		w.version = 42
		w.modes = []int32{0}
		expectPanic(t, func() { weaponXfer4F64A0(object, w.deps()) })
		if !reflect.DeepEqual(w.wordInputs, []uint16{0}) || w.healthLoads != 1 ||
			w.healthMaximumLoads != 1 || w.field34Stores != 0 {
			t.Fatalf("word/health/max/stores = %v/%d/%d/%d, want [0]/1/1/0",
				w.wordInputs, w.healthLoads, w.healthMaximumLoads, w.field34Stores)
		}
	})

	t.Run("nil v64 update faults after HP block", func(t *testing.T) {
		object := newWeaponXferTestObject4F64A0()
		object.update = nil
		w := newWeaponXferTestWorld4F64A0()
		w.modes = []int32{0, 0}
		expectPanic(t, func() { weaponXfer4F64A0(object, w.deps()) })
		if w.updateLoads != 1 || w.updateTransfers != 1 || w.field34Stores != 0 {
			t.Fatalf("update loads/transfers/stores = %d/%d/%d, want 1/1/0",
				w.updateLoads, w.updateTransfers, w.field34Stores)
		}
	})
}
