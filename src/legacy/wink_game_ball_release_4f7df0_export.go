package legacy

/*
#include "wink_game_ball_release_4f7df0.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func winkGameBallReleaseCall4F7DF0(player *server.Object) int32 {
	srv := GetServer()
	return srv.S().WinkGameBallRelease4F7DF0(
		player,
		server.WinkGameBallReleaseRuntime4F7DF0{
			ApplyForce: srv.ApplyForce,
			BallStatus: Sub_4E8290,
		},
	)
}

func winkGameBallReleaseExportCall4F7DF0(player *server.Object) int32 {
	return int32(C.nox_xxx_checkWinkFlags_4F7DF0(asObjectC(player)))
}

//export nox_xxx_checkWinkFlags_4F7DF0
func nox_xxx_checkWinkFlags_4F7DF0(player *C.nox_object_t) C.int32_t {
	return C.int32_t(winkGameBallReleaseCall4F7DF0(
		asObjectS((*nox_object_t)(player)),
	))
}
