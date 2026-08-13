package server

import "math"

// objectColliderState4E7290 is the minimum native-width contract needed by
// GAME.EXE's current-position collider update. Values cross this boundary as
// raw words so the center-shape path remains a literal bit copy, including
// signed zero and NaN payloads.
type objectColliderState4E7290 interface {
	colliderKind4E7290() uint32
	colliderPosXBits4E7290() uint32
	colliderPosYBits4E7290() uint32
	colliderRadiusBits4E7290() uint32
	colliderBoxMinXBits4E7290() uint32
	colliderBoxMinYBits4E7290() uint32
	colliderBoxMaxXBits4E7290() uint32
	colliderBoxMaxYBits4E7290() uint32
	colliderStoreMinXBits4E7290(uint32)
	colliderStoreMinYBits4E7290(uint32)
	colliderStoreMaxXBits4E7290(uint32)
	colliderStoreMaxYBits4E7290(uint32)
}

func colliderAddBits4E7290(a, b uint32) uint32 {
	return math.Float32bits(math.Float32frombits(a) + math.Float32frombits(b))
}

func colliderSubBits4E7290(a, b uint32) uint32 {
	return math.Float32bits(math.Float32frombits(a) - math.Float32frombits(b))
}

// objectUpdateCollider4E7290 preserves the load/store order at 004E7290. The
// original returns the input pointer in EAX on every path, including unknown
// shape kinds that do not touch the four output words.
func objectUpdateCollider4E7290[T objectColliderState4E7290](obj T) T {
	switch obj.colliderKind4E7290() {
	case 1: // center
		x := obj.colliderPosXBits4E7290()
		obj.colliderStoreMinXBits4E7290(x)
		y := obj.colliderPosYBits4E7290()
		obj.colliderStoreMaxXBits4E7290(x)
		obj.colliderStoreMinYBits4E7290(y)
		obj.colliderStoreMaxYBits4E7290(y)
	case 2: // circle
		obj.colliderStoreMinXBits4E7290(colliderSubBits4E7290(
			obj.colliderPosXBits4E7290(), obj.colliderRadiusBits4E7290(),
		))
		obj.colliderStoreMinYBits4E7290(colliderSubBits4E7290(
			obj.colliderPosYBits4E7290(), obj.colliderRadiusBits4E7290(),
		))
		obj.colliderStoreMaxXBits4E7290(colliderAddBits4E7290(
			obj.colliderRadiusBits4E7290(), obj.colliderPosXBits4E7290(),
		))
		obj.colliderStoreMaxYBits4E7290(colliderAddBits4E7290(
			obj.colliderRadiusBits4E7290(), obj.colliderPosYBits4E7290(),
		))
	case 3: // box
		obj.colliderStoreMinXBits4E7290(colliderAddBits4E7290(
			obj.colliderBoxMinXBits4E7290(), obj.colliderPosXBits4E7290(),
		))
		obj.colliderStoreMinYBits4E7290(colliderAddBits4E7290(
			obj.colliderBoxMinYBits4E7290(), obj.colliderPosYBits4E7290(),
		))
		obj.colliderStoreMaxXBits4E7290(colliderAddBits4E7290(
			obj.colliderBoxMaxXBits4E7290(), obj.colliderPosXBits4E7290(),
		))
		obj.colliderStoreMaxYBits4E7290(colliderAddBits4E7290(
			obj.colliderBoxMaxYBits4E7290(), obj.colliderPosYBits4E7290(),
		))
	}
	return obj
}
