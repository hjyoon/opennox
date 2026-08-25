package legacy

/*
#include "GAME1.h"
#include "GAME4_1.h"
extern void* dword_5d4594_251560;
extern uint32_t dword_5d4594_1599656;
*/
import "C"
import (
	"image"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

var (
	Sub_526CA0                               func(a1 string) int
	Nox_xxx_mapSetWallInGlobalDir0pr1_5004D0 func()
	Nox_xxx_map_5004F0                       func()
	Sub_4FF990                               func(a1 uint32)
	Sub_5000B0                               func(a1 *server.Object) int
)

const wallDefNativeSize = 12332 + 2881*(cgoABIPointerSize-4)

var _ = [1]struct{}{}[wallDefNativeSize-unsafe.Sizeof(server.WallDef{})]

func asWallP(p unsafe.Pointer) *server.Wall {
	return (*server.Wall)(p)
}

//export nox_server_getWallAtGrid_410580
func nox_server_getWallAtGrid_410580(x_cgo, y_cgo int32) unsafe.Pointer {
	x := int(x_cgo)
	y := int(y_cgo)
	return GetServer().S().Walls.GetWallAtGrid(image.Pt(x, y)).C()
}

//export nox_xxx_wall_4105E0
func nox_xxx_wall_4105E0(x_cgo, y_cgo int32) unsafe.Pointer {
	x := int(x_cgo)
	y := int(y_cgo)
	return GetServer().S().Walls.GetWallAtGrid2(image.Pt(x, y)).C()
}

//export nox_xxx_wallCreateAt_410250
func nox_xxx_wallCreateAt_410250(x_cgo, y_cgo int32) unsafe.Pointer {
	x := int(x_cgo)
	y := int(y_cgo)
	return GetServer().S().Walls.CreateAtGrid(image.Pt(x, y)).C()
}

//export nox_server_wallAttachDoor
func nox_server_wallAttachDoor(wallPtr unsafe.Pointer, obj *nox_object_t) unsafe.Pointer {
	if wallPtr == nil {
		return nil
	}
	asWallP(wallPtr).AttachDoor(asObjectS(obj))
	return wallPtr
}

//export nox_client_wallAttachDoor
func nox_client_wallAttachDoor(wallPtr, drawable unsafe.Pointer, tile C.uchar) unsafe.Pointer {
	if wallPtr == nil {
		return nil
	}
	asWallP(wallPtr).AttachClientDoor(drawable, byte(tile))
	return wallPtr
}

//export nox_xxx_mapDelWallAtPt_410430
func nox_xxx_mapDelWallAtPt_410430(x_cgo, y_cgo int32) {
	x := int(x_cgo)
	y := int(y_cgo)
	GetServer().S().Walls.DeleteAtGrid(image.Pt(x, y))
}

//export sub_4106A0
func sub_4106A0(y_cgo int32) unsafe.Pointer {
	y := int(y_cgo)
	return GetServer().S().Walls.IndexByY(y).C()
}

//export nox_server_wallNextByY_4106B0
func nox_server_wallNextByY_4106B0(p unsafe.Pointer) unsafe.Pointer {
	if p == nil {
		return nil
	}
	return asWallP(p).NextByY24.C()
}

//export nox_xxx_wallForeachFn_410640
func nox_xxx_wallForeachFn_410640(cfnc unsafe.Pointer, data unsafe.Pointer) {
	GetServer().S().Walls.EachWallXxx(func(it *server.Wall) bool {
		ccall.CallVoidPtr2(cfnc, it.C(), data)
		return true
	})
}

//export sub_57B500
func sub_57B500(x_cgo, y_cgo int32, flags C.char) C.char {
	x := int(x_cgo)
	y := int(y_cgo)
	return C.char(GetServer().S().Sub_57B500(image.Pt(x, y), byte(int8(flags))))
}

//export sub_4D72C0
func sub_4D72C0() C.int {
	return C.int(bool2int(GetServer().S().Doors.Sub_4D72C0()))
}

//export sub_4D72B0
func sub_4D72B0(v C.int) {
	GetServer().S().Doors.Sub_4D72B0(v != 0)
}

//export nox_xxx_wallFlags
func nox_xxx_wallFlags(ind_cgo int32) uint32 {
	ind := int(ind_cgo)
	return GetServer().S().Walls.DefByInd(ind).Flags32
}

//export nox_xxx_getWallSprite_46A3B0
func nox_xxx_getWallSprite_46A3B0(ind_cgo int32, a2_cgo int32, a3_cgo int32, a4_cgo int32) unsafe.Pointer {
	ind := int(ind_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	a4 := int(a4_cgo)
	return GetServer().S().Walls.DefByInd(ind).Sprite(a2, a3, a4)
}

//export nox_xxx_getWallDrawOffset_46A3F0
func nox_xxx_getWallDrawOffset_46A3F0(ind_cgo int32, a2_cgo int32, a3_cgo int32, a4_cgo int32, px, py *C.int) {
	ind := int(ind_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	a4 := int(a4_cgo)
	v := GetServer().S().Walls.DefByInd(ind).DrawOffset(a2, a3, a4)
	*px = C.int(v.X)
	*py = C.int(v.Y)
}

//export nox_xxx_mapWallMaxVariation_410DD0
func nox_xxx_mapWallMaxVariation_410DD0(ind_cgo int32, a2_cgo int32, a3_cgo int32) byte {
	ind := int(ind_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	return GetServer().S().Walls.DefByInd(ind).Variations(a2, a3)
}

//export nox_xxx_map_410E00
func nox_xxx_map_410E00(ind_cgo int32) byte {
	ind := int(ind_cgo)
	return GetServer().S().Walls.DefByInd(ind).Field749
}

//export nox_xxx_mapWallGetHpByTile_410E20
func nox_xxx_mapWallGetHpByTile_410E20(ind_cgo int32) byte {
	ind := int(ind_cgo)
	return GetServer().S().Walls.DefByInd(ind).Health41
}

//export nox_xxx_wallFindOpenSound_410EE0
func nox_xxx_wallFindOpenSound_410EE0(ind_cgo int32) *C.char {
	ind := int(ind_cgo)
	return internCStr(GetServer().S().Walls.DefByInd(ind).OpenSound())
}

//export nox_xxx_wallFindCloseSound_410F20
func nox_xxx_wallFindCloseSound_410F20(ind_cgo int32) *C.char {
	ind := int(ind_cgo)
	return internCStr(GetServer().S().Walls.DefByInd(ind).CloseSound())
}

//export nox_xxx_wallTileByName_410D60
func nox_xxx_wallTileByName_410D60(name *C.char) int32 {
	return int32(GetServer().S().Walls.DefIndByName(GoString(name)))
}

//export sub_526CA0
func sub_526CA0(a1 *C.char) int32 {
	return int32(Sub_526CA0(GoString(a1)))
}

//export nox_xxx_mapSetWallInGlobalDir0pr1_5004D0
func nox_xxx_mapSetWallInGlobalDir0pr1_5004D0() {
	Nox_xxx_mapSetWallInGlobalDir0pr1_5004D0()
}

//export nox_xxx_map_5004F0
func nox_xxx_map_5004F0() {
	Nox_xxx_map_5004F0()
}

//export sub_4FF990
func sub_4FF990(a1 C.uint) {
	Sub_4FF990(uint32(a1))
}

//export sub_5000B0
func sub_5000B0(a1 *nox_object_t) int32 {
	return int32(Sub_5000B0(asObjectS(a1)))
}

//export nox_xxx_mapDamageToWalls_534FC0
func nox_xxx_mapDamageToWalls_534FC0(a1 *C.int4, a2 unsafe.Pointer, a3 C.float, a4_cgo, a5_cgo int32, a6 unsafe.Pointer) C.bool {
	a4 := int(a4_cgo)
	a5 := int(a5_cgo)
	rect := image.Rect(int(a1.field_0), int(a1.field_4), int(a1.field_8), int(a1.field_C))
	return C.bool(GetServer().Nox_xxx_mapDamageToWalls_534FC0(rect, *(*types.Pointf)(a2), float32(a3), a4, object.DamageType(a5), AsObjectP(a6)))
}

//export nox_xxx_damageToMap_534BC0
func nox_xxx_damageToMap_534BC0(a1_cgo, a2_cgo, a3_cgo, a4_cgo int32, a5 *nox_object_t) int32 {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	a4 := int(a4_cgo)
	return int32(GetServer().Nox_xxx_damageToMap_534BC0(a1, a2, a3, object.DamageType(a4), asObjectS(a5)))
}

//export nox_xxx_wallBreackableListAdd_410840
func nox_xxx_wallBreackableListAdd_410840(a1 unsafe.Pointer) {
	GetServer().S().Walls.AddBreakable(asWallP(a1))
}

//export nox_xxx_wall_4DF1E0
func nox_xxx_wall_4DF1E0(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetServer().Nox_xxx_wall_4DF1E0(a1)
}

func Sub_5071C0() bool {
	return C.dword_5d4594_1599656 != 0
}

func Nox_xxx_math_509ED0(pos types.Pointf) int {
	return int(server.DirFromVec(pos))
}

func Nox_xxx_math_509EA0(a1 int) int {
	return int(C.nox_xxx_math_509EA0(C.int(a1)))
}
