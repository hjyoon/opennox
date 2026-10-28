package server

import (
	"image"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// MapRadialDamageContext4E25B0 is the transient native-width callback context.
// It is neither an object update record nor serialized map/save data.
type MapRadialDamageContext4E25B0 struct {
	Center   *types.Pointf
	Outer    float32
	Inner    float32
	Damage   int32
	Type     int32
	Source   *Object
	Excluded *Object
}

// MapRadialDamageRuntime4E25B0 keeps the live center/context and the shared
// source-damage flag across nested object callbacks. DamageWalls is always
// called; the historically named wall flag controls hitting Source itself.
type MapRadialDamageRuntime4E25B0 struct {
	DamageSource      func() bool
	ResetDamageSource func()
	EachInRect        func(*types.Rectf, *MapRadialDamageContext4E25B0, func(*Object, *MapRadialDamageContext4E25B0))
	Trace             func(*types.Rectf) int32
	Damage            func(*Object, *Object, int32, object.DamageType)
	DamageWalls       func(image.Rectangle, *types.Pointf, float32, int32, object.DamageType, *Object)
}

// MapRadialDamageRaw4E25B0 restores the complete original root, including its
// unordered radius selection, directed rectangle spills, signed-DWORD grid
// conversions and original (unswapped) radius passed to wall damage.
func MapRadialDamageRaw4E25B0(center *types.Pointf, outer, inner float32, damage int32, typ object.DamageType, source, excluded *Object, r MapRadialDamageRuntime4E25B0) {
	radius := outer
	if !(outer >= inner) {
		radius = inner
	}
	c := MapRadialDamageContext4E25B0{Center: center, Outer: radius, Inner: inner, Damage: damage, Type: int32(typ), Source: source, Excluded: excluded}
	rect := types.Rectf{
		Min: types.Ptf(
			experienceSpill32TowardZero(objectForceSub53_52DF80(float64(center.X), objectForceMemory32_52DF80(radius))),
			experienceSpill32TowardZero(objectForceSub53_52DF80(float64(center.Y), objectForceMemory32_52DF80(radius)))),
		Max: types.Ptf(
			experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(radius), objectForceMemory32_52DF80(center.X))),
			experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(radius), objectForceMemory32_52DF80(center.Y)))),
	}
	r.EachInRect(&rect, &c, func(target *Object, context *MapRadialDamageContext4E25B0) {
		MapRadialDamageTargetRaw4E2700(target, context, r)
	})
	// image.Rect canonicalizes inverted bounds; the original four signed
	// divisions preserve their order, including invalid FISTP's MinInt32.
	grid := image.Rectangle{
		Min: image.Pt(int(netFXTruncateSignedDWORD419A70(rect.Min.X)/23), int(netFXTruncateSignedDWORD419A70(rect.Min.Y)/23)),
		Max: image.Pt(int(netFXTruncateSignedDWORD419A70(rect.Max.X)/23), int(netFXTruncateSignedDWORD419A70(rect.Max.Y)/23)),
	}
	r.DamageWalls(grid, center, outer, damage, typ, source)
	r.ResetDamageSource()
}

// MapRadialDamageTargetRaw4E2700 retains the extended distance for the outer
// comparison, its binary32 spill for falloff, and the qword conversion's low
// DWORD. Services may replace live context fields and object callback slots.
func MapRadialDamageTargetRaw4E2700(target *Object, c *MapRadialDamageContext4E25B0, r MapRadialDamageRuntime4E25B0) {
	if target == c.Source && !r.DamageSource() {
		return
	}
	if target == c.Excluded {
		return
	}
	center := c.Center
	dx := objectForceSub53_52DF80(float64(target.PosVec.X), objectForceMemory32_52DF80(center.X))
	dy := objectForceSub53_52DF80(float64(target.PosVec.Y), objectForceMemory32_52DF80(center.Y))
	square := objectForceAdd53_52DF80(objectForceMul53_52DF80(dy, dy), objectForceMul53_52DF80(dx, dx))
	distance := monsterMoveForceSqrtChop53_50D50C(square)
	storedDistance := experienceSpill32TowardZero(distance)
	if distance > float64(c.Outer) {
		return
	}
	ray := types.Rectf{Min: *center, Max: target.PosVec}
	if r.Trace(&ray) == 0 {
		return
	}
	damage := c.Damage
	if storedDistance >= c.Inner {
		numerator := objectForceSub53_52DF80(float64(storedDistance), objectForceMemory32_52DF80(c.Inner))
		denominator := objectForceSub53_52DF80(float64(c.Outer), objectForceMemory32_52DF80(c.Inner))
		factor := objectForceSub53_52DF80(1, objectForceDiv53_52DF80(numerator, denominator))
		damage = x87TruncSignedQwordLow566DCC(objectForceMul53_52DF80(factor, float64(c.Damage)))
	}
	if target.Damage == nil {
		panic("nil radial-damage callback at 004E27C1")
	}
	r.Damage(target, c.Source, damage, object.DamageType(c.Type))
}
