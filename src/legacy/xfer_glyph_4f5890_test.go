package legacy

import (
	"fmt"
	"reflect"
	"testing"
)

type glyphXferTestName4F5890 string

type glyphXferTestData4F5890 struct {
	count       uint8
	spells      map[int]uint32
	targetAlive bool
}

type glyphXferTestObject4F5890 struct {
	data       *glyphXferTestData4F5890
	field34    uint32
	direction1 uint16
	direction2 uint16
}

type glyphXferTestInventoryCall4F5890 struct {
	version uint16
	object  *glyphXferTestObject4F5890
	count   int32
}

type glyphXferTestWorld4F5890 struct {
	version         uint16
	mapResult       int32
	readOnlyValues  []int32
	inventoryResult int32
	wireLengths     []uint8
	wireNames       [][]byte
	parsed          map[string]uint32
	names           map[uint32]glyphXferTestName4F5890

	dataLoads       int
	field34Loads    int
	mapVersions     []int32
	readOnlyCalls   int
	legacyDwords    int
	legacySpells    int
	countLoads      int
	spellLoads      []int
	spellNameIDs    []uint32
	nameLengthInput []uint8
	readNames       []string
	writtenNames    []string
	storedSpells    []uint32
	directionCopies int
	targetClears    int
	inventoryCalls  []glyphXferTestInventoryCall4F5890
	field34Stores   int
	events          []string
	after           map[string]func()
}

func newGlyphXferTestWorld4F5890() *glyphXferTestWorld4F5890 {
	return &glyphXferTestWorld4F5890{
		version:         glyphXferCurrentVersion4F5890,
		mapResult:       1,
		readOnlyValues:  []int32{0, 0},
		inventoryResult: 1,
		parsed:          make(map[string]uint32),
		names:           make(map[uint32]glyphXferTestName4F5890),
		after:           make(map[string]func()),
	}
}

