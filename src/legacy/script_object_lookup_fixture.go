package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func callScriptObjectLookupC(name string) *server.Object {
	text := CString(name)
	defer StrFree(text)
	return asObjectS(C.nox_xxx_getObjectByScrName_4DA4F0(text))
}
