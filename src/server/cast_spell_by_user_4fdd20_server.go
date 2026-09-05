package server

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

// CastSpellByUserRuntime4FDD20 supplies the callbacks still owned by the outer
// game runtime. Object and SpellAcceptArg pointers remain native-width here;
// spell IDs, power, flags, enchant IDs, and return codes retain their original
// fixed-width contracts.
type CastSpellByUserRuntime4FDD20 struct {
	SpellGetPower    func(spell.ID, *Object) int32
	DisableEnchant   func(*Object, EnchantID)
	CancelDuration   func(spell.ID, *Object)
	CreateProjectile func(*Object, *Object, spell.ID)
	SpellAccept      func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int32) int32
}

type castSpellByUserNativeDeps4FDD20 struct {
	spellHasFlags func(int32, uint32) int32
	runtime       CastSpellByUserRuntime4FDD20
}

func castSpellByUserNative4FDD20(
	spellID int32,
	caster *Object,
	arg *SpellAcceptArg,
	deps castSpellByUserNativeDeps4FDD20,
) int32 {
	return castSpellByUser4FDD20(castSpellByUserHooks4FDD20[*Object, *SpellAcceptArg]{
		loadCasterArg: func() *Object { return caster },
		loadSpellArg:  func() int32 { return spellID },
		spellGetPower: func(id int32, caster *Object) int32 {
			return deps.runtime.SpellGetPower(spell.ID(id), caster)
		},
		spellHasFlags: deps.spellHasFlags,
		disableEnchant: func(caster *Object, enchant int32) {
			deps.runtime.DisableEnchant(caster, EnchantID(enchant))
		},
		cancelDuration: func(id int32, caster *Object) {
			deps.runtime.CancelDuration(spell.ID(id), caster)
		},
		loadAcceptArg: func() *SpellAcceptArg { return arg },
		loadTarget: func(arg *SpellAcceptArg) *Object {
			return arg.Obj
		},
		createProjectile: func(caster, target *Object, id int32) {
			deps.runtime.CreateProjectile(caster, target, spell.ID(id))
		},
		spellAccept: func(id int32, second, third, fourth *Object, arg *SpellAcceptArg, power int32) int32 {
			return deps.runtime.SpellAccept(spell.ID(id), second, third, fourth, arg, power)
		},
	})
}

func castSpellByUserServerDeps4FDD20(s *Server, runtime CastSpellByUserRuntime4FDD20) castSpellByUserNativeDeps4FDD20 {
	return castSpellByUserNativeDeps4FDD20{
		spellHasFlags: func(spellID int32, mask uint32) int32 {
			if s.Spells.HasFlags(spell.ID(spellID), things.SpellFlags(mask)) {
				return 1
			}
			return 0
		},
		runtime: runtime,
	}
}

// CastSpellByUser4FDD20 binds GAME.EXE 004FDD20 to native-width Object and
// SpellAcceptArg pointers. The original callback order, target load, and
// signed-dword result are preserved without additional guards or coercion.
//
//go:noinline
func (s *Server) CastSpellByUser4FDD20(
	spellID spell.ID,
	caster *Object,
	arg *SpellAcceptArg,
	runtime CastSpellByUserRuntime4FDD20,
) int32 {
	return castSpellByUserNative4FDD20(
		int32(spellID), caster, arg,
		castSpellByUserServerDeps4FDD20(s, runtime),
	)
}
