package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// playerDamageCarry4E20F0 preserves the original binary32 spill before FISTP
// and the live player-before-monster carry accesses. The startup control word
// rounds arithmetic, stores and signed-DWORD conversion toward zero.
func playerDamageCarry4E20F0(target *Object, output *int32, amount float32) {
	if target == nil || output == nil {
		return
	}
	accumulated := amount
	var carry *uint32
	switch {
	case target.Class().Has(object.ClassPlayer):
		carry = &target.UpdateDataPlayer().Field21
	case target.Class().Has(object.ClassMonster):
		carry = &target.UpdateDataMonster().Field1
	}
	if carry != nil {
		accumulated = experienceSpill32TowardZero(experienceAdd53TowardZero(float64(amount), float64(math.Float32frombits(*carry))))
	}
	effective := netFXTruncateSignedDWORD419A70(accumulated)
	*output = effective
	// The original reloads both class and update after writing the integer
	// output. Keep the integer exact until the final fractional-carry spill.
	switch {
	case target.Class().Has(object.ClassPlayer):
		remaining := experienceSpill32TowardZero(experienceAdd53TowardZero(float64(accumulated), -float64(effective)))
		target.UpdateDataPlayer().Field21 = math.Float32bits(remaining)
	case target.Class().Has(object.ClassMonster):
		remaining := experienceSpill32TowardZero(experienceAdd53TowardZero(float64(accumulated), -float64(effective)))
		target.UpdateDataMonster().Field1 = math.Float32bits(remaining)
	}
}
