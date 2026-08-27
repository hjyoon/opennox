package server

import (
	"fmt"
	"math"
	"testing"
	"unsafe"
)

type questSpellAllowedTestRow4F2E70 struct {
	spellID uint32
	slots   uint32
}

func questSpellAllowedTestHooks4F2E70(
	rows []questSpellAllowedTestRow4F2E70,
	trace *[]string,
) questSpellAllowedHooks4F2E70 {
	return questSpellAllowedHooks4F2E70{
		loadSpellID: func(index int) uint32 {
			if trace != nil {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
			}
			return rows[index].spellID
		},
		loadSlots: func(index int) uint32 {
			if trace != nil {
				*trace = append(*trace, fmt.Sprintf("slots:%d", index))
			}
			return rows[index].slots
		},
	}
}

func TestQuestSpellAllowedTableLayout4F2E70(t *testing.T) {
	row := rewardSpellDefinition4F09F0{}
	if got := unsafe.Sizeof(row); got != 12 {
		t.Fatalf("reward-spell row size = %d, want 12", got)
	}
	if weight, spellID, slots := unsafe.Offsetof(row.Weight), unsafe.Offsetof(row.SpellID), unsafe.Offsetof(row.Slots); weight != 0 || spellID != 4 || slots != 8 {
		t.Fatalf("reward-spell row offsets = %d/%d/%d, want 0/4/8", weight, spellID, slots)
	}
	if got := unsafe.Sizeof(rewardSpellDefinitions4F09F0); got != 684 {
		t.Fatalf("reward-spell table size = %d, want 684", got)
	}
}

func TestQuestSpellAllowedMatchesGAMEEXE4F2E70(t *testing.T) {
	tableAllowed := map[int32]bool{
		1: true, 4: true, 5: true, 8: true, 9: true, 10: true,
		12: true, 13: true, 14: true, 16: true, 21: true, 22: true,
		23: true, 24: true, 26: true, 27: true, 29: true, 34: true,
		35: true, 36: true, 37: true, 38: true, 39: true, 41: true,
		42: true, 43: true, 50: true, 51: true, 52: true, 54: true,
		58: true, 60: true, 61: true, 62: true, 64: true, 67: true,
		71: true, 72: true, 74: true, 128: true, 129: true, 130: true,
		132: true, 134: true, 135: true, 136: true,
	}
	for spellID := int32(-3); spellID <= 140; spellID++ {
		want := tableAllowed[spellID] || questSpellExplicitAllowed4F2E70(spellID) ||
			(spellID >= 75 && spellID <= 114)
		var wantValue int32
		if want {
			wantValue = 1
		}
		if got := QuestSpellAllowed4F2E70(spellID); got != wantValue {
			t.Fatalf("spell %d admission = %d, want %d", spellID, got, wantValue)
		}
	}
}

func TestQuestSpellAllowedSignedBoundaries4F2E70(t *testing.T) {
	tests := []struct {
		spellID int32
		want    int32
	}{
		{math.MinInt32, 0}, {-1, 0}, {0, 0},
		{45, 0}, {46, 1}, {49, 1}, {50, 1},
		{74, 1}, {75, 1}, {114, 1}, {115, 0},
		{121, 0}, {122, 1}, {125, 1}, {126, 0},
		{math.MaxInt32, 0},
	}
	for _, test := range tests {
		if got := QuestSpellAllowed4F2E70(test.spellID); got != test.want {
			t.Errorf("spell %d admission = %d, want %d", test.spellID, got, test.want)
		}
	}
}

func TestQuestSpellAllowedScansLiveRowsAndShortCircuitsSlots4F2E70(t *testing.T) {
	rows := []questSpellAllowedTestRow4F2E70{
		{spellID: 7},
		{spellID: 8, slots: 2},
		{spellID: 7, slots: 4},
		{},
	}
	var trace []string
	got := questSpellAllowed4F2E70(7, questSpellAllowedTestHooks4F2E70(rows, &trace))
	if got != 1 {
		t.Fatalf("admission = %d, want 1", got)
	}
	want := []string{"id:0", "slots:0", "id:1", "id:2", "slots:2"}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}

	trace = nil
	got = questSpellAllowed4F2E70(6, questSpellAllowedTestHooks4F2E70(rows, &trace))
	if got != 0 {
		t.Fatalf("missing admission = %d, want 0", got)
	}
	want = []string{"id:0", "id:1", "id:2", "id:3"}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("missing trace = %v, want %v", trace, want)
	}
}

func TestQuestSpellAllowedFallbacksRunAfterTableScan4F2E70(t *testing.T) {
	rows := []questSpellAllowedTestRow4F2E70{{spellID: 7}, {}}
	for _, spellID := range []int32{46, 122, 75, 114} {
		var trace []string
		if got := questSpellAllowed4F2E70(spellID, questSpellAllowedTestHooks4F2E70(rows, &trace)); got != 1 {
			t.Fatalf("fallback spell %d admission = %d, want 1", spellID, got)
		}
		if want := []string{"id:0", "id:1"}; fmt.Sprint(trace) != fmt.Sprint(want) {
			t.Fatalf("fallback spell %d trace = %v, want %v", spellID, trace, want)
		}
	}

	rows = []questSpellAllowedTestRow4F2E70{{}}
	var trace []string
	if got := questSpellAllowed4F2E70(75, questSpellAllowedTestHooks4F2E70(rows, &trace)); got != 1 || fmt.Sprint(trace) != "[id:0]" {
		t.Fatalf("first-sentinel range result/trace = %d/%v, want 1/[id:0]", got, trace)
	}
}

func TestQuestSpellAllowedUsesRawTargetBits4F2E70(t *testing.T) {
	rows := []questSpellAllowedTestRow4F2E70{{spellID: math.MaxUint32, slots: 1}, {}}
	if got := questSpellAllowed4F2E70(-1, questSpellAllowedTestHooks4F2E70(rows, nil)); got != 1 {
		t.Fatalf("raw -1 target admission = %d, want 1", got)
	}
	rows[0].slots = 0
	if got := questSpellAllowed4F2E70(-1, questSpellAllowedTestHooks4F2E70(rows, nil)); got != 0 {
		t.Fatalf("zero-slot raw -1 target admission = %d, want 0", got)
	}
}

func TestQuestSpellAllowedFaultPrefixes4F2E70(t *testing.T) {
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

	fault("first ID before explicit fallback", []string{"id:0"}, func(trace *[]string) {
		questSpellAllowed4F2E70(46, questSpellAllowedHooks4F2E70{
			loadSpellID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				panic("fault")
			},
		})
	})

	fault("matching slots", []string{"id:0", "slots:0"}, func(trace *[]string) {
		questSpellAllowed4F2E70(7, questSpellAllowedHooks4F2E70{
			loadSpellID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				return 7
			},
			loadSlots: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("slots:%d", index))
				panic("fault")
			},
		})
	})

	fault("next ID", []string{"id:0", "id:1"}, func(trace *[]string) {
		questSpellAllowed4F2E70(7, questSpellAllowedHooks4F2E70{
			loadSpellID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				if index == 0 {
					return 8
				}
				panic("fault")
			},
			loadSlots: func(int) uint32 {
				panic("unexpected slots load")
			},
		})
	})
}
