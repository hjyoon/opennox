package server

import (
	"fmt"
	"testing"
)

type randomFieldGuideLossTestRow4F2530 struct {
	guideID uint32
	slots   uint32
}

func randomFieldGuideLossTestHooks4F2530(
	rows []randomFieldGuideLossTestRow4F2530,
	trace *[]string,
) randomFieldGuideLossEligibilityHooks4F2530 {
	return randomFieldGuideLossEligibilityHooks4F2530{
		loadGuideID: func(index int) uint32 {
			if trace != nil {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
			}
			return rows[index].guideID
		},
		loadSlots: func(index int) uint32 {
			if trace != nil {
				*trace = append(*trace, fmt.Sprintf("slots:%d", index))
			}
			return rows[index].slots
		},
	}
}

func TestRandomFieldGuideLossEligibilityMatchesGAMEEXETable4F2530(t *testing.T) {
	eligible := map[int32]bool{
		2: true, 3: true, 4: true, 9: true, 10: true, 11: true,
		13: true, 14: true, 15: true, 16: true, 17: true, 18: true,
		19: true, 20: true, 21: true, 22: true, 23: true, 24: true,
		26: true, 27: true, 29: true, 31: true, 32: true, 33: true,
		34: true, 35: true, 36: true, 37: true, 38: true, 39: true,
		40: true,
	}
	for guideID := int32(-3); guideID <= 45; guideID++ {
		var want int32
		if eligible[guideID] {
			want = 1
		}
		if got := RandomFieldGuideLossEligible4F2530(guideID); got != want {
			t.Fatalf("guide %d eligibility = %d, want %d", guideID, got, want)
		}
	}
}

func TestRandomFieldGuideLossEligibilityScansLiveRowsAndShortCircuitsSlots4F2530(t *testing.T) {
	rows := []randomFieldGuideLossTestRow4F2530{
		{guideID: 7},
		{guideID: 8, slots: 2},
		{guideID: 7, slots: 4},
		{},
	}
	var trace []string
	got := randomFieldGuideLossEligible4F2530(7, randomFieldGuideLossTestHooks4F2530(rows, &trace))
	if got != 1 {
		t.Fatalf("eligibility = %d, want 1", got)
	}
	want := []string{"id:0", "slots:0", "id:1", "id:2", "slots:2"}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}

	trace = nil
	got = randomFieldGuideLossEligible4F2530(99, randomFieldGuideLossTestHooks4F2530(rows, &trace))
	if got != 0 {
		t.Fatalf("missing eligibility = %d, want 0", got)
	}
	want = []string{"id:0", "id:1", "id:2", "id:3"}
	if fmt.Sprint(trace) != fmt.Sprint(want) {
		t.Fatalf("missing trace = %v, want %v", trace, want)
	}
}

func TestRandomFieldGuideLossEligibilityStopsAtFirstSentinel4F2530(t *testing.T) {
	rows := []randomFieldGuideLossTestRow4F2530{{slots: 0xffffffff}, {guideID: 7, slots: 1}}
	var trace []string
	got := randomFieldGuideLossEligible4F2530(0, randomFieldGuideLossTestHooks4F2530(rows, &trace))
	if got != 0 || fmt.Sprint(trace) != "[id:0]" {
		t.Fatalf("result/trace = %d/%v, want 0/[id:0]", got, trace)
	}

	trace = nil
	got = randomFieldGuideLossEligible4F2530(7, randomFieldGuideLossTestHooks4F2530(rows, &trace))
	if got != 0 || fmt.Sprint(trace) != "[id:0]" {
		t.Fatalf("post-sentinel result/trace = %d/%v, want 0/[id:0]", got, trace)
	}
}

func TestRandomFieldGuideLossEligibilityFaultPrefixes4F2530(t *testing.T) {
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

	fault("first guide ID", []string{"id:0"}, func(trace *[]string) {
		randomFieldGuideLossEligible4F2530(7, randomFieldGuideLossEligibilityHooks4F2530{
			loadGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				panic("fault")
			},
		})
	})

	fault("matching slots", []string{"id:0", "slots:0"}, func(trace *[]string) {
		randomFieldGuideLossEligible4F2530(7, randomFieldGuideLossEligibilityHooks4F2530{
			loadGuideID: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("id:%d", index))
				return 7
			},
			loadSlots: func(index int) uint32 {
				*trace = append(*trace, fmt.Sprintf("slots:%d", index))
				panic("fault")
			},
		})
	})

	fault("next guide ID", []string{"id:0", "id:1"}, func(trace *[]string) {
		randomFieldGuideLossEligible4F2530(7, randomFieldGuideLossEligibilityHooks4F2530{
			loadGuideID: func(index int) uint32 {
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
