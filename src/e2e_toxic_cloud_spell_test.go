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

	"gopkg.in/yaml.v2"
)

func TestE2EToxicCloudModesAndBoundedSchedule(t *testing.T) {
	for _, mode := range []string{"player-script", "script-pos-pos", "npc-animated", "", "npc-direct"} {
		for level := -1; level <= 6; level++ {
			t.Run(fmt.Sprintf("%s/%d", mode, level), func(t *testing.T) {
				want := mode == "player-script" && level >= 1 && level <= 5 ||
					(mode == "script-pos-pos" || mode == "npc-animated") && level == 0
				if got := e2eToxicCloudMode(level, mode); got != want {
					t.Fatalf("mode accepted=%t want=%t", got, want)
				}
				var sc e2eScenario
				if !want {
					defer func() {
						if recover() == nil || len(sc.steps) != 0 {
							t.Fatal("invalid Toxic Cloud mode scheduled gameplay")
						}
					}()
					sc.CheckToxicCloudSpell(level, mode, "invalid")
					return
				}
				sc.CheckToxicCloudSpell(level, mode, "Toxic Cloud")
				if len(sc.steps) != 10 {
					t.Fatalf("scheduled steps=%d want=10", len(sc.steps))
				}
				for suffix, timeout := range map[string]time.Duration{
					" prepare": 1200, " published cloud": 180,
					" actual damage and client replay": 300, " natural expiry": 2100,
				} {
					found := 0
					for _, step := range sc.steps {
						if strings.HasSuffix(step.name, suffix) {
							found++
							if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
								t.Fatalf("%s lacks bounded live observation", suffix)
							}
						}
					}
					if found != 1 {
						t.Fatalf("%s found=%d want=1", suffix, found)
					}
				}
			})
		}
	}
}

func TestE2EToxicCloudFixtureRejectsPoisonImmuneTargets(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_toxic_cloud_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	stockWolf, immunityGate := false, false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if assignment, ok := n.(*ast.AssignStmt); ok {
				for _, lhs := range assignment.Lhs {
					if field, ok := lhs.(*ast.SelectorExpr); ok && field.Sel.Name == "ObjSubClass" {
						t.Error("Toxic Cloud fixture writes immunity flags")
					}
				}
			}
			call, ok := n.(*ast.CallExpr)
			if !ok || fn.Name.Name != "prepare" {
				return true
			}
			method, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || len(call.Args) != 1 {
				return true
			}
			arg, ok := call.Args[0].(*ast.BasicLit)
			if !ok {
				return true
			}
			stockWolf = stockWolf || method.Sel.Name == "NewObjectByTypeID" && arg.Kind == token.STRING && arg.Value == `"Wolf"`
			if method.Sel.Name == "Has" && arg.Kind == token.INT && arg.Value == "0x200" {
				query, ok := method.X.(*ast.CallExpr)
				if ok && len(query.Args) == 0 {
					get, ok := query.Fun.(*ast.SelectorExpr)
					if ok && get.Sel.Name == "SubClass" {
						field, ok := get.X.(*ast.SelectorExpr)
						immunityGate = immunityGate || ok && field.Sel.Name == "target"
					}
				}
			}
			return true
		})
	}
	if !stockWolf || !immunityGate {
		t.Fatalf("stock susceptible target=%t live immunity guard=%t", stockWolf, immunityGate)
	}
}

