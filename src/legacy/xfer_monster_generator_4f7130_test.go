package legacy

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

type monsterGeneratorXferTestObject4F7130 struct {
	id       string
	data     *monsterGeneratorXferTestData4F7130
	field34  uint32
	scriptID int
}

type monsterGeneratorXferTestData4F7130 struct {
	spawn      []uint8
	quest      []uint8
	active     uint8
	max        uint8
	frame88    uint32
	field92    uint32
	prototypes []*monsterGeneratorXferTestObject4F7130
}

type monsterGeneratorXferTestScriptCall4F7130 struct {
	data       *monsterGeneratorXferTestData4F7130
	slot       monsterGeneratorScriptSlot4F7130
	scriptData int
	offset     uintptr
}

type monsterGeneratorXferTestStore4F7130 struct {
	data   *monsterGeneratorXferTestData4F7130
	index  int
	object *monsterGeneratorXferTestObject4F7130
}

type monsterGeneratorXferTestInventoryCall4F7130 struct {
	version uint16
	object  *monsterGeneratorXferTestObject4F7130
	count   int32
}

type monsterGeneratorXferTestWorld4F7130 struct {
	version         uint16
	mapResult       int32
	spawnCount      uint8
	questCount      uint8
	modes           []int32
	groupCount      uint8
	prototypeCounts []uint8
	readNameLengths []uint8
	readNames       [][]byte
	created         map[string]*monsterGeneratorXferTestObject4F7130
	xferResults     map[*monsterGeneratorXferTestObject4F7130]int32
	inventoryResult int32
	scriptResult    int32
	saveResult      int32
	writeNameLength *uint8

	events             []string
	after              map[string]func()
	modeCalls          int
	readBranch         bool
	prototypeCountCall int
	readNameCall       int
	readBytesCall      int
	updatePointers     []*monsterGeneratorXferTestData4F7130
	scriptCalls        []monsterGeneratorXferTestScriptCall4F7130
	writtenNames       [][]byte
	typeNameCalls      map[*monsterGeneratorXferTestObject4F7130]int
	saved              []*monsterGeneratorXferTestObject4F7130
	tagCalls           int
	crcCalls           int
	xferCalls          []*monsterGeneratorXferTestObject4F7130
	stores             []monsterGeneratorXferTestStore4F7130
	inventoryCalls     []monsterGeneratorXferTestInventoryCall4F7130
	field34Stores      int
	mapVersions        []int32
}

func newMonsterGeneratorXferTestWorld4F7130() *monsterGeneratorXferTestWorld4F7130 {
	return &monsterGeneratorXferTestWorld4F7130{
		version:         monsterGeneratorXferCurrentVersion4F7130,
		mapResult:       1,
		spawnCount:      monsterGeneratorXferGroups4F7130,
		questCount:      monsterGeneratorXferGroups4F7130,
		inventoryResult: 1,
		scriptResult:    1,
		saveResult:      1,
		created:         make(map[string]*monsterGeneratorXferTestObject4F7130),
		xferResults:     make(map[*monsterGeneratorXferTestObject4F7130]int32),
		after:           make(map[string]func()),
		typeNameCalls:   make(map[*monsterGeneratorXferTestObject4F7130]int),
	}
}

func (w *monsterGeneratorXferTestWorld4F7130) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *monsterGeneratorXferTestWorld4F7130) useData(
	name string,
	data *monsterGeneratorXferTestData4F7130,
) {
	w.updatePointers = append(w.updatePointers, data)
	w.event(name)
}

