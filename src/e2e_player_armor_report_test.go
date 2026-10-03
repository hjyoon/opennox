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

func TestE2EPlayerArmorReportOutcomeSensitivity(t *testing.T) {
	baseline := math.Float32bits(0.125)
	valid := e2ePlayerArmorReportOutcome{
		current: math.Float32bits(0.375), cached: math.Float32bits(0.375), client: math.Float32bits(0.375),
		equipped: true, held: true,
	}
	if err := valid.validate(baseline, true); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*e2ePlayerArmorReportOutcome)
	}{
		{"stale report cache", func(o *e2ePlayerArmorReportOutcome) { o.cached = baseline }},
		{"stale client receiver", func(o *e2ePlayerArmorReportOutcome) { o.client = baseline }},
		{"one-bit cache corruption", func(o *e2ePlayerArmorReportOutcome) { o.cached ^= 1 }},
		{"one-bit client corruption", func(o *e2ePlayerArmorReportOutcome) { o.client ^= 1 }},
		{"no server armor change", func(o *e2ePlayerArmorReportOutcome) { o.current, o.cached, o.client = baseline, baseline, baseline }},
		{"armor decreased", func(o *e2ePlayerArmorReportOutcome) { o.current, o.cached, o.client = 0, 0, 0 }},
		{"not equipped", func(o *e2ePlayerArmorReportOutcome) { o.equipped = false }},
		{"item lost", func(o *e2ePlayerArmorReportOutcome) { o.held = false }},
		{"infinite armor", func(o *e2ePlayerArmorReportOutcome) {
			o.current, o.cached, o.client = 0x7f800000, 0x7f800000, 0x7f800000
		}},
		{"unordered armor", func(o *e2ePlayerArmorReportOutcome) {
			o.current, o.cached, o.client = 0x7fc00001, 0x7fc00001, 0x7fc00001
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			corrupted := valid
			tc.mutate(&corrupted)
			if err := corrupted.validate(baseline, true); err == nil {
				t.Fatal("invalid armor report result passed")
			}
		})
	}
	for _, bad := range []uint32{0x7f800000, 0xff800000, 0x7fc00000, 0xff800001} {
		if err := valid.validate(bad, true); err == nil {
			t.Fatalf("invalid baseline %08x passed", bad)
		}
	}
	dequipped := e2ePlayerArmorReportOutcome{current: baseline, cached: baseline, client: baseline, held: true}
	if err := dequipped.validate(baseline, false); err != nil {
		t.Fatal(err)
	}
	dequipped.current, dequipped.cached, dequipped.client = valid.current, valid.cached, valid.client
	if err := dequipped.validate(baseline, false); err == nil {
		t.Fatal("dequip without restoration passed")
	}
}

func TestE2EPlayerArmorReportSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckPlayerArmorReport("LeatherArmor", "armor")
	if len(sc.steps) != 21 || sc.steps[0].ready == nil || sc.steps[0].waitTimeout != 1200 {
		t.Fatalf("armor observer steps=%d, want bounded preparation and two full cycles", len(sc.steps))
	}
	for _, index := range []int{3, 5, 8, 10, 13, 15, 18, 20} {
		step := sc.steps[index]
		if step.ready == nil || step.fnc == nil || step.waitTimeout != 120 || !strings.HasSuffix(step.name, " report") {
			t.Fatalf("step %d lacks a bounded live report check", index)
		}
	}
	for _, index := range []int{1, 6, 11, 16} {
		if !strings.HasSuffix(sc.steps[index].name, " actual inventory input") || sc.steps[index].fnc == nil ||
			sc.steps[index+1].fnc == nil || sc.steps[index+1].time-sc.steps[index].time != 1 {
			t.Fatalf("step %d lacks real inventory press/release", index)
		}
	}
	for _, index := range []int{4, 9, 14, 19} {
		if sc.steps[index].time-sc.steps[index-1].time != 12 || sc.steps[index].fnc != nil ||
			!strings.HasSuffix(sc.steps[index+1].name, " stable report") {
			t.Fatalf("step %d lacks independent settling and recheck", index)
		}
	}
}

func TestE2EPlayerArmorReportPublicScenario(t *testing.T) {
	const path = "../scripts/e2e/host-game-armor-report.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	grants, checks, inventories, initialDequips := 0, 0, 0, 0
	for _, step := range file.Steps {
		switch step.Action {
		case "grant-item":
			if step.Item != "LeatherArmor" || step.Count != 1 {
				t.Fatal("armor report scenario must grant one stock LeatherArmor")
			}
			grants++
		case "check-player-armor-report":
			if step.Item != "LeatherArmor" {
				t.Fatal("observer must use the granted armor")
			}
			checks++
		case "inventory":
			inventories++
		case "click-inventory-item":
			if step.Item != "LeatherArmor" {
				t.Fatal("stock grant must be dequipped through actual inventory input")
			}
			initialDequips++
		case "click", "wait", "slow", "quit":
		default:
			t.Fatalf("armor report scenario uses an unapproved action: %s", step.Action)
		}
	}
	if grants != 1 || checks != 1 || inventories != 1 || initialDequips != 1 {
		t.Fatalf("scenario grant/check/inventory/initial-dequip=%d/%d/%d/%d", grants, checks, inventories, initialDequips)
	}
	var sc e2eScenario
	sc.Load(path)
	stable := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " stable report") {
			stable++
		}
	}
	if stable != 4 {
		t.Fatalf("loaded stable reports=%d, want two equip/dequip cycles", stable)
	}
}

func TestE2EPlayerArmorReportDoesNotSupplyEquipmentOrReportResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_player_armor_report.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	forbiddenFields := map[string]bool{
		"Field57": true, "Field58": true, "ObjFlags": true, "ArmorEquip": true, "ArmorEquipFlags": true,
		"InvHolder": true, "UpdateData": true, "Player": true,
	}
	forbiddenCalls := map[string]bool{
		"PlayerArmorReport4D992A": true, "NetSendPacketXxx0": true, "NetSendPacketXxx": true,
		"nox_xxx_playerEquipArmor_53E650": true, "sub_53E430": true,
		"Nox_xxx_playerTryEquip_4F2F70": true, "SetFlags": true,
		"SetArmor": true, "Uint32Ptr": true, "PtrOff": true,
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if selector, ok := lhs.(*ast.SelectorExpr); ok && forbiddenFields[selector.Sel.Name] {
					t.Errorf("observer writes equipment/report result field %s", selector.Sel.Name)
				}
				if _, ok := lhs.(*ast.StarExpr); ok {
					t.Error("observer writes through a raw pointer")
				}
			}
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok && forbiddenCalls[selector.Sel.Name] {
				t.Errorf("observer supplies equipment/report result via %s", selector.Sel.Name)
			}
		}
		return true
	})
}
