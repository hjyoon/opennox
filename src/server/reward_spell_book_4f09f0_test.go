package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"
	"testing"
	"unsafe"
)

type rewardSpellBookTestData4F09F0 struct {
	flags  uint8
	spells [rewardSpellBookSpellCount4F09F0]uint8
}

type rewardSpellBookTestMarker4F09F0 struct {
	data *rewardSpellBookTestData4F09F0
}

type rewardSpellBookTestObject4F09F0 struct {
	typeName string
	spell    uint8
}

func rewardSpellBookTestHooks4F09F0(
	rows []rewardSpellDefinition4F09F0,
) rewardSpellBookHooks4F09F0[
	*rewardSpellBookTestMarker4F09F0,
	*rewardSpellBookTestData4F09F0,
	*rewardSpellBookTestObject4F09F0,
] {
	return rewardSpellBookHooks4F09F0[
		*rewardSpellBookTestMarker4F09F0,
		*rewardSpellBookTestData4F09F0,
		*rewardSpellBookTestObject4F09F0,
	]{
		loadInitData: func(marker *rewardSpellBookTestMarker4F09F0) *rewardSpellBookTestData4F09F0 {
			return marker.data
		},
		loadFlags: func(data *rewardSpellBookTestData4F09F0) uint8 {
			return data.flags
		},
		loadExplicitSpell: func(data *rewardSpellBookTestData4F09F0, index int) uint8 {
			return data.spells[index]
		},
		pickSlots: func(uint32) uint32 { return 1 },
		rows:      rows,
		randomInt: func(int32, int32) int32 { return 0 },
		checkSpellClass: func(uint32, uint32) int32 {
			return 0
		},
		createObjectByType: func(typeName string) *rewardSpellBookTestObject4F09F0 {
			return &rewardSpellBookTestObject4F09F0{typeName: typeName}
		},
		isNilObject: func(object *rewardSpellBookTestObject4F09F0) bool {
			return object == nil
		},
		storeSpell: func(object *rewardSpellBookTestObject4F09F0, spell uint8) {
			object.spell = spell
		},
	}
}

func TestRewardSpellDefinitionsMatchGAMEEXE4F09F0(t *testing.T) {
	if got := unsafe.Sizeof(rewardSpellDefinitions4F09F0); got != 684 {
		t.Fatalf("reward spell table size = %d, want 684", got)
	}
	if unsafe.Sizeof(rewardSpellDefinition4F09F0{}) != 12 ||
		unsafe.Offsetof(rewardSpellDefinition4F09F0{}.Weight) != 0 ||
		unsafe.Offsetof(rewardSpellDefinition4F09F0{}.SpellID) != 4 ||
		unsafe.Offsetof(rewardSpellDefinition4F09F0{}.Slots) != 8 {
		t.Fatalf("reward spell row layout = size %d offsets %d/%d/%d, want 12 and 0/4/8",
			unsafe.Sizeof(rewardSpellDefinition4F09F0{}),
			unsafe.Offsetof(rewardSpellDefinition4F09F0{}.Weight),
			unsafe.Offsetof(rewardSpellDefinition4F09F0{}.SpellID),
			unsafe.Offsetof(rewardSpellDefinition4F09F0{}.Slots),
		)
	}
	raw := unsafe.Slice(
		(*byte)(unsafe.Pointer(&rewardSpellDefinitions4F09F0[0])),
		int(unsafe.Sizeof(rewardSpellDefinitions4F09F0)),
	)
	if got, want := fmt.Sprintf("%x", sha256.Sum256(raw)), "e3f0fbda591a355bfdfcf4ef23af861c11187c152ef5c2071a9bcf05a8b8c8f9"; got != want {
		t.Fatalf("raw reward spell table SHA-256 = %s, want %s", got, want)
	}
	var semantic bytes.Buffer
	positive, disabled := 0, 0
	var slotTotals [5]uint32
	for _, row := range rewardSpellDefinitions4F09F0 {
		semantic.WriteByte(row.Weight)
		if err := binary.Write(&semantic, binary.LittleEndian, row.SpellID); err != nil {
			t.Fatal(err)
		}
		if err := binary.Write(&semantic, binary.LittleEndian, row.Slots); err != nil {
			t.Fatal(err)
		}
		if row.Weight != 0 {
			positive++
		} else if row.SpellID != 0 {
			disabled++
		}
		for index := range slotTotals {
			if row.SpellID != 0 && row.Slots&(uint32(1)<<index) != 0 {
				slotTotals[index] += uint32(row.Weight)
			}
		}
	}
	if semantic.Len() != 513 {
		t.Fatalf("semantic table size = %d, want 513", semantic.Len())
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(semantic.Bytes())), "41448fe2c0d7ba60ce7be75819954942708df438ab0cad0217294436c680dd45"; got != want {
		t.Fatalf("semantic reward spell table SHA-256 = %s, want %s", got, want)
	}
	if positive != 46 || disabled != 10 || slotTotals != [5]uint32{48, 78, 97, 113, 113} {
		t.Fatalf("row counts/totals = %d/%d/%v, want 46/10/[48 78 97 113 113]", positive, disabled, slotTotals)
	}
	if got := rewardSpellDefinitions4F09F0[len(rewardSpellDefinitions4F09F0)-1]; got != (rewardSpellDefinition4F09F0{Slots: 0x1f}) {
		t.Fatalf("sentinel = %#v, want zero ID with slots 0x1f", got)
	}
}

