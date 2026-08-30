package legacy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type rewardMarkerXferTestData4F74D0 struct {
	lists [3][]uint8
}

func newRewardMarkerXferTestData4F74D0() *rewardMarkerXferTestData4F74D0 {
	return &rewardMarkerXferTestData4F74D0{lists: [3][]uint8{
		make([]uint8, rewardMarkerXferSpellCount4F74D0),
		make([]uint8, rewardMarkerXferAbilityCount4F74D0),
		make([]uint8, rewardMarkerXferGuideCount4F74D0),
	}}
}

type rewardMarkerXferTestObject4F74D0 struct {
	data    *rewardMarkerXferTestData4F74D0
	field34 uint32
}

type rewardMarkerXferTestInventoryCall4F74D0 struct {
	version uint16
	object  *rewardMarkerXferTestObject4F74D0
	count   int32
}

type rewardMarkerXferTestNameKey4F74D0 struct {
	list  rewardMarkerXferList4F74D0
	index int
}

type rewardMarkerXferTestWorld4F74D0 struct {
	version         uint16
	mapResult       int32
	countResults    []*uint16
	modes           []int32
	lengthResults   []*uint8
	payloadResults  [][]byte
	names           map[rewardMarkerXferTestNameKey4F74D0][][]byte
	nameIDs         map[string]int
	inventoryResult int32
	onMap           func()
	onField         func(rewardMarkerXferField4F74D0)

	initDataLoads    int
	field34Loads     int
	field34Stores    int
	mapVersions      []int32
	countInputs      []uint16
	modeCalls        int
	lengthInputs     []uint8
	payloadCalls     [][]byte
	nameLoadCalls    []rewardMarkerXferTestNameKey4F74D0
	nameResolveCalls []string
	stored           []rewardMarkerXferTestNameKey4F74D0
	fields           []rewardMarkerXferField4F74D0
	inventoryCalls   []rewardMarkerXferTestInventoryCall4F74D0
	events           []string
}

func newRewardMarkerXferTestWorld4F74D0() *rewardMarkerXferTestWorld4F74D0 {
	return &rewardMarkerXferTestWorld4F74D0{
		version:         rewardMarkerXferCurrentVersion4F74D0,
		mapResult:       1,
		names:           make(map[rewardMarkerXferTestNameKey4F74D0][][]byte),
		nameIDs:         make(map[string]int),
		inventoryResult: 1,
	}
}

func rewardMarkerXferTestU8_4F74D0(value uint8) *uint8    { return &value }
func rewardMarkerXferTestU16_4F74D0(value uint16) *uint16 { return &value }

func rewardMarkerXferTestListName4F74D0(list rewardMarkerXferList4F74D0) string {
	return [...]string{"spell", "ability", "guide"}[list]
}

func rewardMarkerXferTestFieldName4F74D0(field rewardMarkerXferField4F74D0) string {
	return [...]string{"196", "192", "200", "204", "208", "212", "216-low"}[field]
}

func rewardMarkerXferTestNameIDKey4F74D0(list rewardMarkerXferList4F74D0, name string) string {
	return fmt.Sprintf("%d:%s", list, name)
}

