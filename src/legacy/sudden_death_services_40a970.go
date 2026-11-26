package legacy

/*
#include "GAME1.h"
#include "GAME1_1.h"
#include <stdint.h>

_Static_assert(sizeof(int)==4 && sizeof(float)==4 && sizeof(double)==8, "sudden death scalar ABI widths");
typedef double (*sudden_death_gamedata_fn_40a970)(char*);
typedef nox_playerInfo* (*sudden_death_first_fn_40a970)(void);
typedef nox_playerInfo* (*sudden_death_next_fn_40a970)(nox_playerInfo*);
typedef char (*sudden_death_unset_status_fn_40a970)(nox_playerInfo*, int);
typedef void (*sudden_death_unset_flags_fn_40a970)(unsigned int);
_Static_assert(_Generic(&nox_xxx_gamedataGetFloat_419D40,sudden_death_gamedata_fn_40a970:1,default:0), "gamedata double(char*) ABI");
_Static_assert(_Generic(&nox_common_playerInfoGetFirst_416EA0,sudden_death_first_fn_40a970:1,default:0), "first native player pointer ABI");
_Static_assert(_Generic(&nox_common_playerInfoGetNext_416EE0,sudden_death_next_fn_40a970:1,default:0), "next native player pointer ABI");
_Static_assert(_Generic(&nox_xxx_playerUnsetStatus_417530,sudden_death_unset_status_fn_40a970:1,default:0), "status char return ABI, ignored by caller");
_Static_assert(_Generic(&nox_common_gameFlags_unset_40A540,sudden_death_unset_flags_fn_40a970:1,default:0), "flags void(unsigned int) ABI");

*/
import "C"

import (
	"encoding/binary"
	"math"
	"runtime"
	"unsafe"
)

// Observer handling precedes every typed service call. The live gamedata
// value uses the double bits. The observer carries diagnostic raw80 values
// in its explicit input transport.
func suddenDeathNativeServiceDefault40A970(site uint32, arguments []suddenDeathNativeArg40A970) suddenDeathNativeReply40A970 {
	defer func() { runtime.KeepAlive(arguments) }()
	switch site {
	case 0x40A98A, 0x40A9A1:
		var reply suddenDeathNativeReply40A970
		reply.FloatFormat = "f64"
		key := internCStr(arguments[0].String)
		binary.LittleEndian.PutUint64(reply.FloatPayload[:8], math.Float64bits(float64(nox_xxx_gamedataGetFloat_419D40(key))))
		return reply
	case 0x40A9BA:
		return suddenDeathNativeReply40A970{Pointer: unsafe.Pointer(nox_common_playerInfoGetFirst_416EA0())}
	case 0x40A9D6:
		nox_xxx_playerUnsetStatus_417530((*nox_playerInfo)(arguments[0].Pointer), C.int(arguments[1].Word))
		return suddenDeathNativeReply40A970{}
	case 0x40A9DF:
		return suddenDeathNativeReply40A970{Pointer: unsafe.Pointer(nox_common_playerInfoGetNext_416EE0((*nox_playerInfo)(arguments[0].Pointer)))}
	case 0x40A9F2:
		nox_common_gameFlags_unset_40A540(C.uint(arguments[0].Word))
		return suddenDeathNativeReply40A970{}
	default:
		panic("unknown sudden death service site")
	}
}
