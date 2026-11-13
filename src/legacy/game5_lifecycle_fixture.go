package legacy

/*
#include <stdint.h>
#include "GAME5.h"
static uintptr_t game5_lifecycle_probe(int entry,nox_object_t* unit) {
 switch(entry) {
 case 0:return (uintptr_t)nox_xxx_createFnObelisk_54CA10(unit);
 case 1:nox_xxx_createFnAnim_54CA50(unit);return 0;
 case 2:return (uintptr_t)nox_xxx_createTrigger_54CA60(unit);
 case 3:return (uintptr_t)nox_xxx_createMonsterGen_54CA90(unit);
 case 4:return (uintptr_t)nox_xxx_createRewardMarker_54CAC0(unit);
 case 5:return (uint16_t)nox_xxx_monsterAutoSpells_54C0C0(unit);
 case 6:nox_xxx_createWeapon_54C710(unit);return 0;
 case 7:sub_54C950(unit);return 0;
 case 8:nox_xxx_monsterMainAIFn_547210(unit);return 0;
 case 9:return nox_xxx_diePlayer_54D2B0(unit);
 default:return (uintptr_t)-1;
 }
}
static int game5_definition_probe(int entry,void* def,char* name) {
 switch(entry){case 0:return nox_xxx_monsterLoadStrikeFn_549040(def,name);case 1:return nox_xxx_monsterLoadDieFn_5490E0(def,name);default:return nox_xxx_monsterLoadDeadFn_549180(def,name);}
}
static void* game5_definition_expected(int entry) {
 switch(entry){case 0:return (void*)nox_xxx_strikeOgre_549220;case 1:return (void*)sub_54A7D0;default:return (void*)sub_54A310;}
}
static int game5_strike_probe(int entry,nox_object_t* unit) {
 switch(entry){
 case 0:return nox_xxx_strikeOgre_549220(unit);
 case 1:return nox_xxx_strikeMonsterDefault_549380(unit);
 case 2:return nox_xxx_strikeScorpion_5495B0(unit);
 case 3:return nox_xxx_strikeVileZombie_549700(unit);
 case 4:return nox_xxx_strikeStoneGolem_5497E0(unit);
 case 5:return nox_xxx_strikeMechGolem_549960(unit);
 case 6:return nox_xxx_strikeWasp_549980(unit);
 case 7:return nox_xxx_strikeGhost_549A60(unit);
 case 8:return nox_xxx_strikeBomber_549BB0();
 case 9:return nox_xxx_strikeSpider_549BC0(unit);
 default:return nox_xxx_strikeSpittingSpider_549CA0(unit);
 }
}
static uintptr_t game5_object_identity(nox_object_t* o){return (uintptr_t)o;}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func game5LifecycleProbe(entry int, unit *server.Object) uintptr {
	return uintptr(C.game5_lifecycle_probe(C.int(entry), asObjectC(unit)))
}
func game5DefinitionProbe(entry int, def *server.MonsterDef, name unsafe.Pointer) int {
	return int(C.game5_definition_probe(C.int(entry), unsafe.Pointer(def), (*C.char)(name)))
}
func game5DefinitionExpected(entry int) unsafe.Pointer {
	return C.game5_definition_expected(C.int(entry))
}
func game5ObjectIdentity(unit *server.Object) uintptr {
	return uintptr(C.game5_object_identity(asObjectC(unit)))
}

func game5StrikeProbe(entry int, obj *server.Object) int {
	return int(C.game5_strike_probe(C.int(entry), asObjectC(obj)))
}
