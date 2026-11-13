package legacy

/*
#include "defs.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func game5Strike(unit *nox_object_t, kind server.MonsterStrikeKind549220) C.int {
	return C.int(GetServer().S().MonsterStrikeSpecial549220(asObjectS(unit), kind, monsterStrikeNativeSpecialRuntime549220()))
}

//export nox_xxx_strikeOgre_549220
func nox_xxx_strikeOgre_549220(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeOgre549220)
}

//export nox_xxx_strikeMonsterDefault_549380
func nox_xxx_strikeMonsterDefault_549380(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeDefault549220)
}

//export nox_xxx_strikeScorpion_5495B0
func nox_xxx_strikeScorpion_5495B0(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeScorpion549220)
}

//export nox_xxx_strikeVileZombie_549700
func nox_xxx_strikeVileZombie_549700(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeZombie549220)
}

//export nox_xxx_strikeStoneGolem_5497E0
func nox_xxx_strikeStoneGolem_5497E0(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeStoneGolem549220)
}

//export nox_xxx_strikeMechGolem_549960
func nox_xxx_strikeMechGolem_549960(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeMechGolem549220)
}

//export nox_xxx_strikeWasp_549980
func nox_xxx_strikeWasp_549980(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeWasp549220)
}

//export nox_xxx_strikeGhost_549A60
func nox_xxx_strikeGhost_549A60(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeGhost549220)
}

//export nox_xxx_strikeBomber_549BB0
func nox_xxx_strikeBomber_549BB0() C.int { return 1 }

//export nox_xxx_strikeSpider_549BC0
func nox_xxx_strikeSpider_549BC0(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeSpider549220)
}

//export nox_xxx_strikeSpittingSpider_549CA0
func nox_xxx_strikeSpittingSpider_549CA0(unit *nox_object_t) C.int {
	return game5Strike(unit, server.MonsterStrikeSpider549220)
}
