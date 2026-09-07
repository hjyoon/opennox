package server

import (
	"image"
	"runtime"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// ProjectileCollisionRuntime537770 contains the three remaining effects that
// still belong to legacy scalar state. Object traversal, filtering, geometry,
// type caches, and callback dispatch stay native-width in server.
type ProjectileCollisionRuntime537770 struct {
	SetTraceReady  func(uint32)
	StoreTraceGrid func(image.Point)
	WallNormal     func(image.Point, types.Rectf) (types.Pointf, bool)
}

type projectileCollisionNativeDeps537770 struct {
	traceHit      func(*Object) (*Object, types.Pointf, bool)
	setTraceReady func(uint32)
	callCollide   func(unsafe.Pointer, uintptr, uintptr, uintptr)
}

func (s *Server) projectileCanCollideNative54E730(source, candidate *Object) bool {
	return projectileCanCollide54E730(source, candidate, projectileCanCollideHooks54E730[*Object]{
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadSubclassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjSubClass)
		},
		loadFlags: func(obj *Object) object.Flags {
			return obj.ObjFlags
		},
		hasCollide: func(obj *Object) bool {
			return obj.Collide != nil
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		sameTeam: UnitsHaveSameTeam4EC520,
		isEnemy:  s.IsEnemyTo,
	})
}

func (s *Server) projectileSampleObjectNative54E810(
	source *Object,
	current, previous *types.Pointf,
) *Object {
	return projectileSampleObject54E810(source, current, previous, projectileSampleObjectHooks54E810[
		*Object,
		*DoorUpdateData,
	]{
		eachAt: func(pos types.Pointf, callback func(*Object)) {
			s.Map.Sub517B70(pos, callback)
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadSubclassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjSubClass)
		},
		loadDoorUpdate: func(obj *Object) *DoorUpdateData {
			return obj.UpdateDataDoor()
		},
		loadDoorDirection: func(update *DoorUpdateData) int32 {
			return update.CurrentDirection
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		doorSize:  DoorSize,
		lineTrace: LineTraceXxx,
		canCollide: func(first, second *Object) bool {
			return s.projectileCanCollideNative54E730(first, second)
		},
		intersects: func(obj *Object, point *types.Pointf) bool {
			return obj.Sub547DB0(point)
		},
	})
}

func (s *Server) projectileTraceHitNative537850(
	source *Object,
	runtime ProjectileCollisionRuntime537770,
) (*Object, types.Pointf, bool) {
	return projectileTraceHit537850(source, projectileTraceHitHooks537850[*Object]{
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadNewPosX: func(obj *Object) float32 {
			return obj.NewPos.X
		},
		loadNewPosY: func(obj *Object) float32 {
			return obj.NewPos.Y
		},
		storeNewPosX: func(obj *Object, value float32) {
			obj.NewPos.X = value
		},
		storeNewPosY: func(obj *Object, value float32) {
			obj.NewPos.Y = value
		},
		loadObjectPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadObjectPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		sampleObject: s.projectileSampleObjectNative54E810,
		mapTrace: func(from, to types.Pointf, flags MapTraceFlags) projectileMapTraceResult537850 {
			var point types.Pointf
			var grid image.Point
			return projectileMapTraceResult537850{
				clear: s.MapTraceRayAt(from, to, &point, &grid, flags),
				point: point,
				grid:  grid,
			}
		},
		storeTraceGrid: runtime.StoreTraceGrid,
		setTraceReady:  runtime.SetTraceReady,
		wallNormal:     runtime.WallNormal,
	})
}

func (s *Server) projectileCollisionDispatchNative537770(
	source *Object,
	deps projectileCollisionNativeDeps537770,
) {
	projectileCollisionDispatch537770(source, projectileCollisionDispatchHooks537770[
		*Object,
		unsafe.Pointer,
	]{
		loadSmallCache: func() uint32 {
			return s.Types.fast.projectileSmallFist537770
		},
		storeSmallCache: func(value uint32) {
			s.Types.fast.projectileSmallFist537770 = value
		},
		loadMediumCache: func() uint32 {
			return s.Types.fast.projectileMediumFist537770
		},
		storeMediumCache: func(value uint32) {
			s.Types.fast.projectileMediumFist537770 = value
		},
		loadLargeCache: func() uint32 {
			return s.Types.fast.projectileLargeFist537770
		},
		storeLargeCache: func(value uint32) {
			s.Types.fast.projectileLargeFist537770 = value
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		loadFlagsLow: func(obj *Object) uint8 {
			return uint8(obj.ObjFlags)
		},
		traceHit: deps.traceHit,
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadCollide: func(obj *Object) unsafe.Pointer {
			return obj.Collide
		},
		callCollide: func(callback unsafe.Pointer, first, second *Object, normal *types.Pointf) {
			var pin runtime.Pinner
			pin.Pin(normal)
			defer pin.Unpin()
			deps.callCollide(
				callback,
				uintptr(first.CObj()),
				uintptr(toObjectC(second)),
				uintptr(unsafe.Pointer(normal)),
			)
			runtime.KeepAlive(normal)
		},
		setTraceReady: deps.setTraceReady,
	})
}

// ProjectileCollisionDispatch537770 binds the missile update loop to the
// native-width implementation of GAME.EXE 00537770/00537850.
func (s *Server) ProjectileCollisionDispatch537770(
	source *Object,
	runtime ProjectileCollisionRuntime537770,
) {
	s.projectileCollisionDispatchNative537770(source, projectileCollisionNativeDeps537770{
		traceHit: func(obj *Object) (*Object, types.Pointf, bool) {
			return s.projectileTraceHitNative537850(obj, runtime)
		},
		setTraceReady: runtime.SetTraceReady,
		callCollide:   ccall.CallVoidUPtr3,
	})
}