func (w *glyphXferTestWorld4F5890) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *glyphXferTestWorld4F5890) deps() glyphXferDeps4F5890[
	*glyphXferTestObject4F5890,
	*glyphXferTestData4F5890,
	glyphXferTestName4F5890,
] {
	return glyphXferDeps4F5890[
		*glyphXferTestObject4F5890,
		*glyphXferTestData4F5890,
		glyphXferTestName4F5890,
	]{
		loadGlyphData: func(object *glyphXferTestObject4F5890) *glyphXferTestData4F5890 {
			w.dataLoads++
			value := object.data
			w.event("load-glyph-data")
			return value
		},
		loadField34: func(object *glyphXferTestObject4F5890) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *glyphXferTestObject4F5890, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event("map-read-write")
			return w.mapResult
		},
		rwLegacyDword: func() {
			w.legacyDwords++
			w.event("rw-legacy-dword")
		},
		rwDirection1: func(*glyphXferTestObject4F5890) { w.event("rw-direction1") },
		rwTargetX:    func(*glyphXferTestData4F5890) { w.event("rw-target-x") },
		rwTargetY:    func(*glyphXferTestData4F5890) { w.event("rw-target-y") },
		rwSpellCount: func(*glyphXferTestData4F5890) { w.event("rw-spell-count") },
		readOnly: func() int32 {
			call := w.readOnlyCalls
			w.readOnlyCalls++
			w.event(fmt.Sprintf("read-only:%d", call+1))
			if call >= len(w.readOnlyValues) {
				return 0
			}
			return w.readOnlyValues[call]
		},
		loadSpellCount: func(data *glyphXferTestData4F5890) uint8 {
			w.countLoads++
			value := data.count
			w.event("load-spell-count")
			return value
		},
		rwLegacySpells: func(*glyphXferTestData4F5890) {
			w.legacySpells++
			w.event("rw-legacy-spells")
		},
		rwNameLength: func(value uint8) uint8 {
			w.nameLengthInput = append(w.nameLengthInput, value)
			w.event("rw-name-length")
			if len(w.wireLengths) == 0 {
				return value
			}
			out := w.wireLengths[0]
			w.wireLengths = w.wireLengths[1:]
			return out
		},
		rwNameBytes: func(value []byte) {
			if len(w.wireNames) != 0 {
				copy(value, w.wireNames[0])
				w.wireNames = w.wireNames[1:]
			}
			w.readNames = append(w.readNames, string(value))
			w.event("rw-read-name")
		},
		spellID: func(name string) uint32 {
			w.event("spell-id:" + name)
			return w.parsed[name]
		},
		storeSpell: func(data *glyphXferTestData4F5890, index int, value uint32) {
			if data.spells == nil {
				data.spells = make(map[int]uint32)
			}
			data.spells[index] = value
			w.storedSpells = append(w.storedSpells, value)
			w.event(fmt.Sprintf("store-spell:%d", index))
		},
		loadSpell: func(data *glyphXferTestData4F5890, index int) uint32 {
			w.spellLoads = append(w.spellLoads, index)
			value := data.spells[index]
			w.event(fmt.Sprintf("load-spell:%d", index))
			return value
		},
		spellName: func(value uint32) glyphXferTestName4F5890 {
			w.spellNameIDs = append(w.spellNameIDs, value)
			name := w.names[value]
			w.event(fmt.Sprintf("spell-name:%d", value))
			return name
		},
		spellNameLength: func(name glyphXferTestName4F5890) uint8 {
			w.event("spell-name-length:" + string(name))
			return uint8(len(name))
		},
		rwSpellNameBytes: func(name glyphXferTestName4F5890, length uint8) {
			w.writtenNames = append(w.writtenNames, string(name)[:int(length)])
			w.event("rw-write-name:" + string(name))
		},
		copyDirection: func(object *glyphXferTestObject4F5890) {
			w.directionCopies++
			object.direction2 = object.direction1
			w.event("copy-direction")
		},
		clearSpellTargetObject: func(data *glyphXferTestData4F5890) {
			w.targetClears++
			data.targetAlive = false
			w.event("clear-spell-target")
		},
		transferInventory: func(version uint16, object *glyphXferTestObject4F5890, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, glyphXferTestInventoryCall4F5890{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *glyphXferTestObject4F5890, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestGlyphXfer4F5890WriteReloadsCountSpellAndName(t *testing.T) {
	entryData := &glyphXferTestData4F5890{
		count: 2, spells: map[int]uint32{0: 10, 1: 20}, targetAlive: true,
	}
	liveData := &glyphXferTestData4F5890{count: 5, spells: map[int]uint32{0: 99}}
	object := &glyphXferTestObject4F5890{
		data: entryData, field34: 0xa1b2c3d4, direction1: 0xabcd, direction2: 0x1234,
	}
	w := newGlyphXferTestWorld4F5890()
	w.names[10] = "FIRST"
	w.names[11] = "LIVE_NAME"
	w.readOnlyValues = []int32{0, 0, 1}
	w.after["load-glyph-data"] = func() { object.data = liveData }
	w.after["rw-name-length"] = func() { entryData.spells[0] = 11 }
	w.after["rw-write-name:LIVE_NAME"] = func() {
		entryData.count = 1
		object.field34 = 0x80000003
	}

	if got := glyphXfer4F5890(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.data != liveData || object.field34 != 0xa1b2c3d4 {
		t.Fatalf("object data/Field34 = %p/%#x, want live/%#x", object.data, object.field34, uint32(0xa1b2c3d4))
	}
	if !reflect.DeepEqual(w.spellLoads, []int{0, 0}) {
		t.Fatalf("spell loads = %v, want index zero twice", w.spellLoads)
	}
	if !reflect.DeepEqual(w.spellNameIDs, []uint32{10, 11}) {
		t.Fatalf("spell name IDs = %v, want entry then reloaded", w.spellNameIDs)
	}
	if !reflect.DeepEqual(w.writtenNames, []string{"LIVE_"}) {
		t.Fatalf("written names = %v, want reloaded name", w.writtenNames)
	}
	if w.countLoads != 2 {
		t.Fatalf("spell count loads = %d, want initial and post-item reload", w.countLoads)
	}
	if w.directionCopies != 0 || w.targetClears != 0 {
		t.Fatalf("write post effects = %d/%d, want none", w.directionCopies, w.targetClears)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []glyphXferTestInventoryCall4F5890{{
		version: glyphXferCurrentVersion4F5890, object: object, count: -2147483645,
	}}) {
		t.Fatalf("inventory calls = %#v, want live signed count", w.inventoryCalls)
	}
}

func TestGlyphXfer4F5890ModernReadReloadsCountAndAppliesPostState(t *testing.T) {
	data := &glyphXferTestData4F5890{count: 2, spells: make(map[int]uint32), targetAlive: true}
	object := &glyphXferTestObject4F5890{
		data: data, field34: 7, direction1: 0xabcd, direction2: 0x1234,
	}
	w := newGlyphXferTestWorld4F5890()
	w.readOnlyValues = []int32{1, 1, 1}
	w.wireLengths = []uint8{7, 6}
	w.wireNames = [][]byte{{'F', 'I', 'R', 'S', 'T', 0, 'X'}, []byte("SECOND")}
	w.parsed["FIRST"] = 101
	w.parsed["SECOND"] = 202
	w.after["store-spell:0"] = func() { data.count = 1 }

	if got := glyphXfer4F5890(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.storedSpells, []uint32{101}) || data.spells[0] != 101 {
		t.Fatalf("stored spells = %v/%v, want one reloaded-count item", w.storedSpells, data.spells)
	}
	if w.countLoads != 2 || w.readOnlyCalls != 3 {
		t.Fatalf("count/mode loads = %d/%d, want 2/3", w.countLoads, w.readOnlyCalls)
	}
	if object.direction2 != 0xabcd || data.targetAlive {
		t.Fatalf("post state = direction %#x target %v, want 0xabcd/false", object.direction2, data.targetAlive)
	}
	if len(w.inventoryCalls) != 1 || w.inventoryCalls[0].count != 7 {
		t.Fatalf("inventory calls = %#v, want count 7", w.inventoryCalls)
	}
}

func TestGlyphXfer4F5890ModernReadZeroCountSkipsSecondModeLoad(t *testing.T) {
	data := &glyphXferTestData4F5890{targetAlive: true}
	object := &glyphXferTestObject4F5890{
		data: data, field34: 9, direction1: 0x12ef, direction2: 0xabcd,
	}
	w := newGlyphXferTestWorld4F5890()
	// The second value must gate inventory directly. A post-loop mode read
	// here would consume zero and suppress both the copy and inventory call.
	w.readOnlyValues = []int32{1, 1}

	if got := glyphXfer4F5890(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.readOnlyCalls != 2 || w.directionCopies != 1 || w.targetClears != 1 || len(w.inventoryCalls) != 1 {
		t.Fatalf("mode/copy/clear/inventory = %d/%d/%d/%d, want 2/1/1/1",
			w.readOnlyCalls, w.directionCopies, w.targetClears, len(w.inventoryCalls))
	}
}

func TestGlyphXfer4F5890LegacyAndSignedVersionThresholds(t *testing.T) {
	t.Run("legacy raw spell read", func(t *testing.T) {
		data := &glyphXferTestData4F5890{count: 5, targetAlive: true}
		object := &glyphXferTestObject4F5890{data: data, direction1: 0xbeef}
		w := newGlyphXferTestWorld4F5890()
		w.version = 30
		w.readOnlyValues = []int32{1, 1}

		if got := glyphXfer4F5890(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.legacyDwords != 1 || w.legacySpells != 1 || w.countLoads != 0 {
			t.Fatalf("legacy dword/spells/count loads = %d/%d/%d, want 1/1/0",
				w.legacyDwords, w.legacySpells, w.countLoads)
		}
		if object.direction2 != 0xbeef || data.targetAlive {
			t.Fatalf("legacy post state = %#x/%v, want 0xbeef/false", object.direction2, data.targetAlive)
		}
	})

	for _, version := range []uint16{0x8000, 0xffff} {
		t.Run(fmt.Sprintf("signed-%#x", version), func(t *testing.T) {
			object := &glyphXferTestObject4F5890{data: &glyphXferTestData4F5890{}}
			w := newGlyphXferTestWorld4F5890()
			w.version = version
			if got := glyphXfer4F5890(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.mapVersions, []int32{int32(int16(version))}) {
				t.Fatalf("map versions = %v, want signed %d", w.mapVersions, int16(version))
			}
			if w.legacyDwords != 1 || w.legacySpells != 0 {
				t.Fatalf("legacy dword/spells = %d/%d, want write-side 1/0", w.legacyDwords, w.legacySpells)
			}
		})
	}

	t.Run("newer version stops after prefix", func(t *testing.T) {
		object := &glyphXferTestObject4F5890{data: &glyphXferTestData4F5890{}, field34: 0x11223344}
		w := newGlyphXferTestWorld4F5890()
		w.version = 61
		if got := glyphXfer4F5890(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if w.dataLoads != 1 || w.field34Loads != 1 || len(w.mapVersions) != 0 || w.field34Stores != 0 {
			t.Fatalf("prefix loads/map/store = %d/%d/%d/%d, want 1/1/0/0",
				w.dataLoads, w.field34Loads, len(w.mapVersions), w.field34Stores)
		}
	})
}

func TestGlyphXfer4F5890ExactOneModeAndInventoryFailurePrefix(t *testing.T) {
	t.Run("mode two uses write path but later one applies read post state", func(t *testing.T) {
		data := &glyphXferTestData4F5890{
			count: 1, spells: map[int]uint32{0: 10}, targetAlive: true,
		}
		object := &glyphXferTestObject4F5890{data: data, direction1: 0x4567}
		w := newGlyphXferTestWorld4F5890()
		w.names[10] = "SPELL"
		w.readOnlyValues = []int32{2, 1}

		if got := glyphXfer4F5890(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.writtenNames, []string{"SPELL"}) || len(w.readNames) != 0 {
			t.Fatalf("written/read names = %v/%v, want write path only", w.writtenNames, w.readNames)
		}
		if object.direction2 != 0x4567 || data.targetAlive {
			t.Fatalf("post state = %#x/%v, want later exact-one copy/clear", object.direction2, data.targetAlive)
		}
	})

	t.Run("inventory failure leaves live Field34", func(t *testing.T) {
		const live = uint32(0x80000004)
		object := &glyphXferTestObject4F5890{
			data: &glyphXferTestData4F5890{}, field34: 0x10203040,
		}
		w := newGlyphXferTestWorld4F5890()
		w.readOnlyValues = []int32{0, 0, 1}
		w.inventoryResult = 0
		w.after["read-only:2"] = func() { object.field34 = live }

		if got := glyphXfer4F5890(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != live || w.field34Stores != 0 {
			t.Fatalf("Field34/store = %#x/%d, want live %#x/no restore", object.field34, w.field34Stores, live)
		}
		if !reflect.DeepEqual(w.inventoryCalls, []glyphXferTestInventoryCall4F5890{{
			version: glyphXferCurrentVersion4F5890, object: object, count: -2147483644,
		}}) {
			t.Fatalf("inventory calls = %#v, want zero-extended version and signed live count", w.inventoryCalls)
		}
	})
}

func TestGlyphXfer4F5890MapFailureKeepsEntryCachesUnrestored(t *testing.T) {
	entry := &glyphXferTestData4F5890{}
	object := &glyphXferTestObject4F5890{data: entry, field34: 0x11223344}
	w := newGlyphXferTestWorld4F5890()
	w.mapResult = 0
	w.after["load-glyph-data"] = func() {
		object.data = &glyphXferTestData4F5890{count: 9}
		object.field34 = 0x55667788
	}

	if got := glyphXfer4F5890(object, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if w.dataLoads != 1 || w.field34Loads != 1 || w.field34Stores != 0 {
		t.Fatalf("data/Field34 loads/stores = %d/%d/%d, want 1/1/0", w.dataLoads, w.field34Loads, w.field34Stores)
	}
	if object.field34 != 0x55667788 {
		t.Fatalf("Field34 = %#x, want callback-mutated value", object.field34)
	}
}