func (w *monsterGeneratorXferTestWorld4F7130) deps() monsterGeneratorXferDeps4F7130[
	*monsterGeneratorXferTestObject4F7130,
	*monsterGeneratorXferTestData4F7130,
	int,
] {
	return monsterGeneratorXferDeps4F7130[
		*monsterGeneratorXferTestObject4F7130,
		*monsterGeneratorXferTestData4F7130,
		int,
	]{
		loadUpdateData: func(object *monsterGeneratorXferTestObject4F7130) *monsterGeneratorXferTestData4F7130 {
			data := object.data
			w.event("load-update-data")
			return data
		},
		loadField34: func(object *monsterGeneratorXferTestObject4F7130) uint32 {
			w.event("load-field34")
			return object.field34
		},
		storeField34: func(object *monsterGeneratorXferTestObject4F7130, value uint32) {
			object.field34 = value
			w.field34Stores++
			w.event(fmt.Sprintf("store-field34:%d", value))
		},
		loadScriptData: func(object *monsterGeneratorXferTestObject4F7130) int {
			w.event("load-script-data")
			return object.scriptID
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *monsterGeneratorXferTestObject4F7130, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
		},
		rwSpawnSelectorCount: func(value uint8) uint8 {
			w.event(fmt.Sprintf("rw-spawn-count:%d", value))
			return w.spawnCount
		},
		rwSpawnSelector: func(data *monsterGeneratorXferTestData4F7130, index int) {
			w.useData(fmt.Sprintf("rw-spawn:%d", index), data)
			data.spawn[index]++
		},
		rwActiveCount: func(data *monsterGeneratorXferTestData4F7130) {
			w.useData("rw-active", data)
			data.active++
		},
		rwMaxActive: func(data *monsterGeneratorXferTestData4F7130) {
			w.useData("rw-max", data)
			data.max++
		},
		rwFrame88: func(data *monsterGeneratorXferTestData4F7130) {
			w.useData("rw-frame88", data)
			data.frame88++
		},
		transferScript: func(data *monsterGeneratorXferTestData4F7130, slot monsterGeneratorScriptSlot4F7130, scriptData int, offset uintptr) int32 {
			w.updatePointers = append(w.updatePointers, data)
			w.scriptCalls = append(w.scriptCalls, monsterGeneratorXferTestScriptCall4F7130{
				data: data, slot: slot, scriptData: scriptData, offset: offset,
			})
			w.event(fmt.Sprintf("script:%d:%d", slot, offset))
			return w.scriptResult
		},
		readMode: func() int32 {
			index := w.modeCalls
			w.modeCalls++
			var value int32
			if index < len(w.modes) {
				value = w.modes[index]
			}
			if index == 0 {
				w.readBranch = value != 0
			}
			w.event(fmt.Sprintf("read-mode:%d=%d", index+1, value))
			return value
		},
		rwPrototypeGroupCount: func(value uint8) uint8 {
			w.event(fmt.Sprintf("rw-group-count:%d", value))
			if w.readBranch {
				return w.groupCount
			}
			return 0xff
		},
		loadPrototype: func(data *monsterGeneratorXferTestData4F7130, index int) *monsterGeneratorXferTestObject4F7130 {
			w.updatePointers = append(w.updatePointers, data)
			w.event(fmt.Sprintf("load-prototype:%d", index))
			return data.prototypes[index]
		},
		rwPrototypeCount: func(value uint8) uint8 {
			w.event(fmt.Sprintf("rw-prototype-count:%d", value))
			if !w.readBranch {
				return 0xff
			}
			index := w.prototypeCountCall
			w.prototypeCountCall++
			if index < len(w.prototypeCounts) {
				return w.prototypeCounts[index]
			}
			return 0
		},
		loadTypeName: func(object *monsterGeneratorXferTestObject4F7130) []byte {
			w.typeNameCalls[object]++
			w.event(fmt.Sprintf("type-name:%s:%d", object.id, w.typeNameCalls[object]))
			return []byte(object.id)
		},
		rwNameLength: func(value uint8) uint8 {
			w.event(fmt.Sprintf("rw-name-length:%d", value))
			if !w.readBranch {
				if w.writeNameLength != nil {
					return *w.writeNameLength
				}
				return value
			}
			index := w.readNameCall
			w.readNameCall++
			if index < len(w.readNameLengths) {
				return w.readNameLengths[index]
			}
			return 0
		},
		rwNameBytes: func(value []byte) {
			w.event(fmt.Sprintf("rw-name-bytes:%d", len(value)))
			if w.readBranch {
				index := w.readBytesCall
				w.readBytesCall++
				if index < len(w.readNames) {
					copy(value, w.readNames[index])
				}
				return
			}
			w.writtenNames = append(w.writtenNames, bytes.Clone(value))
		},
		saveObject: func(object *monsterGeneratorXferTestObject4F7130) int32 {
			w.saved = append(w.saved, object)
			w.event("save:" + object.id)
			return w.saveResult
		},
		rwPrototypeTag: func(value uint16) uint16 {
			w.tagCalls++
			w.event(fmt.Sprintf("rw-tag:%d", value))
			return 0xbeef
		},
		readPrototypeCRC: func() {
			w.crcCalls++
			w.event("read-crc")
		},
		newObjectByTypeName: func(name []byte) *monsterGeneratorXferTestObject4F7130 {
			trimmed := name
			if index := bytes.IndexByte(trimmed, 0); index >= 0 {
				trimmed = trimmed[:index]
			}
			id := string(trimmed)
			w.event("new-object:" + id)
			return w.created[id]
		},
		callObjectXfer: func(object *monsterGeneratorXferTestObject4F7130) int32 {
			w.xferCalls = append(w.xferCalls, object)
			w.event("call-xfer:" + object.id)
			if value, ok := w.xferResults[object]; ok {
				return value
			}
			return 1
		},
		storePrototype: func(data *monsterGeneratorXferTestData4F7130, index int, object *monsterGeneratorXferTestObject4F7130) {
			w.stores = append(w.stores, monsterGeneratorXferTestStore4F7130{data: data, index: index, object: object})
			w.event(fmt.Sprintf("store-prototype:%d:%s", index, object.id))
			data.prototypes[index] = object
		},
		rwQuestSelectorCount: func(value uint8) uint8 {
			w.event(fmt.Sprintf("rw-quest-count:%d", value))
			return w.questCount
		},
		rwQuestSelector: func(data *monsterGeneratorXferTestData4F7130, index int) {
			w.useData(fmt.Sprintf("rw-quest:%d", index), data)
			data.quest[index]++
		},
		rwField92: func(data *monsterGeneratorXferTestData4F7130) {
			w.useData("rw-field92", data)
			data.field92++
		},
		transferInventory: func(version uint16, object *monsterGeneratorXferTestObject4F7130, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, monsterGeneratorXferTestInventoryCall4F7130{
				version: version, object: object, count: count,
			})
			w.event(fmt.Sprintf("inventory:%d:%d", version, count))
			return w.inventoryResult
		},
	}
}

