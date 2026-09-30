package legacy

/*
#include "game_ball_lifecycle_417f50.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func gameBallResetExportCall417F50(old *server.Object) int {
	return int(C.sub_417F50(asObjectC(old)))
}

//export sub_417F50
func sub_417F50(old *C.nox_object_t) C.int {
	return C.int(gameBallResetCall417F50(GetServer(), asObjectS((*nox_object_t)(old))))
}

func gameBallUpdateExportCall53DF40(ball *server.Object) {
	C.nox_xxx_updateGameBall_53DF40(asObjectC(ball))
}

//export nox_xxx_updateGameBall_53DF40
func nox_xxx_updateGameBall_53DF40(ball *C.nox_object_t) {
	gameBallUpdateCall53DF40(asObjectS((*nox_object_t)(ball)))
}

func gameBallDeathExportCall54E620(ball *server.Object) int {
	return int(C.nox_xxx_dieGameBall_54E620(asObjectC(ball)))
}

//export nox_xxx_dieGameBall_54E620
func nox_xxx_dieGameBall_54E620(ball *C.nox_object_t) C.int {
	return C.int(gameBallResetCall417F50(GetServer(), asObjectS((*nox_object_t)(ball))))
}
