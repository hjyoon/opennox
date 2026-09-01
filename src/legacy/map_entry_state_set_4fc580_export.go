package legacy

/*
#include <stdint.h>

#include "map_entry_state_set_4fc580.h"
*/
import "C"

func mapEntryStateSetExportCall4FC580(value int32) int32 {
	return int32(C.sub_4FC580(C.int32_t(value)))
}

//export sub_4FC580
func sub_4FC580(value C.int32_t) C.int32_t {
	return C.int32_t(GetServer().S().SetMapEntryState4FC580(int32(value)))
}
