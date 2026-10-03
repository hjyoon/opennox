package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestE2EPlayerQuestKeysReportOutcomeSensitivity(t *testing.T) {
	for _, counts := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {2, 1}, {2, 2}} {
		valid := e2ePlayerQuestKeysReportOutcome{inventory: counts}
		for kind, count := range counts {
			valid.clientInventory[kind], valid.clientFound[kind] = uint32(count), count != 0
			if count != 0 {
				valid.cached[kind], valid.client[kind] = 1, 1
			}
		}
		if err := valid.validate(counts); err != nil {
			t.Fatal(err)
		}
		for kind := 0; kind < 2; kind++ {
			for _, corrupt := range []func(*e2ePlayerQuestKeysReportOutcome){
				func(o *e2ePlayerQuestKeysReportOutcome) { o.inventory[kind]++ },
				func(o *e2ePlayerQuestKeysReportOutcome) { o.clientInventory[kind]++ },
				func(o *e2ePlayerQuestKeysReportOutcome) { o.clientFound[kind] = !o.clientFound[kind] },
				func(o *e2ePlayerQuestKeysReportOutcome) { o.cached[kind] ^= 1 },
				func(o *e2ePlayerQuestKeysReportOutcome) { o.client[kind] ^= 1 },
				func(o *e2ePlayerQuestKeysReportOutcome) { o.cached[kind] = 2 },
				func(o *e2ePlayerQuestKeysReportOutcome) { o.client[kind] = 255 },
			} {
				bad := valid
				corrupt(&bad)
				if err := bad.validate(counts); err == nil {
					t.Fatalf("corrupted key report passed: counts=%v kind=%d state=%+v", counts, kind, bad)
				}
			}
		}
	}
	for _, counts := range [][2]int{{-1, 0}, {0, -1}, {-1, -1}} {
		if err := (e2ePlayerQuestKeysReportOutcome{}).validate(counts); err == nil {
			t.Fatalf("negative expected counts passed: %v", counts)
		}
	}
}

func TestE2EPlayerQuestKeysReportSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckPlayerQuestKeysReport(2, 1, "keys")
	if len(sc.steps) != 3 || sc.steps[0].ready == nil || sc.steps[0].fnc == nil || sc.steps[0].waitTimeout != 1200 ||
		sc.steps[2].ready == nil || sc.steps[2].fnc == nil || sc.steps[2].waitTimeout != 120 ||
		sc.steps[1].fnc != nil || sc.steps[1].time-sc.steps[0].time != 12 || sc.steps[2].time != sc.steps[1].time ||
		!strings.HasSuffix(sc.steps[2].name, " stable report") {
		t.Fatalf("Quest key checks lack bounded live/stable observations: %+v", sc.steps)
	}
}

func TestE2EPlayerQuestKeysReportPublicScenario(t *testing.T) {
	const path = "../scripts/e2e/host-quest-key-report.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	grants := make(map[string]int)
	drops := make(map[string]int)
	exits, inventories, questClicks, amountChecks, amountAccepts, amountClosed := 0, 0, 0, 0, 0, 0
	var checks [][2]int
	for _, step := range file.Steps {
		switch step.Action {
		case "grant-item":
			if step.Item != "SilverKey" && step.Item != "GoldKey" || step.Count != 1 {
				t.Fatal("Quest key scenario must only grant normal stock keys one at a time")
			}
			grants[step.Item]++
		case "drag-inventory-item-out":
			if step.Item != "SilverKey" && step.Item != "GoldKey" {
				t.Fatal("Quest key scenario must only attempt normal manual key drops")
			}
			drops[step.Item]++
		case "check-player-quest-keys-report":
			checks = append(checks, [2]int{step.Count, step.Amount})
		case "enter-quest-exit":
			exits++
		case "inventory":
			inventories++
		case "assert-item-amount":
			if step.Amount != 1 || step.Max != 2 || step.Price != 0 {
				t.Fatal("Quest key drop must observe the actual one-of-two stack dialog")
			}
			amountChecks++
		case "click-item-amount-accept":
			amountAccepts++
		case "assert-item-amount-closed":
			amountClosed++
		case "click":
			if step.X == 184 && step.Y == 200 {
				questClicks++
			}
		case "wait", "slow", "quit":
		default:
			t.Fatalf("Quest key scenario supplies an unapproved action: %s", step.Action)
		}
	}
	if !reflect.DeepEqual(grants, map[string]int{"SilverKey": 2, "GoldKey": 2}) ||
		!reflect.DeepEqual(drops, map[string]int{"SilverKey": 1, "GoldKey": 1}) || exits != 1 || inventories != 2 || questClicks != 1 ||
		amountChecks != 2 || amountAccepts != 2 || amountClosed != 2 ||
		!reflect.DeepEqual(checks, [][2]int{{0, 0}, {1, 0}, {1, 1}, {2, 1}, {2, 2}, {2, 2}, {2, 2}, {2, 2}}) {
		t.Fatalf("Quest key scenario contract: grants=%v drops=%v exits=%d inventory=%d Quest-clicks=%d checks=%v",
			grants, drops, exits, inventories, questClicks, checks)
	}
	var sc e2eScenario
	sc.Load(path)
	stable := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " stable report") {
			stable++
		}
	}
	if stable != 8 {
		t.Fatalf("loaded Quest key stable observations=%d want=8", stable)
	}
}

func TestE2EPlayerQuestKeysReportDoesNotSupplyKeyOrPacketResults(t *testing.T) {
	for _, path := range []string{"e2e_player_quest_keys_report.go", "legacy/player_quest_keys_client.go"} {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		forbiddenFields := map[string]bool{
			"QuestPlayerFlagsA": true, "QuestPlayerFlagsB": true, "tail_padding": true,
			"InvFirstItem": true, "InvNextItem": true, "InvHolder": true, "ObjFlags": true,
			"UpdateData": true, "Player": true, "PlayerInd": true, "NetCode": true, "NetCodeVal": true,
		}
		forbiddenCalls := map[string]bool{
			"PlayerQuestKeysReport4D9A3F": true, "questKeyReport4D9DF0": true, "NetSendPacketXxx0": true, "NetSendPacketXxx": true,
			"ClientSetPlayerNetCode": true, "SetFlags": true, "SetGame": true, "UnsetGame": true,
			"Sub_4ED0C0": true, "DelayedDelete": true, "Delete": true, "Nox_xxx_inventoryPutImpl_4F3070": true,
			"Nox_xxx_playerRespawnItem_4EF750": true, "Uint8Ptr": true, "Uint32Ptr": true, "PtrOff": true, "PtrT": true,
		}
		var inspectWrite func(ast.Expr)
		inspectWrite = func(expr ast.Expr) {
			switch expr := expr.(type) {
			case *ast.SelectorExpr:
				if forbiddenFields[expr.Sel.Name] {
					t.Errorf("%s observer writes key/report field %s", path, expr.Sel.Name)
				}
			case *ast.IndexExpr:
				inspectWrite(expr.X)
			case *ast.StarExpr:
				t.Errorf("%s observer writes through a raw pointer", path)
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				for _, lhs := range node.Lhs {
					inspectWrite(lhs)
				}
			case *ast.IncDecStmt:
				inspectWrite(node.X)
			case *ast.CallExpr:
				name := ""
				if selector, ok := node.Fun.(*ast.SelectorExpr); ok {
					name = selector.Sel.Name
				} else if ident, ok := node.Fun.(*ast.Ident); ok {
					name = ident.Name
				}
				if forbiddenCalls[name] {
					t.Errorf("%s observer supplies key/report results via %s", path, name)
				}
			}
			return true
		})
	}
}
