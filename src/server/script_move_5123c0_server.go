package server

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// ScriptMoveRuntime5123C0 supplies services still owned by the legacy root.
// MoverTypeID is a callback rather than a value because GAME.EXE reloads the
// global type ID in the direct branch and on every object-list iteration.
type ScriptMoveRuntime5123C0 struct {
	MoverTypeID func() uint32
	SetOn       func(*Object)
}

func scriptMoveStoreArgU325123C0(action *AIStackItem, index int, value uint32) {
	action.Args[index] = action.Args[index]&^uintptr(0xffffffff) | uintptr(value)
}

func scriptMoveStoreArgLow5123C0(action *AIStackItem, index int, value uint8) {
	action.Args[index] = action.Args[index]&^uintptr(0xff) | uintptr(value)
}

func scriptMoveStoreWaypoint5123C0(action *AIStackItem, index int, waypoint *Waypoint) {
	action.Args[index] = uintptr(unsafe.Pointer(waypoint))
}

func (s *Server) moverGoToNative5124C0(unit *Object, waypoint *Waypoint, runtime ScriptMoveRuntime5123C0) {
	moverGoTo5124C0(unit, waypoint, moverGoToHooks5124C0[*Object, *Waypoint, *MoverUpdateData]{
		loadUpdate: func(unit *Object) *MoverUpdateData {
			return (*MoverUpdateData)(unit.UpdateData)
		},
		setOn: runtime.SetOn,
		storeVelocityXBits: func(unit *Object, bits uint32) {
			unit.VelVec.X = math.Float32frombits(bits)
		},
		storeVelocityYBits: func(unit *Object, bits uint32) {
			unit.VelVec.Y = math.Float32frombits(bits)
		},
		storeStateLow: func(update *MoverUpdateData, value uint8) {
			update.Field_0 = value
		},
		loadWaypointIndex: func(waypoint *Waypoint) uint32 {
			return waypoint.Index
		},
		storeWaypointIndex: func(update *MoverUpdateData, value uint32) {
			update.Field_2 = int32(value)
		},
		addToUpdatable: s.Objs.AddToUpdatable,
	})
}

// ScriptMoveTo5123C0 binds GAME.EXE 005123C0 and 005124C0 to native-width
// Object, Waypoint, MonsterUpdateData, MoverUpdateData, and AIStackItem
// pointers. It intentionally preserves the original fault contract after the
// root-level script API has performed its historical nil/dead checks.
func (s *Server) ScriptMoveTo5123C0(unit *Object, waypoint *Waypoint, runtime ScriptMoveRuntime5123C0) {
	scriptMoveTo5123C0(unit, waypoint, scriptMoveHooks5123C0[
		*Object,
		*Waypoint,
		*MonsterUpdateData,
		*MoverUpdateData,
		*AIStackItem,
	]{
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadMonsterUpdate: func(unit *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(unit.UpdateData)
		},
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		storeActionArgU32: scriptMoveStoreArgU325123C0,
		loadWaypointPoints: func(waypoint *Waypoint) uint8 {
			return waypoint.PointsCnt
		},
		storeActionWaypoint: scriptMoveStoreWaypoint5123C0,
		loadRoamFlagsLow: func(update *MonsterUpdateData) uint8 {
			return uint8(update.Field333)
		},
		storeActionArgLow: scriptMoveStoreArgLow5123C0,
		loadWaypointXBits: func(waypoint *Waypoint) uint32 {
			return math.Float32bits(waypoint.PosVec.X)
		},
		loadWaypointYBits: func(waypoint *Waypoint) uint32 {
			return math.Float32bits(waypoint.PosVec.Y)
		},
		loadMoverType: runtime.MoverTypeID,
		loadTypeInd: func(unit *Object) uint16 {
			return unit.TypeInd
		},
		moverGoTo: func(unit *Object, waypoint *Waypoint) {
			s.moverGoToNative5124C0(unit, waypoint, runtime)
		},
		firstObject: s.Objs.First,
		loadMoverUpdate: func(unit *Object) *MoverUpdateData {
			return (*MoverUpdateData)(unit.UpdateData)
		},
		loadExtent: func(unit *Object) uint32 {
			return unit.Extent
		},
		loadMoverTargetExtent: func(update *MoverUpdateData) uint32 {
			return update.Field_8
		},
		nextObject: func(unit *Object) *Object {
			return unit.Next()
		},
	})
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[2-unsafe.Sizeof(Object{}.TypeInd)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.Extent)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.VelVec.X)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Waypoint{}.Index)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Waypoint{}.PosVec.X)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(Waypoint{}.PointsCnt)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterUpdateData{}.Field333)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MoverUpdateData{}.Field_2)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MoverUpdateData{}.Field_8)]
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(AIStackItem{}.Args[0])]
)
