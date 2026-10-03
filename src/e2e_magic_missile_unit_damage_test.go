package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/types"
	"gopkg.in/yaml.v2"
)

func TestE2EMagicMissileDamageRoundingAndSplash(t *testing.T) {
	for _, tc := range []struct {
		raw          int32
		armor, carry float32
		damage       int32
		next         uint32
	}{
		{9, 0.5, 0, 4, math.Float32bits(0.5)},
		{11, 0.5, 0, 6, math.Float32bits(-0.5)},
		{9, 0.5, 0.4, 5, math.Float32bits(float32(4.5) + float32(0.4) - 5)},
		{1, 1, 0, 1, 0},
		{0, 0, 0, 1, 0},
		{-9, 0.5, 0, -4, math.Float32bits(-0.5)},
	} {
		damage, carry := e2eMagicMissileDamage(tc.raw, tc.armor, tc.carry)
		if damage != tc.damage || math.Float32bits(carry) != tc.next {
			t.Fatalf("raw=%d armor=%g carry=%g got=%d/%#x want=%d/%#x", tc.raw, tc.armor, tc.carry, damage, math.Float32bits(carry), tc.damage, tc.next)
		}
	}
	for _, tc := range []struct {
		distance float32
		damage   int32
		ok       bool
	}{
		{0, 20, true}, {5, 20, true}, {10, 15, true}, {15, 10, true}, {25, 0, true}, {26, 0, false},
	} {
		amount, ok := e2eMagicMissileSplash(20, 25, types.Ptf(100, 100), types.Ptf(100+tc.distance, 100))
		if amount != tc.damage || ok != tc.ok {
			t.Fatalf("splash distance=%g got=%d/%t want=%d/%t", tc.distance, amount, ok, tc.damage, tc.ok)
		}
	}
	if _, ok := e2eMagicMissileSplash(20, 5, types.Pointf{}, types.Pointf{}); ok {
		t.Fatal("invalid splash radius accepted")
	}
}

func TestE2EMagicMissileUnitModesScheduleAndScenario(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player", "", "player-to-monster"} {
		for level := -1; level <= 6; level++ {
			fromNPC, ok := e2eMagicMissileUnitMode(level, direction)
			want := (direction == "player-to-npc" || direction == "npc-to-player") &&
				(level >= 1 && level <= 5 || direction == "npc-to-player" && level == 0)
			if ok != want || ok && fromNPC != (direction == "npc-to-player") {
				t.Fatalf("mode=%s/%d got=%t/%t want=%t", direction, level, fromNPC, ok, want)
			}
			if ok {
				var sc e2eScenario
				sc.CheckMagicMissileUnitDamage(level, direction, "Magic Missile")
				if len(sc.steps) != 5 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 ||
					sc.steps[3].ready == nil || sc.steps[3].waitTimeout != 300 || sc.steps[3].fnc == nil {
					t.Fatal("Magic Missile lacks bounded setup/hit/client/deletion/cleanup")
				}
			}
		}
	}
	const path = "../scripts/e2e/host-game-magic-missile-unit-damage.yaml"
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
		if step.Action != "check-magic-missile-unit-damage" {
			continue
		}
		if _, ok := e2eMagicMissileUnitMode(step.Count, step.Text); !ok {
			t.Fatalf("invalid public fixture: %+v", step)
		}
		key := step.Text + "/" + string(rune('0'+step.Count))
		if seen[key] {
			t.Fatal("duplicate fixture:", key)
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
		if strings.HasSuffix(step.name, " actual hits and client replay") {
			count++
		}
	}
	if count != 11 {
		t.Fatalf("loaded outcomes=%d want=11", count)
	}
}

func TestE2EMagicMissileUnitDamageDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_magic_missile_unit_damage.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Buffs": true, "Target": true, "Field76": true, "Field75": true, "Field547": true,
		"Field546": true, "Field21": true, "Field1": true, "Obj130": true, "Field131": true, "Frame134": true,
		"Damage": true, "Collide": true, "Power": true, "UpdateData": true}
	calls := map[string]bool{"CallDamage": true, "DrawImageAt": true, "BoomCollide4E9770": true,
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
