package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import "unsafe"

func mapTypeQueriesC(record unsafe.Pointer, entry *Nox_map_list_item) (int32, int32) {
	return int32(C.nox_xxx_mapGetTypeMB_4CFFA0(record)),
		int32(C.sub_4CFFC0((*C.nox_map_list_item)(unsafe.Pointer(entry))))
}

func legacyNoRenderingC() int32 {
	return int32(C.sub_4D0D70())
}

func legacyBookQueriesC(index int32) (int32, int32) {
	return int32(C.nox_xxx_bookCreatureTest_4D70C0(C.int(index))), int32(C.sub_4D7100(C.int(index)))
}

func legacyQuestWarpFlagC() int32 {
	return int32(C.sub_4D75E0())
}
