package server

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"
)

type questFieldGuideAllowedTestRow4F2EF0 struct {
	guideID uint32
	slots   uint32
}

func questFieldGuideAllowedTestHooks4F2EF0(
	rows []questFieldGuideAllowedTestRow4F2EF0,
	families []*questFieldGuideFamily4F2EF0,
	trace *[]string,
) questFieldGuideAllowedHooks4F2EF0 {
	familyIndexes := make(map[*questFieldGuideFamily4F2EF0]int, len(families))
	for index, family := range families {
		if family != nil {
			familyIndexes[family] = index
		}
	}
	return questFieldGuideAllowedHooks4F2EF0{
		loadRewardGuideID: func(index int) uint32 {
			if trace != nil {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
			}
			return rows[index].guideID
		},
		loadRewardSlots: func(index int) uint32 {
			if trace != nil {
				*trace = append(*trace, fmt.Sprintf("slots:%d", index))
			}
			return rows[index].slots
		},
		loadFamily: func(index int) *questFieldGuideFamily4F2EF0 {
			if trace != nil {
				*trace = append(*trace, fmt.Sprintf("family:%d", index))
			}
			return families[index]
		},
		loadFamilyValue: func(family *questFieldGuideFamily4F2EF0, index int) uint32 {
			if trace != nil {
				*trace = append(*trace, fmt.Sprintf("value:%d:%d", familyIndexes[family], index))
			}
			return family[index]
		},
	}
}

func TestQuestFieldGuideAllowedNativeData4F2EF0(t *testing.T) {
	row := rewardFieldGuideDefinition4F0D20{}
	if got := unsafe.Sizeof(row); got != 12 {
		t.Fatalf("reward-guide row size = %d, want 12", got)
	}
	if weight, guideID, slots := unsafe.Offsetof(row.Weight), unsafe.Offsetof(row.GuideID), unsafe.Offsetof(row.Slots); weight != 0 || guideID != 4 || slots != 8 {
		t.Fatalf("reward-guide row offsets = %d/%d/%d, want 0/4/8", weight, guideID, slots)
	}
	if got := unsafe.Sizeof(questFieldGuideFamily24Data4F2EF0); got != 24 {
		t.Fatalf("guide-family size = %d, want 24", got)
	}
	if got, want := questFieldGuideFamily24Data4F2EF0, (questFieldGuideFamily4F2EF0{24, 7, 8, 25, 26, 0}); got != want {
		t.Fatalf("guide-family values = %v, want %v", got, want)
	}
	if got := unsafe.Sizeof(questFieldGuideFamilies4F2EF0); got != 2*unsafe.Sizeof(uintptr(0)) {
		t.Fatalf("family pointer-table size = %d, want %d", got, 2*unsafe.Sizeof(uintptr(0)))
	}
	if questFieldGuideFamilies4F2EF0[0] != &questFieldGuideFamily24Data4F2EF0 || questFieldGuideFamilies4F2EF0[1] != nil {
		t.Fatalf("family pointer table = %v, want family-24/null", questFieldGuideFamilies4F2EF0)
	}

	raw := make([]byte, 4*len(questFieldGuideFamily24Data4F2EF0))
	for index, value := range questFieldGuideFamily24Data4F2EF0 {
		binary.LittleEndian.PutUint32(raw[4*index:], value)
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(raw)), "ec40c0cd864bc64fa96e3d1954e5f5a69a7f1bd2dfed7d05e2c98fb593af9d4f"; got != want {
		t.Fatalf("normalized family SHA-256 = %s, want %s", got, want)
	}
}

func TestQuestFieldGuideAllowedMatchesGAMEEXE4F2EF0(t *testing.T) {
	allowed := map[int32]bool{
		0: true, 2: true, 3: true, 4: true, 7: true, 8: true,
		9: true, 10: true, 11: true, 13: true, 14: true, 15: true,
		16: true, 17: true, 18: true, 19: true, 20: true, 21: true,
		22: true, 23: true, 24: true, 25: true, 26: true, 27: true,
		29: true, 31: true, 32: true, 33: true, 34: true, 35: true,
		36: true, 37: true, 38: true, 39: true, 40: true,
	}
	for guideID := int32(-3); guideID <= 45; guideID++ {
		var want int32
		if allowed[guideID] {
			want = 1
		}
		if got := QuestFieldGuideAllowed4F2EF0(guideID); got != want {
			t.Fatalf("guide %d admission = %d, want %d", guideID, got, want)
		}
	}
	for _, guideID := range []int32{math.MinInt32, math.MaxInt32} {
		if got := QuestFieldGuideAllowed4F2EF0(guideID); got != 0 {
			t.Fatalf("guide %d admission = %d, want 0", guideID, got)
		}
	}
}

