package legacy

/*
#include "crown_update_53e1d0.h"
*/
import "C"

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func crownUpdateRuntime53E1D0(s *server.Server) server.CrownUpdateRuntime53E1D0 {
	return server.CrownUpdateRuntime53E1D0{
		Pickup: func(who, crown *server.Object, flag1, flag2 int32) uint32 {
			return crownPickupCall4F3400(s, who, crown, flag1, flag2)
		},
		Move: func(crown *server.Object, destination types.Pointf) {
			Nox_xxx_unitMove_4E7010(crown, destination)
		},
	}
}

//export nox_xxx_updateCrown_53E1D0
func nox_xxx_updateCrown_53E1D0(crown *C.nox_object_t) {
	s := GetServer().S()
	s.CrownUpdate53E1D0(
		asObjectS((*nox_object_t)(crown)),
		crownUpdateRuntime53E1D0(s),
	)
}

//export nox_server_crownUpdateDataSetPickupTarget_53E1D0
func nox_server_crownUpdateDataSetPickupTarget_53E1D0(
	crown, target *C.nox_object_t,
) {
	obj := asObjectS((*nox_object_t)(crown))
	update := (*server.CrownUpdateData)(obj.UpdateData)
	update.PickupTarget = asObjectS((*nox_object_t)(target))
}
