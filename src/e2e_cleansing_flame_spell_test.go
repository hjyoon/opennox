package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/types"

	"gopkg.in/yaml.v2"
)

func TestE2ECleansingFlameModesAndBoundedSchedule(t *testing.T) {
	for _, mode := range []string{"red-player", "red-npc", "blue-npc", "", "npc-direct", "blue-player"} {
		for level := -1; level <= 6; level++ {
			t.Run(fmt.Sprintf("%s/%d", mode, level), func(t *testing.T) {
				want := mode == "red-player" && level >= 1 && level <= 5 ||
					(mode == "red-npc" || mode == "blue-npc") && level == 0
				if got := e2eCleansingFlameMode(level, mode); got != want {
					t.Fatalf("accepted=%t want=%t", got, want)
				}
				var sc e2eScenario
				if !want {
					defer func() {
						if recover() == nil || len(sc.steps) != 0 {
							t.Fatal("invalid flame fixture scheduled gameplay")
						}
					}()
					sc.CheckCleansingFlameSpell(level, mode, "invalid")
					return
				}
				sc.CheckCleansingFlameSpell(level, mode, "flame")
				if len(sc.steps) != 10 {
					t.Fatalf("steps=%d want=10", len(sc.steps))
				}
				for suffix, timeout := range map[string]time.Duration{
					" prepare": 1200, " real moving prediction": 180,
					" actual collision and client replay": 300, " natural expiry": 600,
				} {
					count := 0
					for _, step := range sc.steps {
						if strings.HasSuffix(step.name, suffix) {
							count++
							if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
								t.Errorf("unbounded/missing observation %s", suffix)
							}
						}
					}
					if count != 1 {
						t.Errorf("observation %s count=%d want=1", suffix, count)
					}
				}
			})
		}
	}
}

func TestE2ECleansingFlameTerminationRequiresActualOriginalStop(t *testing.T) {
	for _, tc := range []struct {
		name                string
		frame, created      uint32
		deadline, deletedAt uint32
		current, previous   types.Pointf
		rayClear            bool
		want                string
	}{
		{"actual deadline", 109, 100, 110, 110, types.Ptf(1, 2), types.Ptf(0, 1), true, "deadline"},
		{"deadline before ray", 109, 100, 110, 110, types.Ptf(1, 2), types.Ptf(0, 1), false, "deadline"},
		{"blocked before deadline", 104, 100, 200, 105, types.Ptf(1, 2), types.Ptf(0, 1), false, "blocked-ray"},
		{"stationary age four", 103, 100, 200, 104, types.Ptf(1, 2), types.Ptf(1, 2), true, "stationary"},
		{"stationary age three rejected", 102, 100, 200, 103, types.Ptf(1, 2), types.Ptf(1, 2), true, ""},
		{"moving clear ray rejected", 104, 100, 200, 105, types.Ptf(1, 2), types.Ptf(0, 1), true, ""},
		{"one moving axis rejected", 104, 100, 200, 105, types.Ptf(1, 2), types.Ptf(1, 1), true, ""},
		{"stale snapshot rejected", 103, 100, 105, 105, types.Ptf(1, 2), types.Ptf(1, 2), false, ""},
		{"same frame rejected", 105, 100, 105, 105, types.Ptf(1, 2), types.Ptf(1, 2), false, ""},
		{"unordered x87 stationary", 104, 100, 200, 105, types.Ptf(float32(math.NaN()), 2), types.Ptf(1, 2), true, "stationary"},
		{"wrapped frame and age", math.MaxUint32, math.MaxUint32 - 3, 10, 0, types.Ptf(1, 2), types.Ptf(1, 2), true, "stationary"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := e2eCleansingFlameUpdateInput{
				frame: tc.frame, created: tc.created, current: tc.current, previous: tc.previous, rayClear: tc.rayClear,
			}
			if got := e2eCleansingFlameTermination(input, tc.deadline, tc.deletedAt); got != tc.want {
				t.Fatalf("termination=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestE2ECleansingFlameTerminationsRequireEveryActualStop(t *testing.T) {
	for _, tc := range []struct {
		name           string
		flames         []e2eCleansingFlameRecord
		expired, early int
		want           bool
	}{
		{"no flames", nil, 0, 0, false},
		{"deadline", []e2eCleansingFlameRecord{{removed: true, termination: "deadline"}}, 1, 0, true},
		{"wall before deadline", []e2eCleansingFlameRecord{{removed: true, termination: "blocked-ray"}}, 0, 1, true},
		{"stationary before deadline", []e2eCleansingFlameRecord{{removed: true, termination: "stationary"}}, 0, 1, true},
		{"all original stops", []e2eCleansingFlameRecord{
			{removed: true, termination: "deadline"}, {removed: true, termination: "blocked-ray"},
			{removed: true, termination: "stationary"},
		}, 1, 2, true},
		{"missing actual deletion witness", []e2eCleansingFlameRecord{{removed: true}}, 0, 1, false},
		{"unknown early deletion", []e2eCleansingFlameRecord{{removed: true, termination: "forced"}}, 0, 1, false},
		{"still in world", []e2eCleansingFlameRecord{{termination: "deadline"}}, 1, 0, false},
		{"misclassified removal", []e2eCleansingFlameRecord{{removed: true, termination: "blocked-ray"}}, 1, 0, false},
		{"missing removal count", []e2eCleansingFlameRecord{{removed: true, termination: "deadline"}}, 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eCleansingFlameTerminationsComplete(tc.flames, tc.expired, tc.early); got != tc.want {
				t.Fatalf("complete=%t want=%t", got, tc.want)
			}
		})
	}
}

func TestE2ECleansingFlameSnapshotAccountsForPostIncrementHook(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_cleansing_flame_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	frames := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "e2eCleansingFlameInput" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			field, ok := n.(*ast.KeyValueExpr)
			if !ok {
				return true
			}
			key, ok := field.Key.(*ast.Ident)
			if !ok || key.Name != "frame" {
				return true
			}
			frames++
			binary, ok := field.Value.(*ast.BinaryExpr)
			if !ok || binary.Op != token.SUB {
				t.Fatal("tick hook observes the next frame after server_E.IncFrame; snapshot is off by one")
			}
			frame, ok := binary.X.(*ast.Ident)
			if !ok || frame.Name != "frame" {
				t.Fatal("snapshot does not preserve the actual observed frame")
			}
			one, ok := binary.Y.(*ast.BasicLit)
			if !ok || one.Kind != token.INT || one.Value != "1" {
				t.Fatal("snapshot does not account for exactly one server frame increment")
			}
			return true
		})
	}
	if frames != 1 {
		t.Fatalf("actual snapshot frame fields=%d want=1", frames)
	}
}