func newMonsterGeneratorXferTestData4F7130(size int) *monsterGeneratorXferTestData4F7130 {
	return &monsterGeneratorXferTestData4F7130{
		spawn:      make([]uint8, 256),
		quest:      make([]uint8, 256),
		prototypes: make([]*monsterGeneratorXferTestObject4F7130, size),
	}
}

func TestMonsterGeneratorXfer4F7130WriteOrderAndEntryCaches(t *testing.T) {
	entry := newMonsterGeneratorXferTestData4F7130(12)
	replacement := newMonsterGeneratorXferTestData4F7130(12)
	prototypeA := &monsterGeneratorXferTestObject4F7130{id: "Imp"}
	prototypeB := &monsterGeneratorXferTestObject4F7130{id: "Ogre"}
	entry.prototypes[0] = prototypeA
	entry.prototypes[5] = prototypeB
	object := &monsterGeneratorXferTestObject4F7130{
		data: entry, field34: 0x11223344, scriptID: 17,
	}
	w := newMonsterGeneratorXferTestWorld4F7130()
	w.spawnCount = 2
	w.scriptResult = 0
	w.saveResult = 0
	w.modes = []int32{0, 0}
	w.after["map-read-write:63"] = func() {
		object.data = replacement
		object.scriptID = 99
		object.field34 = 7
	}

	if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.data != replacement || object.scriptID != 99 || object.field34 != 0x11223344 {
		t.Fatalf("live object state = data %p script %d Field34 %#x", object.data, object.scriptID, object.field34)
	}
	for index, data := range w.updatePointers {
		if data != entry {
			t.Fatalf("update pointer[%d] = %p, want cached entry %p", index, data, entry)
		}
	}
	wantScripts := []monsterGeneratorXferTestScriptCall4F7130{
		{data: entry, slot: monsterGeneratorScript48_4F7130, scriptData: 17, offset: 1920},
		{data: entry, slot: monsterGeneratorScript56_4F7130, scriptData: 17, offset: 2048},
		{data: entry, slot: monsterGeneratorScript72_4F7130, scriptData: 17, offset: 2176},
		{data: entry, slot: monsterGeneratorScript64_4F7130, scriptData: 17, offset: 2304},
	}
	if !reflect.DeepEqual(w.scriptCalls, wantScripts) {
		t.Fatalf("script calls = %+v, want %+v", w.scriptCalls, wantScripts)
	}
	if !reflect.DeepEqual(w.writtenNames, [][]byte{[]byte("Imp"), []byte("Ogre")}) {
		t.Fatalf("written names = %q", w.writtenNames)
	}
	if !reflect.DeepEqual(w.saved, []*monsterGeneratorXferTestObject4F7130{prototypeA, prototypeB}) {
		t.Fatalf("saved objects = %v", w.saved)
	}
	if w.typeNameCalls[prototypeA] != 2 || w.typeNameCalls[prototypeB] != 2 {
		t.Fatalf("type-name calls = A:%d B:%d, want 2/2", w.typeNameCalls[prototypeA], w.typeNameCalls[prototypeB])
	}
	if entry.spawn[0] != 1 || entry.spawn[1] != 1 || entry.spawn[2] != 0 ||
		entry.active != 1 || entry.max != 1 || entry.frame88 != 1 ||
		entry.quest[0] != 1 || entry.quest[1] != 1 || entry.quest[2] != 1 || entry.field92 != 1 {
		t.Fatalf("entry data transfer mismatch: %+v", entry)
	}
	if w.field34Stores != 1 || len(w.inventoryCalls) != 0 {
		t.Fatalf("Field34 stores/inventory = %d/%d, want 1/0", w.field34Stores, len(w.inventoryCalls))
	}

	wantPrefix := []string{
		"load-update-data", "load-field34", "load-script-data", "rw-version:63",
		"map-read-write:63", "rw-spawn-count:3", "rw-spawn:0", "rw-spawn:1",
		"rw-active", "rw-max", "rw-frame88", "script:0:1920", "script:1:2048",
		"script:2:2176", "script:3:2304", "read-mode:1=0", "rw-group-count:3",
	}
	if len(w.events) < len(wantPrefix) || !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("event prefix =\n%v\nwant\n%v", w.events, wantPrefix)
	}
	wantTail := []string{
		"rw-quest-count:3", "rw-quest:0", "rw-quest:1", "rw-quest:2",
		"rw-field92", "load-field34", "read-mode:2=0", "store-field34:287454020",
	}
	if !reflect.DeepEqual(w.events[len(w.events)-len(wantTail):], wantTail) {
		t.Fatalf("event tail =\n%v\nwant\n%v", w.events[len(w.events)-len(wantTail):], wantTail)
	}
}

