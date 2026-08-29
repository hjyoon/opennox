package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type triggerXferTestUpdate4F4E50 struct {
	flags        uint32
	state        uint8
	field9       uint8
	classInclude uint32
	classExclude uint32
	teamInclude  uint8
	teamExclude  uint8
	colors       [6]uint8
}

type triggerXferTestObject4F4E50 struct {
	field33    uint32
	field34    uint32
	update     *triggerXferTestUpdate4F4E50
	width      float32
	height     float32
	scriptData uintptr
}

type triggerXferMapCall4F4E50 struct {
	object  *triggerXferTestObject4F4E50
	version int32
}

type triggerXferScriptCall4F4E50 struct {
	update   *triggerXferTestUpdate4F4E50
	callback triggerXferCallback4F4E50
	script   uintptr
	offset   uintptr
}

type triggerXferInventoryCall4F4E50 struct {
	version uint16
	object  *triggerXferTestObject4F4E50
	count   int32
}

type triggerXferTestWorld4F4E50 struct {
	version             uint16
	mapResult           int32
	readOnlyDefault     int32
	readOnlyValues      []int32
	widthResult         int32
	heightResult        int32
	legacyScratchResult []uint32
	legacyCountResult   []uint8
	field33Result       uint32
	inventoryResult     int32

	field34Loads       int
	updateLoads        int
	versionTransfers   []uint16
	mapCalls           []triggerXferMapCall4F4E50
	readOnlyCalls      int
	widthLoads         []float32
	heightLoads        []float32
	truncInputs        []float32
	widthTransfers     []int32
	heightTransfers    []int32
	legacyScratchInput []uint32
	legacyCountInput   []uint8
	seekOffsets        []int32
	scriptDataLoads    int
	scriptCalls        []triggerXferScriptCall4F4E50
	legacyScriptCalls  []triggerXferCallback4F4E50
	field33Loads       int
	markedFrames       []uint32
	inventoryCalls     []triggerXferInventoryCall4F4E50
	field34Stores      int
	events             []string
	after              map[string]func()
	faultAt            int
}

