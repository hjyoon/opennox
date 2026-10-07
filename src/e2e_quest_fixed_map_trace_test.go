package opennox

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/legacy"
)

func TestE2EQuestFixedMapTraceAdmission(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stage int
		mapID string
		armed bool
		want  int
	}{
		{name: "first-fixed-checkpoint", stage: 1, mapID: "g_templd", want: 1},
		{name: "zero-stage", mapID: "g_templd"},
		{name: "negative-stage", stage: -1, mapID: "g_templd"},
		{name: "later-checkpoint", stage: 2, mapID: "g_castld"},
		{name: "no-fixed-map", stage: 1},
		{name: "already-armed", stage: 1, mapID: "g_templd", armed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trace := e2eQuestFixedMapTrace4D0F60{armed: tc.armed}
			installed, stopped, emitted := 0, 0, 0
			trace.arm(tc.stage, tc.mapID, func(callback func(legacy.QuestMapSelection4D0F60)) func() {
				installed++
				if !trace.armed || callback == nil {
					t.Fatal("trace was not armed before installation")
				}
				return func() { stopped++ }
			}, func(string, ...any) { emitted++ })
			if installed != tc.want || emitted != 0 || stopped != 0 || trace.armed != (tc.armed || tc.want == 1) {
				t.Fatalf("installation changed: installed=%d emitted=%d stopped=%d armed=%t", installed, emitted, stopped, trace.armed)
			}
			if trace.stop != nil {
				trace.stop()
				if stopped != 1 {
					t.Fatal("observer cleanup delegate was not retained")
				}
			}
		})
	}
}

func e2eQuestFixedMapTraceSelection(t *testing.T, logicIndex int) legacy.QuestMapSelection4D0F60 {
	t.Helper()
	entries := make([]legacy.QuestMapHistory4D0F60, 3)
	for i, name := range []string{"g_templd", "g_lotdd", "g_crypts"} {
		entries[i].Group = uint32(i)
		copy(entries[i].Name[:], name)
	}
	before := legacy.QuestMapSelectionState4D0F60{
		Count: 3, Clock: 17, LastIndex: 0, Frame: 401, LogicIndex: logicIndex, OtherIndex: 901, Entries: entries,
	}
	rng := prand.New(logicIndex)
	index := rng.Int(0, 2)
	after := before
	after.Entries = append([]legacy.QuestMapHistory4D0F60(nil), entries...)
	after.LogicIndex = rng.Index()
	resultPointer := uintptr(index*32) + uintptr(0x80000000)
	resultPointer += uintptr(0x80000100)
	return legacy.QuestMapSelection4D0F60{Before: before, After: after, ResultIndex: int32(index),
		ResultPointer: resultPointer}
}

func TestE2EQuestFixedMapTraceRecordsOriginalInputsWithoutMutation(t *testing.T) {
	for _, logic := range []int{0, 1, 17, 23, 1000, 4095} {
		t.Run(fmt.Sprint(logic), func(t *testing.T) {
			selection := e2eQuestFixedMapTraceSelection(t, logic)
			before, err := json.Marshal(selection)
			if err != nil {
				t.Fatal(err)
			}
			var callback func(legacy.QuestMapSelection4D0F60)
			var records []string
			trace := e2eQuestFixedMapTrace4D0F60{}
			installations := 0
			install := func(observe func(legacy.QuestMapSelection4D0F60)) func() {
				installations++
				callback = observe
				return func() {}
			}
			emit := func(format string, args ...any) { records = append(records, fmt.Sprintf(format, args...)) }
			trace.arm(1, "g_templd", install, emit)
			trace.arm(1, "g_templd", install, emit)
			trace.arm(2, "g_castld", install, emit)
			if installations != 1 || callback == nil || len(records) != 0 {
				t.Fatal("trace duplicated installation or fabricated a choice")
			}
			callback(selection)
			if len(records) != 2 || records[0] != "QUEST FIXED MAP SELECTION SNAPSHOT: "+string(before) ||
				!strings.HasPrefix(records[1], "QUEST FIXED MAP ORIGINAL MATCH: ") ||
				!strings.Contains(records[1], fmt.Sprintf("native=%#x", selection.ResultPointer)) {
				t.Fatalf("complete raw input/original match missing: %v", records)
			}
			after, err := json.Marshal(selection)
			if err != nil || string(after) != string(before) {
				t.Fatal("observer/reference changed captured inputs")
			}
			callback(selection)
			if len(records) != 4 || records[2] != records[0] || records[3] != records[1] {
				t.Fatal("read-only trace consumed private RNG or changed subsequent output")
			}
		})
	}
}