func (w *rewardMarkerXferTestWorld4F74D0) deps() rewardMarkerXferDeps4F74D0[
	*rewardMarkerXferTestObject4F74D0,
	*rewardMarkerXferTestData4F74D0,
] {
	countCall := 0
	lengthCall := 0
	payloadCall := 0
	nameCalls := make(map[rewardMarkerXferTestNameKey4F74D0]int)
	return rewardMarkerXferDeps4F74D0[
		*rewardMarkerXferTestObject4F74D0,
		*rewardMarkerXferTestData4F74D0,
	]{
		loadInitData: func(object *rewardMarkerXferTestObject4F74D0) *rewardMarkerXferTestData4F74D0 {
			w.initDataLoads++
			w.events = append(w.events, "load-init-data")
			return object.data
		},
		loadField34: func(object *rewardMarkerXferTestObject4F74D0) uint32 {
			w.field34Loads++
			w.events = append(w.events, fmt.Sprintf("load-field34:%d", w.field34Loads))
			return object.field34
		},
		storeField34: func(object *rewardMarkerXferTestObject4F74D0, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.events = append(w.events, "store-field34")
		},
		rwVersion: func(value uint16) uint16 {
			w.events = append(w.events, fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *rewardMarkerXferTestObject4F74D0, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.events = append(w.events, fmt.Sprintf("map-read-write:%d", version))
			if w.onMap != nil {
				w.onMap()
			}
			return w.mapResult
		},
		rwHeader: func(_ *rewardMarkerXferTestData4F74D0, field rewardMarkerXferHeader4F74D0) {
			w.events = append(w.events, fmt.Sprintf("header:%d", field))
		},
		loadListValue: func(data *rewardMarkerXferTestData4F74D0, list rewardMarkerXferList4F74D0, index int) uint8 {
			return data.lists[list][index]
		},
		storeListValue: func(data *rewardMarkerXferTestData4F74D0, list rewardMarkerXferList4F74D0, index int, value uint8) {
			data.lists[list][index] = value
			key := rewardMarkerXferTestNameKey4F74D0{list: list, index: index}
			w.stored = append(w.stored, key)
			w.events = append(w.events, fmt.Sprintf("store-%s:%d=%d", rewardMarkerXferTestListName4F74D0(list), index, value))
		},
		rwCount: func(value uint16) uint16 {
			w.countInputs = append(w.countInputs, value)
			w.events = append(w.events, fmt.Sprintf("count:%d=%d", countCall+1, value))
			result := value
			if countCall < len(w.countResults) && w.countResults[countCall] != nil {
				result = *w.countResults[countCall]
			}
			countCall++
			return result
		},
		readMode: func() int32 {
			call := w.modeCalls
			w.modeCalls++
			var value int32
			if call < len(w.modes) {
				value = w.modes[call]
			}
			w.events = append(w.events, fmt.Sprintf("mode:%d=%d", call+1, value))
			return value
		},
		rwNameLength: func(value uint8) uint8 {
			w.lengthInputs = append(w.lengthInputs, value)
			result := value
			if lengthCall < len(w.lengthResults) && w.lengthResults[lengthCall] != nil {
				result = *w.lengthResults[lengthCall]
			}
			w.events = append(w.events, fmt.Sprintf("length:%d=%d->%d", lengthCall+1, value, result))
			lengthCall++
			return result
		},
		rwNameBytes: func(value []byte) {
			if payloadCall < len(w.payloadResults) && w.payloadResults[payloadCall] != nil {
				if len(w.payloadResults[payloadCall]) < len(value) {
					panic("RewardMarkerXfer test byte-stream underrun")
				}
				copy(value, w.payloadResults[payloadCall][:len(value)])
			}
			w.payloadCalls = append(w.payloadCalls, append([]byte(nil), value...))
			w.events = append(w.events, fmt.Sprintf("payload:%d=%d", payloadCall+1, len(value)))
			payloadCall++
		},
		resolveName: func(list rewardMarkerXferList4F74D0, name []byte) int {
			call := fmt.Sprintf("%s:%s", rewardMarkerXferTestListName4F74D0(list), string(name))
			w.nameResolveCalls = append(w.nameResolveCalls, call)
			w.events = append(w.events, "resolve-"+call)
			return w.nameIDs[rewardMarkerXferTestNameIDKey4F74D0(list, string(name))]
		},
		loadName: func(list rewardMarkerXferList4F74D0, index int) []byte {
			key := rewardMarkerXferTestNameKey4F74D0{list: list, index: index}
			call := nameCalls[key]
			nameCalls[key]++
			w.nameLoadCalls = append(w.nameLoadCalls, key)
			w.events = append(w.events, fmt.Sprintf("name-%s:%d:%d", rewardMarkerXferTestListName4F74D0(list), index, call+1))
			names := w.names[key]
			if call >= len(names) {
				return nil
			}
			return names[call]
		},
		rwField: func(_ *rewardMarkerXferTestData4F74D0, field rewardMarkerXferField4F74D0) {
			w.fields = append(w.fields, field)
			w.events = append(w.events, "field:"+rewardMarkerXferTestFieldName4F74D0(field))
			if w.onField != nil {
				w.onField(field)
			}
		},
		transferInventory: func(version uint16, object *rewardMarkerXferTestObject4F74D0, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, rewardMarkerXferTestInventoryCall4F74D0{
				version: version, object: object, count: count,
			})
			w.events = append(w.events, "inventory")
			return w.inventoryResult
		},
	}
}

