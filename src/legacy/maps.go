package legacy

/*
#include "defs.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "GAME5_2.h"
#include "server__script__script.h"
#include "client__gui__guicon.h"
*/
import "C"
import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/log"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/cnxz"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

var (
	gameLog                           = log.New("game")
	mapLog                            = log.New("map")
	Nox_common_checkMapFile           func(name string) error
	Nox_xxx_mapReadSection_426EA0     func(a1 unsafe.Pointer, name string) (bool, error)
	Nox_xxx_mapWriteSectionsMB_426E20 func(a1 unsafe.Pointer) int
)

//export nox_common_checkMapFile_4CFE10
func nox_common_checkMapFile_4CFE10(name *C.char) int32 {
	if err := Nox_common_checkMapFile(GoString(name)); err != nil {
		gameLog.Println("check map file:", err)
		return int32(0)
	}
	return int32(1)
}

//export nox_xxx_mapReadSection_426EA0
func nox_xxx_mapReadSection_426EA0(a1 unsafe.Pointer, cname *C.char, cerr *C.uint) int32 {
	if cerr != nil {
		*cerr = 0
	}
	if Nox_xxx_mapReadSection_426EA0 == nil {
		if cerr != nil {
			*cerr = 1
		}
		_ = cryptfile.Close()
		return 0
	}
	ok, err := Nox_xxx_mapReadSection_426EA0(a1, GoString(cname))
	if err != nil {
		mapLog.Println(err)
		if cerr != nil {
			*cerr = 1
		}
		_ = cryptfile.Close()
		return 0
	}
	return int32(bool2int(ok))
}

//export nox_xxx_mapWriteSectionsMB_426E20
func nox_xxx_mapWriteSectionsMB_426E20(a1 unsafe.Pointer) int32 {
	return int32(Nox_xxx_mapWriteSectionsMB_426E20(a1))
}

//export nox_xxx_nxzCompressFile_57BDD0
func nox_xxx_nxzCompressFile_57BDD0(a1, a2 *C.char) int32 {
	if err := cnxz.CompressFile(GoString(a1), GoString(a2)); err != nil {
		mapLog.Println(err)
		return int32(0)
	}
	return int32(1)
}

//export nox_xxx_mapReset_5028E0
func nox_xxx_mapReset_5028E0() {
	GetServer().Nox_xxx_mapReset5028E0()
}

//export nox_xxx_free_503F40
func nox_xxx_free_503F40() {
	GetServer().Nox_xxx_free503F40()
}

//export sub_51A100
func sub_51A100() {
	GetServer().S().MapSend.Sub_51A100()
}

func Nox_server_mapRWMapInfo_42A6E0(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWMapInfo_42A6E0, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}
func Nox_server_mapRWWallMap_429B20(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWWallMap_429B20, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}
func Nox_server_mapRWFloorMap_422230(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWFloorMap_422230, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}
func Nox_server_mapRWSecretWalls_4297C0(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWSecretWalls_4297C0, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}
func Nox_server_mapRWDestructableWalls_429530(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWDestructableWalls_429530, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}
func Nox_server_mapRWWaypoints_506260(cf *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if cf == nil {
		return fmt.Errorf("%s: nil crypt file", caller(0))
	}
	if !cf.ReadOnly() {
		return mapWriteWaypoints506260(cf, GetServer().S().WPs.First(), func(wp *server.Waypoint) bool {
			if a1 == nil {
				return true
			}
			pos := C.int2{
				field_0: C.int(int32(wp.PosVec.X)),
				field_4: C.int(int32(wp.PosVec.Y)),
			}
			return C.nox_xxx_wallMath_427F30(&pos, (*C.int)(a1)) != 0
		})
	}
	hooks := mapWaypointReadHooks506260{
		newWaypoint: func(ind int, pos types.Pointf) *server.Waypoint {
			if noxflags.HasGame(noxflags.GameFlag23) {
				node := C.sub_5044B0(C.int32_t(ind), C.float(pos.X), C.float(pos.Y))
				return (*server.Waypoint)(unsafe.Pointer(C.nox_map_waypoint_list_value_5044B0(node)))
			}
			return GetServer().S().WPs.Nox_xxx_waypointNewNotMap_579970(ind, pos)
		},
	}
	if a1 != nil {
		hooks.adjust = func(pos types.Pointf) types.Pointf {
			cpos := C.float2{field_0: C.float(pos.X), field_4: C.float(pos.Y)}
			C.nox_mapgen_adjust_waypoint_506260(a1, &cpos)
			return types.Ptf(float32(cpos.field_0), float32(cpos.field_4))
		}
	}
	return mapReadWaypoints506260(cf, hooks)
}