func TestE2EQuestFixedMapTraceDifferenceIsDiagnosticOnly(t *testing.T) {
	for _, mode := range []string{"result-index", "extra-logic-rng", "other-rng", "history", "invalid-count", "null-result"} {
		t.Run(mode, func(t *testing.T) {
			selection := e2eQuestFixedMapTraceSelection(t, 17)
			switch mode {
			case "result-index":
				selection.ResultIndex = (selection.ResultIndex + 1) % 3
			case "extra-logic-rng":
				selection.After.LogicIndex = (selection.After.LogicIndex + 1) % 4096
			case "other-rng":
				selection.After.OtherIndex++
			case "history":
				selection.After.Entries[0].Uses++
			case "invalid-count":
				selection.Before.Count = 129
			case "null-result":
				selection.ResultPointer = 0
			}
			before, _ := json.Marshal(selection)
			var callback func(legacy.QuestMapSelection4D0F60)
			var records []string
			trace := e2eQuestFixedMapTrace4D0F60{}
			trace.arm(1, "g_templd", func(observe func(legacy.QuestMapSelection4D0F60)) func() {
				callback = observe
				return func() {}
			}, func(format string, args ...any) { records = append(records, fmt.Sprintf(format, args...)) })
			callback(selection)
			after, _ := json.Marshal(selection)
			if len(records) != 2 || !strings.HasPrefix(records[1], "QUEST FIXED MAP REFERENCE DIFFERENCE: ") ||
				string(before) != string(after) {
				t.Fatalf("trace changed the captured outcome or concealed mismatch: %v", records)
			}
		})
	}
}

func TestE2EQuestFixedMapTraceBindingKeepsOriginalAssertion(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_quest.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "AssertQuestStage" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch name := call.Fun.(type) {
			case *ast.Ident:
				if name.Name == "e2eError" || name.Name == "e2eQuestTraceFixedMaps4D0F60" {
					calls = append(calls, name.Name)
					if name.Name == "e2eQuestTraceFixedMaps4D0F60" && (len(call.Args) != 2 ||
						call.Args[0].(*ast.Ident).Name != "stage" || call.Args[1].(*ast.Ident).Name != "mapName") {
						t.Error("trace must receive unchanged checkpoint inputs")
					}
				}
			case *ast.SelectorExpr:
				if name.Sel.Name == "validate" {
					calls = append(calls, "validate")
					if len(call.Args) != 2 || call.Args[0].(*ast.Ident).Name != "stage" || call.Args[1].(*ast.Ident).Name != "mapName" {
						t.Error("fixed map expectation was weakened")
					}
				}
			}
			return true
		})
	}
	if !reflect.DeepEqual(calls, []string{"validate", "e2eError", "e2eQuestTraceFixedMaps4D0F60"}) {
		t.Fatalf("trace must be after the unchanged assertion and error-return path: %v", calls)
	}
	var scenario e2eScenario
	scenario.AssertQuestStage(1, "g_templd", "unchanged fixed checkpoint")
	if len(scenario.steps) != 1 || scenario.steps[0].time != 0 || scenario.steps[0].fnc == nil {
		t.Fatal("read-only trace changed checkpoint timing or step count")
	}
}
