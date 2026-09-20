package server

import (
	"image"

	"github.com/opennox/libs/types"
)

const (
	damageVampirismSound4E0B30   = 163
	damageVampirismFX4E0B30      = 162
	damageVampirismBalance4E0B30 = "VampirismCoeff"
	damageVampirismEnchant4E0B30 = EnchantID(13)
)

// damageVampirismPoint4E0B30 preserves the four independent nox_float2int
// conversions used to build the MSG_FX_VAMP packet in GAME.EXE 004E0B30.
func damageVampirismPoint4E0B30(pos types.Pointf) image.Point {
	return image.Point{
		X: int(playerCollideRound4E8460(pos.X)),
		Y: int(playerCollideRound4E8460(pos.Y)),
	}
}

// damageApplyVampirism4E0B30 restores the Vampirism side effects shared by
// DefaultDamage and PlayerDamage. Callers validate the four services before
// entering this helper so the observable order remains audio, coefficient,
// healing, and network FX.
func damageApplyVampirism4E0B30(
	source, target, weapon *Object,
	damage int32,
	audio func(int, *Object),
	balanceFloatInd func(string, int) float64,
	adjustHP func(*Object, int32),
	vampirismFX func(int, image.Point, image.Point, uint16),
) {
	audio(damageVampirismSound4E0B30, weapon)
	power := source.UnitBuffPower4FF570(int32(damageVampirismEnchant4E0B30))
	coefficient := balanceFloatInd(damageVampirismBalance4E0B30, int(power)-1)
	amount := uint16(playerCollideRound4E8460(float32(coefficient * float64(damage))))
	if amount < 1 {
		amount = 1
	}
	adjustHP(source, int32(amount))
	vampirismFX(
		damageVampirismFX4E0B30,
		damageVampirismPoint4E0B30(source.PosVec),
		damageVampirismPoint4E0B30(target.PosVec),
		amount,
	)
}
