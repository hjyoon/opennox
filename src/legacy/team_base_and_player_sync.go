package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

//export nox_xxx_netReportTeamBase_4D92D0
func nox_xxx_netReportTeamBase_4D92D0(recipient C.int, raw *nox_object_t) C.int {
	cache := memmap.PtrUint32(0x5D4594, 1556320)
	teamBaseType := int32(*cache)
	obj := asObjectS(raw)
	attrs := obj.InitDataModifier()
	s := GetServer().S()
	if teamBaseType == 0 {
		teamBaseType = int32(s.Types.IndByID("TeamBase"))
		*cache = uint32(teamBaseType)
	}
	return C.int(s.SendTeamBaseModifierReport4D92D0(int(recipient), obj, attrs, teamBaseType))
}

//export sub_4DE4D0
func sub_4DE4D0(index C.char) C.int {
	GetServer().S().Sub4DE4D0(int(index))
	return 0
}
