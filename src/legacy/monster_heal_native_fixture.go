package legacy

/*
#include "GAME4_3.h"

typedef int (*heal_fn)(nox_object_t*);
typedef void (*heal_candidate_fn)(nox_object_t*, nox_object_t*);
_Static_assert(_Generic(&nox_xxx_mobHealSomeone_5411A0, heal_fn:1, default:0), "heal ABI");
_Static_assert(_Generic(&nox_xxx_mobMayHealThis_5412A0, heal_candidate_fn:1, default:0), "heal candidate ABI");

static heal_fn volatile heal_indirect = nox_xxx_mobHealSomeone_5411A0;
static heal_candidate_fn volatile heal_candidate_indirect = nox_xxx_mobMayHealThis_5412A0;

static int call_heal_native(int indirect, nox_object_t* caster) {
	return indirect ? heal_indirect(caster) : nox_xxx_mobHealSomeone_5411A0(caster);
}

static void call_heal_candidate_native(int indirect, nox_object_t* candidate, nox_object_t* caster) {
	if (indirect) {
		heal_candidate_indirect(candidate, caster);
	} else {
		nox_xxx_mobMayHealThis_5412A0(candidate, caster);
	}
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func callMonsterHealNative5411A0(route int, caster *server.Object) int {
	return int(C.call_heal_native(C.int(route), asObjectC(caster)))
}

func callMonsterHealCandidateNative5412A0(route int, candidate, caster *server.Object) {
	C.call_heal_candidate_native(C.int(route), asObjectC(candidate), asObjectC(caster))
}
