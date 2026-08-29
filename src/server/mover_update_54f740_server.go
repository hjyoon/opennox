package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// MoverUpdateRuntime54F740 supplies the one legacy movement effect that is
// outside the server package. Object and waypoint identities stay as native
// pointers; the fixed-width update record contains only serialized PE32
// compatibility words.
type MoverUpdateRuntime54F740 struct {
	Move func(*Object, types.Pointf)
}

type moverUpdateNativeDeps54F740 struct {
	objectByExtent  func(uint32) *Object
	waypointByID    func(uint32) *Waypoint
	randomInt       func(int, int) int
	move            func(*Object, types.Pointf)
	removeUpdatable func(*Object)
}

func moverUpdateNative54F740(source *Object, deps moverUpdateNativeDeps54F740) {
	moverUpdate54F740(source, moverUpdateHooks54F740[*Object, *MoverUpdateData, *Waypoint]{
		loadUpdateData: func(obj *Object) *MoverUpdateData {
			return obj.UpdateDataMover()
		},
		loadTargetExtent: func(data *MoverUpdateData) uint32 {
			return data.Field_8
		},
		loadTarget: func(data *MoverUpdateData) *Object {
			return source.MoverTargetFor(data)
		},
		storeTarget: func(data *MoverUpdateData, target *Object) {
			source.SetMoverTargetFor(data, target)
		},
		objectByExtent: deps.objectByExtent,
		loadWaypointID: func(data *MoverUpdateData, slot int) uint32 {
			switch slot {
			case 2:
				return uint32(data.Field_2)
			case 3:
				return data.Field_4
			case 5:
				return data.Field_6
			default:
				panic("Mover waypoint ID slot must be 2, 3, or 5")
			}
		},
		loadWaypoint: func(data *MoverUpdateData, slot int) *Waypoint {
			return source.MoverWaypointFor(data, slot)
		},
		storeWaypoint: func(data *MoverUpdateData, slot int, waypoint *Waypoint) {
			source.SetMoverWaypointFor(data, slot, waypoint)
		},
		waypointByID: deps.waypointByID,
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadState: func(data *MoverUpdateData) uint8 {
			return data.Field_0
		},
		storeState: func(data *MoverUpdateData, state uint8) {
			data.Field_0 = state
		},
		loadSpeedUnits: func(data *MoverUpdateData) int32 {
			// GAME.EXE uses FILD on the raw dword at offset 4 even though the
			// serialized declaration names this field as float.
			return int32(math.Float32bits(data.Field_1))
		},
		storeSpeedBase: func(obj *Object, speed float32) {
			obj.SpeedBase = speed
		},
		storeSpeedCur: func(obj *Object, speed float32) {
			obj.SpeedCur = speed
		},
		loadSpeedCur: func(obj *Object) float32 {
			return obj.SpeedCur
		},
		loadPosition: func(obj *Object) types.Pointf {
			return obj.PosVec
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadVelocityX: func(obj *Object) float32 {
			return obj.VelVec.X
		},
		loadVelocityY: func(obj *Object) float32 {
			return obj.VelVec.Y
		},
		storeVelocityX: func(obj *Object, value float32) {
			obj.VelVec.X = value
		},
		storeVelocityY: func(obj *Object, value float32) {
			obj.VelVec.Y = value
		},
		loadWaypointPos: func(waypoint *Waypoint) types.Pointf {
			return waypoint.PosVec
		},
		loadWaypointX: func(waypoint *Waypoint) float32 {
			return waypoint.PosVec.X
		},
		loadWaypointY: func(waypoint *Waypoint) float32 {
			return waypoint.PosVec.Y
		},
		waypointPointCnt: func(waypoint *Waypoint) uint8 {
			return waypoint.PointsCnt
		},
		waypointPoint: func(waypoint *Waypoint, index int) *Waypoint {
			return waypoint.Points[index].Waypoint
		},
		randomInt:       deps.randomInt,
		move:            deps.move,
		removeUpdatable: deps.removeUpdatable,
	})
}

// MoverUpdate54F740 binds GAME.EXE 0054F740 to native-width server state.
func (s *Server) MoverUpdate54F740(source *Object, runtime MoverUpdateRuntime54F740) {
	moverUpdateNative54F740(source, moverUpdateNativeDeps54F740{
		objectByExtent: s.ObjectByExtent4ED020,
		waypointByID: func(id uint32) *Waypoint {
			return s.WPs.ByInd(int(id))
		},
		randomInt:       s.Rand.Logic.IntClamp,
		move:            runtime.Move,
		removeUpdatable: s.Objs.RemoveFromUpdatable,
	})
}
