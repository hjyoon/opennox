package legacy

/*
#include "GAME5.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func monsterStrikeNativeKind549220(fnc unsafe.Pointer) (server.MonsterStrikeKind549220, bool) {
	switch fnc {
	case unsafe.Pointer(C.nox_xxx_strikeMonsterDefault_549380):
		return server.MonsterStrikeDefault549220, true
	case unsafe.Pointer(C.nox_xxx_strikeSpider_549BC0), unsafe.Pointer(C.nox_xxx_strikeSpittingSpider_549CA0):
		return server.MonsterStrikeSpider549220, true
	case unsafe.Pointer(C.nox_xxx_strikeOgre_549220):
		return server.MonsterStrikeOgre549220, true
	case unsafe.Pointer(C.nox_xxx_strikeScorpion_5495B0):
		return server.MonsterStrikeScorpion549220, true
	case unsafe.Pointer(C.nox_xxx_strikeVileZombie_549700):
		return server.MonsterStrikeZombie549220, true
	case unsafe.Pointer(C.nox_xxx_strikeStoneGolem_5497E0):
		return server.MonsterStrikeStoneGolem549220, true
	case unsafe.Pointer(C.nox_xxx_strikeMechGolem_549960):
		return server.MonsterStrikeMechGolem549220, true
	case unsafe.Pointer(C.nox_xxx_strikeWasp_549980):
		return server.MonsterStrikeWasp549220, true
	case unsafe.Pointer(C.nox_xxx_strikeGhost_549A60):
		return server.MonsterStrikeGhost549220, true
	case unsafe.Pointer(C.nox_xxx_strikeBomber_549BB0):
		return server.MonsterStrikeBomber549220, true
	}
	return 0, false
}

func monsterStrikeNativeSpecialRuntime549220() server.MonsterStrikeSpecialRuntime549220 {
	return server.MonsterStrikeSpecialRuntime549220{
		MonsterStrikeDefaultRuntime549380: server.MonsterStrikeDefaultRuntime549380{
			Damage: func(target, source, attacker *server.Object, damage int, typ object.DamageType) bool {
				return target.CallDamage(source, attacker, damage, typ)
			},
			ApplyForce: func(target *server.Object, origin types.Pointf, force float64) {
				GetServer().ApplyForce(target, origin, force)
			},
		},
		Direction: func(dir server.Dir16) types.Pointf {
			return *(*types.Pointf)(memmap.PtrOff(0x587000, uintptr(194136+int(int16(dir))*8)))
		},
		Facing:         Nox_server_testTwoPointsAndDirection_4E6E50,
		Distance:       Nox_xxx_calcDistance_4E6C00,
		ActivatePoison: Nox_xxx_activatePoison_4EE7E0,
		PriorityMessage: func(target *server.Object, id strman.ID, value byte) {
			GetServer().S().NetPriMsgToPlayer(target, id, value)
		},
		BuffApply: Nox_xxx_buffApplyTo_4FF380,
		SetAreaResult: func(golem bool, value uint32) {
			offset := uintptr(2491556)
			if golem {
				offset = 2491576
			}
			*memmap.PtrUint32(0x5D4594, offset) = value
		},
		AreaResult: func(golem bool) uint32 {
			offset := uintptr(2491556)
			if golem {
				offset = 2491576
			}
			return *memmap.PtrUint32(0x5D4594, offset)
		},
		SetGolemMode: func(mode uint32) { *memmap.PtrUint32(0x5D4594, 2491560) = mode },
	}
}
