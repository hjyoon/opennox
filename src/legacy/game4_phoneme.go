package legacy

/*
#include "GAME4.h"
*/
import "C"

import (
	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

type game4PhonemeRuntime struct {
	server bool
	class  func(uint32, bool) object.Class
	player func(uint32) *server.Player
}

var game4PhonemeServices = func() game4PhonemeRuntime {
	return game4PhonemeRuntime{
		server: noxflags.HasGame(1),
		class: func(code uint32, isServer bool) object.Class {
			if isServer {
				if o := GetServer().S().ObjectFromNetCode4ECCB0(code); o != nil {
					return o.ObjClass
				}
			} else if o := GetClient().Cli().Objs.ByNetCodeDynamic(int(code)); o != nil {
				return o.ObjClass
			}
			return 0
		},
		player: func(code uint32) *server.Player { return asPlayerS(nox_common_playerInfoGetByID_417040(int32(code))) },
	}
}

//export nox_xxx_spellGetPhoneme_4FE1C0
func nox_xxx_spellGetPhoneme_4FE1C0(code C.uint32_t, phoneme C.int8_t) C.int32_t {
	r := game4PhonemeServices()
	female := false
	if r.class(uint32(code), r.server)&object.ClassPlayer != 0 {
		p := r.player(uint32(code))
		if p == nil {
			return 0
		}
		female = p.Info().IsFemale()
	}
	var result int32
	switch int8(phoneme) {
	case 0:
		result = 193
	case 1:
		result = 186
	case 2:
		result = 187
	case 3:
		result = 192
	case 5:
		result = 188
	case 6:
		result = 191
	case 7:
		result = 190
	case 8:
		result = 189
	default:
		return 0
	}
	if female {
		result += 8
	}
	return C.int32_t(result)
}
