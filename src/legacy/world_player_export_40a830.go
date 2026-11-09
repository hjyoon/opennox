package legacy

/*
#include "GAME1.h"
*/
import "C"
import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

//export nox_xxx_countNonEliminatedPlayersInTeam_40A830
func nox_xxx_countNonEliminatedPlayersInTeam_40A830(team *C.nox_team_t) C.int {
	return C.int(countNonEliminatedNative40A830((*server.Team)(unsafe.Pointer(team))))
}

//export nox_xxx_gamePlayIsAnyPlayers_40A8A0
func nox_xxx_gamePlayIsAnyPlayers_40A8A0() C.int { return C.int(gamePlayAnyNative40A8A0()) }

//export nox_xxx_cliCanTalkMB_4100F0
func nox_xxx_cliCanTalkMB_4100F0(text *C.short) C.int {
	return C.int(cliCanTalkNative4100F0((*uint16)(unsafe.Pointer(text))))
}

//export nox_xxx_doorAttachWall_410360
func nox_xxx_doorAttachWall_410360(object *C.nox_object_t, x, y C.int) unsafe.Pointer {
	return doorAttachWallNative410360((*server.Object)(unsafe.Pointer(object)), int32(x), int32(y))
}

//export nox_xxx_wallSecretBlock_410760
func nox_xxx_wallSecretBlock_410760(node *C.nox_secret_wall_t) *C.nox_secret_wall_t {
	return (*C.nox_secret_wall_t)(unsafe.Pointer(secretWallBlockNative410760((*server.SecretWall)(unsafe.Pointer(node)))))
}

//export nox_xxx_wallSecretGetFirstWall_410780
func nox_xxx_wallSecretGetFirstWall_410780() *C.nox_secret_wall_t {
	return (*C.nox_secret_wall_t)(unsafe.Pointer(secretWallFirstNative410780()))
}

//export nox_xxx_wallSecretNext_410790
func nox_xxx_wallSecretNext_410790(node *C.nox_secret_wall_t) *C.nox_secret_wall_t {
	return (*C.nox_secret_wall_t)(unsafe.Pointer(secretWallNextNative410790((*server.SecretWall)(unsafe.Pointer(node)))))
}

//export nox_xxx_tileAlloc_410F60_init
func nox_xxx_tileAlloc_410F60_init() C.int { return C.int(tileAllocNative410F60()) }

//export nox_xxx_tileFree_410FC0_free
func nox_xxx_tileFree_410FC0_free() { tileFreeNative410FC0() }

//export nox_xxx_tileNFromPoint_411160
func nox_xxx_tileNFromPoint_411160(point *C.float2) C.int {
	return C.int(tileFromPointNative411160((*types.Pointf)(unsafe.Pointer(point))))
}

//export nox_xxx_mapTileAllowTeleport_411A90
func nox_xxx_mapTileAllowTeleport_411A90(point *C.float2) C.int {
	return C.int(tileAllowsTeleportNative411A90((*types.Pointf)(unsafe.Pointer(point))))
}

//export nox_xxx_modifNext_4133C0
func nox_xxx_modifNext_4133C0(value *C.obj_412ae0_t) *C.obj_412ae0_t {
	return (*C.obj_412ae0_t)(unsafe.Pointer(modifierNextNative4133C0((*server.ModifierEff)(unsafe.Pointer(value)))))
}

//export nox_xxx_checkGameFlagPause_413A50
func nox_xxx_checkGameFlagPause_413A50() C.int { return C.int(checkPauseNative413A50()) }

//export nox_xxx_itemApplyDefendEffect_415C00
func nox_xxx_itemApplyDefendEffect_415C00(item *C.nox_object_t) C.double {
	return C.double(itemDefendNative415C00((*server.Object)(unsafe.Pointer(item))))
}

//export nox_xxx_cliGamedataGet_416590
func nox_xxx_cliGamedataGet_416590(index C.int) *C.char {
	return (*C.char)(cliGamedataNative416590(int32(index)))
}

//export nox_xxx_cliSetSettingsAcquired_4169D0
func nox_xxx_cliSetSettingsAcquired_4169D0(value C.int) C.int {
	return C.int(settingsAcquiredNative4169D0(int32(value)))
}

//export nox_xxx_playerForceSendLessons_416E50
func nox_xxx_playerForceSendLessons_416E50(send C.int) *C.char {
	return (*C.char)(playerForceLessonsNative416E50(int32(send)))
}

//export nox_xxx_playerByName_4170D0
func nox_xxx_playerByName_4170D0(name *C.wchar2_t) *C.nox_playerInfo {
	return (*C.nox_playerInfo)(unsafe.Pointer(playerByNameNative4170D0((*uint16)(unsafe.Pointer(name)))))
}

//export nox_xxx_netUnmarkMinimapSpec_417470
func nox_xxx_netUnmarkMinimapSpec_417470(object *C.nox_object_t, flags C.int) *C.char {
	return (*C.char)(minimapAllNative417470((*server.Object)(unsafe.Pointer(object)), uint32(flags), false))
}

//export nox_xxx_netMarkMinimapForAll_4174B0
func nox_xxx_netMarkMinimapForAll_4174B0(object *C.nox_object_t, flags C.int) *C.char {
	return (*C.char)(minimapAllNative417470((*server.Object)(unsafe.Pointer(object)), uint32(flags), true))
}

//export nox_xxx_playerUnsetStatus_417530
func nox_xxx_playerUnsetStatus_417530(player *C.nox_playerInfo, mask C.int) C.char {
	return C.char(playerUnsetNative417530((*server.Player)(unsafe.Pointer(player)), uint32(mask)))
}

//export nox_xxx_sendAllClientStatus_4175C0
func nox_xxx_sendAllClientStatus_4175C0(player *C.nox_playerInfo) *C.char {
	return (*C.char)(sendAllStatusNative4175C0((*server.Player)(unsafe.Pointer(player))))
}

//export nox_xxx_netReportPlayerStatus_417630
func nox_xxx_netReportPlayerStatus_417630(player *C.nox_playerInfo) C.int {
	return C.int(netReportStatusNative417630((*server.Player)(unsafe.Pointer(player))))
}
