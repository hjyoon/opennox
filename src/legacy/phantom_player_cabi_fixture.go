package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func phantomPlayerCABI53B860(source *server.Object) {
	C.nox_xxx_updatePhantomPlayer_53B860(asObjectC(source))
}