func TestRewardMarkerXfer4F74D0WriteCountsExactOneButEmitsAllNonzero(t *testing.T) {
	entry := newRewardMarkerXferTestData4F74D0()
	replacement := newRewardMarkerXferTestData4F74D0()
	entry.lists[rewardMarkerXferSpells4F74D0][1] = 1
	entry.lists[rewardMarkerXferSpells4F74D0][2] = 2
	entry.lists[rewardMarkerXferAbilities4F74D0][3] = 1
	entry.lists[rewardMarkerXferGuides4F74D0][4] = 0xff
	object := &rewardMarkerXferTestObject4F74D0{data: entry, field34: 0x10203040}
	w := newRewardMarkerXferTestWorld4F74D0()
	w.mapResult = -7
	w.modes = []int32{0, 0, 0, 0}
	w.names[rewardMarkerXferTestNameKey4F74D0{rewardMarkerXferSpells4F74D0, 1}] = [][]byte{[]byte("S1"), []byte("Q1")}
	w.names[rewardMarkerXferTestNameKey4F74D0{rewardMarkerXferSpells4F74D0, 2}] = [][]byte{[]byte(strings.Repeat("x", 260)), []byte("ABCD-tail")}
	w.names[rewardMarkerXferTestNameKey4F74D0{rewardMarkerXferAbilities4F74D0, 3}] = [][]byte{[]byte("ABILITY"), []byte("ability")}
	w.names[rewardMarkerXferTestNameKey4F74D0{rewardMarkerXferGuides4F74D0, 4}] = [][]byte{[]byte("GUIDE"), []byte("guide")}
	w.onMap = func() { object.data = replacement }

	if got := rewardMarkerXfer4F74D0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if object.data != replacement || object.field34 != 0x10203040 {
		t.Fatalf("live InitData/Field34 = %p/%#x, want replacement/restored", object.data, object.field34)
	}
	if len(w.events) < 4 || !reflect.DeepEqual(w.events[:4], []string{
		"load-init-data", "load-field34:1", "rw-version:63", "map-read-write:63",
	}) {
		t.Fatalf("entry events = %v, want InitData then Field34 then version/common", w.events)
	}
	if !reflect.DeepEqual(w.countInputs, []uint16{1, 1, 0}) {
		t.Fatalf("count inputs = %v, want exact-one counts [1 1 0]", w.countInputs)
	}
	if !reflect.DeepEqual(w.lengthInputs, []uint8{2, 4, 7, 5}) {
		t.Fatalf("length inputs = %v, want low-byte first-name lengths [2 4 7 5]", w.lengthInputs)
	}
	wantPayloads := [][]byte{[]byte("Q1"), []byte("ABCD"), []byte("ability"), []byte("guide")}
	if !reflect.DeepEqual(w.payloadCalls, wantPayloads) {
		t.Fatalf("payloads = %q, want second-name bytes %q", w.payloadCalls, wantPayloads)
	}
	if len(w.nameLoadCalls) != 8 || len(w.inventoryCalls) != 0 || w.modeCalls != 4 || w.field34Stores != 1 {
		t.Fatalf("name loads/inventory/modes/stores = %d/%d/%d/%d, want 8/0/4/1",
			len(w.nameLoadCalls), len(w.inventoryCalls), w.modeCalls, w.field34Stores)
	}
	wantFields := []rewardMarkerXferField4F74D0{
		rewardMarkerXferField196_4F74D0, rewardMarkerXferField192_4F74D0,
		rewardMarkerXferField200_4F74D0, rewardMarkerXferField204_4F74D0,
		rewardMarkerXferField208_4F74D0, rewardMarkerXferField212_4F74D0,
		rewardMarkerXferField216Low_4F74D0,
	}
	if !reflect.DeepEqual(w.fields, wantFields) {
		t.Fatalf("fields = %v, want original non-address order %v", w.fields, wantFields)
	}
}

