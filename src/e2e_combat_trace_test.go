package opennox

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestE2ECombatTraceSnapshotAdmission(t *testing.T) {
	for _, name := range []string{"nil", "other-class", "player-no-update", "monster-no-update", "no-health"} {
		t.Run(name, func(t *testing.T) {
			unit := &server.Object{ObjClass: object.ClassPlayer, HealthData: &server.HealthData{Cur: 17, Max: 20, Field2: 20}}
			switch name {
			case "nil":
				unit = nil
			case "other-class":
				unit.ObjClass = object.ClassWeapon
			case "monster-no-update":
				unit.ObjClass = object.ClassMonster
			case "no-health":
				unit.HealthData = nil
			}
			state, ok := e2eCombatTraceSnapshot(unit)
			want := name != "nil" && name != "other-class"
			if ok != want {
				t.Fatalf("admission=%t want=%t", ok, want)
			}
			if !ok {
				if state != (e2eCombatTraceUnit{}) {
					t.Fatal("rejected unit produced a partial observation")
				}
			} else if state.Pointer != uintptr(unsafe.Pointer(unit)) || state.Update != 0 || state.StackIndex != -1 {
				t.Fatal("missing update must stay observable without inventing an action")
			}
		})
	}
}

func TestE2ECombatTraceSnapshotPlayerNativeGraph(t *testing.T) {
	player := &server.Player{WeaponEquip: 0x100}
	player.Info().SetField2239(37)
	weapon, cursor, source := new(server.Object), new(server.Object), new(server.Object)
	update := &server.PlayerUpdateData{Player: player, EquippedWeapon: weapon, CursorObj: cursor,
		State: server.PlayerState1, Field59_0: 3, Stamina: 83, Field0: 708, Field57: 0x3cf5c28f}
	unit := &server.Object{ObjClass: object.ClassPlayer, TypeInd: 713, UpdateData: unsafe.Pointer(update),
		HealthData: &server.HealthData{Cur: 14, Max: 20, Field2: 20}, Obj130: source, Field34: 687, Frame134: 700}
	unit.PosVec.X, unit.PosVec.Y = math.Float32frombits(0x4589a400), math.Float32frombits(0x45038800)
	unit.PrevPos, unit.VelVec, unit.Direction1, unit.Direction2 = unit.PosVec, unit.PosVec, 7, 8
	beforeUnit, beforeUpdate, beforePlayer, beforeHealth := *unit, *update, *player, *unit.HealthData
	state, ok := e2eCombatTraceSnapshot(unit)
	if !ok || state.Player != uintptr(unsafe.Pointer(player)) || state.Weapon != uintptr(unsafe.Pointer(weapon)) ||
		state.Cursor != uintptr(unsafe.Pointer(cursor)) || state.DamageSource != uintptr(unsafe.Pointer(source)) ||
		state.Update != uintptr(unsafe.Pointer(update)) || state.HealthPointer != uintptr(unsafe.Pointer(unit.HealthData)) ||
		state.Health != [3]uint16{14, 20, 20} || state.PositionBits != [2]uint32{0x4589a400, 0x45038800} ||
		state.PreviousBits != state.PositionBits || state.VelocityBits != state.PositionBits || state.Direction != [2]uint16{7, 8} ||
		state.State != uint8(server.PlayerState1) || state.Animation != 3 || state.Stamina != 83 || state.Strength != 37 ||
		state.AttackStart != 687 || state.DamageFrame != 700 || state.StateDeadline != 708 || state.ArmorBits != 0x3cf5c28f || state.Equipment != 0x100 {
		t.Fatalf("player graph observation=%+v", state)
	}
	if !reflect.DeepEqual(*unit, beforeUnit) || !reflect.DeepEqual(*update, beforeUpdate) ||
		!reflect.DeepEqual(*player, beforePlayer) || *unit.HealthData != beforeHealth {
		t.Fatal("observation modified live player records")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && state.Pointer <= math.MaxUint32 {
		t.Fatal("fixture did not exercise a native high address")
	}
}

func TestE2ECombatTraceSnapshotMonsterNativeWords(t *testing.T) {
	for _, name := range []string{"live-head", "empty-stack", "invalid-stack"} {
		t.Run(name, func(t *testing.T) {
			update := &server.MonsterUpdateData{AIStackInd: 2, Field120_0: 8, Field120_1: 5,
				Field120_2: 1, Field120_3: 0, Field128: 731, StatusFlags: 5, Stamina: 57,
				WeaponEquipFlags: 0x100, Field282_1: 1, CurrentEnemy: new(server.Object), PreferredEnemy: new(server.Object)}
			update.AIStack[0].Action, update.AIStack[1].Action, update.AIStack[2].Action = uint32(ai.ACTION_IDLE), uint32(ai.ACTION_FIGHT), uint32(ai.ACTION_MELEE_ATTACK)
			// Retain malformed/high action words as evidence, never dereference.
			high := ^uintptr(0) - 15
			update.AIStack[2].Args = [4]uintptr{0x3f800000, 0x80000000, high, high - 16}
			switch name {
			case "empty-stack":
				update.AIStackInd = -1
			case "invalid-stack":
				update.AIStackInd = int8(len(update.AIStack))
			}
			unit := &server.Object{ObjClass: object.ClassMonster, TypeInd: 1314, ObjSubClass: 70210,
				UpdateData: unsafe.Pointer(update), HealthData: &server.HealthData{Cur: 29, Max: 40, Field2: 40}, Field34: 689}
			beforeUnit, beforeUpdate, beforeHealth := *unit, *update, *unit.HealthData
			state, ok := e2eCombatTraceSnapshot(unit)
			if !ok || state.StackIndex != update.AIStackInd || state.Animation != 5 || state.AnimationFlags != [3]uint8{8, 1, 0} ||
				state.AttackDeadline != 731 || state.Status != 5 || state.Stamina != 57 || state.Health != [3]uint16{29, 40, 40} ||
				state.Current != uintptr(unsafe.Pointer(update.CurrentEnemy)) || state.Preferred != uintptr(unsafe.Pointer(update.PreferredEnemy)) || state.Seen != 1 {
				t.Fatalf("monster snapshot=%+v", state)
			}
			if name == "live-head" {
				if state.HeadArgs != update.AIStack[2].Args || state.Stack[2] != uint32(ai.ACTION_MELEE_ATTACK) {
					t.Fatal("native stack/head arguments changed")
				}
			} else if state.HeadArgs != [4]uintptr{} || state.Stack != [24]uint32{} {
				t.Fatal("invalid/empty stack was dereferenced")
			}
			if !reflect.DeepEqual(*unit, beforeUnit) || *update != beforeUpdate || *unit.HealthData != beforeHealth {
				t.Fatal("observer modified monster graph")
			}
			data, err := json.Marshal(state)
			if err != nil {
				t.Fatal(err)
			}
			update.AIStack[2].Args[2], unit.HealthData.Cur = 0, 0
			after, err := json.Marshal(state)
			if err != nil || string(data) != string(after) {
				t.Fatal("snapshot aliases live health/action memory")
			}
		})
	}
}

func TestE2ECombatTraceDisabledDoesNotInspectWorld(t *testing.T) {
	oldEnabled, oldPlay, oldState := e2eCombatTraceEnabled, e2ePlay, e2eCombatTrace
	t.Cleanup(func() { e2eCombatTraceEnabled, e2ePlay, e2eCombatTrace = oldEnabled, oldPlay, oldState })
	for _, mode := range []string{"disabled", "no-playback", "no-server"} {
		t.Run(mode, func(t *testing.T) {
			e2eCombatTraceEnabled, e2ePlay = mode != "disabled", "fixture.yaml"
			if mode == "no-playback" {
				e2ePlay = ""
			}
			before := e2eCombatTrace
			e2eCombatTraceTick()
			if e2eCombatTrace != before {
				t.Fatal("inactive observer changed its state")
			}
		})
	}
}

func TestE2ECombatTraceWritesOnlyOwnedEvidence(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_combat_trace.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var ownedLHS func(ast.Expr) bool
	ownedLHS = func(expr ast.Expr) bool {
		switch expr := expr.(type) {
		case *ast.Ident:
			return expr.Name == "state" || expr.Name == "e2eCombatTrace"
		case *ast.SelectorExpr:
			return ownedLHS(expr.X)
		case *ast.IndexExpr:
			return ownedLHS(expr.X)
		}
		return false
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if _, local := lhs.(*ast.Ident); !local && !ownedLHS(lhs) {
					t.Errorf("observer writes outside owned evidence: %T", lhs)
				}
			}
		case *ast.IncDecStmt:
			if id, ok := node.X.(*ast.Ident); !ok || id.Name != "i" {
				t.Error("observer advances a live counter")
			}
		case *ast.CallExpr:
			if id, ok := node.Fun.(*ast.Ident); ok {
				switch id.Name {
				case "uintptr", "uint16", "uint32", "uint8", "int", "len", "e2eCombatTraceSnapshot":
				default:
					t.Errorf("observer calls unreviewed function: %s", id.Name)
				}
			}
			if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
				switch sel.Sel.Name {
				case "Class", "Has", "HasAny", "Pointer", "Float32bits", "UpdateDataPlayer", "UpdateDataMonster", "Info", "Field2239",
					"Getenv", "Printf", "HostUnit", "Frame", "TickRate", "Index", "Marshal", "Nox_xxx_getProjectileClassById413250":
				default:
					t.Errorf("observer calls gameplay/input/RNG mutation: %s", sel.Sel.Name)
				}
			}
		}
		return true
	})
}
