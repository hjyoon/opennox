package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/types"
	"gopkg.in/yaml.v2"
)

func TestE2EMeteorModesAndBoundedSchedule(t *testing.T) {
	for _, mode := range []string{"player-script", "npc-animated", "", "npc-direct"} {
		for level := -1; level <= 6; level++ {
			t.Run(fmt.Sprintf("%s/%d", mode, level), func(t *testing.T) {
				want := mode == "player-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
				if got := e2eMeteorMode(level, mode); got != want {
					t.Fatalf("mode accepted=%t want=%t", got, want)
				}
				var sc e2eScenario
				if !want {
					defer func() {
						if recover() == nil || len(sc.steps) != 0 {
							t.Fatal("invalid Meteor mode scheduled gameplay")
						}
					}()
					sc.CheckMeteorSpell(level, mode, "invalid")
					return
				}
				sc.CheckMeteorSpell(level, mode, "Meteor")
				if len(sc.steps) != 9 {
					t.Fatalf("scheduled steps=%d want=9", len(sc.steps))
				}
				for suffix, timeout := range map[string]time.Duration{
					" prepare": 1200, " falling drawable": 180, " natural impact and client replay": 300,
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

func TestE2EMeteorRadialDamageStockFalloff(t *testing.T) {
	for _, tc := range []struct {
		raw      int32
		distance float32
		want     int32
		inside   bool
	}{
		{100, 0, 100, true}, {100, 29, 100, true}, {100, 30, 100, true},
		{100, 55, 50, true}, {100, 79, 1, true}, {100, 80, 0, true}, {100, 81, 0, false},
		{201, 55, 100, true}, {-201, 55, -100, true}, {0, 0, 0, true},
	} {
		t.Run(fmt.Sprintf("%d/%g", tc.raw, tc.distance), func(t *testing.T) {
			got, inside := e2eMeteorRadialDamage(tc.raw, types.Ptf(100, 200), types.Ptf(100+tc.distance, 200))
			if got != tc.want || inside != tc.inside {
				t.Fatalf("radial=%d/%t want=%d/%t", got, inside, tc.want, tc.inside)
			}
		})
	}
}

func TestE2EMeteorPublicScenarioAndDispatch(t *testing.T) {
	const path = "../scripts/e2e/host-game-meteor-spell.yaml"
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
		if step.Action != "check-meteor-spell" {
			continue
		}
		key := fmt.Sprintf("%s/%d", step.Text, step.Count)
		if !e2eMeteorMode(step.Count, step.Text) || seen[key] {
			t.Fatalf("invalid/duplicate Meteor public fixture: %+v", step)
		}
		seen[key] = true
	}
	if len(seen) != 6 || !seen["npc-animated/0"] {
		t.Fatalf("Meteor public modes=%v, want five player levels plus animated NPC", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	count := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " natural impact and client replay") {
			count++
		}
	}
	if count != 6 {
		t.Fatalf("loaded Meteor gameplay checks=%d want=6", count)
	}
}

func TestE2EMeteorObserverDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_meteor_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Buffs": true, "Field120_1": true, "Field120_2": true, "Field120_3": true,
		"Field330": true, "Damage": true, "ObjOwner": true, "UpdateData": true, "ZVal": true, "Field27": true,
		"Field29": true, "Field5": true, "ObjFlags": true, "Obj130": true, "Field131": true}
	calls := map[string]bool{"CastMeteor52D9D0": true, "Nox_xxx_castMeteor_52D9D0": true, "MonsterActionCast5413B0": true,
		"MeteorUpdate53D6E0": true, "CallDamage": true, "DefaultDamageWorld4E0B30": true, "Raise": true,
		"DrawImageAt": true, "NetSendPacketXxx": true, "NetSendPacketXxx0": true, "SetHealth": true,
		"SetMaxHealth": true, "DelayedDelete": true}
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
						t.Errorf("Meteor observer writes result %s", sel.Sel.Name)
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok && calls[sel.Sel.Name] {
					setup := fn.Name.Name == "prepare" && sel.Sel.Name == "SetMaxHealth"
					cleanup := fn.Name.Name == "cleanup" && sel.Sel.Name == "DelayedDelete"
					if !setup && !cleanup {
						t.Errorf("Meteor observer supplies result via %s in %s", sel.Sel.Name, fn.Name.Name)
					}
				}
			}
			return true
		})
	}
}