func TestRewardSpellBookExplicitUsesCachedDataAndTwoLivePasses4F09F0(t *testing.T) {
	entry := &rewardSpellBookTestData4F09F0{flags: 1}
	entry.spells[2], entry.spells[5], entry.spells[136] = 1, 1, 1
	entry.spells[7] = 2
	replacement := &rewardSpellBookTestData4F09F0{flags: 0}
	marker := &rewardSpellBookTestMarker4F09F0{data: entry}
	hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
	reads := 0
	hooks.loadFlags = func(data *rewardSpellBookTestData4F09F0) uint8 {
		if data != entry {
			t.Fatal("flags did not use entry-cached InitData")
		}
		marker.data = replacement
		return data.flags
	}
	hooks.loadExplicitSpell = func(data *rewardSpellBookTestData4F09F0, index int) uint8 {
		if data != entry {
			t.Fatal("explicit spell did not use entry-cached InitData")
		}
		reads++
		return data.spells[index]
	}
	hooks.pickSlots = func(uint32) uint32 {
		t.Fatal("explicit path called slot selector")
		return 0
	}
	hooks.randomInt = func(minimum, maximum int32) int32 {
		if minimum != 0 || maximum != 2 || reads != 137 {
			t.Fatalf("RNG bounds/read prefix = %d..%d/%d, want 0..2/137", minimum, maximum, reads)
		}
		entry.spells[5] = 0
		entry.spells[9] = 1
		return 1
	}
	var classes []uint32
	hooks.checkSpellClass = func(class, spell uint32) int32 {
		if spell != 9 {
			t.Fatalf("class check spell = %d, want live second-pass index 9", spell)
		}
		classes = append(classes, class)
		return 0
	}
	result := rewardSpellBook4F09F0(marker, 0xfeedbeef, hooks)
	if result == nil || result.typeName != rewardSpellBookCommonType4F09F0 || result.spell != 9 {
		t.Fatalf("result = %#v, want CommonSpellBook spell 9", result)
	}
	if reads != 147 {
		t.Fatalf("explicit reads = %d, want first 137 plus second indices 0..9", reads)
	}
	if !slices.Equal(classes, []uint32{1, 2}) || marker.data != replacement {
		t.Fatalf("class calls/data = %v/%p, want [1 2]/replacement", classes, marker.data)
	}
}

