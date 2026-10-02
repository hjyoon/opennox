package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestE2EShockRetaliationAmount(t *testing.T) {
	for _, tc := range []struct {
		balance float64
		carry   float32
		want    uint16
		ok      bool
	}{{8, 0, 8, true}, {8.5, 0, 8, true}, {9.5, 0, 10, true},
		{8.500000001, 0, 8, true}, {9, 0.5, 10, true}, {9, -0.5, 8, true},
		{1, -0.5, 1, true}, {0, 0, 0, false}, {-1, 0, 0, false},
		{math.MaxUint16, 0, math.MaxUint16, true}, {math.MaxUint16, 0.5, 0, false},
		{math.NaN(), 0, 0, false}, {math.Inf(1), 0, 0, false},
		{8, float32(math.NaN()), 0, false}, {8, float32(math.Inf(-1)), 0, false}, {8, 0.6, 0, false}} {
		if got, ok := e2eShockRetaliationAmount(tc.balance, tc.carry); got != tc.want || ok != tc.ok {
			t.Fatalf("balance=%g carry=%g damage=%d/%t want=%d/%t", tc.balance, tc.carry, got, ok, tc.want, tc.ok)
		}
	}
}

func TestE2EShockRetaliationSchedule(t *testing.T) {
	for _, mode := range []string{"to-npc", "to-player", "", "npc", "player"} {
		valid := mode == "to-npc" || mode == "to-player"
		if e2eShockRetaliationMode(mode) != valid {
			t.Fatal("invalid retaliation mode admission")
		}
		var sc e2eScenario
		if !valid {
			func() {
				defer func() {
					if recover() == nil || len(sc.steps) != 0 {
						t.Errorf("invalid Shock retaliation was scheduled: %q", mode)
					}
				}()
				sc.CheckShockRetaliation(mode, "invalid")
			}()
			continue
		}
		sc.CheckShockRetaliation(mode, "Shock")
		if len(sc.steps) != 10 {
			t.Fatalf("retaliation steps=%d want=10", len(sc.steps))
		}
		for _, index := range []int{0, 3, 7} {
			step := sc.steps[index]
			if step.ready == nil || step.fnc == nil || step.waitTimeout <= 0 || step.waitTimeout > 1200 {
				t.Fatalf("retaliation %q is not bounded", step.name)
			}
		}
		if !strings.HasSuffix(sc.steps[5].name, " actual attack") || !strings.HasSuffix(sc.steps[6].name, " release input") ||
			!strings.HasSuffix(sc.steps[8].name, " consumed and damaged") || !strings.HasSuffix(sc.steps[9].name, " cleanup") {
			t.Fatal("retaliation input/observation/screenshot/cleanup order")
		}
	}
}

func TestE2EShockRetaliationScenario(t *testing.T) {
	path := "../scripts/e2e/host-game-shock-retaliation.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	warrior := false
	seen := map[string]bool{}
	for _, step := range file.Steps {
		if step.Action == "click" && step.Name == "select warrior" && step.X == 256 && step.Y == 208 {
			warrior = true
		}
		if step.Action == "check-shock-retaliation" {
			if !e2eShockRetaliationMode(step.Text) || seen[step.Text] {
				t.Fatal("invalid/repeated retaliation scenario")
			}
			seen[step.Text] = true
		}
	}
	if !warrior || !seen["to-player"] || !seen["to-npc"] || len(seen) != 2 {
		t.Fatalf("retaliation coverage=%v warrior=%t", seen, warrior)
	}
	var sc e2eScenario
	sc.Load(path)
	complete := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " complete") {
			complete++
		}
	}
	if complete != 2 {
		t.Fatalf("loaded retaliation checks=%d want=2", complete)
	}
	for _, action := range []string{"cast", "damage-monster", "damage-player", "set-monster-health", "set-player-health", "set-player-buff"} {
		if strings.Contains(string(raw), "action: "+action+"\n") {
			t.Fatalf("retaliation scenario injects a result: %s", action)
		}
	}
}

func TestE2EShockRetaliationDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_shock_retaliation.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	forbiddenFields := map[string]bool{
		"Cur": true, "Max": true, "Buffs": true, "State": true, "Field40_0": true,
		"Field76": true, "Field75": true, "Field21": true, "Field1": true,
		"Field547": true, "Field546": true, "Field523_2": true,
	}
	forbiddenCalls := map[string]bool{
		"CallDamage": true, "Damage": true, "BuffOn": true, "BuffOff": true,
		"SetHealth": true, "SetMaxHealth": true, "TriggerTrap": true,
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if selector, ok := lhs.(*ast.SelectorExpr); ok && forbiddenFields[selector.Sel.Name] {
					t.Errorf("observer writes result field %s", selector.Sel.Name)
				}
			}
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok && forbiddenCalls[selector.Sel.Name] {
				t.Errorf("observer supplies result via %s", selector.Sel.Name)
			}
		}
		return true
	})
}
