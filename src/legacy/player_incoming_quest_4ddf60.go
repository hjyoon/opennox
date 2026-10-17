package legacy

/*
#include "player_incoming_quest_4ddf60.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

// playerIncomingQuest4DDF60 owns the pending-load field at the native update
// layout. GAME.EXE writes PE32 +552 before the incoming Quest callbacks; that
// offset is inside RespawnMarkers after native pointers widen the structure.
func playerIncomingQuest4DDF60(unit *server.Object) {
	(*server.PlayerUpdateData)(unit.UpdateData).Field138 = 1
}

func playerIncomingQuestCEntry4DDF60(unit *server.Object) {
	C.nox_xxx_playerIncomingQuest_native_4DDF60(asObjectC(unit))
}

//export nox_xxx_playerIncomingQuest_native_4DDF60
func nox_xxx_playerIncomingQuest_native_4DDF60(unit *C.nox_object_t) {
	playerIncomingQuest4DDF60(asObjectS((*nox_object_t)(unit)))
}
