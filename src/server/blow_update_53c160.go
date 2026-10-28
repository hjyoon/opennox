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
	directionFloat   func(int32) float32
}

func blowUpdateDirectionAllowed53C240(indexed image.Point, delta types.Pointf) bool {
	// FCOM's C0 test treats unordered as negative; FST/FCHS keeps signed zero.
	absX, absY := delta.X, delta.Y
	xLess, yLess := !(delta.X >= 0), !(delta.Y >= 0)
	if xLess {
		absX = math.Float32frombits(math.Float32bits(absX) ^ 0x80000000)
	}
	if yLess {
		absY = math.Float32frombits(math.Float32bits(absY) ^ 0x80000000)
	}
	ratio := func() float64 {
		return objectForceDiv53_52DF80(objectForceMemory32_52DF80(absY), objectForceMemory32_52DF80(absX))
	}
	within := func(maximum float32) bool { slope := ratio(); return slope <= float64(maximum) || math.IsNaN(slope) }
	diagonal := func() bool {
		slope := ratio()
		return slope >= float64(blowDiagonalMin53C240) && (slope <= float64(blowDiagonalMax53C240) || math.IsNaN(slope))
	}
	// LEA builds a wrapping signed DWORD before the unsigned jump-table bound.
	switch int32(indexed.X) + 3*int32(indexed.Y) + 4 {
	case 0:
		return xLess && yLess && diagonal()
	case 1:
		return yLess && within(blowVerticalSlope53C240)
	case 2:
		return delta.X > 0 && yLess && diagonal()
	case 3:
		return xLess && within(blowHorizontal53C240)
	case 5:
		return delta.X > 0 && within(blowHorizontal53C240)
	case 6:
		return xLess && delta.Y > 0 && diagonal()
	case 7:
		return delta.Y > 0 && within(blowVerticalSlope53C240)
	case 8:
		return delta.X > 0 && delta.Y > 0 && diagonal()
	default:
		return true
	}
}

// BlowUpdateRaw53C160 retains the original required source pointer, live
// direction classification, PC53/ToZero rectangle spills and callback scan.
// The public guarded adapter below keeps its existing nil preflights.
func BlowUpdateRaw53C160(source *Object, runtime BlowUpdateRuntime53C160) {
	if !source.ObjFlags.Has(object.FlagEnabled) {
		return
	}
	indexed := runtime.IndexedDirection(int16(source.Direction1))
	rect := types.Rectf{Min: source.PosVec, Max: source.PosVec}
	if indexed.X < 0 {
		rect.Min.X = experienceSpill32TowardZero(objectForceSub53_52DF80(float64(source.PosVec.X), 400))
	} else {
		rect.Max.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(source.PosVec.X), 400))
	}
	if indexed.Y < 0 {
		rect.Min.Y = experienceSpill32TowardZero(objectForceSub53_52DF80(float64(source.PosVec.Y), 400))
	} else {
		rect.Max.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(source.PosVec.Y), 400))
	}
	runtime.EachInRect(rect, func(candidate *Object) bool {
		BlowCandidateRaw53C240(candidate, source, runtime)
		return true
	})
}

// BlowCandidateRaw53C240 preserves the complete original candidate callback.
// Filtering precedes the required source read. Distance retains the unspilled
// Y difference; classification and both signed-WORD vector indices are live.
// The original diagonal upper bound remains below its lower bound.
func BlowCandidateRaw53C240(candidate, source *Object, runtime BlowUpdateRuntime53C160) {
	if candidate.ObjFlags.Has(object.FlagDestroyed) || candidate.ObjClass.Has(object.ClassImmobile) {
		return
	}
	dx := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(candidate.PosVec.X), objectForceMemory32_52DF80(source.PosVec.X)))
	dyWide := objectForceSub53_52DF80(float64(candidate.PosVec.Y), objectForceMemory32_52DF80(source.PosVec.Y))
	dy := experienceSpill32TowardZero(dyWide)
	squareY := objectForceMul53_52DF80(dyWide, float64(dy))
	squareX := objectForceMul53_52DF80(float64(dx), float64(dx))
	square := objectForceAdd53_52DF80(squareY, squareX)
	distanceWide := objectForceAdd53_52DF80(monsterMoveForceSqrtChop53_50D50C(square), float64(blowDistanceBias53C240))
	distance := experienceSpill32TowardZero(distanceWide)
	// C0 is set for both less-than and unordered.
	if distanceWide >= 400 {
		return
	}
	indexed := runtime.IndexedDirection(int16(source.Direction1))
	if !blowUpdateDirectionAllowed53C240(indexed, types.Ptf(dx, dy)) || !runtime.CanInteract(source, candidate) {
		return
	}
	difference := objectForceSub53_52DF80(400, float64(distance))
	force := objectForceMul53_52DF80(difference, difference)
	force = objectForceMul53_52DF80(force, difference)
	force = objectForceMul53_52DF80(force, float64(blowForceScale53C240))
	force32 := experienceSpill32TowardZero(force)
	// GAME.EXE's complete 004E4A70 getter is FLD of this live float field.
	factor := objectForceDiv53_52DF80(float64(force32), objectForceMemory32_52DF80(candidate.Mass))
	read := runtime.DirectionFloat
	if read == nil {
		read = magicMissileDirectionFloat53BDA0
	}
	cosine := objectForceMemory32_52DF80(read(int32(int16(source.Direction1)) * 2))
	x := objectForceMul53_52DF80(factor, cosine)
	candidate.ForceVec.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(x, objectForceMemory32_52DF80(candidate.ForceVec.X)))
	sine := objectForceMemory32_52DF80(read(int32(int16(source.Direction1))*2 + 1))
	y := objectForceMul53_52DF80(factor, sine)
	candidate.ForceVec.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(y, objectForceMemory32_52DF80(candidate.ForceVec.Y)))
}

func blowUpdateNative53C160(source *Object, deps blowUpdateDeps53C160) {
	if source == nil || !source.ObjFlags.Has(object.FlagEnabled) || deps.indexedDirection == nil || deps.eachInRect == nil || deps.canInteract == nil || deps.directionVector == nil {
		return
	}
	read := deps.directionFloat
	if read == nil {
		read = func(index int32) float32 {
			cosine, sine := deps.directionVector(byte(index >> 1))
			if index&1 == 0 {
				return cosine
			}
			return sine
		}
	}
	BlowUpdateRaw53C160(source, BlowUpdateRuntime53C160{
		IndexedDirection: deps.indexedDirection,
		EachInRect: func(rect types.Rectf, fn func(*Object) bool) {
			deps.eachInRect(rect, func(candidate *Object) bool {
				if candidate == nil {
					return true
				}
				return fn(candidate)
			})
		},
		CanInteract:    deps.canInteract,
		DirectionFloat: read,
	})
}

// BlowUpdate53C160 binds the guarded Go API to native-width world services.
func (s *Server) BlowUpdate53C160(source *Object) {
	if s == nil {
		return
	}
	blowUpdateNative53C160(source, blowUpdateDeps53C160{
		indexedDirection: indexedDirection509E20,
		eachInRect:       s.Map.EachObjAndMissileInRect,
		canInteract:      func(source, candidate *Object) bool { return s.CanInteract(source, candidate, 0) },
		directionVector:  SinCosDir,
		directionFloat:   magicMissileDirectionFloat53BDA0,
	})
}
