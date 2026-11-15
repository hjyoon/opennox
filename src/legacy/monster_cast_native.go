package legacy

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

var monsterCastRuntimeForCall541300 = monsterCastDefaultRuntime541300

func monsterCastDefaultRuntime541300() server.MonsterCastRuntime541300 {
	return server.MonsterCastRuntime541300{
		MorphTo:   (*server.Object).MorphPlayerBotToPlayer4FAAF0,
		MorphFrom: (*server.Object).MorphPlayerBotFromPlayer4FAAC0,
		Direction: server.MonsterCastDirection533CC0,
		Cast:      Nox_xxx_castSpellByUser_4FDD20,
		Buff: func(unit *server.Object, buff int32) int32 {
			if unit.HasEnchant(server.EnchantID(buff)) {
				return 1
			}
			return 0
		},
		Voice: Nox_xxx_monsterGetSoundSet_424300,
		Audio: func(id uint32, unit *server.Object) {
			GetServer().S().Audio.EventObj(sound.ID(id), unit, 0, 0)
		},
		RandomFloat: func(min, max float32) float64 {
			return GetServer().S().RandomFloat416030(min, max)
		},
	}
}

func monsterCastCall541300(id int32, unit *server.Object, arg *server.SpellAcceptArg) byte {
	return server.MonsterCastNative541300(id, unit, arg, monsterCastRuntimeForCall541300())
}

func monsterActionCastCall5413B0(unit *server.Object, mode int32) byte {
	return server.MonsterActionCastNative5413B0(unit, mode, monsterCastRuntimeForCall541300())
}

func monsterCastRandomRecoilCall541490(unit, target *server.Object, out *types.Pointf) {
	server.MonsterCastRandomRecoilNative541490(unit, target, out, monsterCastRuntimeForCall541300())
}

func Nox_xxx_mobCast_541300(id int32, unit *server.Object, arg *server.SpellAcceptArg) byte {
	return monsterCastCall541300(id, unit, arg)
}

func Nox_xxx_mobCastRandomRecoil_541490(unit, target *server.Object, out *types.Pointf) {
	monsterCastRandomRecoilCall541490(unit, target, out)
}
