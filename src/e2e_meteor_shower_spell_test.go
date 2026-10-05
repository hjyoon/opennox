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

func TestE2EMeteorShowerModesAndBoundedSchedule(t *testing.T) {
	for _, mode := range []string{"player-script", "npc-animated", "", "npc-direct"} {
		for level := -1; level <= 6; level++ {
			t.Run(fmt.Sprintf("%s/%d", mode, level), func(t *testing.T) {
				want := mode == "player-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
				var sc e2eScenario
				if !want {
					defer func() {
						if recover() == nil || len(sc.steps) != 0 {
							t.Fatal("invalid MeteorShower mode scheduled gameplay")
						}
					}()
					sc.CheckMeteorShowerSpell(level, mode, "invalid")
					return
				}
				sc.CheckMeteorShowerSpell(level, mode, "MeteorShower")
				if len(sc.steps) != 9 {
					t.Fatalf("scheduled steps=%d want=9", len(sc.steps))
				}
				for suffix, timeout := range map[string]time.Duration{
					" prepare": 1200, " falling drawable": 180, " natural shower expiry and client replay": 600,
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

func TestE2EMeteorShowerPublicScenarioAndDispatch(t *testing.T) {
	const path = "../scripts/e2e/host-game-meteor-shower-spell.yaml"
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
		if step.Action != "check-meteor-shower-spell" {
			continue
		}
		key := fmt.Sprintf("%s/%d", step.Text, step.Count)
		if !e2eMeteorMode(step.Count, step.Text) || seen[key] {
			t.Fatalf("invalid/duplicate MeteorShower public fixture: %+v", step)
		}
		seen[key] = true
	}
	if len(seen) != 6 || !seen["npc-animated/0"] {
		t.Fatalf("MeteorShower public modes=%v", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	count := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " natural shower expiry and client replay") {
			count++
		}
	}
	if count != 6 {
		t.Fatalf("loaded MeteorShower checks=%d want=6", count)
	}
}

func TestE2EMeteorShowerObserverDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_meteor_shower_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Buffs": true, "Field120_1": true, "Field120_2": true, "Field120_3": true,
		"Field330": true, "Damage": true, "ObjOwner": true, "UpdateData": true, "ZVal": true, "Field27": true,
		"Field29": true, "Field5": true, "ObjFlags": true, "Obj130": true, "Field131": true,
		"Field32": true, "Field34": true, "NetCode": true, "ScriptIDVal": true}
	calls := map[string]bool{"CastMeteorShower52D8A0": true, "Nox_xxx_castMeteorShower_52D8A0": true,
		"CastMeteor52D9D0": true, "Nox_xxx_castMeteor_52D9D0": true, "MonsterActionCast5413B0": true,
		"MeteorShowerUpdate53D5A0": true, "MeteorUpdate53D6E0": true, "CallDamage": true,
		"DefaultDamageWorld4E0B30": true, "Raise": true, "DrawImageAt": true, "NetSendPacketXxx": true,
		"NetSendPacketXxx0": true, "SetHealth": true, "SetMaxHealth": true, "DelayedDelete": true}
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
						t.Errorf("MeteorShower observer writes result %s", sel.Sel.Name)
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok && calls[sel.Sel.Name] {
					if fn.Name.Name != "prepare" || sel.Sel.Name != "SetMaxHealth" {
						t.Errorf("MeteorShower observer supplies result via %s in %s", sel.Sel.Name, fn.Name.Name)
					}
				}
			}
			return true
		})
	}
}

func TestE2EMeteorShowerCleanupRequiresObservedCompletion(t *testing.T) {
	f := new(e2eMeteorShowerFixture)
	defer func() {
		if recover() == nil {
			t.Fatal("MeteorShower cleanup accepted an unobserved gameplay result")
		}
	}()
	f.cleanup()
}

func TestE2EMeteorShowerIndependentRegenerationModel(t *testing.T) {
	for _, tc := range []struct {
		name               string
		frame, injury, fps uint32
		maximum, current   int32
		want               int32
	}{
		{"durable-target", 1185, 1153, 30, 30000, 29790, 5},
		{"last-pause-frame", 1183, 1153, 30, 30000, 29790, 0},
		{"first-after-pause", 1184, 1153, 30, 30000, 29790, 5},
		{"injured-now", 1185, 1185, 30, 30000, 29790, 0},
		{"cap-maximum", 1185, 1153, 30, 30000, 29999, 1},
		{"full-health", 1185, 1153, 30, 30000, 30000, 0},
		{"zero-max", 1185, 1153, 30, 0, 0, 0},
		{"zero-fps", 1185, 1153, 0, 30000, 29790, 0},
		{"classic-not-frame", 1185, 1153, 30, 50, 40, 0},
		{"classic-frame", 1188, 1153, 30, 50, 40, 1},
		{"classic-paused", 1188, 1180, 30, 50, 40, 0},
		{"wrapped-clock", 0, math.MaxUint32 - 30, 30, 30000, 29790, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eMeteorShowerRegenAmount(tc.frame, tc.injury, tc.fps, tc.maximum, tc.current); got != tc.want {
				t.Fatalf("regen=%d want=%d", got, tc.want)
			}
		})
	}
}
