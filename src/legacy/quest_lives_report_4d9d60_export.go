package legacy

/*
#include "GAME3_2.h"
_Static_assert(sizeof(int) == 4, "Quest life report has a signed DWORD ABI");
_Static_assert(_Generic(&sub_4D9D60, int (*)(int, nox_object_t*): 1, default: 0),
	"Quest life report must retain its native unit pointer");
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_server_questLivesReport_native_4D9D60
func nox_server_questLivesReport_native_4D9D60(recipient C.int, unit *nox_object_t) C.int {
	return C.int(GetServer().S().QuestLivesReport4D9D60(int32(recipient), asObjectS(unit)))
}

func questLivesReportCEntry4D9D60(recipient int32, unit *server.Object) int32 {
	return int32(C.sub_4D9D60(C.int(recipient), asObjectC(unit)))
}
