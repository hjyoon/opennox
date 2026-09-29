package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
)

const pickupOblivionPlayerClass53A9C0 = uint32(object.ClassPlayer)

var pickupOblivionRules53A9C0 = [...]struct {
	mask    uint32
	message string
	sound   uint32
}{
	{uint32(object.WeaponStaffOblivionHalberd), "weapon.c:PickupHalberdOblivion", uint32(sound.SoundStaffOblivionAchieve1)},
	{uint32(object.WeaponStaffOblivionHeart), "weapon.c:PickupHeartOblivion", uint32(sound.SoundStaffOblivionAchieve2)},
	{uint32(object.WeaponStaffOblivionWierdling), "weapon.c:PickupWierdlingOblivion", uint32(sound.SoundStaffOblivionAchieve3)},
	{uint32(object.WeaponStaffOblivionOrb), "weapon.c:PickupOrbOblivion", uint32(sound.SoundStaffOblivionAchieve4)},
}

// pickupOblivionHooks53A9C0 exposes every field read and call made by
// GAME.EXE 0053A9C0 while keeping object handles native-width.
type pickupOblivionHooks53A9C0[O any] struct {
	weaponPickup     func(O, O, int32, int32) int32
	loadOwnerClass   func(O) uint32
	playerState      func(O) int32
	loadItemSubclass func(O) uint32
	priorityMessage  func(O, string, uint8)
	audio            func(uint32, O, int32, uint32)
	pauseFX          func(O, int32)
	tryEquip         func(O, O)
}

// pickupOblivion53A9C0 preserves the short-circuit order and exact int32
// result of GAME.EXE 0053A9C0. The original has no nil guards.
func pickupOblivion53A9C0[O any](
	owner, item O,
	arg3, arg4 int32,
	hooks pickupOblivionHooks53A9C0[O],
) int32 {
	result := hooks.weaponPickup(owner, item, arg3, arg4)
	if result != 1 {
		return result
	}
	if hooks.loadOwnerClass(owner)&pickupOblivionPlayerClass53A9C0 == 0 {
		return result
	}
	if hooks.playerState(owner) != 0 {
		return result
	}

	for _, rule := range pickupOblivionRules53A9C0 {
		if hooks.loadItemSubclass(item)&rule.mask == 0 {
			continue
		}
		hooks.priorityMessage(owner, rule.message, 0)
		hooks.audio(rule.sound, owner, 0, 0)
		break
	}
	hooks.pauseFX(owner, 1)
	hooks.tryEquip(owner, item)
	return result
}
