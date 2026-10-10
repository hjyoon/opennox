package legacy

/*
#include "player_handle_kotr_death_54dc40.h"
#include "GAME3_3.h"
*/
import "C"

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func playerHandleKotrDeathRuntime54DC40(srv Server) server.PlayerHandleKotrDeathRuntime54DC40 {
	s := srv.S()
	return server.PlayerHandleKotrDeathRuntime54DC40{
		PlayerUpdateScoreRuntime54D980: playerUpdateScoreRuntime54D980(srv),
		IsCrown: func(unit *server.Object) bool {
			return C.nox_xxx_unitIsCrown_4E7BE0(asObjectC(unit)) != 0
		},
		BalanceFloat: s.Balance.Float,
		FloatToInt:   server.PlayerKotrPoints54DC40,
		GameplayFlag: func(mask uint32) bool {
			return noxflags.HasGamePlay(noxflags.GameplayFlag(mask))
		},
		DropCrowns: func(owner, target *server.Object) {
			s.DropOwnedCrowns4ED050(owner, target, dropOwnedCrownsRuntime4ED050(s))
		},
	}
}

func Nox_xxx_playerHandleKotrDeathNative54DC40(victim, killer *server.Object) {
	C.nox_player_handle_kotr_death_native_call_54DC40(asObjectC(victim), asObjectC(killer))
}

//export nox_server_player_handle_kotr_death_native_54DC40
func nox_server_player_handle_kotr_death_native_54DC40(victim, killer *nox_object_t) {
	server.PlayerHandleKotrDeath54DC40(asObjectS(victim), asObjectS(killer), playerHandleKotrDeathRuntime54DC40(GetServer()))
}
