package opennox

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
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
	if spl.Caster16 != nil {
		snd := sp.s.Spells.DefByInd(spell.ID(spl.Spell)).GetOffSound()
		sp.s.Audio.EventObj(snd, spl.Caster16, 0, 0)
	}
	if destroy := spl.Destroy; destroy != nil {
		ccall.CallVoidPtr(destroy, spl.C())
	}
	if u := spl.Caster16; u != nil {
		if u.Class().Has(object.ClassPlayer) {
			ud := u.UpdateDataPlayer()
			if ud.Player.PlayerClass() != player.Warrior || !sp.s.Abils.IsActive(u, server.AbilityBerserk) {
				nox_xxx_playerSetState_4FA020(u, server.PlayerState13)
			}
		} else if u.Class().Has(object.ClassMonster) {
			u.MonsterCancelDurSpell(spell.ID(spl.Spell))
		}
	}
	sp.Unlink(spl)
	sp.FreeRecursive(spl)
}

func (sp *spellsDuration) spellCastByPlayer() {
	var next *server.DurSpell
	for it := sp.List; it != nil; it = next {
		next = it.Next
		if it.Flags88&0x1 != 0 {
			sp.destroyDurSpell(it)
			continue
		}
		if obj16 := it.Caster16; obj16 != nil && obj16.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			it.Caster16 = nil
		}

		if obj12 := it.Obj12; obj12 != nil && obj12.Flags().Has(object.FlagDestroyed) {
			it.Obj12 = nil
		}
		if it.Caster16 == nil && it.Flag20 == 0 {
			sp.CancelSpell(it)
			continue
		}
		if obj24 := it.Obj24; obj24 != nil && obj24.Flags().Has(object.FlagDestroyed) {
			it.Obj24 = nil
		}
		if it.Frame68 != it.Frame60 && it.Frame68 <= sp.s.Frame() || it.Update != nil && ccall.CallIntPtr(it.Update, it.C()) != 0 {
			sp.CancelSpell(it)
		}
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
