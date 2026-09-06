package opennox

import (
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
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
		CallDestroy: func(callback unsafe.Pointer, record *server.DurSpell) {
			ccall.CallVoidPtr(callback, record.C())
		},
		SetPlayerState: func(unit *server.Object, state server.PlayerState) {
			_ = nox_xxx_playerSetState_4FA020(unit, state)
		},
	})
}

func (sp *spellsDuration) process4FEEF0() {
	sp.SpellsDuration.SpellDurationProcess4FEEF0(server.SpellDurationProcessRuntime4FEEF0{
		Destroy: sp.destroyDurSpell,
		CallUpdate: func(callback unsafe.Pointer, record *server.DurSpell) int32 {
			return int32(ccall.CallIntPtr(callback, record.C()))
		},
	})
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
			CallCreate: func(callback unsafe.Pointer, record *server.DurSpell) int32 {
				return int32(ccall.CallIntPtr(callback, record.C()))
			},
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
