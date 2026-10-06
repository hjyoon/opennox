package opennox

import (
	"os"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/legacy"
	"gopkg.in/yaml.v2"
)

func e2eQuestMapSelectionCaptured4D0F60() legacy.QuestMapSelectionState4D0F60 {
	state := legacy.QuestMapSelectionState4D0F60{
		Count: 13, Clock: 1000, LastIndex: 0xffffffff, Frame: 31,
		LogicIndex: 3605, OtherIndex: 2501,
	}
	for index, name := range []string{
		"G_CastlD.map", "G_Castle.map", "G_CryptD.map", "G_Crypts.map",
		"G_ForesD.map", "G_Forest.map", "G_LOTD.map", "G_LOTDD.map",
		"G_Lava.map", "G_Mines.map", "G_Swamp.map", "G_TemplD.map", "G_Temple.map",
	} {
		entry := legacy.QuestMapHistory4D0F60{Group: []uint32{1, 1, 2, 2, 3, 3, 4, 4, 5, 6, 7, 8, 8}[index]}
		copy(entry.Name[:], name)
		state.Entries = append(state.Entries, entry)
	}
	return state
}

func e2eQuestMapSelectionResult4D0F60(state legacy.QuestMapSelectionState4D0F60, index int32, logic int) legacy.QuestMapSelection4D0F60 {
	after := state
	after.Entries = append([]legacy.QuestMapHistory4D0F60(nil), state.Entries...)
	after.LogicIndex = logic
	return legacy.QuestMapSelection4D0F60{Before: state, After: after, ResultIndex: index, ResultPointer: 0x1000}
}

func TestQuestMapSelectionReference4D0F60CapturedHistory(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stage int
		logic int
		index int32
		mapID string
	}{
		{name: "first-unused-table", stage: 1, logic: 3605, index: 11, mapID: "G_TemplD.map"},
		{name: "second-family-and-cooldown", stage: 2, logic: 3826, index: 0, mapID: "G_CastlD.map"},
		{name: "third-after-original-AI-wait-draws", stage: 3, logic: 2754, index: 3, mapID: "G_Crypts.map"},
		{name: "third-old-no-op-AI-checkpoint", stage: 3, logic: 2712, index: 7, mapID: "G_LOTDD.map"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := e2eQuestMapSelectionCaptured4D0F60()
			state.LogicIndex = tc.logic
			if tc.stage >= 2 {
				state.Clock, state.LastIndex = 1001, 11
				state.Frame, state.OtherIndex = 928, 1960
				state.Entries[11].Uses, state.Entries[11].LastUsed = 1, 1000
				state.Entries[12].Uses, state.Entries[12].LastUsed = 1, 1000
			}
			if tc.stage >= 3 {
				state.Clock, state.LastIndex = 1002, 0
				state.Frame, state.OtherIndex = 1480, 3905
				state.Entries[0].Uses, state.Entries[0].LastUsed = 1, 1001
				state.Entries[1].Uses, state.Entries[1].LastUsed = 1, 1001
			}
			index, logic, err := e2eQuestMapReference4D0F60(state)
			if err != nil || index != tc.index || logic != tc.logic+1 {
				t.Fatalf("reference=%d/%d/%v, want %d/%d", index, logic, err, tc.index, tc.logic+1)
			}
			mapID, err := e2eQuestMapSelectionValidate4D0F60(e2eQuestMapSelectionResult4D0F60(state, tc.index, tc.logic+1))
			if err != nil || mapID != tc.mapID {
				t.Fatalf("selected map=%q/%v, want %q", mapID, err, tc.mapID)
			}
		})
	}
}