func TestRewardSpellBookExplicitRequiresExactOneAndRejectsZeroID4F09F0(t *testing.T) {
	tests := []struct {
		name   string
		spells map[int]uint8
		rng    bool
	}{
		{"no exact one", map[int]uint8{1: 2, 2: 0xff}, false},
		{"selected index zero", map[int]uint8{0: 1}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := &rewardSpellBookTestData4F09F0{flags: 1}
			for index, value := range test.spells {
				data.spells[index] = value
			}
			hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
			rngCalls := 0
			hooks.randomInt = func(minimum, maximum int32) int32 {
				rngCalls++
				if minimum != 0 || maximum != 0 {
					t.Fatalf("RNG bounds = %d..%d, want 0..0", minimum, maximum)
				}
				return 0
			}
			hooks.checkSpellClass = func(uint32, uint32) int32 {
				t.Fatal("zero/unselected spell reached class check")
				return 0
			}
			got := rewardSpellBook4F09F0(&rewardSpellBookTestMarker4F09F0{data: data}, 3, hooks)
			if got != nil || (rngCalls != 0) != test.rng {
				t.Fatalf("result/RNG calls = %#v/%d, want nil/RNG %t", got, rngCalls, test.rng)
			}
		})
	}
}

func TestRewardSpellBookExplicitCanExhaustChangedSecondPass4F09F0(t *testing.T) {
	data := &rewardSpellBookTestData4F09F0{flags: 1}
	data.spells[3], data.spells[8] = 1, 1
	hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
	reads := 0
	hooks.loadExplicitSpell = func(data *rewardSpellBookTestData4F09F0, index int) uint8 {
		reads++
		return data.spells[index]
	}
	hooks.randomInt = func(minimum, maximum int32) int32 {
		data.spells[8] = 0
		return 1
	}
	hooks.checkSpellClass = func(uint32, uint32) int32 {
		t.Fatal("exhausted second pass reached class check")
		return 0
	}
	got := rewardSpellBook4F09F0(&rewardSpellBookTestMarker4F09F0{data: data}, 0, hooks)
	if got != nil || reads != 274 {
		t.Fatalf("result/reads = %#v/%d, want nil/274", got, reads)
	}
}

func TestRewardSpellBookAutomaticReloadsRowsAfterRNG4F09F0(t *testing.T) {
	rows := []rewardSpellDefinition4F09F0{
		{Weight: 2, SpellID: 11, Slots: 1},
		{Weight: 3, SpellID: 22, Slots: 1},
		{SpellID: 0, Slots: 0x1f},
	}
	data := &rewardSpellBookTestData4F09F0{}
	hooks := rewardSpellBookTestHooks4F09F0(rows)
	hooks.pickSlots = func(stage uint32) uint32 {
		if stage != 0xfeedbeef {
			t.Fatalf("slot selector stage = %#x", stage)
		}
		return 1
	}
	hooks.randomInt = func(minimum, maximum int32) int32 {
		if minimum != 0 || maximum != 4 {
			t.Fatalf("weighted RNG bounds = %d..%d, want 0..4", minimum, maximum)
		}
		rows[0].Weight = 5
		rows[0].SpellID = 0x123
		return 3
	}
	var checked []uint32
	hooks.checkSpellClass = func(class, spell uint32) int32 {
		checked = append(checked, spell)
		return 0
	}
	result := rewardSpellBook4F09F0(
		&rewardSpellBookTestMarker4F09F0{data: data},
		0xfeedbeef,
		hooks,
	)
	if result == nil || result.typeName != rewardSpellBookCommonType4F09F0 || result.spell != 0x23 {
		t.Fatalf("result = %#v, want CommonSpellBook low byte 0x23", result)
	}
	if !slices.Equal(checked, []uint32{0x123, 0x123}) {
		t.Fatalf("full spell IDs checked = %#v, want two 0x123", checked)
	}
}

