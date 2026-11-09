package legacy

/*
#include "GAME1.h"
#include "common__system__team.h"
#include <stddef.h>
static size_t team_entries_layout_417cf0(int f) {
    switch(f) {
    case 0:return sizeof(nox_team_t);
    case 1:return offsetof(nox_team_t,field_60);
    case 2:return offsetof(nox_team_t,field_72);
    case 3:return sizeof(nox_object_t);
    case 4:return offsetof(nox_object_t,data_update);
    case 5:return offsetof(nox_player_update_data_t,player);
    case 6:return offsetof(nox_playerInfo,netCode);
    case 7:return offsetof(nox_playerInfo,field_3680);
    default:return (size_t)-1;
    }
}
static void team_entries_packet_call_417cf0(void* p) { sub_4184D0((nox_team_t*)p); }
*/
import "C"
import "unsafe"

func teamEntriesCall417CF0() int32           { return int32(C.sub_417CF0()) }
func teamEntriesCall417DC0() int32           { return int32(C.sub_417DC0()) }
func teamEntriesCall417DE0() byte            { return byte(C.sub_417DE0()) }
func teamEntriesCall418390() int32           { return int32(C.sub_418390()) }
func teamEntriesCall4183C0() int32           { return int32(C.sub_4183C0()) }
func teamEntriesCall4184D0(p unsafe.Pointer) { C.team_entries_packet_call_417cf0(p) }
func teamEntriesLayout417CF0(field int) uintptr {
	return uintptr(C.team_entries_layout_417cf0(C.int(field)))
}
