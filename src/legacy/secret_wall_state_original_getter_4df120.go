package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "map_temp_lists_503f40.h"
static int secret_wall_old_4df120(int (*fn)(void*), nox_map_wall_504290* wall) { return fn(wall); }
static int secret_wall_native_4df120(int (*fn)(nox_map_wall_504290*), nox_map_wall_504290* wall) { return fn(wall); }
static int secret_wall_original_public_4df120(int entry, nox_map_wall_504290* wall) {
 if (!entry) return _Generic(&sub_4DF120,int (*)(void*):secret_wall_old_4df120,int (*)(nox_map_wall_504290*):secret_wall_native_4df120)(&sub_4DF120,wall);
 return _Generic(&sub_4DF180,int (*)(void*):secret_wall_old_4df120,int (*)(nox_map_wall_504290*):secret_wall_native_4df120)(&sub_4DF180,wall);
}
static int secret_wall_original_typed_4df120(int entry) {
 if (!entry) return _Generic(&sub_4DF120,int (*)(void*):0,int (*)(nox_map_wall_504290*):1);
 return _Generic(&sub_4DF180,int (*)(void*):0,int (*)(nox_map_wall_504290*):1);
}
static uintptr_t secret_wall_original_layout_4df120(int field) {
 switch(field) {
 case 0:return offsetof(nox_map_wall_504290,field_10);
 case 1:return sizeof(nox_map_wall_504290);
 case 2:return offsetof(nox_object_t,obj_class);
 case 3:return offsetof(nox_object_t,data_update);
 case 4:return sizeof(nox_object_t);
 case 5:return offsetof(nox_player_update_data_t,player);
 case 6:return offsetof(nox_playerInfo,playerInd);
 default:return sizeof(nox_player_update_data_t);
 }
}
static nox_object_t* secret_wall_first_control_4df120(void) { return nox_xxx_getFirstPlayerUnit_4DA7C0(); }
static nox_object_t* secret_wall_next_control_4df120(nox_object_t* unit) { return nox_xxx_getNextPlayerUnit_4DA7F0(unit); }
static uint16_t secret_wall_id_control_4df120(nox_map_wall_504290* wall) { return wall->field_10; }
static unsigned int secret_wall_recipient_control_4df120(nox_object_t* unit) {
 return ((nox_player_update_data_t*)unit->data_update)->player->playerInd;
}
static void* secret_wall_unmapped_control_4df120(void) { return (void*)(uintptr_t)UINT32_C(0x41000000); }
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func secretWallOriginalC4DF120(entry int, wall *server.Wall) int {
	return int(C.secret_wall_original_public_4df120(C.int(entry), (*C.nox_map_wall_504290)(unsafe.Pointer(wall))))
}
func secretWallOriginalTyped4DF120(entry int) bool {
	return C.secret_wall_original_typed_4df120(C.int(entry)) != 0
}
func secretWallOriginalLayout4DF120() (out [8]uintptr) {
	for i := range out {
		out[i] = uintptr(C.secret_wall_original_layout_4df120(C.int(i)))
	}
	return
}
func secretWallOriginalFirst4DF120() *server.Object {
	return (*server.Object)(unsafe.Pointer(C.secret_wall_first_control_4df120()))
}
func secretWallOriginalNext4DF120(unit *server.Object) *server.Object {
	return (*server.Object)(unsafe.Pointer(C.secret_wall_next_control_4df120((*C.nox_object_t)(unsafe.Pointer(unit)))))
}
func secretWallOriginalControls4DF120(wall *server.Wall, unit *server.Object) (uint16, byte) {
	return uint16(C.secret_wall_id_control_4df120((*C.nox_map_wall_504290)(unsafe.Pointer(wall)))),
		byte(C.secret_wall_recipient_control_4df120((*C.nox_object_t)(unsafe.Pointer(unit))))
}

var secretWallNativeRuntimeOverride4DF120 *server.SecretWallPacketRuntime4DF120

func secretWallOriginalUnmapped4DF120() unsafe.Pointer {
	return C.secret_wall_unmapped_control_4df120()
}