func TestE2ECleansingFlameStockTypes(t *testing.T) {
	for _, name := range []string{"SmallFlameCleanse", "MediumFlameCleanse", "FlameCleanse", "LargeFlameCleanse",
		"SmallBlueFlameCleanse", "MediumBlueFlameCleanse", "BlueFlameCleanse", "LargeBlueFlameCleanse"} {
		if !e2eCleansingFlameType(name) {
			t.Errorf("stock flame %q rejected", name)
		}
	}
	for _, name := range []string{"Flame", "BurnMediumFlame", "CustomFlameCleanse", "SmallFlameCleanseExtra", ""} {
		if e2eCleansingFlameType(name) {
			t.Errorf("unrelated flame %q admitted", name)
		}
	}
}

func TestE2ECleansingFlamePublicScenarioAndDispatch(t *testing.T) {
	const path = "../scripts/e2e/host-game-cleansing-flame-spell.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, step := range file.Steps {
		if step.Action != "check-cleansing-flame-spell" {
			continue
		}
		key := fmt.Sprintf("%s/%d", step.Text, step.Count)
		if !e2eCleansingFlameMode(step.Count, step.Text) || seen[key] {
			t.Fatalf("invalid/duplicate fixture %+v", step)
		}
		seen[key] = true
	}
	if len(seen) != 7 || !seen["red-npc/0"] || !seen["blue-npc/0"] {
		t.Fatalf("public modes=%v", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	count := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " natural expiry") {
			count++
		}
	}
	if count != 7 {
		t.Fatalf("loaded natural expiry checks=%d want=7", count)
	}
}

func TestE2ECleansingFlameObserversNeverSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_cleansing_flame_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{
		"Cur": true, "Buffs": true, "ObjClass": true, "ObjSubClass": true, "ObjFlags": true,
		"ObjOwner": true, "Obj130": true, "Field131": true, "Field34": true, "Field32": true,
		"Update": true, "ManaCur": true, "ManaMax": true, "Field120_1": true, "Field120_2": true,
		"Field_115": true, "Field_117": true, "Field_118": true, "Field_119": true,
		"Field_81": true, "Field_82": true, "AnimStart": true, "InClientUpdateList": true,
		"nox_input_seq": true, "nox_input_seq_prev": true,
	}
	calls := map[string]bool{
		"CallDamage": true, "FlameCleanseUpdate53D510": true, "CastCleansingFlame52D5C0": true,
		"Nox_xxx_spellCastCleansingFlame_52D5C0": true, "MonsterActionCast5413B0": true,
		"NetClientPredictLinear523530": true, "AddToMsgListCli": true, "List5Add": true,
		"handleClientPredictLinearPacketNative48EA70": true, "Seed": true, "IntClamp": true,
		"SetHealth": true, "SetMaxHealth": true, "SetMana": true, "DelayedDelete": true,
		"SetPos": true, "ResetInput": true,
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.AssignStmt:
				for _, lhs := range n.Lhs {
					if sel, ok := lhs.(*ast.SelectorExpr); ok && fields[sel.Sel.Name] {
						t.Errorf("fixture writes result %s", sel.Sel.Name)
					}
				}
			case *ast.IncDecStmt:
				if sel, ok := n.X.(*ast.SelectorExpr); ok && fields[sel.Sel.Name] {
					t.Errorf("fixture changes result %s", sel.Sel.Name)
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok && calls[sel.Sel.Name] {
					setup := fn.Name.Name == "prepare" && (sel.Sel.Name == "SetMaxHealth" || sel.Sel.Name == "SetPos")
					cleanup := fn.Name.Name == "cleanup" && (sel.Sel.Name == "DelayedDelete" ||
						sel.Sel.Name == "SetPos" || sel.Sel.Name == "SetMaxHealth" ||
						sel.Sel.Name == "SetHealth" || sel.Sel.Name == "SetMana")
					if !setup && !cleanup {
						t.Errorf("fixture supplies result via %s in %s", sel.Sel.Name, fn.Name.Name)
					}
				}
			}
			return true
		})
	}
}

func TestE2ECleansingFlamePredictionRequiresServerAndClientMovement(t *testing.T) {
	var f e2eCleansingFlameFixture
	for _, tc := range []e2eCleansingFlameRecord{
		{}, {predicted: true}, {predicted: true, moved: true}, {moved: true, clientMoved: true},
	} {
		f.flames = []e2eCleansingFlameRecord{tc}
		if f.predicted() {
			t.Fatal("prediction accepted a missing network/server/client observation")
		}
	}
	f.flames = []e2eCleansingFlameRecord{{predicted: true, moved: true, clientMoved: true}}
	if !f.predicted() {
		t.Fatal("actual network and server/client movement observations rejected")
	}
}

func TestE2ECleansingFlameCastFollowsServerPacketReset(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_cleansing_flame_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var begin *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "beginCast" {
			begin = fn
		}
	}
	if begin == nil {
		t.Fatal("missing actual cast step")
	}
	queued, casts, publishes := 0, 0, 0
	ast.Inspect(begin.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "CastSpellLvl" || sel.Sel.Name == "ObjectsAddPending" {
			t.Fatal("cast/publication precedes the real server packet reset")
		}
		if sel.Sel.Name != "TickCallback" {
			return true
		}
		queued++
		if len(call.Args) != 1 {
			t.Fatal("invalid server callback")
		}
		callback, ok := call.Args[0].(*ast.FuncLit)
		if !ok {
			t.Fatal("missing queued cast body")
		}
		ast.Inspect(callback.Body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					switch sel.Sel.Name {
					case "CastSpellLvl":
						casts++
					case "ObjectsAddPending":
						publishes++
					}
				}
			}
			return true
		})
		return false
	})
	if queued != 1 || casts != 1 || publishes != 1 {
		t.Fatalf("queued=%d actual-casts=%d publications=%d", queued, casts, publishes)
	}
}

func TestE2ECleansingFlameRemovalNeverReadsPlayerAI(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_cleansing_flame_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	isNPCGuard := func(expr ast.Expr) bool {
		binary, ok := expr.(*ast.BinaryExpr)
		if !ok || binary.Op != token.NEQ {
			return false
		}
		mode, ok := binary.X.(*ast.SelectorExpr)
		if !ok || mode.Sel.Name != "mode" {
			return false
		}
		value, ok := binary.Y.(*ast.BasicLit)
		return ok && value.Kind == token.STRING && value.Value == `"red-player"`
	}
	guarded := make(map[*ast.CallExpr]bool)
	var complete *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "complete" {
			complete = fn
		}
	}
	if complete == nil {
		t.Fatal("missing actual removal observer")
	}
	ast.Inspect(complete.Body, func(n ast.Node) bool {
		var body ast.Node
		switch n := n.(type) {
		case *ast.IfStmt:
			if isNPCGuard(n.Cond) {
				body = n.Body
			}
		case *ast.BinaryExpr:
			if n.Op == token.LAND && isNPCGuard(n.X) {
				body = n.Y
			}
		}
		if body != nil {
			ast.Inspect(body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "MonsterActionIsScheduled" {
						guarded[call] = true
					}
				}
				return true
			})
		}
		return true
	})
	calls := 0
	ast.Inspect(complete.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "MonsterActionIsScheduled" {
				calls++
				if !guarded[call] {
					t.Error("removal observer reads MonsterUpdateData for a player")
				}
			}
		}
		return true
	})
	if calls != 2 {
		t.Fatalf("NPC removal/action observations=%d want=2", calls)
	}
}
