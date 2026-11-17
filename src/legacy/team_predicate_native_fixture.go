package legacy

/*
#include "GAME1_1.h"
#include <stddef.h>
typedef int (*team_predicate_has_fn)(nox_object_team_t*);
typedef int (*team_predicate_same_fn)(nox_object_team_t*,nox_object_team_t*);
_Static_assert(_Generic(&nox_xxx_servObjectHasTeam_419130,team_predicate_has_fn:1,default:0),"unchanged team predicate unary ABI");
_Static_assert(_Generic(&nox_xxx_servCompareTeams_419150,team_predicate_same_fn:1,default:0),"unchanged team predicate binary ABI");
static void team_predicate_layout_C(uintptr_t* v) {
 v[0]=sizeof(void*);v[1]=sizeof(nox_object_team_t);v[2]=_Alignof(nox_object_team_t);
 v[3]=offsetof(nox_object_team_t,legacy_next);v[4]=offsetof(nox_object_team_t,id);v[5]=offsetof(nox_object_team_t,reserved);
 v[6]=sizeof(((nox_object_team_t*)0)->legacy_next);v[7]=sizeof(((nox_object_team_t*)0)->id);v[8]=sizeof(((nox_object_team_t*)0)->reserved);
 v[9]=sizeof(team_predicate_has_fn);v[10]=sizeof(team_predicate_same_fn);
}
static int team_predicate_call_C(int route,int kind,nox_object_team_t* first,nox_object_team_t* second) {
 if(kind==0) {
  if(route==1){team_predicate_has_fn volatile call=nox_xxx_servObjectHasTeam_419130;return call(first);}
  return nox_xxx_servObjectHasTeam_419130(first);
 }
 if(route==1){team_predicate_same_fn volatile call=nox_xxx_servCompareTeams_419150;return call(first,second);}
 return nox_xxx_servCompareTeams_419150(first,second);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func teamPredicateNativeLayoutC() (out [11]uintptr) {
	C.team_predicate_layout_C((*C.uintptr_t)(unsafe.Pointer(&out[0])))
	return
}

func teamPredicateNativeCall(route, kind int, first, second *server.ObjectTeam) int32 {
	a := (*C.nox_object_team_t)(unsafe.Pointer(first))
	b := (*C.nox_object_team_t)(unsafe.Pointer(second))
	if route == 2 {
		if kind == 0 {
			return int32(nox_xxx_servObjectHasTeam_419130(a))
		}
		return int32(nox_xxx_servCompareTeams_419150(a, b))
	}
	return int32(C.team_predicate_call_C(C.int(route), C.int(kind), a, b))
}