func TestRewardMarkerXfer4F74D0SamplesEachReadModeAndKeepsExistingFlags(t *testing.T) {
	data := newRewardMarkerXferTestData4F74D0()
	data.lists[rewardMarkerXferSpells4F74D0][2] = 7
	data.lists[rewardMarkerXferAbilities4F74D0][1] = 2
	object := &rewardMarkerXferTestObject4F74D0{data: data, field34: 0x11223344}
	w := newRewardMarkerXferTestWorld4F74D0()
	w.countResults = []*uint16{
		rewardMarkerXferTestU16_4F74D0(2), nil, rewardMarkerXferTestU16_4F74D0(1),
	}
	w.modes = []int32{2, 0, -1, 1}
	w.lengthResults = []*uint8{
		rewardMarkerXferTestU8_4F74D0(3), rewardMarkerXferTestU8_4F74D0(3), nil,
		rewardMarkerXferTestU8_4F74D0(3),
	}
	w.payloadResults = [][]byte{[]byte("S\x00x"), []byte("TWO"), nil, []byte("G07")}
	w.nameIDs[rewardMarkerXferTestNameIDKey4F74D0(rewardMarkerXferSpells4F74D0, "S")] = 5
	w.nameIDs[rewardMarkerXferTestNameIDKey4F74D0(rewardMarkerXferSpells4F74D0, "TWO")] = 6
	w.nameIDs[rewardMarkerXferTestNameIDKey4F74D0(rewardMarkerXferGuides4F74D0, "G07")] = 7
	w.names[rewardMarkerXferTestNameKey4F74D0{rewardMarkerXferAbilities4F74D0, 1}] = [][]byte{[]byte("A"), []byte("B")}
	w.onField = func(field rewardMarkerXferField4F74D0) {
		if field == rewardMarkerXferField216Low_4F74D0 {
			object.field34 = 0x80000002
		}
	}

	if got := rewardMarkerXfer4F74D0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if data.lists[rewardMarkerXferSpells4F74D0][2] != 7 ||
		data.lists[rewardMarkerXferSpells4F74D0][5] != 1 ||
		data.lists[rewardMarkerXferSpells4F74D0][6] != 1 ||
		data.lists[rewardMarkerXferGuides4F74D0][7] != 1 {
		t.Fatalf("read list state = spells[%d,%d,%d], guide=%d; want existing 7 plus resolved ones",
			data.lists[0][2], data.lists[0][5], data.lists[0][6], data.lists[2][7])
	}
	wantResolves := []string{"spell:S", "spell:TWO", "guide:G07"}
	if !reflect.DeepEqual(w.nameResolveCalls, wantResolves) {
		t.Fatalf("resolved names = %q, want C-string prefixes %q", w.nameResolveCalls, wantResolves)
	}
	wantInventory := []rewardMarkerXferTestInventoryCall4F74D0{{
		version: 63, object: object, count: -2147483646,
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) || object.field34 != 0x11223344 {
		t.Fatalf("inventory/Field34 = %#v/%#x, want live signed count and restored entry", w.inventoryCalls, object.field34)
	}
}

func TestRewardMarkerXfer4F74D0SignedVersionGateAndSuffixVersions(t *testing.T) {
	for _, version := range []uint16{0, 64, 0x7fff, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("reject-%#x", version), func(t *testing.T) {
			object := &rewardMarkerXferTestObject4F74D0{data: newRewardMarkerXferTestData4F74D0(), field34: 7}
			w := newRewardMarkerXferTestWorld4F74D0()
			w.version = version
			if got := rewardMarkerXfer4F74D0(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if w.initDataLoads != 1 || w.field34Loads != 1 || len(w.mapVersions) != 0 || w.field34Stores != 0 {
				t.Fatalf("entry loads/map/stores = %d/%d/%v/%d, want 1/1/none/0",
					w.initDataLoads, w.field34Loads, w.mapVersions, w.field34Stores)
			}
		})
	}

	for _, tc := range []struct {
		version uint16
		fields  []rewardMarkerXferField4F74D0
	}{
		{61, []rewardMarkerXferField4F74D0{0, 1, 2, 3, 4}},
		{62, []rewardMarkerXferField4F74D0{0, 1, 2, 3, 4, 5}},
		{63, []rewardMarkerXferField4F74D0{0, 1, 2, 3, 4, 5, 6}},
	} {
		t.Run(fmt.Sprintf("suffix-%d", tc.version), func(t *testing.T) {
			object := &rewardMarkerXferTestObject4F74D0{data: newRewardMarkerXferTestData4F74D0()}
			w := newRewardMarkerXferTestWorld4F74D0()
			w.version = tc.version
			w.mapResult = -1
			w.modes = []int32{0, 0, 0}
			if got := rewardMarkerXfer4F74D0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if !reflect.DeepEqual(w.mapVersions, []int32{int32(tc.version)}) || !reflect.DeepEqual(w.fields, tc.fields) {
				t.Fatalf("map versions/fields = %v/%v, want signed version/%v", w.mapVersions, w.fields, tc.fields)
			}
			if w.modeCalls != 3 {
				t.Fatalf("mode calls = %d, want three list calls and no suffix call for Field34 zero", w.modeCalls)
			}
		})
	}
}

func TestRewardMarkerXfer4F74D0FailurePrefixesDoNotRestoreField34(t *testing.T) {
	t.Run("common failure", func(t *testing.T) {
		object := &rewardMarkerXferTestObject4F74D0{data: newRewardMarkerXferTestData4F74D0(), field34: 0x11}
		w := newRewardMarkerXferTestWorld4F74D0()
		w.mapResult = 0
		w.onMap = func() { object.field34 = 0x22 }
		if got := rewardMarkerXfer4F74D0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 0x22 || w.field34Stores != 0 || len(w.countInputs) != 0 {
			t.Fatalf("Field34/stores/counts = %#x/%d/%v, want callback state/0/none", object.field34, w.field34Stores, w.countInputs)
		}
	})

	t.Run("unbounded read count reaches first invalid name", func(t *testing.T) {
		object := &rewardMarkerXferTestObject4F74D0{data: newRewardMarkerXferTestData4F74D0(), field34: 0x11}
		w := newRewardMarkerXferTestWorld4F74D0()
		w.countResults = []*uint16{rewardMarkerXferTestU16_4F74D0(0xffff)}
		w.modes = []int32{1}
		w.lengthResults = []*uint8{rewardMarkerXferTestU8_4F74D0(3)}
		w.payloadResults = [][]byte{[]byte("BAD")}
		w.onMap = func() { object.field34 = 0x22 }
		if got := rewardMarkerXfer4F74D0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if !reflect.DeepEqual(w.nameResolveCalls, []string{"spell:BAD"}) || w.field34Stores != 0 || object.field34 != 0x22 {
			t.Fatalf("resolves/stores/Field34 = %q/%d/%#x, want one invalid prefix/no restore", w.nameResolveCalls, w.field34Stores, object.field34)
		}
	})

	t.Run("inventory failure", func(t *testing.T) {
		object := &rewardMarkerXferTestObject4F74D0{data: newRewardMarkerXferTestData4F74D0(), field34: 7}
		w := newRewardMarkerXferTestWorld4F74D0()
		w.modes = []int32{0, 0, 0, 1}
		w.inventoryResult = 0
		w.onField = func(field rewardMarkerXferField4F74D0) {
			if field == rewardMarkerXferField216Low_4F74D0 {
				object.field34 = 9
			}
		}
		if got := rewardMarkerXfer4F74D0(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 9 || w.field34Stores != 0 || len(w.inventoryCalls) != 1 || w.inventoryCalls[0].count != 9 {
			t.Fatalf("Field34/stores/inventory = %d/%d/%#v, want live 9/no restore/one call", object.field34, w.field34Stores, w.inventoryCalls)
		}
	})

	for _, mode := range []int32{2, -1} {
		t.Run(fmt.Sprintf("suffix-mode-%d-skips-inventory", mode), func(t *testing.T) {
			object := &rewardMarkerXferTestObject4F74D0{data: newRewardMarkerXferTestData4F74D0(), field34: 7}
			w := newRewardMarkerXferTestWorld4F74D0()
			w.modes = []int32{0, 0, 0, mode}
			if got := rewardMarkerXfer4F74D0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if len(w.inventoryCalls) != 0 || object.field34 != 7 || w.field34Stores != 1 {
				t.Fatalf("inventory/Field34/stores = %d/%d/%d, want none/restored/1", len(w.inventoryCalls), object.field34, w.field34Stores)
			}
		})
	}
}
