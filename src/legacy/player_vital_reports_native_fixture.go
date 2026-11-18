package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"

typedef int (*team_health_fn)(int, nox_object_t*);
typedef uintptr_t (*owner_health_fn)(int);
typedef short (*throttled_health_fn)(nox_object_t*);
_Static_assert(_Generic(&nox_xxx_netSendTeam_4D8670, team_health_fn:1, default:0), "team health reporter ABI");
_Static_assert(_Generic(&nox_xxx_netSendPlrHealthToTeam_4D86E0, owner_health_fn:1, default:0), "player health reporter ABI");
_Static_assert(_Generic(&sub_528030, throttled_health_fn:1, default:0), "throttled vitals ABI");

static team_health_fn volatile team_health_indirect = nox_xxx_netSendTeam_4D8670;
static owner_health_fn volatile owner_health_indirect = nox_xxx_netSendPlrHealthToTeam_4D86E0;
static throttled_health_fn volatile throttled_health_indirect = sub_528030;

static int call_team_health(int indirect, int recipient, nox_object_t* unit) {
	return indirect ? team_health_indirect(recipient, unit) : nox_xxx_netSendTeam_4D8670(recipient, unit);
}

static uintptr_t call_owner_health(int indirect, int recipient) {
	return indirect ? owner_health_indirect(recipient) : nox_xxx_netSendPlrHealthToTeam_4D86E0(recipient);
}

static short call_vitals_throttled(int indirect, nox_object_t* unit) {
	return indirect ? throttled_health_indirect(unit) : sub_528030(unit);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func callTeamHealthNative4D8670(route int, recipient uint32, unit *server.Object) uint32 {
	return uint32(C.call_team_health(C.int(route), C.int(int32(recipient)), asObjectC(unit)))
}

func callOwnerHealthNative4D86E0(route int, recipient uint32) uintptr {
	return uintptr(C.call_owner_health(C.int(route), C.int(int32(recipient))))
}

func callVitalsThrottledNative528030(route int, unit *server.Object) uint16 {
	return uint16(C.call_vitals_throttled(C.int(route), asObjectC(unit)))
}
