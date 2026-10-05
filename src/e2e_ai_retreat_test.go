package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EAIRetreatSchedule(t *testing.T) {
	for _, kind := range []string{"Troll", "NPC"} {
		for _, slope := range []string{"ascending", "descending"} {
			mode := kind + "/" + slope
			t.Run(mode, func(t *testing.T) {
				got, descending, ok := e2eAIRetreatMode(mode)
				if !ok || got != kind || descending != (slope == "descending") {
					t.Fatal("RETREAT mode changed")
				}
				var sc e2eScenario
				sc.CheckAIRetreat(mode, "Retreat")
				if len(sc.steps) != 4 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 ||
					sc.steps[1].ready == nil || sc.steps[1].waitTimeout != 600 ||
					sc.steps[2].ready == nil || sc.steps[2].waitTimeout != 600 ||
					!strings.HasSuffix(sc.steps[1].name, " natural first attack then configure retreat ability") ||
					!strings.HasSuffix(sc.steps[2].name, " natural incoming hit and RETREAT self-buff") ||
					sc.steps[1].time-sc.steps[0].time != 1 || sc.steps[2].time-sc.steps[1].time != 1 ||
					sc.steps[3].time-sc.steps[2].time != 12 {
					t.Fatal("bounded natural first-attack/incoming damage/RETREAT schedule changed")
				}
			})
		}
	}
}

func TestE2EAIRetreatInvalidMode(t *testing.T) {
	for _, mode := range []string{"", "Troll", "Spider/ascending", "troll/ascending", "NPC/other", "NPC/ascending/forced"} {
		t.Run(mode, func(t *testing.T) {
			var sc e2eScenario
			defer func() {
				if recover() == nil || len(sc.steps) != 0 {
					t.Fatal("invalid mode scheduled a RETREAT probe")
				}
			}()
			sc.CheckAIRetreat(mode, "invalid")
		})
	}
}

func TestE2EAIRetreatSelfCastIdentity(t *testing.T) {
	for _, mode := range []string{"valid", "nil-update", "nil-unit", "no-retreat", "no-head", "other-action", "other-spell", "other-target", "nil-target"} {
		t.Run(mode, func(t *testing.T) {
			unit, other := new(server.Object), new(server.Object)
			update := &server.MonsterUpdateData{AIStackInd: 2}
			update.AIStack[0].Action = uint32(ai.ACTION_RETREAT)
			update.AIStack[1].Action = uint32(ai.DEPENDENCY_UNINTERRUPTABLE)
			update.AIStack[2].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
			update.AIStack[2].SetArgs(uint32(e2eAIRetreatSpell), uint32(0), unit)
			switch mode {
			case "nil-update":
				update = nil
			case "nil-unit":
				unit = nil
			case "no-retreat":
				update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
			case "no-head":
				update.AIStackInd = -1
			case "other-action":
				update.AIStack[2].Action = uint32(ai.ACTION_CAST_SPELL_ON_LOCATION)
			case "other-spell":
				update.AIStack[2].SetArgs(uint32(spell.SPELL_HASTE))
			case "other-target":
				update.AIStack[2].SetArgs(uint32(e2eAIRetreatSpell), uint32(0), other)
			case "nil-target":
				update.AIStack[2].SetArgs(uint32(e2eAIRetreatSpell), uint32(0), (*server.Object)(nil))
			}
			var before server.MonsterUpdateData
			if update != nil {
				before = *update
			}
			if got := e2eAIRetreatSelfCast(update, unit); got != (mode == "valid") {
				t.Fatalf("native self-cast accepted=%t mode=%s", got, mode)
			}
			if update != nil && *update != before {
				t.Fatal("self-cast observer altered the action/update record")
			}
			runtime.KeepAlive(unit)
			runtime.KeepAlive(other)
		})
	}
}

func TestE2EAIRetreatScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-ai-retreat.yaml"
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
		if step.Action != "check-ai-retreat" {
			continue
		}
		if _, _, ok := e2eAIRetreatMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/repeated RETREAT mode %q", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 4 {
		t.Fatalf("RETREAT modes=%d want=4", len(seen))
	}
	var sc e2eScenario
	sc.Load(path)
	checks := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " natural incoming hit and RETREAT self-buff") {
			checks++
		}
	}
	if checks != 4 {
		t.Fatalf("natural RETREAT observers=%d want=4", checks)
	}
}

func TestE2EAIRetreatDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_ai_retreat.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Max": true, "Buffs": true, "HealthData": true, "CurrentEnemy": true,
		"PreferredEnemy": true, "AIStack": true, "AIStackInd": true, "Action": true, "Args": true, "Field371": true,
		"Field124": true, "Field137": true, "Field120_1": true, "Field120_2": true, "PosVec": true, "UpdateData": true,
		"Obj130": true, "Frame134": true, "Damage": true, "EnchantData": true}
	calls := map[string]bool{"CallDamage": true, "ApplyEnchant": true, "BuffOn": true, "BuffOff": true,
		"SetHealth": true, "SetMaxHealth": true, "MonsterPushAction": true, "MonsterPopAction": true,
		"MonsterCast": true, "MonsterActionRetreat545440": true, "DrawImageAt": true, "NetSendPacketXxx": true,
		"SetPos": true, "SetArgs": true}
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
						t.Errorf("RETREAT observer writes result %s", sel.Sel.Name)
					}
					if _, ok := lhs.(*ast.StarExpr); ok && fn.Name.Name != "configure" {
						t.Errorf("RETREAT observer writes through pointer outside ordinary ability setup")
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
					if calls[sel.Sel.Name] {
						t.Errorf("RETREAT observer supplies result via %s", sel.Sel.Name)
					}
					if sel.Sel.Name == "CastSpellLvl" && fn.Name.Name != "fire" {
						t.Errorf("RETREAT observer casts outside ordinary incoming attack")
					}
				}
			}
			return true
		})
	}
}
