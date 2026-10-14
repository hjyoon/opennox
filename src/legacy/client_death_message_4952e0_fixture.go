package legacy

/*
#include "GAME2_3.h"
#include "client__network__deathmsg.h"
extern uint32_t dword_5d4594_1203836;
extern uint32_t dword_5d4594_1203840;
*/
import "C"

func deathMessageRingIndexes495210() (uint32, uint32) {
	return uint32(C.dword_5d4594_1203836), uint32(C.dword_5d4594_1203840)
}

func setDeathMessageRingIndexes495210(next, oldest uint32) {
	C.dword_5d4594_1203836 = C.uint32_t(next)
	C.dword_5d4594_1203840 = C.uint32_t(oldest)
}
