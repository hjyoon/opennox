package server

import (
	"fmt"
	"testing"
)

type randomSpellLossTestRow4F24E0 struct {
	spellID uint32
	slots   uint32
}

func randomSpellLossTestHooks4F24E0(
	rows []randomSpellLossTestRow4F24E0,
	trace *[]string,
) randomSpellLossEligibilityHooks4F24E0 {
	return randomSpellLossEligibilityHooks4F24E0{
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

func TestRandomSpellLossEligibilityMatchesGAMEEXETable4F24E0(t *testing.T) {
	eligible := map[int32]bool{
		1: true, 4: true, 5: true, 8: true, 10: true, 12: true,
		13: true, 14: true, 16: true, 21: true, 22: true, 23: true,
		24: true, 26: true, 29: true, 35: true, 36: true, 37: true,
		38: true, 39: true, 42: true, 43: true, 50: true, 51: true,
		52: true, 54: true, 58: true, 60: true, 61: true, 62: true,
		64: true, 67: true, 71: true, 72: true, 74: true, 128: true,
		129: true, 130: true, 132: true, 134: true, 135: true, 136: true,
	}
	for spellID := int32(-3); spellID <= 140; spellID++ {
		var want int32
		if eligible[spellID] {
			want = 1
		}
		if got := RandomSpellLossEligible4F24E0(spellID); got != want {
			t.Fatalf("spell %d eligibility = %d, want %d", spellID, got, want)
		}
	}
}

func TestRandomSpellLossEligibilityScansLiveRowsAndShortCircuitsSlots4F24E0(t *testing.T) {
	rows := []randomSpellLossTestRow4F24E0{
		{spellID: 7},
		{spellID: 8, slots: 2},
		{spellID: 7, slots: 4},
		{},
	}
	var trace []string
	got := randomSpellLossEligible4F24E0(7, randomSpellLossTestHooks4F24E0(rows, &trace))
	if got != 1 {
		t.Fatalf("eligibility = %d, want 1", got)
	}
	want := []string{"id:0", "slots:0", "id:1", "id:2", "slots:2"}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}

	trace = nil
	got = randomSpellLossEligible4F24E0(99, randomSpellLossTestHooks4F24E0(rows, &trace))
	if got != 0 {
		t.Fatalf("missing eligibility = %d, want 0", got)
	}
	want = []string{"id:0", "id:1", "id:2", "id:3"}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("missing trace = %v, want %v", trace, want)
	}
}

func TestRandomSpellLossEligibilityStopsAtFirstSentinel4F24E0(t *testing.T) {
	rows := []randomSpellLossTestRow4F24E0{{}, {spellID: 7, slots: 1}}
	var trace []string
	got := randomSpellLossEligible4F24E0(7, randomSpellLossTestHooks4F24E0(rows, &trace))
	if got != 0 || fmt.Sprint(trace) != "[id:0]" {
		t.Fatalf("result/trace = %d/%v, want 0/[id:0]", got, trace)
	}
}

func TestRandomSpellLossEligibilityProtectsExactFourSpells4F24E0(t *testing.T) {
	for _, spellID := range []int32{34, 27, 9, 41} {
		rows := []randomSpellLossTestRow4F24E0{{spellID: uint32(spellID), slots: 0xffffffff}, {}}
		if got := randomSpellLossEligible4F24E0(spellID, randomSpellLossTestHooks4F24E0(rows, nil)); got != 0 {
			t.Fatalf("protected spell %d eligibility = %d, want 0", spellID, got)
		}
	}
	rows := []randomSpellLossTestRow4F24E0{{spellID: 33, slots: 1}, {}}
	if got := randomSpellLossEligible4F24E0(33, randomSpellLossTestHooks4F24E0(rows, nil)); got != 1 {
		t.Fatalf("adjacent spell eligibility = %d, want 1", got)
	}
}

func TestRandomSpellLossEligibilityFaultPrefixes4F24E0(t *testing.T) {
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

	fault("first spell ID", []string{"id:0"}, func(trace *[]string) {
		randomSpellLossEligible4F24E0(7, randomSpellLossEligibilityHooks4F24E0{
			loadSpellID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				panic("fault")
			},
		})
	})

	fault("matching slots", []string{"id:0", "slots:0"}, func(trace *[]string) {
		randomSpellLossEligible4F24E0(7, randomSpellLossEligibilityHooks4F24E0{
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

	fault("next spell ID", []string{"id:0", "id:1"}, func(trace *[]string) {
		randomSpellLossEligible4F24E0(7, randomSpellLossEligibilityHooks4F24E0{
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
