package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type playerInvokeAbilityHooks4FBAF0[O any] struct {
	loadFlags    func(O) object.Flags
	berserk      func(O)
	warcry       func(O)
	harpoon      func(O)
	loadDuration func(server.Ability) int32
	treadLightly func(O, int32)
	infravis     func(O, int32)
}

// playerInvokeAbility4FBAF0 preserves GAME.EXE 004FBAF0. The original reads
// the unit flags before inspecting the signed 32-bit ability ID and has no nil
// guard before that read. Dead or destroyed units return immediately. Only the
// exact IDs 1..5 dispatch, and only Tread Lightly and Infravision read their
// signed 32-bit duration before invoking the matching implementation.
func playerInvokeAbility4FBAF0[O any](
	unit O,
	ability server.Ability,
	hooks playerInvokeAbilityHooks4FBAF0[O],
) {
	if hooks.loadFlags(unit).HasAny(object.FlagDestroyed | object.FlagDead) {
		return
	}
	switch ability {
	case server.AbilityBerserk:
		hooks.berserk(unit)
	case server.AbilityWarcry:
		hooks.warcry(unit)
	case server.AbilityHarpoon:
		hooks.harpoon(unit)
	case server.AbilityTreadLightly:
		duration := hooks.loadDuration(ability)
		hooks.treadLightly(unit, duration)
	case server.AbilityInfravis:
		duration := hooks.loadDuration(ability)
		hooks.infravis(unit, duration)
	}
}
