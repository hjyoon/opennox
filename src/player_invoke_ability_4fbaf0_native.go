package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type playerInvokeAbilityNativeDeps4FBAF0 struct {
	berserk      func(*server.Object)
	warcry       func(*server.Object)
	harpoon      func(*server.Object)
	loadDuration func(server.Ability) int
	treadLightly func(*server.Object, int)
	infravis     func(*server.Object, int)
}

func playerInvokeAbilityNative4FBAF0(
	unit *server.Object,
	ability server.Ability,
	deps playerInvokeAbilityNativeDeps4FBAF0,
) {
	playerInvokeAbility4FBAF0(unit, ability, playerInvokeAbilityHooks4FBAF0[*server.Object]{
		loadFlags: func(unit *server.Object) object.Flags {
			// Deliberately read the field directly: Object.Flags is nil-safe,
			// while GAME.EXE faults on a nil unit before ability dispatch.
			return unit.ObjFlags
		},
		berserk: deps.berserk,
		warcry:  deps.warcry,
		harpoon: deps.harpoon,
		loadDuration: func(ability server.Ability) int32 {
			return int32(deps.loadDuration(ability))
		},
		treadLightly: func(unit *server.Object, duration int32) {
			deps.treadLightly(unit, int(duration))
		},
		infravis: func(unit *server.Object, duration int32) {
			deps.infravis(unit, int(duration))
		},
	})
}

// nox_xxx_playerInvokeAbility_4FBAF0 is the active native-width replacement
// for GAME.EXE 004FBAF0. Its sole original caller is already owned by Do, so
// this leaf needs no public C ABI.
//
//go:noinline
func (a *serverAbilities) nox_xxx_playerInvokeAbility_4FBAF0(unit *server.Object, ability server.Ability) {
	playerInvokeAbilityNative4FBAF0(unit, ability, playerInvokeAbilityNativeDeps4FBAF0{
		berserk:      nox_xxx_warriorBerserker_53FEB0,
		warcry:       nox_xxx_warriorWarcry_53FF40,
		harpoon:      a.harpoon.Do,
		loadDuration: a.getDuration,
		treadLightly: nox_xxx_warriorTreadLightly_5400B0,
		infravis:     nox_xxx_warriorInfravis_540110,
	})
}
