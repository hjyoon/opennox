package server

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
)

// SpellBuffOffRuntime4FF5B0 supplies the protection-table effect that remains
// owned by the legacy layer. SetBuffFlags invokes it only for player objects.
type SpellBuffOffRuntime4FF5B0 struct {
	ResetPlayerProtection func(*Player, uint32)
}

type spellBuffOffNativeDeps4FF5B0 struct {
	loadBuffArg   func(int32) int32
	loadUnitArg   func(*Object) *Object
	loadBuffs     func(*Object) uint32
	setBuffFlags  func(*Object, uint32)
	storeDuration func(*Object, int32, uint16)
	storePower    func(*Object, int32, uint8)
	enchantSpell  func(int32) int32
	spellAudio    func(int32, int32) int32
	audio         func(int32, *Object, int32, int32)
}

func spellBuffOffNative4FF5B0(
	unit *Object,
	buff int32,
	deps spellBuffOffNativeDeps4FF5B0,
) int32 {
	return SpellBuffOff4FF5B0(SpellBuffOffHooks4FF5B0[*Object]{
		LoadBuffArg: func() int32 {
			return deps.loadBuffArg(buff)
		},
		LoadUnitArg: func() *Object {
			return deps.loadUnitArg(unit)
		},
		LoadBuffs:     deps.loadBuffs,
		SetBuffFlags:  deps.setBuffFlags,
		StoreDuration: deps.storeDuration,
		StorePower:    deps.storePower,
		EnchantSpell:  deps.enchantSpell,
		SpellAudio:    deps.spellAudio,
		Audio:         deps.audio,
	})
}

func spellBuffOffServerDeps4FF5B0(
	s *Server,
	runtime SpellBuffOffRuntime4FF5B0,
) spellBuffOffNativeDeps4FF5B0 {
	return spellBuffOffNativeDeps4FF5B0{
		loadBuffArg: func(buff int32) int32 {
			return buff
		},
		loadUnitArg: func(unit *Object) *Object {
			return unit
		},
		loadBuffs: func(unit *Object) uint32 {
			return unit.Buffs
		},
		setBuffFlags: func(unit *Object, flags uint32) {
			unit.SetBuffFlags(flags, runtime.ResetPlayerProtection)
		},
		storeDuration: func(unit *Object, buff int32, value uint16) {
			unit.BuffsDur[int(buff)] = value
		},
		storePower: func(unit *Object, buff int32, value uint8) {
			unit.BuffsPower[int(buff)] = value
		},
		enchantSpell: func(buff int32) int32 {
			return int32(EnchantID(buff).Spell())
		},
		spellAudio: func(spellID, selector int32) int32 {
			return int32(s.Spells.DefByInd(spell.ID(spellID)).GetAudio(int(selector)))
		},
		audio: func(id int32, unit *Object, kind, code int32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), uint32(code))
		},
	}
}

// SpellBuffOff4FF5B0 binds GAME.EXE 004FF5B0 to native-width Object state.
// Buff flags are cleared before the selected duration word and power byte,
// preserving SetBuffFlags synchronization and player-protection boundaries.
// Valid callers use buff slots 0 through 31; an invalid active slot reaches a
// checked Go array access after the original flag-clear prefix instead of
// reproducing an out-of-object write. A nil object deliberately faults at the
// original buff-dword load.
//
//go:noinline
func (s *Server) SpellBuffOff4FF5B0(
	unit *Object,
	buff int32,
	runtime SpellBuffOffRuntime4FF5B0,
) int32 {
	return spellBuffOffNative4FF5B0(unit, buff, spellBuffOffServerDeps4FF5B0(s, runtime))
}