func newTriggerXferTestWorld4F4E50() *triggerXferTestWorld4F4E50 {
	return &triggerXferTestWorld4F4E50{
		version:         triggerXferCurrentVersion4F4E50,
		mapResult:       1,
		readOnlyDefault: 1,
		widthResult:     70,
		heightResult:    -5,
		field33Result:   77,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func (w *triggerXferTestWorld4F4E50) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *triggerXferTestWorld4F4E50) deps() triggerXferDeps4F4E50[
	*triggerXferTestObject4F4E50,
	*triggerXferTestUpdate4F4E50,
	uintptr,
] {
	return triggerXferDeps4F4E50[
		*triggerXferTestObject4F4E50,
		*triggerXferTestUpdate4F4E50,
		uintptr,
	]{
		loadField34: func(object *triggerXferTestObject4F4E50) uint32 {
			w.field34Loads++
			value := object.field34
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return value
		},
		loadUpdateData: func(object *triggerXferTestObject4F4E50) *triggerXferTestUpdate4F4E50 {
			w.updateLoads++
			value := object.update
			w.event("load-update")
			return value
		},
		rwVersion: func(value uint16) uint16 {
			w.versionTransfers = append(w.versionTransfers, value)
			w.event("rw-version")
			return w.version
		},
		mapReadWrite: func(object *triggerXferTestObject4F4E50, version int32) int32 {
			w.mapCalls = append(w.mapCalls, triggerXferMapCall4F4E50{object: object, version: version})
			w.event("map-read-write")
			return w.mapResult
		},
		readOnly: func() int32 {
			w.readOnlyCalls++
			value := w.readOnlyDefault
			if w.readOnlyCalls <= len(w.readOnlyValues) {
				value = w.readOnlyValues[w.readOnlyCalls-1]
			}
			w.event(fmt.Sprintf("read-only:%d", w.readOnlyCalls))
			return value
		},
		loadBoxWidth: func(object *triggerXferTestObject4F4E50) float32 {
			value := object.width
			w.widthLoads = append(w.widthLoads, value)
			w.event("load-width")
			return value
		},
		loadBoxHeight: func(object *triggerXferTestObject4F4E50) float32 {
			value := object.height
			w.heightLoads = append(w.heightLoads, value)
			w.event("load-height")
			return value
		},
		truncFloatDword: func(value float32) int32 {
			w.truncInputs = append(w.truncInputs, value)
			w.event(fmt.Sprintf("trunc:%d", len(w.truncInputs)))
			return triggerXferTruncFloatDword4F4E50(value)
		},
		rwBoxWidth: func(value int32) int32 {
			w.widthTransfers = append(w.widthTransfers, value)
			w.event("rw-width")
			return w.widthResult
		},
		rwBoxHeight: func(value int32) int32 {
			w.heightTransfers = append(w.heightTransfers, value)
			w.event("rw-height")
			return w.heightResult
		},
		storeBoxWidth: func(object *triggerXferTestObject4F4E50, value float32) {
			w.event(fmt.Sprintf("store-width:%g", value))
			object.width = value
		},
		storeBoxHeight: func(object *triggerXferTestObject4F4E50, value float32) {
			w.event(fmt.Sprintf("store-height:%g", value))
			object.height = value
		},
		calcBox: func(_ *triggerXferTestObject4F4E50) {
			w.event("calc-box")
		},
		rwLegacyScratch3: func(value uint32) uint32 {
			index := len(w.legacyScratchInput)
			w.legacyScratchInput = append(w.legacyScratchInput, value)
			w.event(fmt.Sprintf("rw-legacy-scratch:%d", index))
			if index < len(w.legacyScratchResult) {
				return w.legacyScratchResult[index]
			}
			return value
		},
		rwColor: func(update *triggerXferTestUpdate4F4E50, index int) {
			update.colors[index] = uint8(0xa0 + index)
			w.event(fmt.Sprintf("rw-color:%d", index))
		},
		rwFlags: func(update *triggerXferTestUpdate4F4E50) {
			update.flags = 0x11223344
			w.event("rw-flags")
		},
		loadScriptData: func(object *triggerXferTestObject4F4E50) uintptr {
			w.scriptDataLoads++
			value := object.scriptData
			w.event("load-script-data")
			return value
		},
		transferScript: func(update *triggerXferTestUpdate4F4E50, callback triggerXferCallback4F4E50, script uintptr, offset uintptr) {
			w.scriptCalls = append(w.scriptCalls, triggerXferScriptCall4F4E50{
				update: update, callback: callback, script: script, offset: offset,
			})
			w.event(fmt.Sprintf("transfer-script:%d", callback))
		},
		initLegacyScript: func(_ *triggerXferTestUpdate4F4E50, callback triggerXferCallback4F4E50) {
			w.legacyScriptCalls = append(w.legacyScriptCalls, callback)
			w.event(fmt.Sprintf("init-legacy-script:%d", callback))
		},
		rwLegacyCount: func(value uint8) uint8 {
			index := len(w.legacyCountInput)
			w.legacyCountInput = append(w.legacyCountInput, value)
			w.event(fmt.Sprintf("rw-legacy-count:%d", index))
			if index < len(w.legacyCountResult) {
				return w.legacyCountResult[index]
			}
			return value
		},
		seekCurrent: func(offset int32) {
			w.seekOffsets = append(w.seekOffsets, offset)
			w.event("seek-current")
		},
		rwClassInclude: func(update *triggerXferTestUpdate4F4E50) {
			update.classInclude = 0x55667788
			w.event("rw-class-include")
		},
		rwClassExclude: func(update *triggerXferTestUpdate4F4E50) {
			update.classExclude = 0x99aabbcc
			w.event("rw-class-exclude")
		},
		storeTeamInclude: func(update *triggerXferTestUpdate4F4E50, value uint8) {
			update.teamInclude = value
			w.event("store-team-include")
		},
		storeTeamExclude: func(update *triggerXferTestUpdate4F4E50, value uint8) {
			update.teamExclude = value
			w.event("store-team-exclude")
		},
		rwTeamInclude: func(update *triggerXferTestUpdate4F4E50) {
			update.teamInclude = 0x5a
			w.event("rw-team-include")
		},
		rwTeamExclude: func(update *triggerXferTestUpdate4F4E50) {
			update.teamExclude = 0xa5
			w.event("rw-team-exclude")
		},
		rwState: func(update *triggerXferTestUpdate4F4E50) {
			update.state = 3
			w.event("rw-state")
		},
		rwField9: func(update *triggerXferTestUpdate4F4E50) {
			update.field9 = 4
			w.event("rw-field9")
		},
		rwField33: func(object *triggerXferTestObject4F4E50) {
			object.field33 = w.field33Result
			w.event("rw-field33")
		},
		loadField33: func(object *triggerXferTestObject4F4E50) uint32 {
			w.field33Loads++
			value := object.field33
			w.event("load-field33")
			return value
		},
		markAnimationFrame: func(_ *triggerXferTestObject4F4E50, frame uint32) {
			w.markedFrames = append(w.markedFrames, frame)
			w.event("mark-animation")
		},
		transferInventory: func(version uint16, object *triggerXferTestObject4F4E50, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, triggerXferInventoryCall4F4E50{
				version: version, object: object, count: count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *triggerXferTestObject4F4E50, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestTriggerXfer4F4E50PreservesEntryCachesAndReadOrder(t *testing.T) {
	entryScriptBits := uint64(0x100000000)
	liveScriptBits := uint64(0x200000000)
	if ^uintptr(0) == uintptr(^uint32(0)) {
		entryScriptBits = 0x10000000
		liveScriptBits = 0x20000000
	}
	entryScript := uintptr(entryScriptBits)
	liveScript := uintptr(liveScriptBits)
	entryUpdate := &triggerXferTestUpdate4F4E50{teamInclude: 7, teamExclude: 8}
	liveUpdate := &triggerXferTestUpdate4F4E50{flags: 9, teamInclude: 10, teamExclude: 11}
	object := &triggerXferTestObject4F4E50{
		field33: 1, field34: 0x11223344, update: entryUpdate,
		width: 9, height: 10, scriptData: entryScript,
	}
	w := newTriggerXferTestWorld4F4E50()
	w.after["load-field34:1"] = func() { object.field34 = 0x55667788 }
	w.after["load-update"] = func() { object.update = liveUpdate }
	w.after["load-script-data"] = func() { object.scriptData = liveScript }
	w.after["rw-field33"] = func() { object.field33 = 0x89abcdef }
	w.after["mark-animation"] = func() { object.field34 = 0x80000003 }
	w.after["load-field34:2"] = func() { object.field34 = 9 }

	if got := triggerXfer4F4E50(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.field34 != 0x11223344 {
		t.Fatalf("Field34 = %#08x, want entry cache %#08x", object.field34, uint32(0x11223344))
	}
	if object.update != liveUpdate {
		t.Fatalf("live UpdateData pointer changed: got %p, want %p", object.update, liveUpdate)
	}
	if object.width != 60 || object.height != -5 {
		t.Fatalf("shape = %g/%g, want clamped 60/-5", object.width, object.height)
	}
	if entryUpdate.flags != 0x11223344 || entryUpdate.state != 3 || entryUpdate.field9 != 4 ||
		entryUpdate.classInclude != 0x55667788 || entryUpdate.classExclude != 0x99aabbcc ||
		entryUpdate.teamInclude != 0x5a || entryUpdate.teamExclude != 0xa5 ||
		entryUpdate.colors != [6]uint8{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5} {
		t.Fatalf("cached update = %+v, want transferred current record", *entryUpdate)
	}
	if *liveUpdate != (triggerXferTestUpdate4F4E50{flags: 9, teamInclude: 10, teamExclude: 11}) {
		t.Fatalf("live update mutated: %+v", *liveUpdate)
	}
	if !reflect.DeepEqual(w.widthTransfers, []int32{0}) || !reflect.DeepEqual(w.heightTransfers, []int32{0}) {
		t.Fatalf("read-mode shape inputs = %v/%v, want zero locals", w.widthTransfers, w.heightTransfers)
	}
	wantScripts := []triggerXferScriptCall4F4E50{
		{entryUpdate, triggerXferActivate4F4E50, entryScript, 256},
		{entryUpdate, triggerXferDeactivate4F4E50, entryScript, 384},
		{entryUpdate, triggerXferCollide4F4E50, entryScript, 512},
	}
	if !reflect.DeepEqual(w.scriptCalls, wantScripts) {
		t.Fatalf("script calls = %#v, want cached update/context %#v", w.scriptCalls, wantScripts)
	}
	if !reflect.DeepEqual(w.markedFrames, []uint32{0x89abcdef}) {
		t.Fatalf("marked frames = %#v, want live Field33", w.markedFrames)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []triggerXferInventoryCall4F4E50{{
		version: 61, object: object, count: -2147483645,
	}}) {
		t.Fatalf("inventory calls = %#v, want raw version and signed live count", w.inventoryCalls)
	}

	wantEvents := triggerXferReadEvents4F4E50()
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %v, want %v", w.events, wantEvents)
	}
}

func triggerXferReadEvents4F4E50() []string {
	return []string{
		"load-field34:1", "load-update", "rw-version", "map-read-write",
		"read-only:1", "rw-width", "rw-height", "store-width:70", "store-height:-5",
		"store-width:60", "calc-box",
		"rw-color:0", "rw-color:1", "rw-color:2", "rw-color:3", "rw-color:4", "rw-color:5",
		"rw-flags", "load-script-data", "transfer-script:0", "transfer-script:1", "transfer-script:2",
		"read-only:2", "rw-class-include", "rw-class-exclude", "read-only:3",
		"store-team-include", "store-team-exclude", "read-only:4", "rw-team-include", "rw-team-exclude",
		"rw-state", "rw-field9", "rw-field33", "read-only:5", "load-field33", "mark-animation",
		"load-field34:2", "read-only:6", "transfer-inventory", "store-field34",
	}
}

func TestTriggerXfer4F4E50SignedVersionAndLegacyScratch(t *testing.T) {
	update := &triggerXferTestUpdate4F4E50{teamInclude: 1, teamExclude: 2}
	object := &triggerXferTestObject4F4E50{field34: 0, update: update, width: 12.75, height: -2.9}
	w := newTriggerXferTestWorld4F4E50()
	w.version = 0xffff
	w.readOnlyDefault = 0
	w.widthResult = 12
	w.heightResult = -2
	w.legacyScratchResult = []uint32{0xaa030201, 0xaa060504, 0xaa090807}

	if got := triggerXfer4F4E50(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.mapCalls, []triggerXferMapCall4F4E50{{object: object, version: -1}}) {
		t.Fatalf("map calls = %#v, want signed version -1", w.mapCalls)
	}
	if !reflect.DeepEqual(w.widthLoads, []float32{12.75}) || !reflect.DeepEqual(w.heightLoads, []float32{-2.9}) ||
		!reflect.DeepEqual(w.truncInputs, []float32{12.75, -2.9}) ||
		!reflect.DeepEqual(w.widthTransfers, []int32{12}) || !reflect.DeepEqual(w.heightTransfers, []int32{-2}) {
		t.Fatalf("write shape order/values = loads %v/%v trunc %v transfer %v/%v",
			w.widthLoads, w.heightLoads, w.truncInputs, w.widthTransfers, w.heightTransfers)
	}
	if !reflect.DeepEqual(w.legacyScratchInput, []uint32{12, 0xaa030201, 0xaa060504}) {
		t.Fatalf("legacy scratch inputs = %#v, want one shared local", w.legacyScratchInput)
	}
	if !reflect.DeepEqual(w.legacyScriptCalls, []triggerXferCallback4F4E50{
		triggerXferActivate4F4E50, triggerXferDeactivate4F4E50,
	}) || len(w.scriptCalls) != 0 {
		t.Fatalf("legacy/current script calls = %v/%v", w.legacyScriptCalls, w.scriptCalls)
	}
	if update.teamInclude != 0x5a || update.teamExclude != 0xa5 {
		t.Fatalf("write-mode teams = %#x/%#x, want transferred", update.teamInclude, update.teamExclude)
	}
	if w.readOnlyCalls != 4 || w.field34Stores != 1 {
		t.Fatalf("mode calls/restores = %d/%d, want 4/1", w.readOnlyCalls, w.field34Stores)
	}
}

func TestTriggerXfer4F4E50VersionAndModeBoundaries(t *testing.T) {
	t.Run("too new stops before base and does not restore", func(t *testing.T) {
		object := &triggerXferTestObject4F4E50{field34: 7, update: &triggerXferTestUpdate4F4E50{}}
		w := newTriggerXferTestWorld4F4E50()
		w.version = 62
		w.after["load-field34:1"] = func() { object.field34 = 9 }
		if got := triggerXfer4F4E50(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 9 || w.field34Stores != 0 || len(w.mapCalls) != 0 {
			t.Fatalf("failure state = field34 %d stores %d map %v", object.field34, w.field34Stores, w.mapCalls)
		}
	})

	t.Run("base failure does not restore", func(t *testing.T) {
		object := &triggerXferTestObject4F4E50{field34: 7, update: &triggerXferTestUpdate4F4E50{}}
		w := newTriggerXferTestWorld4F4E50()
		w.mapResult = 0
		w.after["load-field34:1"] = func() { object.field34 = 9 }
		if got := triggerXfer4F4E50(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 9 || w.field34Stores != 0 || w.readOnlyCalls != 0 {
			t.Fatalf("failure state = field34 %d stores %d modes %d", object.field34, w.field34Stores, w.readOnlyCalls)
		}
	})

	t.Run("version two reads skips and suppresses old teams", func(t *testing.T) {
		update := &triggerXferTestUpdate4F4E50{teamInclude: 7, teamExclude: 8}
		object := &triggerXferTestObject4F4E50{update: update}
		w := newTriggerXferTestWorld4F4E50()
		w.version = 2
		w.legacyCountResult = []uint8{1, 2, 0xff, 4}
		if got := triggerXfer4F4E50(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.legacyCountInput, []uint8{0, 1, 2, 0xff}) ||
			!reflect.DeepEqual(w.seekOffsets, []int32{4, 8, 1020, 16}) {
			t.Fatalf("legacy skip inputs/offsets = %v/%v", w.legacyCountInput, w.seekOffsets)
		}
		if update.teamInclude != 0 || update.teamExclude != 0 {
			t.Fatalf("old read teams = %#x/%#x, want zero without transfer", update.teamInclude, update.teamExclude)
		}
		if len(w.scriptCalls) != 0 || len(w.legacyScriptCalls) != 2 {
			t.Fatalf("script calls = current %v legacy %v", w.scriptCalls, w.legacyScriptCalls)
		}
	})

	for _, tc := range []struct {
		name        string
		version     uint16
		wantScripts int
		wantCounts  int
		wantColors  bool
	}{
		{"v3", 3, 2, 4, false},
		{"v30", 30, 2, 4, false},
		{"v31", 31, 3, 0, false},
		{"v40", 40, 3, 0, false},
		{"v41", 41, 3, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			object := &triggerXferTestObject4F4E50{update: &triggerXferTestUpdate4F4E50{}, scriptData: 1}
			w := newTriggerXferTestWorld4F4E50()
			w.version = tc.version
			if got := triggerXfer4F4E50(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if len(w.scriptCalls) != tc.wantScripts || len(w.legacyCountInput) != tc.wantCounts {
				t.Fatalf("script/count calls = %d/%d, want %d/%d", len(w.scriptCalls), len(w.legacyCountInput), tc.wantScripts, tc.wantCounts)
			}
			colors := 0
			for _, event := range w.events {
				if len(event) >= len("rw-color:") && event[:len("rw-color:")] == "rw-color:" {
					colors++
				}
			}
			if got := colors == 6; got != tc.wantColors {
				t.Fatalf("color branch = %v (%d calls), want %v", got, colors, tc.wantColors)
			}
		})
	}

	t.Run("truthy read is not exact-one read", func(t *testing.T) {
		update := &triggerXferTestUpdate4F4E50{teamInclude: 7, teamExclude: 8}
		object := &triggerXferTestObject4F4E50{field34: 3, update: update, field33: 5}
		w := newTriggerXferTestWorld4F4E50()
		w.readOnlyDefault = 2
		if got := triggerXfer4F4E50(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if object.width != 60 || object.height != -5 {
			t.Fatalf("truthy shape = %g/%g, want read/clamp", object.width, object.height)
		}
		if update.teamInclude != 0x5a || update.teamExclude != 0xa5 {
			t.Fatalf("v61 teams = %#x/%#x, want transferred without zeroing", update.teamInclude, update.teamExclude)
		}
		if len(w.markedFrames) != 0 || len(w.inventoryCalls) != 0 {
			t.Fatalf("exact-one side effects = marks %v inventory %v, want none", w.markedFrames, w.inventoryCalls)
		}
	})

	t.Run("mode is reloaded at every gate", func(t *testing.T) {
		update := &triggerXferTestUpdate4F4E50{teamInclude: 7, teamExclude: 8}
		object := &triggerXferTestObject4F4E50{field34: 3, update: update, width: 7.9, height: 8.9}
		w := newTriggerXferTestWorld4F4E50()
		w.version = 20
		w.readOnlyValues = []int32{0, 1, 1, 2, 1}
		if got := triggerXfer4F4E50(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !reflect.DeepEqual(w.widthTransfers, []int32{7}) || !reflect.DeepEqual(w.heightTransfers, []int32{8}) {
			t.Fatalf("shape transfers = %v/%v, want write-mode 7/8", w.widthTransfers, w.heightTransfers)
		}
		if len(w.legacyCountInput) != 4 || update.teamInclude != 0 || update.teamExclude != 0 {
			t.Fatalf("dynamic gates = counts %d teams %#x/%#x", len(w.legacyCountInput), update.teamInclude, update.teamExclude)
		}
		if len(w.inventoryCalls) != 1 || w.readOnlyCalls != 5 {
			t.Fatalf("inventory/mode calls = %v/%d, want one/5", w.inventoryCalls, w.readOnlyCalls)
		}
	})
}

func TestTriggerXfer4F4E50InventoryFailureDoesNotRestore(t *testing.T) {
	object := &triggerXferTestObject4F4E50{field34: 5, update: &triggerXferTestUpdate4F4E50{}}
	w := newTriggerXferTestWorld4F4E50()
	w.inventoryResult = 0
	w.after["load-field34:1"] = func() { object.field34 = 3 }

	if got := triggerXfer4F4E50(object, w.deps()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if object.field34 != 3 || w.field34Stores != 0 {
		t.Fatalf("Field34/stores = %d/%d, want live 3 and no restore", object.field34, w.field34Stores)
	}
	if !reflect.DeepEqual(w.inventoryCalls, []triggerXferInventoryCall4F4E50{{
		version: 61, object: object, count: 3,
	}}) {
		t.Fatalf("inventory calls = %#v, want live count 3", w.inventoryCalls)
	}
}

func TestTriggerXferTruncFloatDword4F4E50(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float32
		want  int32
	}{
		{"positive", 12.9, 12},
		{"negative", -12.9, -12},
		{"signed-dword-wrap", 0x1p31, math.MinInt32},
		{"unsigned-dword-wrap", 0x1p32, 0},
		{"minimum-qword", -0x1p63, 0},
		{"positive-overflow", 0x1p63, 0},
		{"negative-overflow", float32(math.Inf(-1)), 0},
		{"positive-infinity", float32(math.Inf(1)), 0},
		{"nan", float32(math.NaN()), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := triggerXferTruncFloatDword4F4E50(tc.value); got != tc.want {
				t.Fatalf("trunc(%g) = %#x, want %#x", tc.value, uint32(got), uint32(tc.want))
			}
		})
	}
}

func TestTriggerXfer4F4E50FaultPrefixes(t *testing.T) {
	wantEvents := triggerXferReadEvents4F4E50()
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			object := &triggerXferTestObject4F4E50{
				field34: 3, update: &triggerXferTestUpdate4F4E50{}, scriptData: 1,
			}
			w := newTriggerXferTestWorld4F4E50()
			w.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatalf("fault %d did not panic", faultAt)
					}
				}()
				triggerXfer4F4E50(object, w.deps())
			}()
			if !reflect.DeepEqual(w.events, wantEvents[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", w.events, wantEvents[:faultAt])
			}
			if faultAt < len(wantEvents) && w.field34Stores != 0 {
				t.Fatalf("fault prefix restored Field34 early at event %d", faultAt)
			}
		})
	}
}
