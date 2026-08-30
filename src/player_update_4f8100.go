package opennox

import "github.com/opennox/opennox/v1/common/sound"

type playerUpdateHurtHooks4F8100[P any] struct {
	loadPlayer     func() P
	isFemale       func(P) bool
	loadDamageType func() uint32
	audio          func(sound.ID)
}

// playerUpdateHurt4F8100 preserves 004F8292..004F832B. Player is loaded from
// the cached update record only after NeedSync has returned at the call site.
// The original then reads the gender byte before it reads Object.Field131.
func playerUpdateHurt4F8100[P any](damage int32, h playerUpdateHurtHooks4F8100[P]) {
	if damage <= 0 {
		return
	}
	player := h.loadPlayer()
	female := h.isFemale(player)
	damageType := h.loadDamageType()
	h.audio(playerUpdateHurtSound4F8100(damage, damageType, female))
}

func playerUpdateHurtSound4F8100(damage int32, damageType uint32, female bool) sound.ID {
	if female {
		if damageType == 5 {
			return sound.SoundHumanFemaleHurtPoison
		}
		if damage > 450 {
			return sound.SoundHumanFemaleHurtHeavy
		}
		if damage > 70 {
			return sound.SoundHumanFemaleHurtMedium
		}
		return sound.SoundHumanFemaleHurtLight
	}
	if damageType == 5 {
		return sound.SoundHumanMaleHurtPoison
	}
	if damage > 450 {
		return sound.SoundHumanMaleHurtHeavy
	}
	if damage > 70 {
		return sound.SoundHumanMaleHurtMedium
	}
	return sound.SoundHumanMaleHurtLight
}

// playerUpdateHarpoonHooks4F8100 exposes the observable loads and calls in the
// Harpoon tail of GAME.EXE 004F8100. The update-data pointer itself remains
// cached by the caller, while its target slot is deliberately reloaded around
// callbacks that may mutate it.
type playerUpdateHarpoonHooks4F8100[T comparable] struct {
	loadTarget  func() T
	loadForce   func() float64
	destroyed   func(T) bool
	breakOwner  func()
	attribution func(T)
	applyForce  func(T, float64)
}

// playerUpdateHarpoon4F8100 preserves 004F83B1..004F840C. In particular, the
// force lookup occurs before the destroyed-bit load, and the target is loaded
// again both after that lookup and after attribution. There are intentionally
// no nil checks on either reloaded target: the original immediately
// dereferences or passes those values after the corresponding callback.
func playerUpdateHarpoon4F8100[T comparable](h playerUpdateHarpoonHooks4F8100[T]) {
	target := h.loadTarget()
	var zero T
	if target == zero {
		return
	}

	force := h.loadForce()
	target = h.loadTarget()
	if h.destroyed(target) {
		h.breakOwner()
		return
	}

	h.attribution(target)
	target = h.loadTarget()
	// GAME.EXE spills HarpoonForce through m32real before negating and passing
	// it to ObjectApplyForce, so retain that float32 rounding on native hosts.
	h.applyForce(target, -float64(float32(force)))
}
