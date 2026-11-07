package legacy

/*
#include "GAME1.h"
#include "GAME1_1.h"
#include <stdint.h>
#include <string.h>

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

static uint64_t sudden_death_gamedata_bits_40a970(uint32_t site) {
    static char threshold[] = "SuddenDeathPlayerThreshold";
    static char lifetime[] = "SuddenDeathLifeTime";
    char* key = site == 0x40a98a ? threshold : lifetime;
    sudden_death_gamedata_fn_40a970 fn = &nox_xxx_gamedataGetFloat_419D40;
    double value = fn(key);
    uint64_t bits;
    memcpy(&bits, &value, sizeof(bits));
    return bits;
}
static void* sudden_death_first_40a970(void) { return nox_common_playerInfoGetFirst_416EA0(); }
static void* sudden_death_next_40a970(void* player) { return nox_common_playerInfoGetNext_416EE0((nox_playerInfo*)player); }
static void sudden_death_unset_status_40a970(void* player, uint32_t mask) { (void)nox_xxx_playerUnsetStatus_417530((nox_playerInfo*)player, (int32_t)mask); }
*/
import "C"

import (
	"encoding/binary"
	"runtime"
)

// Observer handling precedes every typed helper conversion. The gamedata
// bridge returns the actual linked C.double bits; it cannot carry a raw80
// service value or claim the original physical x87 return ABI. Diagnostic
// raw80 packets remain in the observer's explicit input transport.
func suddenDeathNativeServiceDefault40A970(site uint32, arguments []suddenDeathNativeArg40A970) suddenDeathNativeReply40A970 {
	defer func() { runtime.KeepAlive(arguments) }()
	switch site {
	case 0x40A98A, 0x40A9A1:
		var reply suddenDeathNativeReply40A970
		reply.FloatFormat = "f64"
		binary.LittleEndian.PutUint64(reply.FloatPayload[:8], uint64(C.sudden_death_gamedata_bits_40a970(C.uint32_t(site))))
		return reply
	case 0x40A9BA:
		return suddenDeathNativeReply40A970{Pointer: C.sudden_death_first_40a970()}
	case 0x40A9D6:
		C.sudden_death_unset_status_40a970(arguments[0].Pointer, C.uint32_t(arguments[1].Word))
		return suddenDeathNativeReply40A970{}
	case 0x40A9DF:
		return suddenDeathNativeReply40A970{Pointer: C.sudden_death_next_40a970(arguments[0].Pointer)}
	case 0x40A9F2:
		C.nox_common_gameFlags_unset_40A540(C.uint(arguments[0].Word))
		return suddenDeathNativeReply40A970{}
	default:
		panic("unknown sudden death service site")
	}
}
