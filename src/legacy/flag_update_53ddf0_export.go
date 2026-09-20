package legacy

/*
#include "flag_update_53ddf0.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func flagUpdateCall53DDF0(flag *server.Object) int32 {
	s := GetServer().S()
	return s.FlagUpdate53DDF0(flag, server.FlagUpdateRuntime53DDF0{
		AudioEvent: func(id uint32, obj *server.Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		FlagStatus: Sub_4E82C0,
		Move:       Nox_xxx_unitMove_4E7010,
		InformHome: func(flagIndex uint32) int32 {
			flagPickupInformOne4DA180(s, 8, flagIndex)
			return 0
		},
	})
}

//export nox_xxx_updateFlag_53DDF0
func nox_xxx_updateFlag_53DDF0(flag *C.nox_object_t) C.int {
	return C.int(flagUpdateCall53DDF0(asObjectS((*nox_object_t)(flag))))
}
