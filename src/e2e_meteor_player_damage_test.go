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

	"gopkg.in/yaml.v2"
)

func TestE2EMeteorPlayerModesAndBoundedSchedule(t *testing.T) {
	for _, mode := range []string{"npc-script", "npc-animated", "player-script", "npc-direct", ""} {
		for level := -1; level <= 6; level++ {
			t.Run(fmt.Sprintf("%s/%d", mode, level), func(t *testing.T) {
				want := mode == "npc-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
				if got := e2eMeteorPlayerMode(level, mode); got != want {
					t.Fatalf("mode accepted=%t want=%t", got, want)
				}
				var sc e2eScenario
				if !want {
					defer func() {
						if recover() == nil || len(sc.steps) != 0 {
							t.Fatal("invalid Meteor player mode scheduled gameplay")
						}
					}()
					sc.CheckMeteorPlayerDamage(level, mode, "invalid")
					return
				}
				sc.CheckMeteorPlayerDamage(level, mode, "Meteor player")
				if len(sc.steps) != 9 {
					t.Fatalf("scheduled steps=%d want=9", len(sc.steps))
				}
				for suffix, timeout := range map[string]time.Duration{
					" prepare": 1200, " falling drawable": 180, " natural player impact and HUD": 300,
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

func TestE2EMeteorPlayerPublicScenarioAndDispatch(t *testing.T) {
	const path = "../scripts/e2e/host-game-meteor-player-damage.yaml"
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
		if step.Action != "check-meteor-player-damage" {
			continue
		}
		key := fmt.Sprintf("%s/%d", step.Text, step.Count)
		if !e2eMeteorPlayerMode(step.Count, step.Text) || seen[key] {
			t.Fatalf("invalid/duplicate Meteor player fixture: %+v", step)
		}
		seen[key] = true
	}
	if len(seen) != 6 || !seen["npc-animated/0"] {
		t.Fatalf("Meteor player modes=%v, want five NPC levels plus natural animated cast", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	count := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " natural player impact and HUD") {
			count++
		}
	}
	if count != 6 {
		t.Fatalf("loaded Meteor player checks=%d want=6", count)
	}
}

func TestE2EMeteorPlayerCleanupRequiresVerifiedHitAndHUD(t *testing.T) {
	var f e2eMeteorPlayerFixture
	defer func() {
		if err := recover(); err == nil || !strings.Contains(fmt.Sprint(err), "before verified damage/HUD") {
			t.Fatalf("unverified cleanup reached gameplay: %v", err)
		}
	}()
	f.cleanup()
}

func TestE2EMeteorPlayerObserverDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_meteor_player_damage.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Buffs": true, "Field120_1": true, "Field120_2": true, "Field120_3": true,
		"Field330": true, "Damage": true, "ObjOwner": true, "UpdateData": true, "ZVal": true, "Field27": true,
		"Field29": true, "Field5": true, "ObjFlags": true, "Obj130": true, "Field131": true,
		"Field21": true, "Field57": true, "Field75": true, "Field76": true}
	calls := map[string]bool{"CastMeteor52D9D0": true, "Nox_xxx_castMeteor_52D9D0": true, "MonsterActionCast5413B0": true,
		"MeteorUpdate53D6E0": true, "CallDamage": true, "DefaultDamageWorld4E0B30": true, "Raise": true,
		"DrawImageAt": true, "NetSendPacketXxx": true, "NetSendPacketXxx0": true, "SetHealth": true,
		"SetMaxHealth": true, "DelayedDelete": true, "SetOwner": true}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == "cleanup" {
			guard, ok := fn.Body.List[0].(*ast.IfStmt)
			if !ok {
				t.Fatal("cleanup lacks an entry verification guard")
			}
			not, ok := guard.Cond.(*ast.UnaryExpr)
			if !ok || not.Op != token.NOT {
				t.Fatal("cleanup does not reject unverified state")
			}
			verified, ok := not.X.(*ast.SelectorExpr)
			if !ok || verified.Sel.Name != "verified" {
				t.Fatal("cleanup does not guard verified damage/HUD")
			}
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.AssignStmt:
				for _, lhs := range n.Lhs {
					if sel, ok := lhs.(*ast.SelectorExpr); ok && fields[sel.Sel.Name] {
						t.Errorf("Meteor player observer writes result %s", sel.Sel.Name)
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
					if calls[sel.Sel.Name] {
						setup := fn.Name.Name == "prepare" && sel.Sel.Name == "SetMaxHealth"
						cleanup := fn.Name.Name == "cleanup" && (sel.Sel.Name == "DelayedDelete" || sel.Sel.Name == "SetMaxHealth" || sel.Sel.Name == "SetHealth")
						if !setup && !cleanup {
							t.Errorf("Meteor player observer supplies result via %s in %s", sel.Sel.Name, fn.Name.Name)
						}
					}
					if base, ok := sel.X.(*ast.SelectorExpr); ok && base.Sel.Name == "cast" &&
						(sel.Sel.Name == "prepare" || sel.Sel.Name == "observeHit" || sel.Sel.Name == "complete" || sel.Sel.Name == "cleanup") {
						t.Errorf("Meteor player observer uses the Troll-specific %s", sel.Sel.Name)
					}
				}
			}
			return true
		})
	}
}
