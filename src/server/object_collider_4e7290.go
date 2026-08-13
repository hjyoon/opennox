package server

import "math"

// objectColliderShapeState4E72x0 is the common native-width shape and output
// contract shared by GAME.EXE's current- and next-position collider updates.
// Values cross this boundary as raw words so center-shape paths remain literal
// bit copies, including signed zero and NaN payloads.
type objectColliderShapeState4E72x0 interface {
	colliderKind4E7290() uint32
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

type objectColliderState4E7290 interface {
	objectColliderShapeState4E72x0
	colliderPosXBits4E7290() uint32
	colliderPosYBits4E7290() uint32
}

type objectColliderState4E7350 interface {
	objectColliderShapeState4E72x0
	colliderNewPosXBits4E7350() uint32
	colliderNewPosYBits4E7350() uint32
}

func colliderAddBits4E7290(a, b uint32) uint32 {
	return math.Float32bits(math.Float32frombits(a) + math.Float32frombits(b))
}

func colliderSubBits4E7290(a, b uint32) uint32 {
	return math.Float32bits(math.Float32frombits(a) - math.Float32frombits(b))
}

func objectUpdateColliderAt4E72x0[T objectColliderShapeState4E72x0](
	obj T, posX, posY func() uint32,
) T {
	switch obj.colliderKind4E7290() {
	case 1: // center
		x := posX()
		obj.colliderStoreMinXBits4E7290(x)
		y := posY()
		obj.colliderStoreMaxXBits4E7290(x)
		obj.colliderStoreMinYBits4E7290(y)
		obj.colliderStoreMaxYBits4E7290(y)
	case 2: // circle
		obj.colliderStoreMinXBits4E7290(colliderSubBits4E7290(
			posX(), obj.colliderRadiusBits4E7290(),
		))
		obj.colliderStoreMinYBits4E7290(colliderSubBits4E7290(
			posY(), obj.colliderRadiusBits4E7290(),
		))
		obj.colliderStoreMaxXBits4E7290(colliderAddBits4E7290(
			obj.colliderRadiusBits4E7290(), posX(),
		))
		obj.colliderStoreMaxYBits4E7290(colliderAddBits4E7290(
			obj.colliderRadiusBits4E7290(), posY(),
		))
	case 3: // box
		obj.colliderStoreMinXBits4E7290(colliderAddBits4E7290(
			obj.colliderBoxMinXBits4E7290(), posX(),
		))
		obj.colliderStoreMinYBits4E7290(colliderAddBits4E7290(
			obj.colliderBoxMinYBits4E7290(), posY(),
		))
		obj.colliderStoreMaxXBits4E7290(colliderAddBits4E7290(
			obj.colliderBoxMaxXBits4E7290(), posX(),
		))
		obj.colliderStoreMaxYBits4E7290(colliderAddBits4E7290(
			obj.colliderBoxMaxYBits4E7290(), posY(),
		))
	}
	return obj
}

// objectUpdateCollider4E7290 preserves the load/store order at 004E7290. The
// original returns the input pointer in EAX on every path, including unknown
// shape kinds that do not touch the four output words.
func objectUpdateCollider4E7290[T objectColliderState4E7290](obj T) T {
	return objectUpdateColliderAt4E72x0(obj, obj.colliderPosXBits4E7290, obj.colliderPosYBits4E7290)
}

// objectUpdateCollider4E7350 is the byte-for-byte structural twin at 004E7350
// that reads NewPos instead of PosVec and otherwise preserves the same return,
// branch, arithmetic, and store contract.
func objectUpdateCollider4E7350[T objectColliderState4E7350](obj T) T {
	return objectUpdateColliderAt4E72x0(obj, obj.colliderNewPosXBits4E7350, obj.colliderNewPosYBits4E7350)
}