func TestMonsterGeneratorXfer4F7130SignedVersionsAndSuffixes(t *testing.T) {
	tests := []struct {
		name        string
		version     uint16
		mapResult   int32
		wantResult  int32
		wantMap     []int32
		wantQuest   int
		wantField92 int
		wantRestore int
	}{
		{name: "version 1", version: 1, mapResult: 1, wantResult: 1, wantMap: []int32{1}, wantRestore: 1},
		{name: "version 61", version: 61, mapResult: -1, wantResult: 1, wantMap: []int32{61}, wantRestore: 1},
		{name: "version 62", version: 62, mapResult: 1, wantResult: 1, wantMap: []int32{62}, wantQuest: 3, wantRestore: 1},
		{name: "version 63", version: 63, mapResult: 1, wantResult: 1, wantMap: []int32{63}, wantQuest: 3, wantField92: 1, wantRestore: 1},
		{name: "zero", version: 0, mapResult: 1},
		{name: "too new", version: 64, mapResult: 1},
		{name: "negative minimum", version: 0x8000, mapResult: 1},
		{name: "minus one", version: 0xffff, mapResult: 1},
		{name: "common failure", version: 63, mapResult: 0, wantMap: []int32{63}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := newMonsterGeneratorXferTestData4F7130(12)
			object := &monsterGeneratorXferTestObject4F7130{data: data, field34: 5}
			w := newMonsterGeneratorXferTestWorld4F7130()
			w.version = test.version
			w.mapResult = test.mapResult
			w.modes = []int32{0, 0}

			if got := monsterGeneratorXfer4F7130(object, w.deps()); got != test.wantResult {
				t.Fatalf("result = %d, want %d", got, test.wantResult)
			}
			if !reflect.DeepEqual(w.mapVersions, test.wantMap) {
				t.Fatalf("map versions = %v, want %v", w.mapVersions, test.wantMap)
			}
			questTransfers := 0
			for _, value := range data.quest {
				questTransfers += int(value)
			}
			if questTransfers != test.wantQuest || int(data.field92) != test.wantField92 || w.field34Stores != test.wantRestore {
				t.Fatalf("quest/Field92/restores = %d/%d/%d, want %d/%d/%d",
					questTransfers, data.field92, w.field34Stores,
					test.wantQuest, test.wantField92, test.wantRestore)
			}
		})
	}
}

