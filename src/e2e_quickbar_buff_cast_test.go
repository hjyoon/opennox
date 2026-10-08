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

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EQuickbarBuffSpecsAndBoundedSchedule(t *testing.T) {
	for id := spell.ID(0); id <= 137; id++ {
		buff, key, ok := e2eQuickbarBuffSpec(id)
		want := map[spell.ID]server.EnchantID{
			spell.SPELL_HASTE:                  server.ENCHANT_HASTED,
			spell.SPELL_PROTECTION_FROM_FIRE:   server.ENCHANT_PROTECT_FROM_FIRE,
			spell.SPELL_PROTECTION_FROM_POISON: server.ENCHANT_PROTECT_FROM_POISON,
		}
		_, allowed := want[id]
		if ok != allowed || ok && (buff != want[id] || key == "") {
			t.Fatalf("spell %d spec = %d/%q/%t, want supported=%t", id, buff, key, ok, allowed)
		}
		var sc e2eScenario
		if !allowed {
			func() {
				defer func() {
					if recover() == nil || len(sc.steps) != 0 {
						t.Errorf("unsupported spell %d scheduled gameplay", id)
					}
				}()
				sc.CheckQuickbarBuffCast(id, "invalid")
			}()
			continue
		}
		sc.CheckQuickbarBuffCast(id, "live cast")
		for suffix, timeout := range map[string]time.Duration{
			"prepare ordinary waiting NPC fixtures":                 1200,
			"self=true observe natural cursor selection":            120,
			"self=false observe natural cursor selection":           120,
			"self=true observe incantation mana and client effect":  120,
			"self=false observe incantation mana and client effect": 120,
			"observe natural buff expiry":                           6200,
		} {
			count := 0
			for _, step := range sc.steps {
				if strings.HasSuffix(step.name, suffix) {
					count++
					if step.ready == nil || step.fnc == nil || step.waitTimeout != timeout {
						t.Errorf("spell %d missing bounded observation %s", id, suffix)
					}
				}
			}
			if count != 1 {
				t.Errorf("spell %d observation %s count=%d", id, suffix, count)
			}
		}
		for _, mode := range []string{"true", "false"} {
			for _, suffix := range []string{"release real spell shortcut", "actual quickbar effect"} {
				count := 0
				for _, step := range sc.steps {
					if strings.Contains(step.name, "self="+mode+" "+suffix) {
						count++
					}
				}
				if count != 1 {
					t.Errorf("spell %d %s %s count=%d", id, mode, suffix, count)
				}
			}
		}
	}
}

func TestE2EQuickbarBuffPublicScenariosAndDispatch(t *testing.T) {
	for _, tc := range []struct {
		class  string
		ids    []int
		classX int
	}{
		{"wizard", []int{36, 62}, 768}, {"conjurer", []int{64}, 512},
	} {
		t.Run(tc.class, func(t *testing.T) {
			path := "../scripts/e2e/solo-" + tc.class + "-quickbar-buff.yaml"
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var file e2eFileYML
			if err := yaml.Unmarshal(raw, &file); err != nil {
				t.Fatal(err)
			}
			seen := map[int]int{}
			classClicks := 0
			for _, step := range file.Steps {
				if step.Action == "check-quickbar-buff-cast" {
					seen[step.Spell]++
				}
				if step.Action == "click" && step.Name == "select "+tc.class && step.X == tc.classX && step.Y == 208 {
					classClicks++
				}
				switch step.Action {
				case "cast", "set-quickbar-spell", "set-player-health", "set-player-mana", "set-player-buff", "damage-player", "damage-monster":
					t.Errorf("public scenario bypasses the observer via %s", step.Action)
				}
			}
			if classClicks != 1 || len(seen) != len(tc.ids) {
				t.Fatalf("class clicks=%d cases=%v", classClicks, seen)
			}
			for _, id := range tc.ids {
				if seen[id] != 1 {
					t.Errorf("missing/repeated real caster/spell %s/%d", tc.class, id)
				}
			}
			if file.Steps[len(file.Steps)-1].Action != "quit" {
				t.Fatal("missing native terminal shutdown")
			}
			var sc e2eScenario
			sc.Load(path)
			for _, mode := range []string{"true", "false"} {
				n := 0
				for _, step := range sc.steps {
					if strings.Contains(step.name, "self="+mode+" observe incantation mana and client effect") {
						n++
					}
				}
				if n != len(tc.ids) {
					t.Errorf("actual target mode %s scheduled %d casts, want %d", mode, n, len(tc.ids))
				}
			}
		})
	}
}

