package legacy

/*
#include <stdint.h>
#include "GAME1_1.h"
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type interestingXferNativeDeps4D0010 struct {
	firstPending        func() *server.Object
	nextPending         func(*server.Object) *server.Object
	typeXfer            func(*server.Object) unsafe.Pointer
	findPendingObject   func(uint32) *server.Object
	findPendingWaypoint func(uint32) *server.Waypoint
	mapWallSize         func() *[2]uint32
	setElevatorLink     func(*server.Object, unsafe.Pointer, *server.Object)
	setTransporterLink  func(*server.Object, *server.TransporterUpdateData, *server.Object)

	elevatorXfer      unsafe.Pointer
	elevatorShaftXfer unsafe.Pointer
	transporterXfer   unsafe.Pointer
	holeXfer          unsafe.Pointer
	exitXfer          unsafe.Pointer
	moverXfer         unsafe.Pointer
	glyphXfer         unsafe.Pointer
}

func interestingXferDelta4D0010(origin, wallSize uint32) int32 {
	return int32(origin - 23*wallSize)
}

func interestingXferTranslateFloat4D0010(value float32, delta int32) float32 {
	// GAME.EXE uses FILD m32int, FADD m32real, then one FSTP m32real.
	// Nox configures x87 for 53-bit intermediates, which binary64 models
	// before the single final binary32 spill.
	return float32(float64(delta) + float64(value))
}

// interestingXferNative4D0010 restores GAME.EXE's two-pass random-map
// placement fixup without storing native pointers in PE32 records. The first
// pass preserves every old extent as ScriptIDVal and assigns canonical new
// extents. The second pass resolves references and translates embedded map
// coordinates according to the transfer callback registered on each type.
func interestingXferNative4D0010(
	origin *[2]uint32,
	nextExtent int32,
	deps interestingXferNativeDeps4D0010,
) int32 {
	for obj := deps.firstPending(); obj != nil; obj = deps.nextPending(obj) {
		obj.ScriptIDVal = int32(obj.Extent)
		obj.Extent = uint32(nextExtent)
		nextExtent++
	}

	for obj := deps.firstPending(); obj != nil; obj = deps.nextPending(obj) {
		switch deps.typeXfer(obj) {
		case deps.elevatorXfer:
			dataPtr := obj.UpdateData
			data := (*server.ElevatorUpdateData)(dataPtr)
			target := deps.findPendingObject(data.Field_2)
			if target == nil {
				// Original failure order is extent first, pointer second.
				data.Field_2 = 0
				data.Field_1 = 0
				deps.setElevatorLink(obj, dataPtr, nil)
				break
			}
			targetExtent := target.Extent
			data.Field_1 = 0
			deps.setElevatorLink(obj, dataPtr, target)
			data.Field_2 = targetExtent

		case deps.elevatorShaftXfer:
			dataPtr := obj.UpdateData
			data := (*server.ElevatorShaftUpdateData)(dataPtr)
			target := deps.findPendingObject(data.Field_2)
			if target == nil {
				data.Field_2 = 0
				data.Field_1 = 0
				deps.setElevatorLink(obj, dataPtr, nil)
				break
			}
			targetExtent := target.Extent
			data.Field_1 = 0
			deps.setElevatorLink(obj, dataPtr, target)
			data.Field_2 = targetExtent

		case deps.transporterXfer:
			data := (*server.TransporterUpdateData)(obj.UpdateData)
			target := deps.findPendingObject(data.TargetExtent)
			if target == nil {
				data.TargetExtent = 0
				data.TargetPE32 = 0
				deps.setTransporterLink(obj, data, nil)
				break
			}
			targetExtent := target.Extent
			data.TargetPE32 = 0
			deps.setTransporterLink(obj, data, target)
			data.TargetExtent = targetExtent

		case deps.holeXfer:
			data := (*server.HoleCollideData)(obj.CollideData)
			wallSize := deps.mapWallSize()
			wallWidth := wallSize[0]
			originX := origin[0]
			oldX := data.DestinationX
			dx := interestingXferDelta4D0010(originX, wallWidth)
			oldY := data.DestinationY
			data.DestinationX = int32(uint32(oldX) + uint32(dx))
			wallHeight := wallSize[1]
			originY := origin[1]
			dy := interestingXferDelta4D0010(originY, wallHeight)
			data.DestinationY = int32(uint32(oldY) + uint32(dy))

		case deps.exitXfer:
			data := (*server.ExitCollideData)(obj.CollideData)
			wallSize := deps.mapWallSize()
			wallWidth := wallSize[0]
			originX := origin[0]
			dx := interestingXferDelta4D0010(originX, wallWidth)
			oldX := data.DestinationX
			data.DestinationX = interestingXferTranslateFloat4D0010(oldX, dx)
			wallHeight := wallSize[1]
			originY := origin[1]
			dy := interestingXferDelta4D0010(originY, wallHeight)
			oldY := data.DestinationY
			data.DestinationY = interestingXferTranslateFloat4D0010(oldY, dy)

		case deps.moverXfer:
			data := (*server.MoverUpdateData)(obj.UpdateData)
			if waypoint := deps.findPendingWaypoint(uint32(data.Field_2)); waypoint != nil {
				data.Field_2 = int32(waypoint.Index)
			} else {
				data.Field_2 = 0
			}
			if target := deps.findPendingObject(data.Field_8); target != nil {
				data.Field_8 = target.Extent
			} else {
				data.Field_8 = 0
			}

		case deps.glyphXfer:
			data := (*server.GlyphInitData)(obj.InitData)
			wallSize := deps.mapWallSize()
			wallWidth := wallSize[0]
			originX := origin[0]
			dx := interestingXferDelta4D0010(originX, wallWidth)
			oldX := data.SpellArg.Pos.X
			data.SpellArg.Pos.X = interestingXferTranslateFloat4D0010(oldX, dx)
			wallHeight := wallSize[1]
			originY := origin[1]
			dy := interestingXferDelta4D0010(originY, wallHeight)
			oldY := data.SpellArg.Pos.Y
			data.SpellArg.Pos.Y = interestingXferTranslateFloat4D0010(oldY, dy)
		}
	}
	return nextExtent
}

func interestingXferRuntimeDeps4D0010() interestingXferNativeDeps4D0010 {
	s := GetServer().S()
	return interestingXferNativeDeps4D0010{
		firstPending: func() *server.Object {
			return s.Objs.Pending
		},
		nextPending: (*server.Object).Next,
		typeXfer: func(obj *server.Object) unsafe.Pointer {
			// GAME.EXE obtains the type name from TypeInd, then performs a
			// second lookup by that name before reading the XFer callback.
			// Keep both lookups on the current server's table as the legacy
			// exports do; pending objects need not resolve their own handle.
			id := s.Types.ByInd(int(obj.TypeInd)).ID()
			return s.Types.ByID(id).XferFunc()
		},
		findPendingObject: func(scriptID uint32) *server.Object {
			return s.Objs.PendingByScriptID(int(int32(scriptID)))
		},
		findPendingWaypoint: s.WPs.PendingByIndTmp,
		mapWallSize: func() *[2]uint32 {
			ptr := unsafe.Pointer(C.nox_xxx_mapGetWallSize_426A70())
			return (*[2]uint32)(ptr)
		},
		setElevatorLink: func(obj *server.Object, data unsafe.Pointer, target *server.Object) {
			obj.SetElevatorLinkFor(data, target)
		},
		setTransporterLink: func(obj *server.Object, data *server.TransporterUpdateData, target *server.Object) {
			obj.SetTransporterTargetFor(data, target)
		},
		elevatorXfer:      Get_nox_xxx_XFerElevator_4F53D0(),
		elevatorShaftXfer: Get_nox_xxx_XFerElevatorShaft_4F54A0(),
		transporterXfer:   Get_nox_xxx_XFerTransporter_4F5300(),
		holeXfer:          Get_nox_xxx_XFerHole_4F51D0(),
		exitXfer:          Get_nox_xxx_XFerExit_4F4B90(),
		moverXfer:         Get_nox_xxx_XFerMover_4F5730(),
		glyphXfer:         Get_nox_xxx_XFerGlyph_4F5890(),
	}
}

//export nox_xxx_interesting_xfer_native_4D0010
func nox_xxx_interesting_xfer_native_4D0010(origin *C.uint32_t, nextExtent C.int32_t) C.int32_t {
	return C.int32_t(interestingXferNative4D0010(
		(*[2]uint32)(unsafe.Pointer(origin)),
		int32(nextExtent),
		interestingXferRuntimeDeps4D0010(),
	))
}
