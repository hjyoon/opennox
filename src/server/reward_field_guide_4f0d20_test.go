package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"
)

type rewardFieldGuideTestData4F0D20 struct {
	flags  uint8
	guides [rewardFieldGuideCount4F0D20]uint8
}

type rewardFieldGuideTestMarker4F0D20 struct {
	data *rewardFieldGuideTestData4F0D20
}

type rewardFieldGuideTestUse4F0D20 struct {
	name string
}

type rewardFieldGuideTestObject4F0D20 struct {
	typeName string
	useData  *rewardFieldGuideTestUse4F0D20
}

func rewardFieldGuideTestHooks4F0D20(rows []rewardFieldGuideDefinition4F0D20) rewardFieldGuideHooks4F0D20[
	*rewardFieldGuideTestMarker4F0D20,
	*rewardFieldGuideTestData4F0D20,
	[]rewardFieldGuideDefinition4F0D20,
	*rewardFieldGuideTestObject4F0D20,
	*rewardFieldGuideTestUse4F0D20,
] {
	return rewardFieldGuideHooks4F0D20[
		*rewardFieldGuideTestMarker4F0D20,
		*rewardFieldGuideTestData4F0D20,
		[]rewardFieldGuideDefinition4F0D20,
		*rewardFieldGuideTestObject4F0D20,
		*rewardFieldGuideTestUse4F0D20,
	]{
		loadInitData: func(marker *rewardFieldGuideTestMarker4F0D20) *rewardFieldGuideTestData4F0D20 {
			return marker.data
		},
		loadFlags: func(data *rewardFieldGuideTestData4F0D20) uint8 {
			return data.flags
		},
		loadExplicitGuide: func(data *rewardFieldGuideTestData4F0D20, index int) uint8 {
			return data.guides[index]
		},
		pickSlots: func(uint32) uint32 { return 1 },
		rows:      rows,
		loadRowWeight: func(rows []rewardFieldGuideDefinition4F0D20, index int) uint8 {
			return rows[index].Weight
		},
		loadRowGuideID: func(rows []rewardFieldGuideDefinition4F0D20, index int) uint32 {
			return rows[index].GuideID
		},
		loadRowSlots: func(rows []rewardFieldGuideDefinition4F0D20, index int) uint32 {
			return rows[index].Slots
		},
		randomInt: func(int32, int32) int32 { return 0 },
		createObjectByType: func(typeName string) *rewardFieldGuideTestObject4F0D20 {
			return &rewardFieldGuideTestObject4F0D20{
				typeName: typeName,
				useData:  &rewardFieldGuideTestUse4F0D20{},
			}
		},
		isNilObject: func(object *rewardFieldGuideTestObject4F0D20) bool {
			return object == nil
		},
		loadUseData: func(object *rewardFieldGuideTestObject4F0D20) *rewardFieldGuideTestUse4F0D20 {
			return object.useData
		},
		guideName: func(guideID uint32) string {
			return rewardFieldGuideNames4F0D20[guideID]
		},
		storeGuide: func(useData *rewardFieldGuideTestUse4F0D20, name string) {
			useData.name = name
		},
	}
}

