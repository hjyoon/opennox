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
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EEarthquakeModesAndBoundedSchedule(t *testing.T) {
	for _, mode := range []string{"player-script", "npc-animated", "", "npc-direct"} {
		for level := -1; level <= 6; level++ {
			t.Run(fmt.Sprintf("%s/%d", mode, level), func(t *testing.T) {
				want := mode == "player-script" && level >= 1 && level <= 5 || mode == "npc-animated" && level == 0
				if got := e2eEarthquakeMode(level, mode); got != want {
					t.Fatalf("accepted=%t want=%t", got, want)
				}
				var sc e2eScenario
				if !want {
					defer func() {
						if recover() == nil || len(sc.steps) != 0 {
							t.Fatal("invalid Earthquake scheduled gameplay")
						}
					}()
					sc.CheckEarthquakeSpell(level, mode, "invalid")
					return
				}
				sc.CheckEarthquakeSpell(level, mode, "Earthquake")
				if len(sc.steps) != 8 {
					t.Fatalf("steps=%d want=8", len(sc.steps))
				}
				for suffix, timeout := range map[string]time.Duration{
					" prepare": 1200, " actual damage and client replay": 300, " real quake packet and natural settling": 300,
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
						t.Errorf("observation %s count=%d", suffix, count)
					}
				}
			})
		}
	}
}

func TestE2EEarthquakeDamageWitnessNativeIdentityAndType(t *testing.T) {
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	other, freeOther := alloc.New(server.Object{})
	defer freeOther()
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(caster)) <= math.MaxUint32 || uintptr(unsafe.Pointer(other)) <= math.MaxUint32) {
		t.Fatal("native attribution fixtures below 4 GiB")
	}
	for _, tc := range []struct {
		name          string
		before, after uint16
		by, source    *server.Object
		typ           uint32
		want          bool
	}{
		{"actual-impact", 2000, 1975, caster, caster, 11, true},
		{"one-hp", 2000, 1999, caster, caster, 11, true},
		{"no-damage", 2000, 2000, caster, caster, 11, false},
		{"healing", 1975, 2000, caster, caster, 11, false},
		{"dead", 2000, 0, caster, caster, 11, false},
		{"wrong-native-source", 2000, 1975, other, caster, 11, false},
		{"missing-attribution", 2000, 1975, nil, caster, 11, false},
		{"nil-pair", 2000, 1975, nil, nil, 11, false},
		{"crush-not-impact", 2000, 1975, caster, caster, 2, false},
		{"poison-not-impact", 2000, 1975, caster, caster, 5, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eEarthquakeDamageObserved(tc.before, tc.after, tc.by, tc.source, tc.typ); got != tc.want {
				t.Fatalf("witness=%t want=%t", got, tc.want)
			}
		})
	}
}

func TestE2EEarthquakePublicScenarioAndDispatch(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/host-game-earthquake-spell.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	var cases []string
	for _, step := range file.Steps {
		if step.Action == "check-earthquake-spell" {
			cases = append(cases, fmt.Sprintf("%s/%d", step.Text, step.Count))
		}
	}
	if strings.Join(cases, ",") != "npc-animated/0,player-script/1,player-script/2,player-script/3,player-script/4,player-script/5" {
		t.Fatalf("public cases=%v", cases)
	}
	if len(file.Steps) == 0 || file.Steps[len(file.Steps)-1].Action != "quit" {
		t.Fatal("public world lacks native terminal shutdown")
	}
	var sc e2eScenario
	sc.Load("../scripts/e2e/host-game-earthquake-spell.yaml")
	if len(sc.steps) < 6*8 {
		t.Fatal("public dispatcher did not schedule all six bounded casts")
	}
}

func TestE2EEarthquakeFixtureDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_earthquake_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	protected := map[string]bool{
		"ObjOwner": true, "Obj130": true, "Field131": true, "Frame134": true,
		"Buffs": true, "Cur": true, "Jiggle12": true, "Field120_1": true, "Field120_2": true,
		"NetCode": true, "ScriptIDVal": true,
	}
	ast.Inspect(file, func(n ast.Node) bool {
		if assignment, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignment.Lhs {
				ast.Inspect(lhs, func(part ast.Node) bool {
					if field, ok := part.(*ast.SelectorExpr); ok && protected[field.Sel.Name] {
						t.Errorf("fixture writes live outcome %s", field.Sel.Name)
					}
					return true
				})
			}
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch method.Sel.Name {
		case "CallDamage", "CastEarthquake52DE40", "Nox_xxx_castEquake_52DE40", "Nox_xxx_earthquakeSend_4D9110", "NetSendPacket", "AdjustHP", "BuffApply":
			t.Errorf("fixture bypasses real cast/damage/network: %s", method.Sel.Name)
		}
		return true
	})
}
