package opennox

import (
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type spellsDuration struct {
	s *Server
	*server.SpellsDuration
}

func (sp *spellsDuration) Init(s *Server) {
	sp.s = s
	sp.SpellsDuration = &s.Server.Spells.Dur
}

func (sp *spellsDuration) Free() {
}

func (sp *spellsDuration) destroyDurSpell(spl *server.DurSpell) {
	sp.SpellsDuration.SpellDurationDestroy4FEDA0(spl, server.SpellDurationDestroyRuntime4FEDA0{
		CallDestroy: sp.callDestroy4FEDA0,
		SetPlayerState: func(unit *server.Object, state server.PlayerState) {
			_ = nox_xxx_playerSetState_4FA020(unit, state)
		},
	})
}

func (sp *spellsDuration) callDestroy4FEDA0(callback unsafe.Pointer, record *server.DurSpell) {
	if callback == legacy.Get_sub_530270() {
		sp.s.S().SpellTagDestroy530270(record)
		return
	}
	if callback == legacy.Get_sub_531560() {
		sp.s.S().SpellOvalShieldDestroy531560(record, ovalShieldRuntime531490())
		return
	}
	ccall.CallVoidPtr(callback, record.C())
}

func (sp *spellsDuration) process4FEEF0() {
	sp.SpellsDuration.SpellDurationProcess4FEEF0(server.SpellDurationProcessRuntime4FEEF0{
		Destroy:    sp.destroyDurSpell,
		CallUpdate: sp.callUpdate4FEEF0,
	})
}

func (sp *spellsDuration) callUpdate4FEEF0(callback unsafe.Pointer, record *server.DurSpell) int32 {
	if callback == legacy.Get_sub_530250() {
		return server.SpellTagUpdate530250(record)
	}
	if callback == legacy.Get_sub_5314F0() {
		return sp.s.S().SpellOvalShieldUpdate5314F0(record)
	}
	return int32(ccall.CallIntPtr(callback, record.C()))
}

func (sp *spellsDuration) callCreate4FEBA0(callback unsafe.Pointer, record *server.DurSpell) int32 {
	if callback == legacy.Get_nox_xxx_spellTagCreature_530160() {
		return sp.s.S().SpellTagCreate530160(record)
	}
	if callback == legacy.Get_sub_531490() {
		return sp.s.S().SpellOvalShieldCreate531490(record, ovalShieldRuntime531490())
	}
	return int32(ccall.CallIntPtr(callback, record.C()))
}

func ovalShieldRuntime531490() server.SpellOvalShieldRuntime531490 {
	return server.SpellOvalShieldRuntime531490{
		ApplyBuff: func(target *server.Object, buff int32, duration int16, power int8) {
			legacy.Nox_xxx_buffApplyTo_4FF380(target, server.EnchantID(buff), int(duration), int(power))
		},
		BuffOff: func(target *server.Object, buff int32) {
			legacy.Nox_xxx_spellBuffOff_4FF5B0(target, server.EnchantID(buff))
		},
	}
}

func (sp *spellsDuration) New(spellID spell.ID, u1, u2, u3 *server.Object, sa *server.SpellAcceptArg, lvl int, create, update, destroy unsafe.Pointer, dt uint32) bool {
	return sp.SpellsDuration.SpellDurationCreate4FEBA0(
		int32(spellID),
		u1,
		u2,
		u3,
		sa,
		int32(lvl),
		create,
		update,
		destroy,
		dt,
		server.SpellDurationCreateRuntime4FEBA0{
			DestroySpell: sp.destroyDurSpell,
			CallCreate:   sp.callCreate4FEBA0,
			AudioEvent: func(id sound.ID, object *server.Object, kind int, code uint32) {
				sp.s.Audio.EventObj(id, object, kind, code)
			},
		},
	) != 0
}

// SpellDurationCreate4FEBA0 is the legacy C ABI bridge for GAME.EXE
// 004FEBA0. Pointer arguments remain native-width, while duration is converted
// from its signed C dword by bit pattern before the original wrapping frame
// addition is performed by the server model.
func (s *Server) SpellDurationCreate4FEBA0(
	spellID int32,
	second, third, fourth *server.Object,
	arg *server.SpellAcceptArg,
	level int32,
	create, update, destroy unsafe.Pointer,
	duration int32,
) int32 {
	return spellAcceptBool4FD400(s.spells.duration.New(
		spell.ID(spellID),
		second,
		third,
		fourth,
		arg,
		int(level),
		create,
		update,
		destroy,
		uint32(duration),
	))
}