func TestE2EToxicCloudIndependentLifetime(t *testing.T) {
	for _, tc := range []struct {
		name string
		life float32
		fps  uint32
		want int32
		ok   bool
	}{
		{"fractional-chop", 1.5, 31, 46, true},
		{"negative-signed-FPS", -1.5, 0xffffffe1, 46, true},
		{"single-frame", 0.0625, 16, 1, true},
		{"maximum-bounded", 60, 30, 1800, true},
		{"above-bound", 60, 31, 0, false},
		{"zero-life", 0, 30, 0, false},
		{"zero-FPS", 1.5, 0, 0, false},
		{"subframe", 0.03125, 16, 0, false},
		{"negative-life", -1.5, 31, 0, false},
		{"signed-FPS", 1.5, math.MaxUint32, 0, false},
		{"NaN", float32(math.NaN()), 30, 0, false},
		{"infinity", float32(math.Inf(1)), 30, 0, false},
		{"huge", math.MaxFloat32, math.MaxInt32, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := e2eToxicCloudLifetime(tc.life, tc.fps)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("lifetime=%d/%t want=%d/%t", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestE2EToxicCloudPublicScenarioAndDispatch(t *testing.T) {
	const path = "../scripts/e2e/host-game-toxic-cloud-spell.yaml"
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
		if step.Action != "check-toxic-cloud-spell" {
			continue
		}
		key := fmt.Sprintf("%s/%d", step.Text, step.Count)
		if !e2eToxicCloudMode(step.Count, step.Text) || seen[key] {
			t.Fatalf("invalid/duplicate Toxic Cloud public fixture: %+v", step)
		}
		seen[key] = true
	}
	if len(seen) != 7 || !seen["npc-animated/0"] || !seen["script-pos-pos/0"] {
		t.Fatalf("public modes=%v, want position script, five player levels and animated NPC", seen)
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
		t.Fatalf("loaded Toxic Cloud gameplay checks=%d want=7", count)
	}
}

func TestE2EToxicCloudObserverDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_toxic_cloud_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Buffs": true, "Poison540": true, "Duration": true,
		"Field120_1": true, "Field120_2": true, "Field120_3": true, "Field330": true,
		"Damage": true, "ObjOwner": true, "UpdateData": true, "ObjFlags": true, "Obj130": true, "Field131": true}
	calls := map[string]bool{"CastToxicCloud52DB60": true, "Nox_xxx_castToxicCloud_52DB60": true,
		"MonsterActionCast5413B0": true, "ToxicCloudUpdate53D850": true, "CallDamage": true,
		"DefaultDamageWorld4E0B30": true, "Poison": true, "PoisonActivate": true, "Raise": true,
		"DrawImageAt": true, "NetSendPacketXxx": true, "NetSendPacketXxx0": true,
		"SetHealth": true, "SetMaxHealth": true, "SetPos": true, "DelayedDelete": true}
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
						t.Errorf("Toxic Cloud observer writes result %s", sel.Sel.Name)
					}
				}
			case *ast.IncDecStmt:
				if sel, ok := n.X.(*ast.SelectorExpr); ok && fields[sel.Sel.Name] {
					t.Errorf("Toxic Cloud observer changes result %s", sel.Sel.Name)
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok && calls[sel.Sel.Name] {
					setup := fn.Name.Name == "prepare" && (sel.Sel.Name == "SetMaxHealth" || sel.Sel.Name == "SetPos")
					cleanup := fn.Name.Name == "cleanup" && (sel.Sel.Name == "DelayedDelete" || sel.Sel.Name == "SetPos")
					if !setup && !cleanup {
						t.Errorf("Toxic Cloud observer supplies result via %s in %s", sel.Sel.Name, fn.Name.Name)
					}
				}
			}
			return true
		})
	}
}

func TestE2EToxicCloudFixtureUsesNormalScriptAndNPCAcceptance(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_toxic_cloud_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	field := func(expr ast.Expr, name string) bool {
		sel, ok := expr.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != name {
			return false
		}
		base, ok := sel.X.(*ast.Ident)
		return ok && base.Name == "f"
	}
	owned, enemy, imaginary, pospos, objectCast, animation, frameGate := false, false, false, false, false, false, false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if assignment, ok := n.(*ast.AssignStmt); ok && fn.Name.Name == "prepare" {
				for i, lhs := range assignment.Lhs {
					if field(lhs, "caster") && i < len(assignment.Rhs) {
						id, ok := assignment.Rhs[i].(*ast.Ident)
						imaginary = imaginary || ok && id.Name == "nox_xxx_imagCasterUnit_1569664"
					}
				}
			}
			if sel, ok := n.(*ast.SelectorExpr); ok && fn.Name.Name == "observeSound" && sel.Sel.Name == "MissileAttackFrame216" {
				frameGate = true
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if fn.Name.Name == "prepare" {
				if sel.Sel.Name == "CreateObjectAt" && len(call.Args) == 3 && field(call.Args[0], "caster") {
					owned = field(call.Args[1], "host")
				}
				if sel.Sel.Name == "IsEnemyTo" && len(call.Args) == 2 && field(call.Args[0], "target") {
					chain, ok := call.Args[1].(*ast.CallExpr)
					if ok && len(chain.Args) == 0 {
						get, ok := chain.Fun.(*ast.SelectorExpr)
						enemy = ok && get.Sel.Name == "FindOwnerChainPlayer" && field(get.X, "caster")
					}
				}
			}
			if fn.Name.Name == "beginCast" {
				pospos = pospos || sel.Sel.Name == "CastSpell" && len(call.Args) == 3
				objectCast = objectCast || sel.Sel.Name == "CastSpellLvl" && len(call.Args) == 4
				if sel.Sel.Name == "MonsterPushAction" && len(call.Args) == 4 {
					kind, ok := call.Args[0].(*ast.SelectorExpr)
					animation = ok && kind.Sel.Name == "ACTION_CAST_SPELL_ON_OBJECT" && field(call.Args[3], "target")
				}
			}
			return true
		})
	}
	if !owned || !enemy || !imaginary || !pospos || !objectCast || !animation || !frameGate {
		t.Fatalf("normal fixture owned=%t enemy=%t imaginary=%t pospos=%t object=%t animation=%t frame=%t", owned, enemy, imaginary, pospos, objectCast, animation, frameGate)
	}
}
