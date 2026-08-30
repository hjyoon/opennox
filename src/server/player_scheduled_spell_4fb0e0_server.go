package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

// PlayerScheduledSpellRuntime4FB0E0 supplies the services whose concrete
// implementations remain in the outer server. Object and SpellAcceptArg
// pointers retain native width throughout this boundary.
type PlayerScheduledSpellRuntime4FB0E0 struct {
	InformText func(ntype.PlayerInd, byte, int)
	AudioEvent func(sound.ID, *Object, int, uint32)
	CastSpell  func(spell.ID, *Object, *SpellAcceptArg)
}

type playerScheduledSpellNativeDeps4FB0E0 struct {
	checkSpell func(*Object, uint32, int32) int32
	informText func(ntype.PlayerInd, byte, int)
	audioEvent func(sound.ID, *Object, int, uint32)
	castSpell  func(spell.ID, *Object, *SpellAcceptArg)
}

func playerScheduledSpellNativeHooks4FB0E0(
	deps playerScheduledSpellNativeDeps4FB0E0,
) playerScheduledSpellHooks4FB0E0[*Object, *PlayerUpdateData, *Player] {
	return playerScheduledSpellHooks4FB0E0[*Object, *PlayerUpdateData, *Player]{
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadCountLow: func(update *PlayerUpdateData) uint8 {
			return uint8(update.TrapSpellsCnt)
		},
		loadSpell: func(update *PlayerUpdateData, index int) uint32 {
			return update.TrapSpells[index]
		},
		checkSpell: deps.checkSpell,
		loadPosX: func(update *PlayerUpdateData) int32 {
			return int32(update.Field55)
		},
		loadPosY: func(update *PlayerUpdateData) int32 {
			return int32(update.Field56)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerInd: func(player *Player) uint8 {
			return player.PlayerInd
		},
		informText: func(index, code uint8, value int32) {
			deps.informText(ntype.PlayerInd(index), byte(code), int(value))
		},
		audioEvent: func(id int32, unit *Object, kind, code int32) {
			deps.audioEvent(sound.ID(id), unit, int(kind), uint32(code))
		},
		castSpell: func(id uint32, unit *Object, arg playerScheduledSpellArg4FB0E0[*Object]) {
			nativeArg := SpellAcceptArg{
				Obj: arg.target,
				Pos: types.Pointf{X: arg.posX, Y: arg.posY},
			}
			deps.castSpell(spell.ID(int32(id)), unit, &nativeArg)
		},
		storeSpell: func(update *PlayerUpdateData, index int, value uint32) {
			update.TrapSpells[index] = value
		},
		storeCountLow: func(update *PlayerUpdateData, value uint8) {
			update.TrapSpellsCnt = update.TrapSpellsCnt&^0xff | uint32(value)
		},
	}
}

func playerDoScheduledSpellNative4FB0E0(
	unit, target *Object,
	deps playerScheduledSpellNativeDeps4FB0E0,
) int32 {
	return playerDoScheduledSpell4FB0E0(unit, target, playerScheduledSpellNativeHooks4FB0E0(deps))
}

func playerDoScheduledSpellQueueNative4FB1D0(
	unit, target *Object,
	deps playerScheduledSpellNativeDeps4FB0E0,
) int32 {
	return playerDoScheduledSpellQueue4FB1D0(unit, target, playerScheduledSpellNativeHooks4FB0E0(deps))
}

func (s *Server) playerScheduledSpellDeps4FB0E0(
	runtime PlayerScheduledSpellRuntime4FB0E0,
) playerScheduledSpellNativeDeps4FB0E0 {
	return playerScheduledSpellNativeDeps4FB0E0{
		checkSpell: func(unit *Object, id uint32, bypass int32) int32 {
			return s.CheckPlayerCantCastSpell4FD150(unit, spell.ID(int32(id)), int(bypass))
		},
		informText: runtime.InformText,
		audioEvent: runtime.AudioEvent,
		castSpell:  runtime.CastSpell,
	}
}

// PlayerDoScheduledSpell4FB0E0 casts and consumes the oldest queued spell.
//
//go:noinline
func (s *Server) PlayerDoScheduledSpell4FB0E0(
	unit, target *Object,
	runtime PlayerScheduledSpellRuntime4FB0E0,
) int32 {
	return playerDoScheduledSpellNative4FB0E0(unit, target, s.playerScheduledSpellDeps4FB0E0(runtime))
}

// PlayerDoScheduledSpellQueue4FB1D0 casts and consumes the newest queued
// spell while leaving the consumed slot intact, matching GAME.EXE.
//
//go:noinline
func (s *Server) PlayerDoScheduledSpellQueue4FB1D0(
	unit, target *Object,
	runtime PlayerScheduledSpellRuntime4FB0E0,
) int32 {
	return playerDoScheduledSpellQueueNative4FB1D0(unit, target, s.playerScheduledSpellDeps4FB0E0(runtime))
}