func TestRewardSpellBookAutomaticZeroBranches4F09F0(t *testing.T) {
	tests := []struct {
		name   string
		rows   []rewardSpellDefinition4F09F0
		mutate func([]rewardSpellDefinition4F09F0)
		rng    bool
	}{
		{
			name: "zero first ID",
			rows: []rewardSpellDefinition4F09F0{{Slots: 1}, {Slots: 0x1f}},
		},
		{
			name: "zero total",
			rows: []rewardSpellDefinition4F09F0{{SpellID: 7, Slots: 2}, {Slots: 0x1f}},
		},
		{
			name: "RNG clears first ID",
			rows: []rewardSpellDefinition4F09F0{{Weight: 1, SpellID: 7, Slots: 1}, {Slots: 0x1f}},
			mutate: func(rows []rewardSpellDefinition4F09F0) {
				rows[0].SpellID = 0
			},
			rng: true,
		},
		{
			name: "RNG removes selectable weight",
			rows: []rewardSpellDefinition4F09F0{{Weight: 1, SpellID: 7, Slots: 1}, {Slots: 0x1f}},
			mutate: func(rows []rewardSpellDefinition4F09F0) {
				rows[0].Weight = 0
			},
			rng: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hooks := rewardSpellBookTestHooks4F09F0(test.rows)
			rngCalls := 0
			hooks.randomInt = func(minimum, maximum int32) int32 {
				rngCalls++
				if test.mutate != nil {
					test.mutate(test.rows)
				}
				return 0
			}
			hooks.checkSpellClass = func(uint32, uint32) int32 {
				t.Fatal("zero automatic branch reached class check")
				return 0
			}
			got := rewardSpellBook4F09F0(
				&rewardSpellBookTestMarker4F09F0{data: &rewardSpellBookTestData4F09F0{}},
				0,
				hooks,
			)
			if got != nil || (rngCalls != 0) != test.rng {
				t.Fatalf("result/RNG calls = %#v/%d, want nil/RNG %t", got, rngCalls, test.rng)
			}
		})
	}
}

func TestRewardSpellBookClassCheckOrder4F09F0(t *testing.T) {
	tests := []struct {
		name      string
		responses []int32
		calls     []uint32
		typeName  string
	}{
		{"common", []int32{0, 0}, []uint32{1, 2}, rewardSpellBookCommonType4F09F0},
		{"wizard", []int32{0, 9, 0}, []uint32{1, 2, 1}, rewardSpellBookWizardType4F09F0},
		{"conjurer", []int32{9, 9, 0}, []uint32{1, 1, 2}, rewardSpellBookConjurerType4F09F0},
		{"rejected", []int32{9, 9, 9}, []uint32{1, 1, 2}, ""},
		{"live second wizard", []int32{9, 0}, []uint32{1, 1}, rewardSpellBookWizardType4F09F0},
		{"live second conjurer", []int32{0, 9, 9, 0}, []uint32{1, 2, 1, 2}, rewardSpellBookConjurerType4F09F0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
			var calls []uint32
			hooks.checkSpellClass = func(class, spell uint32) int32 {
				if spell != 0x12345678 {
					t.Fatalf("class check spell = %#x", spell)
				}
				calls = append(calls, class)
				return test.responses[len(calls)-1]
			}
			created := 0
			hooks.createObjectByType = func(typeName string) *rewardSpellBookTestObject4F09F0 {
				created++
				return &rewardSpellBookTestObject4F09F0{typeName: typeName}
			}
			got := rewardSpellBookCreate4F09F0(0x12345678, hooks)
			if !slices.Equal(calls, test.calls) {
				t.Fatalf("class calls = %v, want %v", calls, test.calls)
			}
			if test.typeName == "" {
				if got != nil || created != 0 {
					t.Fatalf("rejected result/created = %#v/%d", got, created)
				}
			} else if got == nil || got.typeName != test.typeName || got.spell != 0x78 || created != 1 {
				t.Fatalf("result/created = %#v/%d, want %s/1", got, created, test.typeName)
			}
		})
	}
}