func TestQuestFieldGuideAllowedScansRewardRowsThenEveryFamily4F2EF0(t *testing.T) {
	rows := []questFieldGuideAllowedTestRow4F2EF0{
		{guideID: 7},
		{guideID: 8, slots: 2},
		{guideID: 7, slots: 4},
		{},
	}
	family := questFieldGuideFamily4F2EF0{24, 10, 11, 0}
	families := []*questFieldGuideFamily4F2EF0{&family, nil}
	var trace []string
	if got := questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedTestHooks4F2EF0(rows, families, &trace)); got != 1 {
		t.Fatalf("admission = %d, want 1", got)
	}
	want := []string{
		"id:0", "slots:0", "id:1", "id:2", "slots:2",
		"family:0", "value:0:0", "value:0:1", "value:0:2", "value:0:3", "family:1",
	}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}

	trace = nil
	if got := questFieldGuideAllowed4F2EF0(6, questFieldGuideAllowedTestHooks4F2EF0(rows, families, &trace)); got != 0 {
		t.Fatalf("missing admission = %d, want 0", got)
	}
	want = []string{
		"id:0", "id:1", "id:2", "id:3",
		"family:0", "value:0:0", "value:0:1", "value:0:2", "value:0:3", "family:1",
	}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("missing trace = %v, want %v", trace, want)
	}
}

func TestQuestFieldGuideAllowedFamilyHeaderAndSentinelOrder4F2EF0(t *testing.T) {
	rows := []questFieldGuideAllowedTestRow4F2EF0{{}}
	family := questFieldGuideFamily4F2EF0{24, 7, 0}
	families := []*questFieldGuideFamily4F2EF0{&family, nil}
	if got := questFieldGuideAllowed4F2EF0(24, questFieldGuideAllowedTestHooks4F2EF0(rows, families, nil)); got != 0 {
		t.Fatalf("header-only guide admission = %d, want 0", got)
	}
	if got := questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedTestHooks4F2EF0(rows, families, nil)); got != 1 {
		t.Fatalf("member guide admission = %d, want 1", got)
	}
	if got := questFieldGuideAllowed4F2EF0(0, questFieldGuideAllowedTestHooks4F2EF0(rows, families, nil)); got != 1 {
		t.Fatalf("zero terminator admission = %d, want 1", got)
	}

	emptyHeader := questFieldGuideFamily4F2EF0{0, 7, 0}
	if got := questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedTestHooks4F2EF0(rows, []*questFieldGuideFamily4F2EF0{&emptyHeader, nil}, nil)); got != 0 {
		t.Fatalf("zero-header family admission = %d, want 0", got)
	}
	if got := questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedTestHooks4F2EF0(rows, []*questFieldGuideFamily4F2EF0{nil, &family}, nil)); got != 0 {
		t.Fatalf("post-null family admission = %d, want 0", got)
	}
}

func TestQuestFieldGuideAllowedMemberMatchStillLoadsNextFamily4F2EF0(t *testing.T) {
	rows := []questFieldGuideAllowedTestRow4F2EF0{{}}
	first := questFieldGuideFamily4F2EF0{24, 7, 8, 0}
	second := questFieldGuideFamily4F2EF0{5, 9, 0}
	families := []*questFieldGuideFamily4F2EF0{&first, &second, nil}
	var trace []string
	if got := questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedTestHooks4F2EF0(rows, families, &trace)); got != 1 {
		t.Fatalf("admission = %d, want 1", got)
	}
	want := []string{
		"id:0", "family:0", "value:0:0", "value:0:1",
		"family:1", "value:1:0", "value:1:1", "value:1:2", "family:2",
	}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
}

