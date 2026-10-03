package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EGreatSwordMissileNPCBlockStateIncludesExecutingHead(t *testing.T) {
	ud := new(server.MonsterUpdateData)
	ud.AIStackInd = 1
	ud.AIStack[0].Action = uint32(ai.ACTION_WAIT)
	ud.AIStack[1].Action = uint32(ai.ACTION_WEAPON_BLOCK)
	if ud.HasScheduledAction(ai.ACTION_WEAPON_BLOCK) || !e2eGreatSwordNPCBlockState(ud) {
		t.Fatal("the executing block was confused with a queued action")
	}
	ud.AIStack[0].Action = uint32(ai.ACTION_WEAPON_BLOCK)
	ud.AIStack[1].Action = uint32(ai.ACTION_WAIT)
	if !ud.HasScheduledAction(ai.ACTION_WEAPON_BLOCK) || e2eGreatSwordNPCBlockState(ud) {
		t.Fatal("a queued block was observed as the executing action")
	}
	ud.AIStackInd = -1
	if e2eGreatSwordNPCBlockState(ud) || e2eGreatSwordNPCBlockState(nil) {
		t.Fatal("an absent block state was observed")
	}
}

func TestE2EGreatSwordMissileModesAndSchedule(t *testing.T) {
	for _, direction := range []string{"player-to-npc", "npc-to-player", "", "player-to-monster"} {
		for _, facing := range []string{"front", "rear", "", "side"} {
			for level := -1; level <= 6; level++ {
				mode := direction + "/" + facing
				fromNPC, front, ok := e2eGreatSwordMissileMode(level, mode)
				want := (direction == "player-to-npc" || direction == "npc-to-player") &&
					(facing == "front" || facing == "rear") &&
					(level >= 1 && level <= 5 || level == 0 && direction == "npc-to-player")
				if ok != want || ok && (fromNPC != (direction == "npc-to-player") || front != (facing == "front")) {
					t.Fatalf("mode=%s/%d result=%t/%t/%t want=%t", mode, level, fromNPC, front, ok, want)
				}
				if !ok {
					continue
				}
				var sc e2eScenario
				sc.CheckGreatSwordMissileDefense(level, mode, "GreatSword")
				if len(sc.steps) != 7 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 ||
					sc.steps[5].ready == nil || sc.steps[5].waitTimeout != 300 || sc.steps[5].fnc == nil {
					t.Fatal("GreatSword lacks bounded actual cast/block/hit and cleanup")
				}
			}
		}
	}
}

func TestE2EGreatSwordMissileScenario(t *testing.T) {
	const path = "../scripts/e2e/host-warrior-greatsword-missile-defense.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	greatSwordEquip := false
	for _, step := range file.Steps {
		if step.Action == "click-inventory-item" && step.Item == "GreatSword" {
			greatSwordEquip = true
		}
		if step.Action == "click-inventory-item" && step.Item == "WoodenShield" {
			t.Fatal("GreatSword already dequips the shield; clicking it would re-equip it")
		}
		if step.Action != "check-greatsword-missile-defense" {
			continue
		}
		if _, _, ok := e2eGreatSwordMissileMode(step.Count, step.Text); !ok {
			t.Fatalf("invalid fixture: %+v", step)
		}
		key := step.Text + "/" + string(rune('0'+step.Count))
		if seen[key] {
			t.Fatal("duplicate fixture:", key)
		}
		seen[key] = true
	}
	if len(seen) != 21 || !seen["npc-to-player/front/0"] || !greatSwordEquip {
		t.Fatalf("fixtures=%v want all levels/facings/directions and natural NPC", seen)
	}
	var sc e2eScenario
	sc.Load(path)
	count := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " actual block or hit") {
			count++
		}
	}
	if count != 21 {
		t.Fatalf("loaded outcomes=%d want=21", count)
	}
}

func TestE2EGreatSwordMissileDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_greatsword_missile_defense.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Cur": true, "Buffs": true, "State": true, "Action": true, "WeaponEquip": true,
		"WeaponEquipFlags": true, "ObjOwner": true, "VelVec": true, "Direction1": true,
		"Field76": true, "Field75": true, "Field547": true, "Field546": true, "Field21": true,
		"Field1": true, "Field0": true, "Obj130": true, "Field131": true, "Damage": true, "UpdateData": true}
	calls := map[string]bool{"CallDamage": true, "DrawImageAt": true, "PlayerDamageNative4E17B0": true,
		"SparkExplosionCollide4E9AC0": true, "ApplyEnchant": true, "BuffOff": true, "SetOwner": true,
		"ObjSetOwner": true, "SpellProjectileReflect4E0A70": true, "PlayerSetState": true,
		"NetSendPacketXxx": true, "NetSendPacketXxx0": true, "DelayedDelete": true, "SetHealth": true, "SetMaxHealth": true}
	ast.Inspect(file, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok && fields[sel.Sel.Name] {
					t.Errorf("observer writes result %s", sel.Sel.Name)
				}
			}
		case *ast.CallExpr:
			if sel, ok := n.Fun.(*ast.SelectorExpr); ok && calls[sel.Sel.Name] {
				t.Errorf("observer supplies result via %s", sel.Sel.Name)
			}
		}
		return true
	})
}