func TestMonsterGeneratorXfer4F7130ReadBranchAndInventory(t *testing.T) {
	data := newMonsterGeneratorXferTestData4F7130(12)
	object := &monsterGeneratorXferTestObject4F7130{data: data, field34: 9, scriptID: 23}
	imp := &monsterGeneratorXferTestObject4F7130{id: "Imp"}
	bat := &monsterGeneratorXferTestObject4F7130{id: "Bat"}
	ogre := &monsterGeneratorXferTestObject4F7130{id: "Ogre!"}
	w := newMonsterGeneratorXferTestWorld4F7130()
	w.modes = []int32{2, 1}
	w.groupCount = 2
	w.prototypeCounts = []uint8{2, 1}
	w.readNameLengths = []uint8{3, 3, 5}
	w.readNames = [][]byte{[]byte("Imp"), []byte("Bat"), []byte("Ogre!")}
	w.created["Imp"] = imp
	w.created["Bat"] = bat
	w.created["Ogre!"] = ogre

	if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantStores := []monsterGeneratorXferTestStore4F7130{
		{data: data, index: 0, object: imp},
		{data: data, index: 1, object: bat},
		{data: data, index: 4, object: ogre},
	}
	if !reflect.DeepEqual(w.stores, wantStores) {
		t.Fatalf("stores = %+v, want %+v", w.stores, wantStores)
	}
	if w.tagCalls != 3 || w.crcCalls != 3 || !reflect.DeepEqual(w.xferCalls, []*monsterGeneratorXferTestObject4F7130{imp, bat, ogre}) {
		t.Fatalf("tag/CRC/xfer = %d/%d/%v", w.tagCalls, w.crcCalls, w.xferCalls)
	}
	wantInventory := []monsterGeneratorXferTestInventoryCall4F7130{{
		version: 63, object: object, count: 9,
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) || object.field34 != 9 || w.field34Stores != 1 {
		t.Fatalf("inventory/Field34/stores = %+v/%d/%d", w.inventoryCalls, object.field34, w.field34Stores)
	}
	for _, index := range []int{0, 1, 4} {
		if data.prototypes[index] == nil {
			t.Fatalf("prototype[%d] was not stored", index)
		}
	}
}

func TestMonsterGeneratorXfer4F7130DoesNotClampTransferredCounts(t *testing.T) {
	t.Run("selector counts", func(t *testing.T) {
		data := newMonsterGeneratorXferTestData4F7130(12)
		object := &monsterGeneratorXferTestObject4F7130{data: data}
		w := newMonsterGeneratorXferTestWorld4F7130()
		w.spawnCount = 5
		w.questCount = 4
		w.modes = []int32{0}

		if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		for index := 0; index < 5; index++ {
			if data.spawn[index] != 1 {
				t.Fatalf("spawn[%d] = %d, want transferred", index, data.spawn[index])
			}
		}
		for index := 0; index < 4; index++ {
			if data.quest[index] != 1 {
				t.Fatalf("quest[%d] = %d, want transferred", index, data.quest[index])
			}
		}
	})

	t.Run("outer group count", func(t *testing.T) {
		data := newMonsterGeneratorXferTestData4F7130(16)
		object := &monsterGeneratorXferTestObject4F7130{data: data}
		created := &monsterGeneratorXferTestObject4F7130{id: "Imp"}
		w := newMonsterGeneratorXferTestWorld4F7130()
		w.modes = []int32{1}
		w.groupCount = 4
		w.prototypeCounts = []uint8{0, 0, 0, 1}
		w.readNameLengths = []uint8{3}
		w.readNames = [][]byte{[]byte("Imp")}
		w.created["Imp"] = created

		if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.stores) != 1 || w.stores[0].index != 12 {
			t.Fatalf("stores = %+v, want unclamped index 12", w.stores)
		}
	})

	t.Run("inner prototype count", func(t *testing.T) {
		data := newMonsterGeneratorXferTestData4F7130(8)
		object := &monsterGeneratorXferTestObject4F7130{data: data}
		w := newMonsterGeneratorXferTestWorld4F7130()
		w.modes = []int32{1}
		w.groupCount = 1
		w.prototypeCounts = []uint8{5}
		for index := 0; index < 5; index++ {
			id := fmt.Sprintf("X%d", index)
			w.readNameLengths = append(w.readNameLengths, 2)
			w.readNames = append(w.readNames, []byte(id))
			w.created[id] = &monsterGeneratorXferTestObject4F7130{id: id}
		}

		if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if len(w.stores) != 5 || w.stores[4].index != 4 {
			t.Fatalf("stores = %+v, want unclamped inner index 4", w.stores)
		}
	})
}

