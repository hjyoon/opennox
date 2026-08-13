package opennox

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

// spellProjectileExpireObject4E71F0 binds the independently tested contract
// to native-width server objects and spell-projectile update data.
func spellProjectileExpireObject4E71F0(
	obj *server.Object,
	search func(*server.Object, float32, *targetSearchArg4E6EA0[*server.Object]) *server.Object,
	accept func(spell.ID, *server.Object, *server.Object, *server.Object, *server.SpellAcceptArg, int) bool,
	delayedDelete func(*server.Object),
) {
	spellProjectileExpire4E71F0(obj, spellProjectileExpire4E71F0Hooks[*server.Object, *server.SpellProjectileUpdateData]{
		updateData: func(obj *server.Object) *server.SpellProjectileUpdateData {
			return obj.UpdateDataSpellProjectile()
		},
		search: search,
		level:  func(ud *server.SpellProjectileUpdateData) int32 { return int32(ud.Level16) },
		owner:  func(ud *server.SpellProjectileUpdateData) *server.Object { return ud.Field0 },
		spell:  func(ud *server.SpellProjectileUpdateData) int32 { return int32(ud.Spell12) },
		source: func(ud *server.SpellProjectileUpdateData) *server.Object { return ud.Field8 },
		accept: func(spellID int32, source, owner3, owner4, target *server.Object, level int32) int32 {
			// GAME.EXE initializes only the target word of this argument. Its
			// two position words are indeterminate and must not be consumed while
			// Obj is non-nil; Go zeroes them to avoid propagating undefined data.
			sa := server.SpellAcceptArg{Obj: target}
			if accept(spell.ID(spellID), source, owner3, owner4, &sa, int(level)) {
				return 1
			}
			return 0
		},
		delayedDelete: delayedDelete,
	})
}