func TestE2EQuickbarBuffObservesInsteadOfSupplyingResults(t *testing.T) {
	tree, err := parser.ParseFile(token.NewFileSet(), "e2e_quickbar_buff_cast.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	protected := map[string]bool{
		"CursorObj": true, "CursorVec": true, "Obj3640": true, "SpellCastStart": true, "SpellPhonemeLeaf": true,
		"ManaCur": true, "ManaPrev": true, "ManaMax": true, "ProtUnitManaCur": true,
		"Buffs": true, "BuffsDur": true, "BuffsPower": true, "Cur": true, "NetCode": true,
		"Spells8": true, "SpellInd28": true, "Field32": true, "Field36": true, "Frame40": true, "Field48": true,
	}
	forbiddenCalls := map[string]bool{
		"CastSpell": true, "CastSpellLvl": true, "CastSpellByUser4FDD20": true, "SpellAccept4FD400": true,
		"PlayerSpell": true, "SpellBookInsert4FE340": true, "Nox_xxx_quickBarSetSpell": true,
		"BuffApply": true, "BuffOff": true, "ApplyEnchant": true, "SetMana": true,
		"SpellManaCharge4FCF90": true, "Nox_xxx_playerManaSub_4EEBF0": true,
		"AddToMsgListCli": true, "onPacketTrySpell51BAD0": true, "ChangeMousePos": true,
	}
	ast.Inspect(tree, func(n ast.Node) bool {
		var writes []ast.Expr
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			writes = stmt.Lhs
		case *ast.IncDecStmt:
			writes = []ast.Expr{stmt.X}
		}
		for _, lhs := range writes {
			ast.Inspect(lhs, func(part ast.Node) bool {
				if field, ok := part.(*ast.SelectorExpr); ok && protected[field.Sel.Name] {
					t.Errorf("fixture supplies live result %s", field.Sel.Name)
				}
				if id, ok := part.(*ast.Ident); ok && (id.Name == "magicEntityHead" || id.Name == "gameFrame") {
					t.Errorf("fixture supplies queue/frame %s", id.Name)
				}
				return true
			})
		}
		if call, ok := n.(*ast.CallExpr); ok {
			name := ""
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			case *ast.Ident:
				name = fn.Name
			}
			if forbiddenCalls[name] {
				t.Errorf("fixture bypasses input or supplies result via %s", name)
			}
			if name == "DelayedDelete" {
				if len(call.Args) != 1 {
					t.Error("invalid fixture retirement")
				} else if field, ok := call.Args[0].(*ast.SelectorExpr); !ok || field.Sel.Name != "npc" && field.Sel.Name != "control" {
					t.Error("fixture retires a non-fixture unit")
				}
			}
		}
		return true
	})
	if raw, err := os.ReadFile("e2e_quickbar_buff_cast.go"); err != nil {
		t.Fatal(err)
	} else {
		for _, required := range []string{"noxServer.TickHook(f.observe)", "sc.CheckBookRewardDefault(id, false", "ClientQuickbarNugget(f.slot)", "Pressed: true", "Pressed: false", "magicEntityHead", "SpellPhonemeLeaf", "e2eClientHUDMeter(1)"} {
			if !strings.Contains(string(raw), required) {
				t.Error(fmt.Sprintf("missing real path observation %s", required))
			}
		}
	}
}
