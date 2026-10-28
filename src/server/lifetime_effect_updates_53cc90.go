package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

type LifetimeEffectKind53CC90 uint8

const (
	LifetimeEffectSelfDestruct53CC90 LifetimeEffectKind53CC90 = iota
	LifetimeEffectDeathBallFragment53D220
	LifetimeEffectMoonglow53D270
	LifetimeEffectTelekinesis53D330
	LifetimeEffectFist53D400
	LifetimeEffectFlameCleanse53D510
)

// LifetimeEffectRuntime53CC90 keeps native object/point identities at world
// service boundaries. Frame/FPS reads are live DWORD observations.
type LifetimeEffectRuntime53CC90 struct {
	Frame         func() uint32
	FPS           func() uint32
	DamageAround  func(*types.Pointf, float32, float32, int, object.DamageType, *Object, Obj)
	DelayedDelete func(*Object)
	BuffOff       func(*Object, EnchantID)
	ValidPosition func(types.Pointf) bool
	Move          func(*Object, *types.Pointf)
	SpellSound    func(spell.ID, int) sound.ID
	Audio         func(sound.ID, *Object)
	TraceRay      func(*types.Rectf, MapTraceFlags) bool
	Scorch        func(*types.Pointf, int)
	PointFX       func(uint8, *types.Pointf)
	Earthquake    func(*types.Pointf, int)
}

// LifetimeEffectUpdateRaw53CC90 preserves the six raw C entry contracts.
// Guarded public adapters retain their existing nil preflights separately.
func LifetimeEffectUpdateRaw53CC90(kind LifetimeEffectKind53CC90, source *Object, h LifetimeEffectRuntime53CC90) {
	switch kind {
	case LifetimeEffectSelfDestruct53CC90:
		if h.Frame()-source.Field32 > 2 {
			h.DelayedDelete(source)
		}
	case LifetimeEffectDeathBallFragment53D220:
		age := int32(h.Frame() - source.Field32)
		if age > 10 {
			h.DamageAround(&source.PosVec, 25, 0, 20, object.DamageCrush, source, nil)
		}
		// Both original JLE branches compare signed DWORDs, including FPS overflow.
		if age > int32(2*h.FPS()) {
			h.DelayedDelete(source)
		}
	case LifetimeEffectMoonglow53D270:
		owner := source.ObjOwner
		if owner == nil {
			h.DelayedDelete(source)
			return
		}
		if uint8(owner.ObjClass)&4 == 0 {
			return
		}
		flags := uint8(owner.ObjFlags)
		update := (*PlayerUpdateData)(owner.UpdateData)
		if flags&0x20 != 0 || h.Frame()-source.Field32 > 300*h.FPS() {
			h.DelayedDelete(source)
			h.BuffOff(source.ObjOwner, ENCHANT_MOONGLOW)
			return
		}
		point := lifetimeCursorPoint53D270(update)
		if h.ValidPosition(point) {
			h.Move(source, &point)
		}
	case LifetimeEffectTelekinesis53D330:
		lifetimeTelekinesis53D330(source, h, false)
	case LifetimeEffectFist53D400:
		if (source.ZVal <= 0 || math.IsNaN(float64(source.ZVal))) && int32(source.ObjFlags) >= 0 {
			h.Audio(sound.ID(48), source)
			h.Scorch(&source.PosVec, 2)
			x := objectForceAdd53_52DF80(objectForceMemory32_52DF80(source.Shape.Circle.R), objectForceMemory32_52DF80(source.PosVec.X))
			source.ObjFlags |= object.FlagMarked
			point := types.Ptf(experienceSpill32TowardZero(x), source.PosVec.Y)
			h.PointFX(138, &point)
			x = objectForceSub53_52DF80(float64(source.PosVec.X), objectForceMemory32_52DF80(source.Shape.Circle.R))
			point = types.Ptf(experienceSpill32TowardZero(x), source.PosVec.Y)
			h.PointFX(138, &point)
			y := objectForceAdd53_52DF80(objectForceMemory32_52DF80(source.Shape.Circle.R), objectForceMemory32_52DF80(source.PosVec.Y))
			point = types.Ptf(source.PosVec.X, experienceSpill32TowardZero(y))
			h.PointFX(138, &point)
			h.Earthquake(&source.PosVec, 30)
		}
		if source.ZVal >= 200 && int32(source.ObjFlags) < 0 {
			h.DelayedDelete(source)
		}
		if h.Frame()-source.Field32 > 3*h.FPS() {
			h.DelayedDelete(source)
		}
	case LifetimeEffectFlameCleanse53D510:
		if h.Frame() >= source.Field34 {
			h.DelayedDelete(source)
			return
		}
		ray := types.Rectf{Min: source.Pos39, Max: source.PosVec}
		if !h.TraceRay(&ray, MapTraceFlags(65)) {
			h.DelayedDelete(source)
			return
		}
		if h.Frame()-source.Field32 > 3 && flameCleanseEqual53D510(source.PrevPos.X, source.PosVec.X) && flameCleanseEqual53D510(source.PrevPos.Y, source.PosVec.Y) {
			h.DelayedDelete(source)
		}
	default:
		panic("unknown lifetime effect update")
	}
}

func lifetimeCursorPoint53D270(update *PlayerUpdateData) types.Pointf {
	// FILD signed DWORD, followed by FSTP binary32 in actual PC53/ToZero mode.
	x := experienceSpill32TowardZero(float64(int32(update.Player.CursorVec.X)))
	y := experienceSpill32TowardZero(float64(int32(update.Player.CursorVec.Y)))
	return types.Ptf(x, y)
}

func lifetimeTelekinesis53D330(source *Object, h LifetimeEffectRuntime53CC90, guarded bool) {
	owner := source.ObjOwner
	if guarded && (owner == nil || owner.UpdateData == nil) {
		h.DelayedDelete(source)
		return
	}
	flags := owner.ObjFlags
	update := (*PlayerUpdateData)(owner.UpdateData)
	if uint32(flags)&0x8020 != 0 || h.Frame()-source.Field32 > 20*h.FPS() {
		id := h.SpellSound(spell.SPELL_TELEKINESIS, 2)
		h.Audio(id, owner)
		h.DelayedDelete(source)
		h.BuffOff(source.ObjOwner, ENCHANT_TELEKINESIS)
		return
	}
	if guarded && update.Player == nil {
		h.DelayedDelete(source)
		return
	}
	point := lifetimeCursorPoint53D270(update)
	ray := types.Rectf{Min: owner.PosVec, Max: point}
	if h.TraceRay(&ray, MapTraceFlags(5)) {
		h.Move(source, &ray.Max)
	}
}
