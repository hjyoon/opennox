package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const boxBoxSqrtHalf550F80 = float32(0.70710677)

type boxBoxCollisionHooks550F80 struct {
	pushScale    func() float32
	addCollision func(*Object, *Object, types.Pointf)
	wake         func(*Object)
}

// boxBoxCollision550F80 preserves startup PC53/ToZero operations and loads
// the live C-owned push coefficient after the collision service.
func boxBoxCollision550F80(first, second *Object, hooks boxBoxCollisionHooks550F80) {
	if first == nil || second == nil {
		return
	}
	add, sub, mul := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	k := float64(boxBoxSqrtHalf550F80)
	bounds := func(obj *Object) (x, y, minX, minY, maxX, maxY float32) {
		x = spill(mul(add(float64(obj.NewPos.X), memory(obj.NewPos.Y)), k))
		y = spill(mul(sub(float64(obj.NewPos.Y), memory(obj.NewPos.X)), k))
		halfW := mul(float64(obj.Shape.Box.W), .5)
		halfH := spill(mul(float64(obj.Shape.Box.H), .5))
		return x, y, spill(sub(float64(x), halfW)), spill(sub(float64(y), memory(halfH))),
			spill(add(halfW, memory(x))), spill(add(float64(halfH), memory(y)))
	}
	firstX, firstY, firstMinX, firstMinY, firstMaxX, firstMaxY := bounds(first)
	secondX, secondY, secondMinX, secondMinY, secondMaxX, secondMaxY := bounds(second)
	if firstMinX > secondMaxX || firstMinY > secondMaxY || !(firstMaxX >= secondMinX) || !(firstMaxY >= secondMinY) {
		return
	}
	hooks.addCollision(second, first, types.Ptf(
		spill(sub(float64(second.NewPos.X), memory(first.NewPos.X))),
		spill(sub(float64(second.NewPos.Y), memory(first.NewPos.Y))),
	))
	canPush := first.ObjFlags&8 == 0 && second.ObjFlags&8 == 0
	if (uint8(first.ObjClass)&6 == 0 || second.ObjFlags&0x2000 == 0) && canPush {
		low := func(a, b float32) float32 {
			if a > b {
				return a
			}
			return b
		}
		high := func(a, b float32) float32 {
			if a < b || math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
				return a
			}
			return b
		}
		overlapX := spill(sub(float64(high(firstMaxX, secondMaxX)), float64(low(firstMinX, secondMinX))))
		overlapY := spill(sub(float64(high(firstMaxY, secondMaxY)), float64(low(firstMinY, secondMinY))))
		var pushX, pushY float32
		if !(overlapX >= overlapY) {
			pushX = overlapX
			if !(firstX >= secondX) {
				pushX = -pushX
			}
		} else {
			pushY = overlapY
			if !(firstY >= secondY) {
				pushY = -pushY
			}
		}
		// Neither scaled component is spilled before rotation.
		x := mul(float64(pushX), memory(hooks.pushScale()))
		y := mul(float64(pushY), memory(hooks.pushScale()))
		first.Sub548600(types.Ptf(spill(mul(sub(x, y), k)), spill(mul(add(y, x), k))))
	}
	if first.ObjFlags&0x08000000 != 0 {
		hooks.wake(first)
		first.ObjFlags &^= 0x08000000
	}
	if second.ObjFlags&0x08000000 != 0 {
		hooks.wake(second)
		second.ObjFlags &^= 0x08000000
	}
}

func (s *Server) BoxBoxCollision550F80(first, second *Object, pushScale func() float32, addCollision func(*Object, *Object, types.Pointf), wake func(*Object)) {
	boxBoxCollision550F80(first, second, boxBoxCollisionHooks550F80{
		pushScale:    pushScale,
		addCollision: addCollision,
		wake:         wake,
	})
}
