package server

import (
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// BuffApplyRuntime4FF380 supplies the two effects that still belong to the
// legacy layer. BuffOff is called only for Invisibility by this routine. The
// protection callback is used only when SetBuffFlags updates a player.
type BuffApplyRuntime4FF380 struct {
	BuffOff               func(*Object, EnchantID) int32
	ResetPlayerProtection func(*Player, uint32)
}

type buffApplyNativeDeps4FF380 struct {
	loadHecubahTypeID      func() uint32
	storeHecubahTypeID     func(uint32)
	loadNecromancerTypeID  func() uint32
	storeNecromancerTypeID func(uint32)
	lookupTypeID           func(string) uint32
	gameFlag               func(uint32) int32
	audio                  func(int32, *Object, int32, int32)
	buffOff                func(*Object, EnchantID) int32
	resetPlayerProtection  func(*Player, uint32)
	enchantSpell           func(int32) int32
	spellAudio             func(int32, int32) int32
}

func buffApplyNative4FF380(
	unit *Object,
	buff int32,
	duration int16,
	power int8,
	deps buffApplyNativeDeps4FF380,
) {
	BuffApply4FF380(BuffApplyHooks4FF380[*Object]{
		LoadHecubahTypeID:      deps.loadHecubahTypeID,
		StoreHecubahTypeID:     deps.storeHecubahTypeID,
		LoadNecromancerTypeID:  deps.loadNecromancerTypeID,
		StoreNecromancerTypeID: deps.storeNecromancerTypeID,
		LookupTypeID:           deps.lookupTypeID,
		LoadBuffArg: func() int32 {
			return buff
		},
		LoadUnitArg: func() *Object {
			return unit
		},
		LoadDurationArg: func() int16 {
			return duration
		},
		LoadPowerArg: func() int8 {
			return power
		},
		LoadTypeIndex: func(unit *Object) uint16 {
			return unit.TypeInd
		},
		LoadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		LoadSubclass: func(unit *Object) uint32 {
			return uint32(unit.ObjSubClass)
		},
		LoadObjectFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		GameFlag: deps.gameFlag,
		Audio:    deps.audio,
		TestBuff: func(unit *Object, buff int32) int32 {
			return unit.UnitBuffTest4FF350(buff)
		},
		LoadBuffTimer: func(unit *Object, buff int32) int32 {
			return int32(unit.BuffsDur[int(buff)])
		},
		BuffOff: func(unit *Object, buff int32) int32 {
			return deps.buffOff(unit, EnchantID(buff))
		},
		StoreDuration: func(unit *Object, buff int32, value uint16) {
			unit.BuffsDur[int(buff)] = value
		},
		StorePower: func(unit *Object, buff int32, value uint8) {
			unit.BuffsPower[int(buff)] = value
		},
		LoadBuffs: func(unit *Object) uint32 {
			return unit.Buffs
		},
		SetBuffFlags: func(unit *Object, flags uint32) {
			unit.SetBuffFlags(flags, deps.resetPlayerProtection)
		},
		EnchantSpell: deps.enchantSpell,
		SpellAudio:   deps.spellAudio,
	})
}

func buffApplyServerDeps4FF380(
	s *Server,
	runtime BuffApplyRuntime4FF380,
) buffApplyNativeDeps4FF380 {
	return buffApplyNativeDeps4FF380{
		loadHecubahTypeID: s.Types.buffApplyHecubahIDCached4FF380,
		storeHecubahTypeID: func(value uint32) {
			s.Types.storeBuffApplyHecubahID4FF380(value)
		},
		loadNecromancerTypeID: s.Types.buffApplyNecromancerIDCached4FF380,
		storeNecromancerTypeID: func(value uint32) {
			s.Types.storeBuffApplyNecromancerID4FF380(value)
		},
		lookupTypeID: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		audio: func(id int32, unit *Object, kind, code int32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), uint32(code))
		},
		buffOff:               runtime.BuffOff,
		resetPlayerProtection: runtime.ResetPlayerProtection,
		enchantSpell: func(buff int32) int32 {
			return int32(EnchantID(buff).Spell())
		},
		spellAudio: func(spellID, selector int32) int32 {
			return int32(s.Spells.DefByInd(spell.ID(spellID)).GetAudio(int(selector)))
		},
	}
}

// BuffApply4FF380 binds GAME.EXE 004FF380 to native-width Object state. The
// routine accepts the original signed dword/word/byte argument widths. Valid
// callers use buff slots 0 through 31; an invalid slot reaches a checked Go
// array access instead of reproducing the original out-of-object corruption.
//
//go:noinline
func (s *Server) BuffApply4FF380(
	unit *Object,
	buff int32,
	duration int16,
	power int8,
	runtime BuffApplyRuntime4FF380,
) {
	buffApplyNative4FF380(unit, buff, duration, power, buffApplyServerDeps4FF380(s, runtime))
}
