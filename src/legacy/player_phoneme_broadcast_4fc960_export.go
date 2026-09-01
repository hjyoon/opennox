package legacy

/*
#include <stdint.h>

#include "player_phoneme_broadcast_4fc960.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerPhonemeBroadcastExportCall4FC960(source *server.Object, phoneme int8) int32 {
	return int32(C.sub_4FC960(asObjectC(source), C.int8_t(phoneme)))
}

func Nox_xxx_spellGetPhoneme_4FE1C0(netCode uint32, phoneme int8) int32 {
	return int32(C.nox_xxx_spellGetPhoneme_4FE1C0(
		C.uint32_t(netCode),
		C.int8_t(phoneme),
	))
}

//export sub_4FC960
func sub_4FC960(source *C.nox_object_t, phoneme C.int8_t) C.int32_t {
	return C.int32_t(GetServer().PlayerPhonemeBroadcast4FC960(
		asObjectS((*nox_object_t)(source)),
		int8(phoneme),
	))
}
