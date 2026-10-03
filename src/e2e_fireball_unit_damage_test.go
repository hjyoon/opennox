package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestE2EFireballUnitDamageModesAndSchedule(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player", "", "player-to-monster"} {
		for level := -1; level <= 6; level++ {
			fromNPC, ok := e2eFireballUnitMode(level, direction)
			want := (direction == "player-to-npc" || direction == "npc-to-player") &&
				(level >= 1 && level <= 5 || direction == "npc-to-player" && level == 0)
			if ok != want || ok && fromNPC != (direction == "npc-to-player") {
				t.Fatalf("mode=%s/%d got=%t/%t want=%t", direction, level, fromNPC, ok, want)
			}
			if !ok {
				continue
			}
			var sc e2eScenario
			sc.CheckFireballUnitDamage(level, direction, "Fireball")
			if len(sc.steps) != 5 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 ||
				sc.steps[3].ready == nil || sc.steps[3].waitTimeout != 300 || sc.steps[3].fnc == nil {
				t.Fatal("Fireball lacks bounded setup/hit/client/deletion and cleanup")
			}
		}
	}
	for level, want := range map[int]string{1: "Fireball", 2: "StrongFireball", 3: "TitanFireball", 4: "TitanFireball", 5: "TitanFireball"} {
		if got := e2eFireballUnitType(level); got != want {
			t.Fatalf("level %d type=%q want=%q", level, got, want)
		}
	}
}

func TestE2EFireballUnitDamageScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-fireball-unit-damage.yaml"
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
		if step.Action != "check-fireball-unit-damage" {
			continue
		}
		if _, ok := e2eFireballUnitMode(step.Count, step.Text); !ok {
			t.Fatalf("invalid public fixture: %+v", step)
		}
		key := step.Text + "/" + string(rune('0'+step.Count))
		if seen[key] {
			t.Fatal("duplicate Fireball fixture:", key)
		}
		seen[key] = true
	}
	if len(seen) != 11 || !seen["npc-to-player/0"] {
		t.Fatalf("casts=%v want five levels each direction and natural NPC", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	count := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " actual hit and client replay") {
			count++
		}
	}
	if count != 11 {
		t.Fatalf("loaded Fireball outcomes=%d want=11", count)
	}
}

func TestE2EFireballUnitDamageDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_fireball_unit_damage.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Buffs": true, "Field76": true, "Field75": true, "Field547": true,
		"Field546": true, "Field21": true, "Field1": true, "Obj130": true, "Field131": true, "Frame134": true,
		"Damage": true, "Collide": true, "CollideData": true, "Power": true, "UpdateData": true}
	calls := map[string]bool{"CallDamage": true, "DrawImageAt": true, "SparkExplosionCollide4E9AC0": true,
		"PlayerDamageNative4E17B0": true, "DefaultDamageWorld4E0B30": true, "ApplyEnchant": true,
		"NetSendPacketXxx": true, "NetSendPacketXxx0": true, "BuffOff": true, "DelayedDelete": true,
		"SetHealth": true, "SetMaxHealth": true}
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
						t.Errorf("observer writes result %s", sel.Sel.Name)
					}
				}
			case *ast.CallExpr:
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok && calls[sel.Sel.Name] {
					setup := fn.Name.Name == "prepare" && sel.Sel.Name == "SetMaxHealth"
					cleanup := fn.Name.Name == "cleanup" && (sel.Sel.Name == "DelayedDelete" || sel.Sel.Name == "SetHealth" || sel.Sel.Name == "SetMaxHealth")
					if !setup && !cleanup {
						t.Errorf("observer supplies result via %s in %s", sel.Sel.Name, fn.Name.Name)
					}
				}
			}
			return true
		})
	}
}