func TestRewardFieldGuideDataMatchesGAMEEXE4F0D20(t *testing.T) {
	if got := unsafe.Sizeof(rewardFieldGuideDefinitions4F0D20); got != 384 {
		t.Fatalf("reward field-guide table size = %d, want 384", got)
	}
	if unsafe.Sizeof(rewardFieldGuideDefinition4F0D20{}) != 12 ||
		unsafe.Offsetof(rewardFieldGuideDefinition4F0D20{}.Weight) != 0 ||
		unsafe.Offsetof(rewardFieldGuideDefinition4F0D20{}.GuideID) != 4 ||
		unsafe.Offsetof(rewardFieldGuideDefinition4F0D20{}.Slots) != 8 {
		t.Fatalf("reward field-guide row layout = size %d offsets %d/%d/%d, want 12 and 0/4/8",
			unsafe.Sizeof(rewardFieldGuideDefinition4F0D20{}),
			unsafe.Offsetof(rewardFieldGuideDefinition4F0D20{}.Weight),
			unsafe.Offsetof(rewardFieldGuideDefinition4F0D20{}.GuideID),
			unsafe.Offsetof(rewardFieldGuideDefinition4F0D20{}.Slots),
		)
	}
	raw := unsafe.Slice(
		(*byte)(unsafe.Pointer(&rewardFieldGuideDefinitions4F0D20[0])),
		int(unsafe.Sizeof(rewardFieldGuideDefinitions4F0D20)),
	)
	if got, want := fmt.Sprintf("%x", sha256.Sum256(raw)), "2e1f41cb42b7594cc505480b3ebf3358be40df43051dcf7d2e9dafdf316fb694"; got != want {
		t.Fatalf("raw reward field-guide table SHA-256 = %s, want %s", got, want)
	}
	var semantic bytes.Buffer
	var slotTotals [5]uint32
	weighted := 0
	for _, row := range rewardFieldGuideDefinitions4F0D20 {
		semantic.WriteByte(row.Weight)
		if err := binary.Write(&semantic, binary.LittleEndian, row.GuideID); err != nil {
			t.Fatal(err)
		}
		if err := binary.Write(&semantic, binary.LittleEndian, row.Slots); err != nil {
			t.Fatal(err)
		}
		if row.GuideID != 0 {
			weighted++
		}
		for index := range slotTotals {
			if row.GuideID != 0 && row.Slots&(uint32(1)<<index) != 0 {
				slotTotals[index] += uint32(row.Weight)
			}
		}
	}
	if semantic.Len() != 288 {
		t.Fatalf("semantic reward field-guide table size = %d, want 288", semantic.Len())
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(semantic.Bytes())), "00a23aa9dd119ac1f2348086a646199fcc635d600728c17b16899af2ca365248"; got != want {
		t.Fatalf("semantic reward field-guide table SHA-256 = %s, want %s", got, want)
	}
	if weighted != 31 || slotTotals != [5]uint32{32, 64, 88, 108, 114} {
		t.Fatalf("weighted rows/totals = %d/%v, want 31/[32 64 88 108 114]", weighted, slotTotals)
	}
	if got := rewardFieldGuideDefinitions4F0D20[len(rewardFieldGuideDefinitions4F0D20)-1]; got != (rewardFieldGuideDefinition4F0D20{Slots: 0x1f}) {
		t.Fatalf("sentinel = %#v, want zero ID with slots 0x1f", got)
	}

	var names bytes.Buffer
	for _, name := range rewardFieldGuideNames4F0D20 {
		names.WriteString(name)
		names.WriteByte(0)
	}
	if len(rewardFieldGuideNames4F0D20) != 41 || names.Len() != 394 {
		t.Fatalf("guide names count/semantic bytes = %d/%d, want 41/394", len(rewardFieldGuideNames4F0D20), names.Len())
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(names.Bytes())), "fcc0dc43e9f61b049480cd9faa342063b559eed612cd40ed7387ec71db456575"; got != want {
		t.Fatalf("semantic guide-name SHA-256 = %s, want %s", got, want)
	}
}

func TestRewardFieldGuideExplicitEveryID4F0D20(t *testing.T) {
	for guideID := 0; guideID < rewardFieldGuideCount4F0D20; guideID++ {
		data := &rewardFieldGuideTestData4F0D20{flags: rewardFieldGuideExplicitFlag4F0D20}
		data.guides[guideID] = 1
		hooks := rewardFieldGuideTestHooks4F0D20(nil)
		hooks.pickSlots = func(uint32) uint32 {
			t.Fatal("explicit path called slot selector")
			return 0
		}
		rngCalls := 0
		hooks.randomInt = func(minimum, maximum int32) int32 {
			rngCalls++
			if minimum != 0 || maximum != 0 {
				t.Fatalf("guide %d RNG bounds = %d..%d, want 0..0", guideID, minimum, maximum)
			}
			return 0
		}
		got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: data}, 0xfeedbeef, hooks)
		if guideID == 0 {
			if got != nil || rngCalls != 1 {
				t.Fatalf("guide zero result/RNG = %#v/%d, want nil/1", got, rngCalls)
			}
			continue
		}
		if got == nil || got.typeName != rewardFieldGuideType4F0D20 || got.useData.name != rewardFieldGuideNames4F0D20[guideID] || rngCalls != 1 {
			t.Fatalf("guide %d result/RNG = %#v/%d", guideID, got, rngCalls)
		}
	}
}