type mapWaypointReadHooks506260 struct {
	adjust      func(types.Pointf) types.Pointf
	newWaypoint func(int, types.Pointf) *server.Waypoint
}

func mapReadWaypoints506260(cf *cryptfile.CryptFile, hooks mapWaypointReadHooks506260) error {
	version, err := cf.ReadU16()
	if err != nil {
		return err
	}
	if int16(version) > 4 {
		return fmt.Errorf("unsupported waypoint section version: %d", version)
	}
	count, err := cf.ReadU32()
	if err != nil {
		return err
	}
	if int32(count) <= 0 {
		return nil
	}
	if hooks.newWaypoint == nil {
		return fmt.Errorf("waypoint allocator is not configured")
	}
	for i := uint32(0); i < count; i++ {
		index, err := cf.ReadU32()
		if err != nil {
			return err
		}
		xbits, err := cf.ReadU32()
		if err != nil {
			return err
		}
		ybits, err := cf.ReadU32()
		if err != nil {
			return err
		}
		var pos types.Pointf
		if int16(version) < 4 {
			pos = types.Ptf(float32(xbits), float32(ybits))
		} else {
			pos = types.Ptf(math.Float32frombits(xbits), math.Float32frombits(ybits))
		}
		name := ""
		if int16(version) >= 3 {
			name, err = cf.ReadString8()
			if err != nil {
				return err
			}
		}
		if hooks.adjust != nil {
			pos = hooks.adjust(pos)
		}
		wp := hooks.newWaypoint(int(index), pos)
		if wp == nil {
			return fmt.Errorf("cannot allocate waypoint %d", index)
		}
		wp.SetName(name)
		wp.Flags, err = cf.ReadU32()
		if err != nil {
			return err
		}
		if int16(version) < 4 {
			value, err := cf.ReadU32()
			if err != nil {
				return err
			}
			wp.PointsCnt = byte(value)
		} else {
			wp.PointsCnt, err = cf.ReadU8()
			if err != nil {
				return err
			}
		}
		if int(wp.PointsCnt) > len(wp.Points) {
			return fmt.Errorf("waypoint %d has too many connections: %d", index, wp.PointsCnt)
		}
		for j := 0; j < int(wp.PointsCnt); j++ {
			wp.Field348[j], err = cf.ReadU32()
			if err != nil {
				return err
			}
			if int16(version) >= 2 {
				wp.Points[j].Ind, err = cf.ReadU8()
				if err != nil {
					return err
				}
			} else {
				wp.Points[j].Ind = 2
			}
		}
	}
	return nil
}

func FreeMapgenWaypointList503F40(freePayloads bool) {
	C.nox_mapgen_free_waypoint_list_503F40(C.int32_t(bool2int(freePayloads)))
}

func FreeMapgenTileList503F40() {
	C.nox_mapgen_free_tile_list_503F40()
}

func FreeMapgenWallList503F40(freePayloads bool) {
	C.nox_mapgen_free_wall_list_503F40(C.int32_t(bool2int(freePayloads)))
}

