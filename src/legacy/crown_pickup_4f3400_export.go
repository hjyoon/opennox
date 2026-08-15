package legacy

/*
#include "crown_pickup_4f3400.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func crownPickupRuntime4F3400(_ *server.Server) server.CrownPickupRuntime4F3400 {
	return server.CrownPickupRuntime4F3400{
		ApplyEnchant: func(obj *server.Object, enchant server.EnchantID, duration, power uint32) {
			Nox_xxx_buffApplyTo_4FF380(obj, enchant, int(uint16(duration)), int(uint8(power)))
		},
	}
}

func crownPickupCall4F3400(
	s *server.Server,
	who, crown *server.Object,
	flag1, flag2 int32,
) uint32 {
	return s.CrownPickup4F3400(
		who,
		crown,
		flag1,
		flag2,
		crownPickupRuntime4F3400(s),
	)
}

//export sub_4F3400
func sub_4F3400(
	who, crown *C.nox_object_t,
	flag1, flag2 C.int32_t,
) C.uint32_t {
	s := GetServer().S()
	return C.uint32_t(crownPickupCall4F3400(
		s,
		asObjectS((*nox_object_t)(who)),
		asObjectS((*nox_object_t)(crown)),
		int32(flag1),
		int32(flag2),
	))
}
