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

func TestE2EPlayerHammerTargetPlacement(t *testing.T) {
	radii := [2]float32{15, 12}
	distances, err := e2ePlayerHammerTargetDistances(10, radii, 20)
	if err != nil || distances != [2]float32{61, 30} {
		t.Fatalf("stock lane = %v/%v", distances, err)
	}
	if distances[1] <= 10+radii[1] || distances[0]-distances[1] <= radii[0]+radii[1] {
		t.Fatal("fixture colliders overlap")
	}
	for i, distance := range distances {
		if math.Abs(float64(distance-35))-float64(radii[i]) >= 20 {
			t.Fatal("fixture outside hammer's original +35 center / range")
		}
	}
	if _, err := e2ePlayerHammerTargetDistances(100, radii, 20); err == nil {
		t.Fatal("oversized colliders accepted outside stock range")
	}
	if _, err := e2ePlayerHammerTargetDistances(10, radii, 1); err == nil {
		t.Fatal("out-of-range target accepted")
	}
}

func TestE2EPlayerHammerSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckPlayerHammer("hammer")
	if len(sc.steps) != 12 {
		t.Fatalf("steps=%d, want 12 (two attacks, no new PNG baselines)", len(sc.steps))
	}
	for _, index := range []int{0, 1, 4, 5, 6, 9, 10} {
		step := sc.steps[index]
		if step.ready == nil || step.fnc == nil || step.waitTimeout == 0 {
			t.Fatalf("step %d has no bounded live check", index)
		}
	}
	for _, index := range []int{1, 6} {
		if !strings.HasSuffix(sc.steps[index].name, " naturally recovered attack stamina") ||
			sc.steps[index].waitTimeout != 120 || sc.steps[index].time-sc.steps[index-1].time != 12 {
			t.Fatalf("step %d must await natural stamina without forcing recovery", index)
		}
	}
	for _, index := range []int{4, 9} {
		if !strings.HasSuffix(sc.steps[index].name, " actual monster and NPC damage and client effects") || sc.steps[index].waitTimeout != 120 {
			t.Fatalf("step %d has no bounded actual hit/effects check", index)
		}
	}
	for _, index := range []int{5, 10} {
		if !strings.HasSuffix(sc.steps[index].name, " natural attack completion and equipment") || sc.steps[index].waitTimeout != 360 {
			t.Fatalf("step %d has no bounded natural completion check", index)
		}
	}
}

func TestE2EPlayerHammerOutcomeSensitivity(t *testing.T) {
	valid := e2ePlayerHammerOutcome{
		before: [2]uint16{2000, 2000}, after: [2]uint16{1862, 1901}, attributed: [2]bool{true, true},
		started: true, animation: true, advanced: true, quake: true,
		completed: true, held: true, stable: true, sounds: 1,
	}
	if err := valid.validate(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*e2ePlayerHammerOutcome)
	}{
		{"no monster damage", func(o *e2ePlayerHammerOutcome) { o.after[0] = o.before[0] }},
		{"no NPC damage", func(o *e2ePlayerHammerOutcome) { o.after[1] = o.before[1] }},
		{"NPC healed", func(o *e2ePlayerHammerOutcome) { o.after[1] = o.before[1] + 1 }},
		{"dead target", func(o *e2ePlayerHammerOutcome) { o.after[0] = 0 }},
		{"uninitialized target", func(o *e2ePlayerHammerOutcome) { o.before[1] = 0 }},
		{"unattributed monster", func(o *e2ePlayerHammerOutcome) { o.attributed[0] = false }},
		{"unattributed NPC", func(o *e2ePlayerHammerOutcome) { o.attributed[1] = false }},
		{"attack not started", func(o *e2ePlayerHammerOutcome) { o.started = false }},
		{"missing animation", func(o *e2ePlayerHammerOutcome) { o.animation = false }},
		{"stuck animation", func(o *e2ePlayerHammerOutcome) { o.advanced = false }},
		{"missing quake", func(o *e2ePlayerHammerOutcome) { o.quake = false }},
		{"no sound", func(o *e2ePlayerHammerOutcome) { o.sounds = 0 }},
		{"duplicate sound", func(o *e2ePlayerHammerOutcome) { o.sounds = 2 }},
		{"attack not completed", func(o *e2ePlayerHammerOutcome) { o.completed = false }},
		{"weapon lost", func(o *e2ePlayerHammerOutcome) { o.held = false }},
		{"late extra damage", func(o *e2ePlayerHammerOutcome) { o.stable = false }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			corrupted := valid
			tc.mutate(&corrupted)
			if err := corrupted.validate(); err == nil {
				t.Fatal("invalid hammer result passed")
			}
		})
	}
}

func TestE2EPlayerHammerPublicScenario(t *testing.T) {
	const path = "../scripts/e2e/host-warrior-hammer.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	grant, equip, check := 0, 0, 0
	for _, step := range file.Steps {
		switch step.Action {
		case "grant-item":
			if step.Item != "WarHammer" || step.Count != 1 {
				t.Fatal("hammer scenario must grant only one stock WarHammer")
			}
			grant++
		case "click-inventory-item":
			if step.Item != "WarHammer" {
				t.Fatal("hammer must be equipped through real inventory input")
			}
			equip++
		case "check-player-hammer":
			check++
		case "cast", "damage-monster", "damage-player", "set-player-buff", "set-player-health", "set-monster-health", "screen":
			t.Fatalf("hammer scenario injects a result or replaces a PNG baseline: %s", step.Action)
		}
	}
	if grant != 1 || equip != 1 || check != 1 {
		t.Fatalf("scenario grant/equip/check=%d/%d/%d", grant, equip, check)
	}
	var sc e2eScenario
	sc.Load(path)
	hits, completions := 0, 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " actual monster and NPC damage and client effects") {
			hits++
		}
		if strings.HasSuffix(step.name, " natural attack completion and equipment") {
			completions++
		}
	}
	if hits != 2 || completions != 2 {
		t.Fatalf("loaded real attacks=%d completions=%d, want two each", hits, completions)
	}
}

func TestE2EPlayerHammerDoesNotSupplyAttackResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_player_hammer.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	forbiddenFields := map[string]bool{
		"Cur": true, "State": true, "Stamina": true, "Buffs": true,
		"Field0": true, "Field34": true, "Field59_0": true,
		"Obj130": true, "Field131": true, "Frame134": true,
		"AnimInd": true, "AnimFrameSlave": true, "Jiggle12": true,
	}
	forbiddenCalls := map[string]bool{
		"CallDamage": true, "Damage": true, "SetHealth": true, "SetState": true, "BuffOn": true,
		"EventObj": true, "PlayerSubStamina4F7D30": true, "PlayerAdjustStamina4F7DB0": true,
		"SetMouseState": true, "Nox_xxx_playerAttack_538960": true, "PlayerInputAttack4F9C70": true,
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if selector, ok := lhs.(*ast.SelectorExpr); ok && forbiddenFields[selector.Sel.Name] {
					t.Errorf("observer writes attack result field %s", selector.Sel.Name)
				}
			}
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok && forbiddenCalls[selector.Sel.Name] {
				t.Errorf("observer supplies attack result via %s", selector.Sel.Name)
			}
		}
		return true
	})
}
