package server

import "math"

const (
	chakramRetargetClassMask4EB250 = uint32(0x00020006)
	chakramRetargetFlagsMask4EB250 = uint32(0x00008020)
	chakramRetargetInvisible4EB250 = uint32(0)
	chakramRetargetRange4EB250     = float32(400)
	chakramRetargetRangeSq4EB250   = float32(160000)
	chakramRetargetBestBits4EB250  = uint32(0x4b189680)
	chakramRetargetEpsilon4EB250   = float32(0.1)
)

type chakramRetargetRect4EB250 struct {
	MinX float32
	MinY float32
	MaxX float32
	MaxY float32
}

type chakramRetargetHooks4EB250[O comparable, U any] struct {
	loadUpdateData func(O) U
	loadLastHit    func(U) O
	loadOwner      func(O) O
	loadClass      func(O) uint32
	loadFlags      func(O) uint32
	hasEnchant     func(O, uint32) bool
	mapCheck       func(O, O) bool
	loadPosX       func(O) float32
	loadPosY       func(O) float32
	loadSpeed      func(O) float32
	eachInRect     func(chakramRetargetRect4EB250, func(O))
	storeState     func(U, uint8)
	storeVelocityX func(O, float32)
	storeVelocityY func(O, float32)
	// Ordinary gameplay uses the original startup FPU mode. The generic
	// nearest-profile controls retain their established default contract.
	roundTowardZero bool
}

// chakramRetarget4EB250 preserves GAME.EXE 004EB250 and its 004EB340 search
// callback. The source owner and previous hit are entry-cached exclusions,
// while all positions remain live throughout enumeration. Ordered distances
// above 400 units and ordered distances greater than or equal to the current
// best are rejected; unordered NaNs follow the original x87 acceptance path.
func chakramRetarget4EB250[O comparable, U any](
	source O,
	hooks chakramRetargetHooks4EB250[O, U],
) O {
	f := chakramRetargetArithmetic4EB250{nearest: !hooks.roundTowardZero}
	update := hooks.loadUpdateData(source)
	lastHit := hooks.loadLastHit(update)
	owner := hooks.loadOwner(source)
	bestDistance := math.Float32frombits(chakramRetargetBestBits4EB250)

	rect := chakramRetargetRect4EB250{
		MinX: f.spill(f.sub(float64(hooks.loadPosX(source)), float64(chakramRetargetRange4EB250))),
		MinY: f.spill(f.sub(float64(hooks.loadPosY(source)), float64(chakramRetargetRange4EB250))),
		MaxX: f.spill(f.add(float64(hooks.loadPosX(source)), float64(chakramRetargetRange4EB250))),
		MaxY: f.spill(f.add(float64(hooks.loadPosY(source)), float64(chakramRetargetRange4EB250))),
	}
	var best O
	hooks.eachInRect(rect, func(candidate O) {
		if hooks.loadClass(candidate)&chakramRetargetClassMask4EB250 == 0 {
			return
		}
		if hooks.loadFlags(candidate)&chakramRetargetFlagsMask4EB250 != 0 {
			return
		}
		if hooks.hasEnchant(candidate, chakramRetargetInvisible4EB250) {
			return
		}
		if candidate == owner || candidate == lastHit {
			return
		}
		if !hooks.mapCheck(candidate, source) {
			return
		}
		dx := f.sub(float64(hooks.loadPosX(source)), float64(hooks.loadPosX(candidate)))
		dy := f.sub(float64(hooks.loadPosY(source)), float64(hooks.loadPosY(candidate)))
		distance := f.add(f.mul(dy, dy), f.mul(dx, dx))
		if distance > float64(chakramRetargetRangeSq4EB250) || distance >= float64(bestDistance) {
			return
		}
		bestDistance = f.spill(distance)
		best = candidate
	})

	var zero O
	if best == zero {
		return zero
	}
	hooks.storeState(update, chakramReturnStateSeek4EAF00)
	dx := f.sub(float64(hooks.loadPosX(best)), float64(hooks.loadPosX(source)))
	dyExtended := f.sub(float64(hooks.loadPosY(best)), float64(hooks.loadPosY(source)))
	// 004EB2F3 stores, but does not pop, Y before multiplying the extended
	// value by that binary32 spill. X remains entirely in x87 registers.
	dy := f.spill(dyExtended)
	denominator := f.spill(f.add(f.sqrt(f.add(f.mul(dyExtended, float64(dy)), f.mul(dx, dx))), float64(chakramRetargetEpsilon4EB250)))
	hooks.storeVelocityX(source, f.spill(f.div(f.mul(dx, float64(hooks.loadSpeed(source))), float64(denominator))))
	hooks.storeVelocityY(source, f.spill(f.div(f.mul(float64(dy), float64(hooks.loadSpeed(source))), float64(denominator))))
	return best
}
