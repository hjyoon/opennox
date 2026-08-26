package server

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
)

const boxBoxSqrtHalf550F80 = float32(0.70710677)

type boxBoxCollisionHooks550F80 struct {
	pushScale    float32
	addCollision func(*Object, *Object, types.Pointf)
	wake         func(*Object)
}

// boxBoxCollision550F80 restores GAME.EXE 00550F80 without treating the
// native-width Object records as PE32 float arrays.
func boxBoxCollision550F80(first, second *Object, hooks boxBoxCollisionHooks550F80) {
	if first == nil || second == nil {
		return
	}
	firstX := (first.NewPos.X + first.NewPos.Y) * boxBoxSqrtHalf550F80
	firstY := (first.NewPos.Y - first.NewPos.X) * boxBoxSqrtHalf550F80
	secondX := (second.NewPos.X + second.NewPos.Y) * boxBoxSqrtHalf550F80
	secondY := (second.NewPos.Y - second.NewPos.X) * boxBoxSqrtHalf550F80

	firstHalfW := first.Shape.Box.W * 0.5
	firstHalfH := first.Shape.Box.H * 0.5
	firstMinX := firstX - firstHalfW
	firstMinY := firstY - firstHalfH
	firstMaxX := firstX + firstHalfW
	firstMaxY := firstY + firstHalfH
	secondHalfW := second.Shape.Box.W * 0.5
	secondHalfH := second.Shape.Box.H * 0.5
	secondMinX := secondX - secondHalfW
	secondMinY := secondY - secondHalfH
	secondMaxX := secondX + secondHalfW
	secondMaxY := secondY + secondHalfH

	if firstMinX > secondMaxX || firstMinY > secondMaxY || firstMaxX < secondMinX || firstMaxY < secondMinY {
		return
	}
	hooks.addCollision(second, first, second.NewPos.Sub(first.NewPos))
	canPush := first.ObjFlags&0x8 == 0 && second.ObjFlags&0x8 == 0
	if (uint8(first.ObjClass)&0x6 == 0 || second.ObjFlags&0x2000 == 0) && canPush {
		overlapX := min(firstMaxX, secondMaxX) - max(firstMinX, secondMinX)
		overlapY := min(firstMaxY, secondMaxY) - max(firstMinY, secondMinY)
		var pushX, pushY float32
		if overlapX >= overlapY {
			if firstY >= secondY {
				pushY = overlapY
			} else {
				pushY = -overlapY
			}
		} else {
			pushX = overlapX
			if firstX < secondX {
				pushX = -pushX
			}
		}
		rotatedX := pushX * hooks.pushScale
		rotatedY := pushY * hooks.pushScale
		first.Sub548600(types.Ptf(
			(rotatedX-rotatedY)*boxBoxSqrtHalf550F80,
			(rotatedY+rotatedX)*boxBoxSqrtHalf550F80,
		))
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

func (s *Server) BoxBoxCollision550F80(first, second *Object, addCollision func(*Object, *Object, types.Pointf), wake func(*Object)) {
	boxBoxCollision550F80(first, second, boxBoxCollisionHooks550F80{
		pushScale:    memmap.Float32(0x587000, 292488),
		addCollision: addCollision,
		wake:         wake,
	})
}