func TestQuestMapSelectionReference4D0F60SignedBranches(t *testing.T) {
	entry := func(group, uses, last uint32) legacy.QuestMapHistory4D0F60 {
		return legacy.QuestMapHistory4D0F60{Group: group, Uses: uses, LastUsed: last}
	}
	for _, tc := range []struct {
		name    string
		entries []legacy.QuestMapHistory4D0F60
		clock   uint32
		last    uint32
		index   int32
		draws   int
	}{
		{name: "empty-no-draw", last: 0xffffffff, index: -2},
		{name: "singleton-no-history-or-draw", entries: []legacy.QuestMapHistory4D0F60{entry(9, 0xffffffff, 0xffffffff)}, last: 0xffffffff},
		// Table[0]=0x7304=29444, therefore direct modulo 3 selects index 2.
		{name: "unused-direct-modulo", entries: []legacy.QuestMapHistory4D0F60{entry(9, 0, 0), entry(9, 0, 0), entry(9, 0, 0)}, last: 0xffffffff, index: 2, draws: 1},
		{name: "negative-uses-do-not-become-maximum", entries: []legacy.QuestMapHistory4D0F60{entry(9, 0xffffffff, 0), entry(9, 0x80000000, 0), entry(9, 0xfffffffe, 0)}, last: 0xffffffff, index: 2, draws: 1},
		{name: "negative-uses-eligible", entries: []legacy.QuestMapHistory4D0F60{entry(9, 2, 0), entry(7, 0x80000000, 0), entry(7, 2, 0)}, clock: 1000, index: 1, draws: 1},
		{name: "signed-future-cooldown-excluded", entries: []legacy.QuestMapHistory4D0F60{entry(9, 2, 0), entry(7, 0, 1001), entry(8, 0, 0)}, clock: 1000, index: 2, draws: 1},
		{name: "high-bit-cooldown-excluded", entries: []legacy.QuestMapHistory4D0F60{entry(9, 2, 0), entry(7, 0, 0), entry(8, 0, 0x7ffffffb)}, clock: 0x80000000, index: 2, draws: 1},
		{name: "wrapped-five-ticks-versus-four", entries: []legacy.QuestMapHistory4D0F60{entry(9, 2, 0), entry(7, 0, 0xfffffffd), entry(8, 0, 0xfffffffe)}, clock: 2, index: 1, draws: 1},
		{name: "exact-four-ticks-excluded", entries: []legacy.QuestMapHistory4D0F60{entry(9, 2, 0), entry(7, 0, 996), entry(8, 0, 995)}, clock: 1000, index: 2, draws: 1},
		{name: "same-family-excluded", entries: []legacy.QuestMapHistory4D0F60{entry(9, 2, 0), entry(9, 0, 0), entry(8, 0, 0)}, clock: 1000, index: 2, draws: 1},
		{name: "equal-uses-increment-threshold", entries: []legacy.QuestMapHistory4D0F60{entry(9, 2, 0), entry(7, 2, 0), entry(8, 2, 0)}, clock: 1000, index: 1, draws: 1},
		{name: "equal-maximum-signed-overflow", entries: []legacy.QuestMapHistory4D0F60{entry(9, 0x7fffffff, 0), entry(7, 0x7fffffff, 0), entry(8, 0x7fffffff, 0)}, clock: 1000, index: -1},
		{name: "no-candidates-no-draw-fallback", entries: []legacy.QuestMapHistory4D0F60{entry(9, 2, 0), entry(9, 0, 0), entry(9, 0, 0)}, clock: 1000, index: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := legacy.QuestMapSelectionState4D0F60{Count: uint32(len(tc.entries)), Clock: tc.clock, LastIndex: tc.last, Entries: tc.entries}
			before := state
			before.Entries = append([]legacy.QuestMapHistory4D0F60(nil), state.Entries...)
			index, logic, err := e2eQuestMapReference4D0F60(state)
			if err != nil || index != tc.index || logic != tc.draws || !reflect.DeepEqual(state, before) {
				t.Fatalf("reference=%d/%d/%v unchanged=%t, want %d/%d", index, logic, err, reflect.DeepEqual(state, before), tc.index, tc.draws)
			}
		})
	}
}

func TestQuestMapSelectionReference4D0F60RejectsInvalidSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*legacy.QuestMapSelectionState4D0F60)
	}{
		{name: "too-many-records", change: func(s *legacy.QuestMapSelectionState4D0F60) { s.Count = 129 }},
		{name: "missing-record", change: func(s *legacy.QuestMapSelectionState4D0F60) { s.Entries = s.Entries[:12] }},
		{name: "negative-rng", change: func(s *legacy.QuestMapSelectionState4D0F60) { s.LogicIndex = -1 }},
		{name: "past-rng-table", change: func(s *legacy.QuestMapSelectionState4D0F60) { s.LogicIndex = 4096 }},
		{name: "used-history-invalid-last", change: func(s *legacy.QuestMapSelectionState4D0F60) { s.Entries[0].Uses = 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := e2eQuestMapSelectionCaptured4D0F60()
			tc.change(&state)
			if _, _, err := e2eQuestMapReference4D0F60(state); err == nil {
				t.Fatal("invalid snapshot accepted")
			}
		})
	}
}

