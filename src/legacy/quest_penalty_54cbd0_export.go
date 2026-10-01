package legacy

/*
#include "GAME5.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func questPenaltyCEntry54CBD0(unit *server.Object) { C.sub_54CBD0(asObjectC(unit)) }

//export nox_server_questPenalty_native_54CBD0
func nox_server_questPenalty_native_54CBD0(unit *nox_object_t) {
	GetServer().S().QuestPenalty54CBD0(asObjectS(unit), server.QuestPenaltyRuntime54CBD0{
		SubGold: func(unit *server.Object, amount uint32) {
			Nox_xxx_playerSubGold_4FA5D0(unit, int(amount))
		},
		LoseGems: func(unit *server.Object) {
			nox_server_questLoseGems_native_54D080(asObjectC(unit))
		},
		LoseWeapon: func(unit *server.Object) {
			nox_server_questLoseWeapon_native_54CC40(asObjectC(unit))
		},
		LoseArmor: func(unit *server.Object) {
			nox_server_questLoseArmor_native_54CD30(asObjectC(unit))
		},
	})
}