func TestRewardFieldGuideExplicitCachedDataAndLivePass4F0D20(t *testing.T) {
	entry := &rewardFieldGuideTestData4F0D20{flags: rewardFieldGuideExplicitFlag4F0D20}
	entry.guides[2], entry.guides[5], entry.guides[40] = 1, 1, 1
	entry.guides[4] = 2
	replacement := &rewardFieldGuideTestData4F0D20{}
	marker := &rewardFieldGuideTestMarker4F0D20{data: entry}
	hooks := rewardFieldGuideTestHooks4F0D20(nil)
	reads := 0
	hooks.loadFlags = func(data *rewardFieldGuideTestData4F0D20) uint8 {
		if data != entry {
			t.Fatal("flags did not use entry-cached InitData")
		}
		marker.data = replacement
		return data.flags
	}
	hooks.loadExplicitGuide = func(data *rewardFieldGuideTestData4F0D20, index int) uint8 {
		if data != entry {
			t.Fatal("guide read did not use entry-cached InitData")
		}
		reads++
		return data.guides[index]
	}
	hooks.randomInt = func(minimum, maximum int32) int32 {
		if minimum != 0 || maximum != 2 || reads != 41 {
			t.Fatalf("RNG bounds/read prefix = %d..%d/%d, want 0..2/41", minimum, maximum, reads)
		}
		entry.guides[5] = 0
		entry.guides[7] = 1
		return 1
	}
	got := rewardFieldGuide4F0D20(marker, 0, hooks)
	if got == nil || got.useData.name != rewardFieldGuideNames4F0D20[7] {
		t.Fatalf("result = %#v, want live second-pass guide 7", got)
	}
	if reads != 49 || marker.data != replacement {
		t.Fatalf("guide reads/live marker = %d/%p, want 49/replacement", reads, marker.data)
	}
}

func TestRewardFieldGuideExplicitExactOneAndExhaustion4F0D20(t *testing.T) {
	data := &rewardFieldGuideTestData4F0D20{flags: rewardFieldGuideExplicitFlag4F0D20}
	data.guides[1], data.guides[20], data.guides[40] = 2, 0xff, 3
	hooks := rewardFieldGuideTestHooks4F0D20(nil)
	hooks.randomInt = func(int32, int32) int32 {
		t.Fatal("non-one bytes reached RNG")
		return 0
	}
	if got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: data}, 0, hooks); got != nil {
		t.Fatalf("non-one result = %#v, want nil", got)
	}

	data.guides[2], data.guides[8] = 1, 1
	hooks = rewardFieldGuideTestHooks4F0D20(nil)
	reads := 0
	hooks.loadExplicitGuide = func(data *rewardFieldGuideTestData4F0D20, index int) uint8 {
		reads++
		return data.guides[index]
	}
	hooks.randomInt = func(int32, int32) int32 {
		data.guides[2], data.guides[8] = 0, 0
		return 1
	}
	if got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: data}, 0, hooks); got != nil || reads != 82 {
		t.Fatalf("exhausted result/reads = %#v/%d, want nil/82", got, reads)
	}
}

func TestRewardFieldGuideAutomaticEveryWeightedRow4F0D20(t *testing.T) {
	wantTotals := [...]int32{32, 64, 88, 108, 114}
	for slotIndex, wantTotal := range wantTotals {
		slots := uint32(1) << slotIndex
		cumulative := int32(0)
		for rowIndex, row := range rewardFieldGuideDefinitions4F0D20 {
			if row.GuideID == 0 {
				break
			}
			if slots&row.Slots == 0 {
				continue
			}
			draw := cumulative
			cumulative += int32(row.Weight)
			hooks := rewardFieldGuideTestHooks4F0D20(rewardFieldGuideDefinitions4F0D20[:])
			hooks.pickSlots = func(stage uint32) uint32 {
				if stage != uint32(slotIndex) {
					t.Fatalf("row %d stage = %d, want %d", rowIndex, stage, slotIndex)
				}
				return slots
			}
			hooks.randomInt = func(minimum, maximum int32) int32 {
				if minimum != 0 || maximum != wantTotal-1 {
					t.Fatalf("slot %d row %d RNG bounds = %d..%d, want 0..%d", slotIndex, rowIndex, minimum, maximum, wantTotal-1)
				}
				return draw
			}
			got := rewardFieldGuide4F0D20(
				&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}},
				uint32(slotIndex),
				hooks,
			)
			if got == nil || got.useData.name != rewardFieldGuideNames4F0D20[row.GuideID] {
				t.Fatalf("slot %d row %d result = %#v, want guide %d", slotIndex, rowIndex, got, row.GuideID)
			}
		}
		if cumulative != wantTotal {
			t.Fatalf("slot %d cumulative = %d, want %d", slotIndex, cumulative, wantTotal)
		}
	}
}

