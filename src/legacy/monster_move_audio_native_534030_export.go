package legacy

/*
#include "monster_move_audio_native_534030.h"
_Static_assert(_Generic(&nox_xxx_monsterMoveAudio_534030, void (*)(nox_object_t*): 1, default: 0), "native MoveAudio void update ABI");
*/
import "C"
import "unsafe"

//export nox_xxx_monsterMoveAudio_534030
func nox_xxx_monsterMoveAudio_534030(unit *C.nox_object_t) {
	GetServer().S().MonsterMoveAudio534030(asObjectS((*nox_object_t)(unit)))
}

func monsterMoveAudioPublicC534030() unsafe.Pointer { return C.nox_xxx_monsterMoveAudio_534030 }
