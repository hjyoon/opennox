package server

import "math"

const objectColliderLimit4E7410 = 85.0

type objectColliderState4E7410 interface {
	objectColliderState4E7290
	colliderFlagsByte4E7410() uint8
	colliderMinXBits4E7410() uint32
	colliderMinYBits4E7410() uint32
	colliderMaxXBits4E7410() uint32
	colliderMaxYBits4E7410() uint32
}

// objectColliderExtentBelowLimit4E7410 reproduces the C0-only x87 test used by
// GAME.EXE. Binary32 inputs are widened before subtraction because the x87
// result is not rounded back to binary32. An unordered comparison (NaN) sets
// C0 just like a value below the 85.0f limit.
func objectColliderExtentBelowLimit4E7410(maxBits, minBits uint32) bool {
	maxValue := float64(math.Float32frombits(maxBits))
	minValue := float64(math.Float32frombits(minBits))
	return !(maxValue-minValue >= objectColliderLimit4E7410)
}

// objectColliderAllowed4E7410 preserves the control flow at 004E7410. The
// low-byte NoCollide flag accepts the object without refreshing or reading its
// bounds. Otherwise current-position bounds are refreshed first, width is
// tested before height, and a failed width test short-circuits the height read.
func objectColliderAllowed4E7410[T objectColliderState4E7410](obj T) bool {
	if obj.colliderFlagsByte4E7410()&0x40 != 0 {
		return true
	}
	objectUpdateCollider4E7290(obj)
	if !objectColliderExtentBelowLimit4E7410(
		obj.colliderMaxXBits4E7410(), obj.colliderMinXBits4E7410(),
	) {
		return false
	}
	return objectColliderExtentBelowLimit4E7410(
		obj.colliderMaxYBits4E7410(), obj.colliderMinYBits4E7410(),
	)
}