func TestQuestMapSelectionValidate4D0F60RejectsMutations(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*legacy.QuestMapSelection4D0F60)
	}{
		{name: "wrong-result", change: func(s *legacy.QuestMapSelection4D0F60) { s.ResultIndex = 10 }},
		{name: "unrecognized-native-slot", change: func(s *legacy.QuestMapSelection4D0F60) { s.ResultIndex = -3 }},
		{name: "null-native-pointer", change: func(s *legacy.QuestMapSelection4D0F60) { s.ResultPointer = 0 }},
		{name: "no-logic-draw", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.LogicIndex-- }},
		{name: "extra-logic-draw", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.LogicIndex++ }},
		{name: "other-rng-advanced", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.OtherIndex++ }},
		{name: "frame-advanced", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.Frame++ }},
		{name: "count-edited", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.Count++ }},
		{name: "clock-edited", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.Clock++ }},
		{name: "last-edited", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.LastIndex-- }},
		{name: "group-edited", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.Entries[0].Group++ }},
		{name: "name-edited", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.Entries[0].Name[0] = 'x' }},
		{name: "use-count-edited", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.Entries[0].Uses++ }},
		{name: "last-use-edited", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.Entries[0].LastUsed++ }},
		{name: "record-lost", change: func(s *legacy.QuestMapSelection4D0F60) { s.After.Entries = s.After.Entries[:12] }},
		{name: "empty-name", change: func(s *legacy.QuestMapSelection4D0F60) {
			s.Before.Entries[11].Name = [20]byte{}
			s.After.Entries[11].Name = [20]byte{}
		}},
		{name: "unterminated-name", change: func(s *legacy.QuestMapSelection4D0F60) {
			for i := range s.Before.Entries[11].Name {
				s.Before.Entries[11].Name[i] = 'x'
				s.After.Entries[11].Name[i] = 'x'
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			selection := e2eQuestMapSelectionResult4D0F60(e2eQuestMapSelectionCaptured4D0F60(), 11, 3606)
			tc.change(&selection)
			if _, err := e2eQuestMapSelectionValidate4D0F60(selection); err == nil {
				t.Fatal("mutated selection accepted")
			}
		})
	}
}

func TestQuestMapSelectionValidate4D0F60RejectsStockNullAndFallback(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state legacy.QuestMapSelectionState4D0F60
		index int32
		ptr   uintptr
	}{
		{name: "null", index: -2},
		{name: "fallback", index: -1, ptr: 0x1000, state: legacy.QuestMapSelectionState4D0F60{
			Count: 2, Clock: 1000, Entries: []legacy.QuestMapHistory4D0F60{{Group: 1, Uses: 2}, {Group: 1}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			selection := e2eQuestMapSelectionResult4D0F60(tc.state, tc.index, 0)
			selection.ResultPointer = tc.ptr
			if _, err := e2eQuestMapSelectionValidate4D0F60(selection); err == nil {
				t.Fatal("null/fallback cannot establish a playable stock map")
			}
		})
	}
}

func TestQuestMapSelection4D0F60SupplementKeepsExistingFlow(t *testing.T) {
	read := func(path string) e2eFileYML {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var scenario e2eFileYML
		if err := yaml.UnmarshalStrict(data, &scenario); err != nil {
			t.Fatal(err)
		}
		return scenario
	}
	original := read("../scripts/e2e/host-quest-twenty-stages.yaml")
	supplement := read("../scripts/e2e/host-quest-twenty-stages-map-oracle.yaml")
	if len(supplement.Steps) != len(original.Steps)+1 || !reflect.DeepEqual(supplement.Steps[0], e2eStepYML{
		Action: "observe-quest-map-selection", Name: "observe original native Quest map selections",
	}) {
		t.Fatal("supplement must arm only the read-only observer before existing startup")
	}
	stages, captures, walks, exits, generators, minions := 0, 0, 0, 0, 0, 0
	for index, step := range original.Steps {
		want := step
		switch step.Action {
		case "assert-quest-stage":
			stages++
			if step.Count != stages || step.Map == "" {
				t.Fatal("existing fixed-map checkpoint changed")
			}
			want.Action, want.Map = "assert-quest-selected-map", ""
		case "screen":
			captures++
			want.Action = "capture-quest-map-audit-frame"
		case "walk-quest":
			walks++
		case "enter-quest-exit":
			exits++
		case "assert-quest-generators":
			generators++
		case "assert-quest-minion":
			minions++
		}
		if got := supplement.Steps[index+1]; !reflect.DeepEqual(got, want) {
			t.Fatalf("supplement changed existing input/timing/assertion %d: got %+v want %+v", index, got, want)
		}
	}
	if stages != 20 || captures != 5 || walks != 19 || exits != 19 || generators != 20 || minions != 3 {
		t.Fatalf("existing flow changed: stages=%d captures=%d walks=%d exits=%d generators=%d minions=%d", stages, captures, walks, exits, generators, minions)
	}
}
