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

func TestE2EPlayerItemEnchantmentReportOutcomeSensitivity(t *testing.T) {
	valid := e2ePlayerItemEnchantmentReportOutcome{current: 1, cached: 1, client: 1, equipped: true, held: true, weapon: true}
	if err := valid.validate(0, 1, true); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*e2ePlayerItemEnchantmentReportOutcome)
	}{
		{"stale cache", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.cached = 0 }},
		{"stale client", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.client = 0 }},
		{"wrong server bit", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.current = 2 }},
		{"upper bits changed", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.current |= 0x100 }},
		{"wrong cache bit", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.cached = 2 }},
		{"wrong client bit", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.client = 2 }},
		{"not equipped", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.equipped = false }},
		{"wrong equipped weapon", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.weapon = false }},
		{"item lost", func(o *e2ePlayerItemEnchantmentReportOutcome) { o.held = false }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := valid
			tc.mutate(&bad)
			if err := bad.validate(0, 1, true); err == nil {
				t.Fatal("invalid item enchantment report passed")
			}
		})
	}
	for _, tc := range []struct{ baseline, effect uint32 }{{0, 0}, {1, 1}} {
		if err := valid.validate(tc.baseline, tc.effect, true); err == nil {
			t.Fatal("ineffective fixture passed")
		}
	}
	dequipped := e2ePlayerItemEnchantmentReportOutcome{held: true}
	if err := dequipped.validate(0, 1, false); err != nil {
		t.Fatal(err)
	}
	dequipped.current, dequipped.cached, dequipped.client = 1, 1, 1
	if err := dequipped.validate(0, 1, false); err == nil {
		t.Fatal("dequip without mask restoration passed")
	}
	valid.current = 0xa1b2c301
	if err := valid.validate(0xa1b2c300, 1, true); err != nil {
		t.Fatal("unchanged upper bits rejected:", err)
	}
}

func TestE2EPlayerItemEnchantmentReportSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckPlayerItemEnchantmentReport("Sword", "effect")
	if len(sc.steps) != 29 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 {
		t.Fatalf("steps=%d, want bounded preparation and two full mouse/hover cycles", len(sc.steps))
	}
	for base := 1; base < 29; base += 7 {
		if !strings.HasSuffix(sc.steps[base].name, " actual inventory input") || sc.steps[base].fnc == nil ||
			sc.steps[base+1].fnc == nil || sc.steps[base+1].time-sc.steps[base].time != 1 {
			t.Fatalf("step %d lacks actual mouse press/release", base)
		}
		for _, offset := range []int{2, 4, 6} {
			if step := sc.steps[base+offset]; step.ready == nil || step.waitTimeout != 120 {
				t.Fatalf("step %d lacks bounded outcome observation", base+offset)
			}
		}
		if sc.steps[base+3].time-sc.steps[base+2].time != 12 || sc.steps[base+3].fnc != nil ||
			!strings.HasSuffix(sc.steps[base+4].name, " stable report") ||
			!strings.HasSuffix(sc.steps[base+5].name, " hover input") || sc.steps[base+5].fnc == nil ||
			!strings.HasSuffix(sc.steps[base+6].name, " tooltip") || sc.steps[base+6].time-sc.steps[base+5].time != 1 {
			t.Fatalf("step %d lacks settled report and actual hover validation", base)
		}
	}
}

func TestE2EPlayerItemEnchantmentReportPublicScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-player-item-enchantment-report.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	grants, checks, inventories := 0, 0, 0
	for _, step := range file.Steps {
		switch step.Action {
		case "grant-engage-item":
			if step.Item != "Sword" || step.Modifier != "FireProtect1" || step.Mask != 1 {
				t.Fatal("scenario must grant one stock Sword with native FireProtect1")
			}
			grants++
		case "check-player-item-enchantment-report":
			if step.Item != "Sword" {
				t.Fatal("observer must use the granted Sword")
			}
			checks++
		case "inventory":
			inventories++
		case "click", "wait", "slow", "quit":
		default:
			t.Fatalf("unapproved scenario action: %s", step.Action)
		}
	}
	if grants != 1 || checks != 1 || inventories != 1 {
		t.Fatalf("grant/check/inventory=%d/%d/%d", grants, checks, inventories)
	}
	var sc e2eScenario
	sc.Load(path)
	stable, tooltips := 0, 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " stable report") {
			stable++
		}
		if strings.HasSuffix(step.name, " tooltip") && step.ready != nil {
			tooltips++
		}
	}
	if stable != 4 || tooltips != 4 {
		t.Fatalf("stable reports/tooltips=%d/%d, want two full equip/dequip cycles", stable, tooltips)
	}
}

func TestE2EPlayerItemEnchantmentReportDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_player_item_enchantment_report.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	forbiddenFields := map[string]bool{
		"Field110": true, "Field2172": true, "Poison540": true, "Buffs": true, "ObjFlags": true,
		"InvHolder": true, "UpdateData": true, "Player": true, "EquippedWeapon": true,
	}
	forbiddenCalls := map[string]bool{
		"PlayerReportItemEnchantment4D99A7": true, "NetSendPacketXxx0": true, "NetSendPacketXxx": true,
		"Nox_xxx_playerTryEquip_4F2F70": true, "SetFlags": true, "Uint8Ptr": true, "PtrUint8": true,
		"PtrUint32": true, "DrawImageAt": true, "DrawImage16": true, "Sub_413480": true,
		"nox_xxx_cursorSetTooltip_4776B0": true, "Nox_xxx_gLoadImg_42F970": true,
		"CallEngage": true, "CallDisengage": true,
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if selector, ok := lhs.(*ast.SelectorExpr); ok && forbiddenFields[selector.Sel.Name] {
					t.Errorf("observer writes result field %s", selector.Sel.Name)
				}
				if _, ok := lhs.(*ast.StarExpr); ok {
					t.Error("observer writes through a raw pointer")
				}
			}
		case *ast.CallExpr:
			name := ""
			switch fun := node.Fun.(type) {
			case *ast.SelectorExpr:
				name = fun.Sel.Name
			case *ast.Ident:
				name = fun.Name
			}
			if forbiddenCalls[name] {
				t.Errorf("observer supplies a result via %s", name)
			}
		}
		return true
	})
}
