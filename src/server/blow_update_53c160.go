package server

import (
	"image"
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	blowRadius53C160        = float32(400)
	blowDistanceBias53C240  = float32(0.1)
	blowForceScale53C240    = float32(0.0000005)
	blowVerticalSlope53C240 = float32(0.3732)
	blowDiagonalMin53C240   = float32(0.5773)
	blowDiagonalMax53C240   = float32(0.1732)
	blowHorizontal53C240    = float32(0.2679)
)

type blowUpdateDeps53C160 struct {
	indexedDirection func(int16) image.Point
	eachInRect       func(types.Rectf, func(*Object) bool)
	canInteract      func(*Object, *Object) bool
	directionVector  func(byte) (float32, float32)
}

func blowUpdateDirectionAllowed53C240(indexed image.Point, delta types.Pointf) bool {
	absX := math.Abs(float64(delta.X))
	absY := math.Abs(float64(delta.Y))
	slope := absY / absX
	switch indexed.X + 3*indexed.Y + 4 {
	case 0:
		return delta.X < 0 && delta.Y < 0 &&
			slope >= float64(blowDiagonalMin53C240) && slope <= float64(blowDiagonalMax53C240)
	case 1:
		return delta.Y < 0 && slope <= float64(blowVerticalSlope53C240)
	case 2:
		return delta.X > 0 && delta.Y < 0 &&
			slope >= float64(blowDiagonalMin53C240) && slope <= float64(blowDiagonalMax53C240)
	case 3:
		return delta.X < 0 && slope <= float64(blowHorizontal53C240)
	case 5:
		return delta.X > 0 && slope <= float64(blowHorizontal53C240)
	case 6:
		return delta.X < 0 && delta.Y > 0 &&
			slope >= float64(blowDiagonalMin53C240) && slope <= float64(blowDiagonalMax53C240)
	case 7:
		return delta.Y > 0 && slope <= float64(blowVerticalSlope53C240)
	case 8:
		return delta.X > 0 && delta.Y > 0 &&
			slope >= float64(blowDiagonalMin53C240) && slope <= float64(blowDiagonalMax53C240)
	default:
		return true
	}
}

// blowUpdateNative53C160 restores GAME.EXE 0053C160 and its scan callback
// 0053C240. The original diagonal upper bound really is 0.1732, despite being
// lower than its 0.5773 lower bound; retaining it preserves the original
// binary's behavior.
func blowUpdateNative53C160(source *Object, deps blowUpdateDeps53C160) {
	if source == nil || !source.ObjFlags.Has(object.FlagEnabled) ||
		deps.indexedDirection == nil || deps.eachInRect == nil ||
		deps.canInteract == nil || deps.directionVector == nil {
		return
	}
	indexed := deps.indexedDirection(int16(source.Direction1))
	rect := types.Rectf{Min: source.PosVec, Max: source.PosVec}
	if indexed.X < 0 {
		rect.Min.X -= blowRadius53C160
	} else {
		rect.Max.X += blowRadius53C160
	}
	if indexed.Y < 0 {
		rect.Min.Y -= blowRadius53C160
	} else {
		rect.Max.Y += blowRadius53C160
	}
	deps.eachInRect(rect, func(candidate *Object) bool {
		if candidate == nil || candidate.ObjFlags.Has(object.FlagDestroyed) ||
			candidate.ObjClass.Has(object.ClassImmobile) {
			return true
		}
		delta := candidate.PosVec.Sub(source.PosVec)
		distanceWide := math.Sqrt(float64(delta.X)*float64(delta.X)+float64(delta.Y)*float64(delta.Y)) +
			float64(blowDistanceBias53C240)
		if !(distanceWide < float64(blowRadius53C160)) ||
			!blowUpdateDirectionAllowed53C240(indexed, delta) ||
			!deps.canInteract(source, candidate) {
			return true
		}
		distance := float32(distanceWide)
		difference := float64(blowRadius53C160 - distance)
		force := float32(difference * difference * difference * float64(blowForceScale53C240))
		forcePerMass := float64(force) / float64(candidate.Mass)
		cosine, sine := deps.directionVector(byte(source.Direction1))
		candidate.ForceVec.X = float32(forcePerMass*float64(cosine) + float64(candidate.ForceVec.X))
		candidate.ForceVec.Y = float32(forcePerMass*float64(sine) + float64(candidate.ForceVec.Y))
		return true
	})
}

// BlowUpdate53C160 binds the blow object's world scan to native-width Object
// pointers instead of re-entering the PE32 callback through cgo.
func (s *Server) BlowUpdate53C160(source *Object) {
	if s == nil {
		return
	}
	blowUpdateNative53C160(source, blowUpdateDeps53C160{
		indexedDirection: indexedDirection509E20,
		eachInRect:       s.Map.EachObjAndMissileInRect,
		canInteract: func(source, candidate *Object) bool {
			return s.CanInteract(source, candidate, 0)
		},
		directionVector: SinCosDir,
	})
}