func TestMonsterGeneratorXfer4F7130FailureAndRestoreBoundaries(t *testing.T) {
	t.Run("allocation failure", func(t *testing.T) {
		data := newMonsterGeneratorXferTestData4F7130(12)
		object := &monsterGeneratorXferTestObject4F7130{data: data, field34: 7}
		w := newMonsterGeneratorXferTestWorld4F7130()
		w.modes = []int32{1}
		w.groupCount = 1
		w.prototypeCounts = []uint8{1}
		w.readNameLengths = []uint8{3}
		w.readNames = [][]byte{[]byte("Bad")}
		w.after["new-object:Bad"] = func() { object.field34 = 11 }

		if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 11 || w.field34Stores != 0 || w.tagCalls != 0 || w.crcCalls != 0 || len(w.stores) != 0 {
			t.Fatalf("state/stores/tag/CRC/prototypes = %d/%d/%d/%d/%d",
				object.field34, w.field34Stores, w.tagCalls, w.crcCalls, len(w.stores))
		}
	})

	t.Run("created object xfer failure", func(t *testing.T) {
		data := newMonsterGeneratorXferTestData4F7130(12)
		object := &monsterGeneratorXferTestObject4F7130{data: data, field34: 7}
		created := &monsterGeneratorXferTestObject4F7130{id: "Imp"}
		w := newMonsterGeneratorXferTestWorld4F7130()
		w.modes = []int32{1}
		w.groupCount = 1
		w.prototypeCounts = []uint8{1}
		w.readNameLengths = []uint8{3}
		w.readNames = [][]byte{[]byte("Imp")}
		w.created["Imp"] = created
		w.xferResults[created] = 0
		w.after["call-xfer:Imp"] = func() { object.field34 = 13 }

		if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 13 || w.field34Stores != 0 || w.tagCalls != 1 || w.crcCalls != 1 || len(w.stores) != 0 {
			t.Fatalf("state/stores/tag/CRC/prototypes = %d/%d/%d/%d/%d",
				object.field34, w.field34Stores, w.tagCalls, w.crcCalls, len(w.stores))
		}
	})

	t.Run("inventory failure", func(t *testing.T) {
		data := newMonsterGeneratorXferTestData4F7130(12)
		object := &monsterGeneratorXferTestObject4F7130{data: data, field34: 0x80000003}
		w := newMonsterGeneratorXferTestWorld4F7130()
		w.modes = []int32{0, 1}
		w.inventoryResult = 0
		w.after["inventory:63:-2147483645"] = func() { object.field34 = 17 }

		if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 17 || w.field34Stores != 0 {
			t.Fatalf("Field34/stores = %d/%d, want failure mutation/0", object.field34, w.field34Stores)
		}
		want := []monsterGeneratorXferTestInventoryCall4F7130{{
			version: 63, object: object, count: -2147483645,
		}}
		if !reflect.DeepEqual(w.inventoryCalls, want) {
			t.Fatalf("inventory calls = %+v, want %+v", w.inventoryCalls, want)
		}
	})

	t.Run("non-exact suffix mode skips inventory", func(t *testing.T) {
		for _, mode := range []int32{-1, 0, 2} {
			data := newMonsterGeneratorXferTestData4F7130(12)
			object := &monsterGeneratorXferTestObject4F7130{data: data, field34: 4}
			w := newMonsterGeneratorXferTestWorld4F7130()
			w.modes = []int32{0, mode}
			if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 1 || len(w.inventoryCalls) != 0 || object.field34 != 4 {
				t.Fatalf("mode %d result/inventory/Field34 = %d/%d/%d", mode, got, len(w.inventoryCalls), object.field34)
			}
		}
	})

	t.Run("zero live Field34 skips second mode read", func(t *testing.T) {
		data := newMonsterGeneratorXferTestData4F7130(12)
		object := &monsterGeneratorXferTestObject4F7130{data: data, field34: 7}
		w := newMonsterGeneratorXferTestWorld4F7130()
		w.modes = []int32{0, 1}
		w.after["rw-field92"] = func() { object.field34 = 0 }
		if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 1 || w.modeCalls != 1 || object.field34 != 7 {
			t.Fatalf("result/mode calls/Field34 = %d/%d/%d", got, w.modeCalls, object.field34)
		}
	})
}

func TestMonsterGeneratorXfer4F7130WriteLengthUsesTwoNameLoads(t *testing.T) {
	data := newMonsterGeneratorXferTestData4F7130(12)
	prototype := &monsterGeneratorXferTestObject4F7130{id: string(bytes.Repeat([]byte{'x'}, 260))}
	data.prototypes[0] = prototype
	object := &monsterGeneratorXferTestObject4F7130{data: data}
	w := newMonsterGeneratorXferTestWorld4F7130()
	w.modes = []int32{0}

	if got := monsterGeneratorXfer4F7130(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantLength := uint8(len(prototype.id))
	if w.typeNameCalls[prototype] != 2 || len(w.writtenNames) != 1 || len(w.writtenNames[0]) != int(wantLength) {
		t.Fatalf("name calls/wire length = %d/%d, want 2/%d", w.typeNameCalls[prototype], len(w.writtenNames[0]), wantLength)
	}
}
