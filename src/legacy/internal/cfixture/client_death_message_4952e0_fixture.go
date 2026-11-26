package cfixture

/*
#include "GAME2_3.h"
#include "world_native_storage.h"
#include "client__network__deathmsg.h"
*/
import "C"

func DeathMessageRingIndexes495210() (uint32, uint32) {
	return uint32(C.nox_world_word_storage().death_message_word_1203836), uint32(C.nox_world_word_storage().death_message_word_1203840)
}

func SetDeathMessageRingIndexes495210(next, oldest uint32) {
	C.nox_world_word_storage().death_message_word_1203836 = C.uint32_t(next)
	C.nox_world_word_storage().death_message_word_1203840 = C.uint32_t(oldest)
}
