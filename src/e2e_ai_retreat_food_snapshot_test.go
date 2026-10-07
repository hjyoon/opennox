package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestE2EAIRetreatFoodSnapshotAdmission(t *testing.T) {
	for _, mode := range []string{"valid", "nil-unit", "nil-food", "nil-update", "nil-health", "not-monster", "no-head", "invalid-head"} {
		t.Run(mode, func(t *testing.T) {
			update := &server.MonsterUpdateData{}
			unit := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update), HealthData: &server.HealthData{}}
			food := &server.Object{}
			switch mode {
			case "nil-unit":
				unit = nil
			case "nil-food":
				food = nil
			case "nil-update":
				unit.UpdateData = nil
			case "nil-health":
				unit.HealthData = nil
			case "not-monster":
				unit.ObjClass = object.ClassPlayer
			case "no-head":
				update.AIStackInd = -1
			case "invalid-head":
				update.AIStackInd = int8(len(update.AIStack))
			}
			before := *update
			_, ok := e2eAIRetreatFoodSnapshot(unit, food, 655, 30)
			if ok != (mode == "valid") || *update != before {
				t.Fatalf("diagnostic admission=%t mode=%s; live update must remain untouched", ok, mode)
			}
		})
	}
}

func TestE2EAIRetreatFoodSnapshotImmutableNamedActions(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_WAIT, ai.ACTION_IDLE} {
		for _, wrap := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/wrap-%t", action, wrap), func(t *testing.T) {
				frame, corner := uint32(655), uint32(650)
				if wrap {
					frame, corner = 1, math.MaxUint32-3
				}
				update := &server.MonsterUpdateData{
					AIStackInd: 2, Field124: 630, Field127: corner,
					Field125: math.Float32bits(3425.7507), Field126: math.Float32bits(2310.5396),
					Field2: 3, Field67: 2, Field71: 0xabcdef00, StatusFlags: object.MonStatusFrustrated,
				}
				update.AIStack[0].Action = uint32(ai.DEPENDENCY_NOT_CORNERED)
				update.AIStack[1].Action = uint32(ai.ACTION_RETREAT)
				update.AIStack[2].Action = uint32(action)
				// A diagnostic prints the full native word without casting or
				// dereferencing it. Even malformed targets must be observable.
				tracked := uintptr(^uint32(0))
				if unsafe.Sizeof(tracked) == 8 {
					tracked = tracked<<16 | 0xb110
				}
				update.AIStack[2].Args[2] = tracked
				unit := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update),
					PosVec: types.Ptf(3426.8457, 2309.9163), HealthData: &server.HealthData{Cur: 76, Max: 80}}
				food := &server.Object{PosVec: types.Ptf(3428, 2308)}
				beforeUnit, beforeFood, beforeUpdate, beforeHealth := *unit, *food, *update, *unit.HealthData
				state, ok := e2eAIRetreatFoodSnapshot(unit, food, frame, 30)
				wantTracked := uintptr(0)
				if action == ai.ACTION_MOVE_TO {
					wantTracked = tracked
				}
				if !ok || state.frame != frame || state.fps != 30 || state.progressFrame != 630 || state.cornerFrame != corner ||
					state.position != unit.PosVec || state.target != food.PosVec || state.currentHP != 76 || state.maximumHP != 80 ||
					state.pathCount != 3 || state.pathCursor != 2 || state.pathStatus != 0xabcdef00 ||
					state.status != object.MonStatusFrustrated || state.head != action || state.headTarget != wantTracked ||
					!reflect.DeepEqual(state.stack, []ai.ActionType{ai.DEPENDENCY_NOT_CORNERED, ai.ACTION_RETREAT, action}) {
					t.Fatalf("live diagnostic snapshot changed: %#v", state)
				}
				printed := state.String()
				if !strings.Contains(printed, "head="+action.String()) || !strings.Contains(printed, "corner-age=5 ") ||
					!strings.Contains(printed, fmt.Sprintf("tracked=%#x ", wantTracked)) ||
					!reflect.DeepEqual(*unit, beforeUnit) || !reflect.DeepEqual(*food, beforeFood) || *update != beforeUpdate || *unit.HealthData != beforeHealth {
					t.Fatal("named diagnostic or immutable live-state contract changed")
				}
				update.AIStack[2].Action = uint32(ai.ACTION_DEAD)
				if state.String() != printed {
					t.Fatal("captured stack aliases live action state")
				}
				runtime.KeepAlive(unit)
			})
		}
	}
}

func TestE2EAIRetreatFoodSnapshotReadsOnly(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_ai_retreat_food_snapshot.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				for _, lhs := range node.Lhs {
					if sel, ok := lhs.(*ast.SelectorExpr); ok && sel.Sel.Name != "headTarget" {
						t.Errorf("diagnostic writes live state: %s", sel.Sel.Name)
					}
					if _, ok := lhs.(*ast.StarExpr); ok {
						t.Error("diagnostic writes through a pointer")
					}
				}
			case *ast.CallExpr:
				if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
					switch sel.Sel.Name {
					case "Class", "Has", "UpdateDataMonster", "GetAIStack", "Type", "Float32frombits", "Ptf", "Sprintf":
					default:
						t.Errorf("diagnostic calls gameplay or RNG service: %s", sel.Sel.Name)
					}
				}
			}
			return true
		})
	}
}