func TestRewardFieldGuideAutomaticLiveReloadAndOrder4F0D20(t *testing.T) {
	rows := []rewardFieldGuideDefinition4F0D20{
		{Weight: 2, GuideID: 11, Slots: 1},
		{Weight: 3, GuideID: 22, Slots: 1},
		{Slots: 0x1f},
	}
	hooks := rewardFieldGuideTestHooks4F0D20(rows)
	var events []string
	hooks.pickSlots = func(stage uint32) uint32 {
		events = append(events, "slots")
		if stage != 0xfeedbeef {
			t.Fatalf("stage = %#x", stage)
		}
		return 1
	}
	hooks.loadRowGuideID = func(rows []rewardFieldGuideDefinition4F0D20, index int) uint32 {
		events = append(events, fmt.Sprintf("id:%d", index))
		return rows[index].GuideID
	}
	hooks.loadRowSlots = func(rows []rewardFieldGuideDefinition4F0D20, index int) uint32 {
		events = append(events, fmt.Sprintf("row-slots:%d", index))
		return rows[index].Slots
	}
	hooks.loadRowWeight = func(rows []rewardFieldGuideDefinition4F0D20, index int) uint8 {
		events = append(events, fmt.Sprintf("weight:%d", index))
		return rows[index].Weight
	}
	hooks.randomInt = func(minimum, maximum int32) int32 {
		events = append(events, fmt.Sprintf("rng:%d:%d", minimum, maximum))
		rows[0].Weight = 5
		rows[0].GuideID = 35
		return 3
	}
	hooks.createObjectByType = func(typeName string) *rewardFieldGuideTestObject4F0D20 {
		events = append(events, "create:"+typeName)
		return &rewardFieldGuideTestObject4F0D20{typeName: typeName, useData: &rewardFieldGuideTestUse4F0D20{}}
	}
	hooks.isNilObject = func(object *rewardFieldGuideTestObject4F0D20) bool {
		events = append(events, "nil")
		return object == nil
	}
	hooks.loadUseData = func(object *rewardFieldGuideTestObject4F0D20) *rewardFieldGuideTestUse4F0D20 {
		events = append(events, "use")
		return object.useData
	}
	hooks.guideName = func(guideID uint32) string {
		events = append(events, fmt.Sprintf("name:%d", guideID))
		return rewardFieldGuideNames4F0D20[guideID]
	}
	hooks.storeGuide = func(useData *rewardFieldGuideTestUse4F0D20, name string) {
		events = append(events, "store:"+name)
		useData.name = name
	}
	got := rewardFieldGuide4F0D20(
		&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}},
		0xfeedbeef,
		hooks,
	)
	wantEvents := []string{
		"slots", "id:0", "row-slots:0", "weight:0", "id:1", "row-slots:1", "weight:1", "id:2",
		"rng:0:4", "id:0", "row-slots:0", "weight:0", "id:0",
		"create:FieldGuide", "nil", "use", "name:35", "store:Zombie",
	}
	if got == nil || got.useData.name != "Zombie" || !slices.Equal(events, wantEvents) {
		t.Fatalf("result/events = %#v/%v, want Zombie/%v", got, events, wantEvents)
	}
}

func TestRewardFieldGuideAutomaticZeroBranchesAndSignedMath4F0D20(t *testing.T) {
	if got := rewardFieldGuideAddWeight4F0D20(math.MaxInt32, 1); got != math.MinInt32 {
		t.Fatalf("wrapping add = %d, want MinInt32", got)
	}
	if got := rewardFieldGuideAddWeight4F0D20(-1, 1); got != 0 {
		t.Fatalf("wrapping add -1+1 = %d, want 0", got)
	}

	t.Run("zero first ID", func(t *testing.T) {
		hooks := rewardFieldGuideTestHooks4F0D20([]rewardFieldGuideDefinition4F0D20{{Slots: 1}})
		hooks.randomInt = func(int32, int32) int32 {
			t.Fatal("zero first ID reached RNG")
			return 0
		}
		if got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}}, 0, hooks); got != nil {
			t.Fatalf("zero first ID result = %#v", got)
		}
	})

	t.Run("zero total", func(t *testing.T) {
		rows := []rewardFieldGuideDefinition4F0D20{{Weight: 5, GuideID: 1, Slots: 2}, {Slots: 0x1f}}
		hooks := rewardFieldGuideTestHooks4F0D20(rows)
		hooks.randomInt = func(int32, int32) int32 {
			t.Fatal("zero total reached RNG")
			return 0
		}
		if got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}}, 0, hooks); got != nil {
			t.Fatalf("zero total result = %#v", got)
		}
	})

	t.Run("first ID becomes zero after RNG", func(t *testing.T) {
		rows := []rewardFieldGuideDefinition4F0D20{{Weight: 1, GuideID: 1, Slots: 1}, {Slots: 0x1f}}
		hooks := rewardFieldGuideTestHooks4F0D20(rows)
		hooks.randomInt = func(int32, int32) int32 {
			rows[0].GuideID = 0
			return 0
		}
		if got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}}, 0, hooks); got != nil {
			t.Fatalf("post-RNG zero ID result = %#v", got)
		}
	})

	t.Run("second pass exhausts", func(t *testing.T) {
		rows := []rewardFieldGuideDefinition4F0D20{{Weight: 1, GuideID: 1, Slots: 1}, {Weight: 1, GuideID: 2, Slots: 1}, {Slots: 0x1f}}
		hooks := rewardFieldGuideTestHooks4F0D20(rows)
		hooks.randomInt = func(int32, int32) int32 {
			rows[0].Slots, rows[1].Slots = 0, 0
			return 1
		}
		if got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}}, 0, hooks); got != nil {
			t.Fatalf("exhausted second pass result = %#v", got)
		}
	})

	t.Run("signed negative draw selects first row", func(t *testing.T) {
		rows := []rewardFieldGuideDefinition4F0D20{{Weight: 1, GuideID: 1, Slots: 1}, {Slots: 0x1f}}
		hooks := rewardFieldGuideTestHooks4F0D20(rows)
		hooks.randomInt = func(int32, int32) int32 { return -1 }
		got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}}, 0, hooks)
		if got == nil || got.useData.name != "Bat" {
			t.Fatalf("negative draw result = %#v, want Bat", got)
		}
	})
}