func TestRewardSpellBookNilCreatedObjectSkipsStore4F09F0(t *testing.T) {
	hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
	hooks.createObjectByType = func(typeName string) *rewardSpellBookTestObject4F09F0 {
		return nil
	}
	hooks.storeSpell = func(*rewardSpellBookTestObject4F09F0, uint8) {
		t.Fatal("nil created object reached spell store")
	}
	if got := rewardSpellBookCreate4F09F0(7, hooks); got != nil {
		t.Fatalf("nil creation result = %#v", got)
	}
}

func TestRewardSpellBookFaultPrefixes4F09F0(t *testing.T) {
	tests := []struct {
		name string
		run  func(*[]string)
		want []string
	}{
		{
			name: "nil marker",
			run: func(events *[]string) {
				hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
				hooks.loadInitData = func(marker *rewardSpellBookTestMarker4F09F0) *rewardSpellBookTestData4F09F0 {
					*events = append(*events, "init")
					return marker.data
				}
				rewardSpellBook4F09F0(nil, 0, hooks)
			},
			want: []string{"init"},
		},
		{
			name: "nil InitData",
			run: func(events *[]string) {
				hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
				hooks.loadInitData = func(*rewardSpellBookTestMarker4F09F0) *rewardSpellBookTestData4F09F0 {
					*events = append(*events, "init")
					return nil
				}
				hooks.loadFlags = func(data *rewardSpellBookTestData4F09F0) uint8 {
					*events = append(*events, "flags")
					return data.flags
				}
				rewardSpellBook4F09F0(&rewardSpellBookTestMarker4F09F0{}, 0, hooks)
			},
			want: []string{"init", "flags"},
		},
		{
			name: "nil automatic slot selector",
			run: func(events *[]string) {
				hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
				hooks.loadInitData = func(marker *rewardSpellBookTestMarker4F09F0) *rewardSpellBookTestData4F09F0 {
					*events = append(*events, "init")
					return marker.data
				}
				hooks.loadFlags = func(data *rewardSpellBookTestData4F09F0) uint8 {
					*events = append(*events, "flags")
					return data.flags
				}
				hooks.pickSlots = nil
				rewardSpellBook4F09F0(
					&rewardSpellBookTestMarker4F09F0{data: &rewardSpellBookTestData4F09F0{}}, 0, hooks,
				)
			},
			want: []string{"init", "flags"},
		},
		{
			name: "nil class checker after selection",
			run: func(events *[]string) {
				data := &rewardSpellBookTestData4F09F0{flags: 1}
				data.spells[7] = 1
				hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
				hooks.loadInitData = func(marker *rewardSpellBookTestMarker4F09F0) *rewardSpellBookTestData4F09F0 {
					*events = append(*events, "init")
					return marker.data
				}
				hooks.randomInt = func(int32, int32) int32 {
					*events = append(*events, "rng")
					return 0
				}
				hooks.checkSpellClass = nil
				rewardSpellBook4F09F0(&rewardSpellBookTestMarker4F09F0{data: data}, 0, hooks)
			},
			want: []string{"init", "rng"},
		},
		{
			name: "nil store after creation",
			run: func(events *[]string) {
				hooks := rewardSpellBookTestHooks4F09F0(rewardSpellDefinitions4F09F0[:])
				hooks.checkSpellClass = func(class, spell uint32) int32 {
					*events = append(*events, fmt.Sprintf("class:%d", class))
					return 0
				}
				hooks.createObjectByType = func(typeName string) *rewardSpellBookTestObject4F09F0 {
					*events = append(*events, "create:"+typeName)
					return &rewardSpellBookTestObject4F09F0{}
				}
				hooks.storeSpell = nil
				rewardSpellBookCreate4F09F0(7, hooks)
			},
			want: []string{"class:1", "class:2", "create:CommonSpellBook"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var events []string
			defer func() {
				if recover() == nil {
					t.Fatal("expected fault did not occur")
				}
				if !slices.Equal(events, test.want) {
					t.Fatalf("fault prefix = %q, want %q", events, test.want)
				}
			}()
			test.run(&events)
		})
	}
}