func TestQuestFieldGuideAllowedUsesRawTargetBits4F2EF0(t *testing.T) {
	rows := []questFieldGuideAllowedTestRow4F2EF0{{guideID: math.MaxUint32, slots: 1}, {}}
	if got := questFieldGuideAllowed4F2EF0(-1, questFieldGuideAllowedTestHooks4F2EF0(rows, []*questFieldGuideFamily4F2EF0{nil}, nil)); got != 1 {
		t.Fatalf("raw reward target admission = %d, want 1", got)
	}
	rows[0].slots = 0
	family := questFieldGuideFamily4F2EF0{1, math.MaxUint32, 0}
	if got := questFieldGuideAllowed4F2EF0(-1, questFieldGuideAllowedTestHooks4F2EF0(rows, []*questFieldGuideFamily4F2EF0{&family, nil}, nil)); got != 1 {
		t.Fatalf("raw family target admission = %d, want 1", got)
	}
}

func TestQuestFieldGuideAllowedFaultPrefixes4F2EF0(t *testing.T) {
	fault := func(name string, want []string, run func(*[]string)) {
		t.Helper()
		var trace []string
		defer func() {
			if recover() == nil {
				t.Fatalf("%s did not fault", name)
			}
			if fmt.Sprint(trace) != fmt.Sprint(want) {
				t.Fatalf("%s trace = %v, want %v", name, trace, want)
			}
		}()
		run(&trace)
	}

	fault("first reward ID", []string{"id:0"}, func(trace *[]string) {
		questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedHooks4F2EF0{
			loadRewardGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				panic("fault")
			},
		})
	})

	fault("matching reward slots", []string{"id:0", "slots:0"}, func(trace *[]string) {
		questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedHooks4F2EF0{
			loadRewardGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				return 7
			},
			loadRewardSlots: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("slots:%d", index))
				panic("fault")
			},
		})
	})

	fault("next reward ID", []string{"id:0", "id:1"}, func(trace *[]string) {
		questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedHooks4F2EF0{
			loadRewardGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				if index == 0 {
					return 8
				}
				panic("fault")
			},
			loadRewardSlots: func(int) uint32 {
				panic("unexpected slots load")
			},
		})
	})

	fault("first family after reward match", []string{"id:0", "slots:0", "family:0"}, func(trace *[]string) {
		questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedHooks4F2EF0{
			loadRewardGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				return 7
			},
			loadRewardSlots: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("slots:%d", index))
				return 1
			},
			loadFamily: func(index int) *questFieldGuideFamily4F2EF0 {
				*trace = append(*trace, fmt.Sprintf("family:%d", index))
				panic("fault")
			},
		})
	})

	family := questFieldGuideFamily4F2EF0{24, 7, 0}
	fault("family header", []string{"id:0", "family:0", "value:0"}, func(trace *[]string) {
		questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedHooks4F2EF0{
			loadRewardGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				return 0
			},
			loadFamily: func(index int) *questFieldGuideFamily4F2EF0 {
				*trace = append(*trace, fmt.Sprintf("family:%d", index))
				return &family
			},
			loadFamilyValue: func(*questFieldGuideFamily4F2EF0, int) uint32 {
				*trace = append(*trace, "value:0")
				panic("fault")
			},
		})
	})

	fault("first family member", []string{"id:0", "family:0", "value:0", "value:1"}, func(trace *[]string) {
		questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedHooks4F2EF0{
			loadRewardGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				return 0
			},
			loadFamily: func(index int) *questFieldGuideFamily4F2EF0 {
				*trace = append(*trace, fmt.Sprintf("family:%d", index))
				return &family
			},
			loadFamilyValue: func(family *questFieldGuideFamily4F2EF0, index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("value:%d", index))
				if index == 0 {
					return family[index]
				}
				panic("fault")
			},
		})
	})

	fault("next family after member match", []string{"id:0", "family:0", "value:0", "value:1", "family:1"}, func(trace *[]string) {
		questFieldGuideAllowed4F2EF0(7, questFieldGuideAllowedHooks4F2EF0{
			loadRewardGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				return 0
			},
			loadFamily: func(index int) *questFieldGuideFamily4F2EF0 {
				*trace = append(*trace, fmt.Sprintf("family:%d", index))
				if index == 0 {
					return &family
				}
				panic("fault")
			},
			loadFamilyValue: func(family *questFieldGuideFamily4F2EF0, index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("value:%d", index))
				return family[index]
			},
		})
	})
}