func TestRewardFieldGuideCreationOrderNilAndFaultPrefixes4F0D20(t *testing.T) {
	rows := []rewardFieldGuideDefinition4F0D20{{Weight: 1, GuideID: 1, Slots: 1}, {Slots: 0x1f}}
	t.Run("nil object", func(t *testing.T) {
		hooks := rewardFieldGuideTestHooks4F0D20(rows)
		var events []string
		hooks.createObjectByType = func(typeName string) *rewardFieldGuideTestObject4F0D20 {
			events = append(events, "create:"+typeName)
			return nil
		}
		hooks.isNilObject = func(object *rewardFieldGuideTestObject4F0D20) bool {
			events = append(events, "nil")
			return object == nil
		}
		hooks.loadUseData = func(*rewardFieldGuideTestObject4F0D20) *rewardFieldGuideTestUse4F0D20 {
			t.Fatal("nil object reached use-data load")
			return nil
		}
		got := rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}}, 0, hooks)
		if got != nil || !slices.Equal(events, []string{"create:FieldGuide", "nil"}) {
			t.Fatalf("nil result/events = %#v/%v", got, events)
		}
	})

	t.Run("invalid name faults after use-data load", func(t *testing.T) {
		invalidRows := []rewardFieldGuideDefinition4F0D20{{Weight: 1, GuideID: 41, Slots: 1}, {Slots: 0x1f}}
		hooks := rewardFieldGuideTestHooks4F0D20(invalidRows)
		var events []string
		hooks.loadUseData = func(object *rewardFieldGuideTestObject4F0D20) *rewardFieldGuideTestUse4F0D20 {
			events = append(events, "use")
			return object.useData
		}
		hooks.guideName = func(guideID uint32) string {
			events = append(events, fmt.Sprintf("name:%d", guideID))
			return rewardFieldGuideNames4F0D20[guideID]
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid guide name did not panic")
				}
			}()
			_ = rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}}, 0, hooks)
		}()
		if !slices.Equal(events, []string{"use", "name:41"}) {
			t.Fatalf("invalid-name fault prefix = %v, want [use name:41]", events)
		}
	})

	t.Run("nil use data faults after name lookup", func(t *testing.T) {
		hooks := rewardFieldGuideTestHooks4F0D20(rows)
		var events []string
		hooks.loadUseData = func(*rewardFieldGuideTestObject4F0D20) *rewardFieldGuideTestUse4F0D20 {
			events = append(events, "use")
			return nil
		}
		hooks.guideName = func(guideID uint32) string {
			events = append(events, "name")
			return rewardFieldGuideNames4F0D20[guideID]
		}
		hooks.storeGuide = func(useData *rewardFieldGuideTestUse4F0D20, name string) {
			events = append(events, "store:"+name)
			useData.name = name
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("nil use data did not panic")
				}
			}()
			_ = rewardFieldGuide4F0D20(&rewardFieldGuideTestMarker4F0D20{data: &rewardFieldGuideTestData4F0D20{}}, 0, hooks)
		}()
		if !slices.Equal(events, []string{"use", "name", "store:Bat"}) {
			t.Fatalf("nil-use fault prefix = %v, want [use name store:Bat]", events)
		}
	})
}
