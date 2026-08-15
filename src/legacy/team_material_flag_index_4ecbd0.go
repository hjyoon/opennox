package legacy

/*
typedef struct nox_object_t nox_object_t;
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export sub_4ECBD0
func sub_4ECBD0(obj *C.nox_object_t) C.int {
	return C.int(server.TeamMaterialObjectIndex4ECBD0(
		asObjectS((*nox_object_t)(obj)),
	))
}