func mapWriteWaypoints506260(cf *cryptfile.CryptFile, first *server.Waypoint, accept func(*server.Waypoint) bool) error {
	if err := cf.WriteU16(4); err != nil {
		return err
	}
	var count uint32
	for wp := first; wp != nil; wp = wp.WpNext {
		if accept(wp) {
			count++
		}
	}
	if err := cf.WriteU32(count); err != nil {
		return err
	}
	for wp := first; wp != nil; wp = wp.WpNext {
		if !accept(wp) {
			continue
		}
		if err := cf.WriteU32(wp.Index); err != nil {
			return err
		}
		if err := cf.WriteU32(math.Float32bits(wp.PosVec.X)); err != nil {
			return err
		}
		if err := cf.WriteU32(math.Float32bits(wp.PosVec.Y)); err != nil {
			return err
		}
		if err := cf.WriteString8(wp.Name()); err != nil {
			return err
		}
		if err := cf.WriteU32(wp.Flags & 1); err != nil {
			return err
		}
		if err := cf.WriteU8(wp.PointsCnt); err != nil {
			return err
		}
		for i := 0; i < int(wp.PointsCnt); i++ {
			point := &wp.Points[i]
			if err := cf.WriteU32(point.Waypoint.Index); err != nil {
				return err
			}
			if err := cf.WriteU8(point.Ind); err != nil {
				return err
			}
		}
	}
	return nil
}
func Nox_server_mapRWWindowWalls_4292C0(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWWindowWalls_4292C0, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}
func Nox_server_mapRWGroupData_505C30(cf *cryptfile.CryptFile, _ unsafe.Pointer) error {
	if cf == nil {
		return fmt.Errorf("%s: nil crypt file", caller(0))
	}
	s := GetServer().S()
	if !cf.ReadOnly() {
		return mapWriteGroups505C30(cf, s.MapGroups.GetFirstMapGroup())
	}
	hooks := mapGroupReadHooks505C30{}
	if s.CurrentMapXxx != nil {
		hooks.currentMap = s.CurrentMapXxx()
	}
	hooks.skip = memmap.Uint32(0x5D4594, 739992)&4 != 0
	if noxflags.HasGame(noxflags.GameFlag23) {
		hooks.addGroup = func(name string, index uint32, kind server.MapGroupKind) {
			s.MapGroups.Sub504600(name, index, uint8(kind))
		}
		hooks.addItem = func(index uint32, _ server.MapGroupKind, item mapGroupItemRecord505C30) {
			s.MapGroups.Sub5046A0([]uint32{item.raw0, item.raw4}, index)
		}
	} else {
		if noxflags.HasGame(noxflags.GameHost | noxflags.GameFlag22) {
			hooks.addGroup = func(name string, index uint32, kind server.MapGroupKind) {
				s.MapGroups.MapLoadAddGroup57C0C0(name, index, byte(kind))
			}
		}
		hooks.addItem = func(index uint32, _ server.MapGroupKind, item mapGroupItemRecord505C30) {
			s.MapGroups.Sub57C130([]uint32{item.raw0, item.raw4}, index)
		}
	}
	return mapReadGroups505C30(cf, hooks)
}
func Nox_server_mapRWAmbientData_429200(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWAmbientData_429200, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}
func Nox_server_mapRWPolygons_428CD0(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWPolygons_428CD0, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}
func Nox_server_mapRWObjectTOC_428B30(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
	if ccall.CallIntPtr(C.nox_server_mapRWObjectTOC_428B30, a1) == 0 {
		return fmt.Errorf("%s failed", caller(0))
	}
	return nil
}

func mapObjectListAdd5048A0(obj *server.Object) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_unitAddToList_5048A0(asObjectC(obj)))
}

func MapObjectListNodeObject5048A0(node unsafe.Pointer) *server.Object {
	return asObjectS((*nox_object_t)(C.nox_map_object_list_node_object_5048A0(
		(*C.nox_map_object_list_node_5048A0)(node),
	)))
}

func MapObjectListNodeNext5048A0(node unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nox_map_object_list_node_next_5048A0(
		(*C.nox_map_object_list_node_5048A0)(node),
	))
}

func FreeMapObjectListNode5048A0(node unsafe.Pointer) {
	C.nox_map_object_list_node_free_5048A0((*C.nox_map_object_list_node_5048A0)(node))
}

func Nox_xxx_prepareLightningEffects_4BAB30() {
	C.nox_xxx_prepareLightningEffects_4BAB30()
}

func Sub_4B64C0() {
	C.sub_4B64C0()
}

func Nox_xxx_bookSetColor_45AC40() {
	C.nox_xxx_bookSetColor_45AC40()
}

func Nox_xxx_colorInit_4C4FD0() {
	C.nox_xxx_colorInit_4C4FD0()
}

func Sub_445FF0() {
	C.sub_445FF0()
}

func Sub_470680() {
	C.sub_470680()
}

func Sub_461520() {
	C.sub_461520()
}

func Nox_xxx_tile_486060() {
	C.nox_xxx_tile_486060()
}

func Sub_461400() {
	C.sub_461400()
}

func Sub_461450() int {
	return int(C.sub_461450())
}

func Nox_xxx_cliShowHideTubes_470AA0(v int) {
	C.nox_xxx_cliShowHideTubes_470AA0(C.int(v))
}

func Sub_428170(a1, a2 unsafe.Pointer) {
	C.sub_428170(a1, (*C.int4)(a2))
}

func Nox_xxx_tileNFromPoint_411160(p types.Pointf) int {
	cp, free := alloc.New(types.Pointf{})
	defer free()
	*cp = p
	return int(C.nox_xxx_tileNFromPoint_411160((*C.float2)(unsafe.Pointer(cp))))
}

func Nox_xxx_unitSetDecayTime_511660(obj *server.Object, a2 int) {
	GetServer().S().DecaySetTime511660(obj, uint32(a2))
}

func Nox_xxx_tileFreeTileOne_4221E0(p unsafe.Pointer) {
	C.nox_xxx_tileFreeTileOne_4221E0(p)
}

func Get_nox_client_mapSpecialRWObjectData_4AC610() unsafe.Pointer {
	return C.nox_client_mapSpecialRWObjectData_4AC610
}

func Sub_4DE410(pli ntype.PlayerInd) {
	C.sub_4DE410(C.int(pli))
}
