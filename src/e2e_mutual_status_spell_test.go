package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/server"
	"gopkg.in/yaml.v2"
)

func TestE2EMutualStatusOpaquePixelsRejectsBackgroundAndMissingEffect(t *testing.T) {
	rect := image.Rect(0, 0, 4, 1)
	black, white, live := noximage.NewImage16(rect), noximage.NewImage16(rect), noximage.NewImage16(rect)
	copy(black.Pix, []uint16{0, 0x1234, 0x2222, 0x3333})
	copy(white.Pix, []uint16{0xffff, 0x1234, 0x7777, 0x3333})
	copy(live.Pix, []uint16{0, 0x1234, 0x2222, 0x4444})
	if m, n := e2eMutualStatusOpaquePixels(black, white, live, rect); m != 1 || n != 2 {
		t.Fatalf("opaque pixels = %d/%d, want 1/2 (exclude transparent/partial alpha)", m, n)
	}
	clear(live.Pix)
	if m, n := e2eMutualStatusOpaquePixels(black, white, live, rect); m != 0 || n != 2 {
		t.Fatalf("absent effect = %d/%d, want 0/2", m, n)
	}
}

func TestE2EMutualStatusModesAndSchedule(t *testing.T) {
	for _, kind := range []string{"confused", "stun", "slow", "freeze", "blind"} {
		for _, dir := range []string{"player-to-npc", "npc-to-player"} {
			got, id, fromNPC, ok := e2eMutualStatusMode(kind + "/" + dir)
			if !ok || got != kind || id == 0 || fromNPC != (dir == "npc-to-player") {
				t.Fatalf("mode %s/%s", kind, dir)
			}
			var sc e2eScenario
			sc.CheckMutualStatusSpell(kind+"/"+dir, "status")
			if len(sc.steps) != 7 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 || sc.steps[3].ready == nil ||
				sc.steps[4].ready == nil || sc.steps[5].ready == nil || sc.steps[5].fnc == nil || sc.steps[5].waitTimeout != 3000 {
				t.Fatal("unbounded/absent status observer phases")
			}
		}
	}
	for _, mode := range []string{"", "stun", "stun/player-to-monster", "stun/npc-to-player/extra", "foo/player-to-npc"} {
		if _, _, _, ok := e2eMutualStatusMode(mode); ok {
			t.Fatalf("invalid mode accepted: %s", mode)
		}
	}
}

func TestE2EMutualStatusIndependentMechanics(t *testing.T) {
	for _, tc := range []struct {
		kind            string
		player, warrior bool
		mass            float32
		balance         float64
		buff            server.EnchantID
		dur             uint16
	}{
		{"confused", false, false, 30, 89.5, server.ENCHANT_CONFUSED, 90},
		{"stun", true, false, 30, 60.5, server.ENCHANT_HELD, 60},
		{"stun", true, true, 1, 60.5, server.ENCHANT_SLOWED, 60},
		{"stun", false, false, 15, 60, server.ENCHANT_HELD, 60},
		{"stun", false, false, math.Nextafter32(15, 16), 60, server.ENCHANT_SLOWED, 60},
		{"stun", false, false, float32(math.NaN()), 60, server.ENCHANT_HELD, 60},
		{"slow", false, false, 1, 60.9, server.ENCHANT_SLOWED, 60},
		{"freeze", true, false, 1, 0, server.ENCHANT_FREEZE, 120},
		{"blind", false, false, 1, 0, server.ENCHANT_BLINDED, 120},
	} {
		buff, dur := e2eMutualStatusExpected(tc.kind, tc.player, tc.warrior, tc.mass, tc.balance, 30)
		if buff != tc.buff || dur != tc.dur {
			t.Fatalf("%+v got=%d/%d", tc, buff, dur)
		}
	}
}

func TestE2EMutualStatusPublicScenario(t *testing.T) {
	raw, err := os.ReadFile("../scripts/e2e/host-game-mutual-status-spells.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, step := range file.Steps {
		if step.Action != "check-mutual-status-spell" {
			continue
		}
		if _, _, _, ok := e2eMutualStatusMode(step.Text); !ok || seen[step.Text] {
			t.Fatalf("invalid/duplicate status mode %s", step.Text)
		}
		seen[step.Text] = true
	}
	if len(seen) != 10 {
		t.Fatalf("status directions=%v", seen)
	}
	var sc e2eScenario
	sc.Load("../scripts/e2e/host-game-mutual-status-spells.yaml")
	count := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " natural expiry") {
			count++
		}
	}
	if count != 10 {
		t.Fatalf("loaded status outcomes=%d", count)
	}
}

func TestE2EMutualStatusDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_mutual_status_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]bool{"Buffs": true, "Cur": true, "Dur": true, "Power": true, "Level16": true, "Target": true, "Field0": true, "Field8": true, "UpdateData": true, "Obj130": true, "Damage": true, "Frame134": true}
	calls := map[string]bool{"ApplyEnchant": true, "BuffOff": true, "BuffApply4FF380": true, "SpellAccept4FD400": true, "CallDamage": true, "SetHealth": true, "SetMaxHealth": true, "DrawImageAt": true, "NetSendPacketXxx": true, "NetSendPacketXxx0": true, "DelayedDelete": true}
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
				if sel, ok := n.Fun.(*ast.SelectorExpr); ok && calls[sel.Sel.Name] && !(fn.Name.Name == "cleanup" && sel.Sel.Name == "DelayedDelete") {
					t.Errorf("observer supplies result via %s", sel.Sel.Name)
				}
			}
			return true
		})
	}
}
