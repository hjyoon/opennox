package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// MonsterCastRuntime541300 retains the external services of the three raw
// cast entries. Object, argument and output owners keep their native widths.
type MonsterCastRuntime541300 struct {
	MorphTo     func(*Object) byte
	MorphFrom   func(*Object) byte
	Direction   func(*Object, *types.Pointf)
	Cast        func(int32, *Object, *SpellAcceptArg) int32
	Buff        func(*Object, int32) int32
	Voice       func(*Object) unsafe.Pointer
	Audio       func(uint32, *Object)
	RandomFloat func(float32, float32) float64
}

// MonsterCastNative541300 deliberately caches the initial monster record.
// The second BOT test follows live changes to that record after the cast.
func MonsterCastNative541300(id int32, unit *Object, arg *SpellAcceptArg, rt MonsterCastRuntime541300) byte {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if update.StatusFlags.Has(object.MonStatusBot) {
		rt.MorphTo(unit)
	}
	rt.Direction(unit, &arg.Pos)
	rt.Cast(id, unit, arg)
	flags := update.StatusFlags
	if flags.Has(object.MonStatusBot) {
		return rt.MorphFrom(unit)
	}
	return byte(flags)
}

// MonsterActionCastNative5413B0 keeps the head captured before any service,
// then reloads its spell after recoil and the full-width buff result.
func MonsterActionCastNative5413B0(unit *Object, mode int32, rt MonsterCastRuntime541300) byte {
	update := (*MonsterUpdateData)(unit.UpdateData)
	head := &update.AIStack[int(update.AIStackInd)]
	if ready := update.Field120_2; ready != 0 {
		return ready
	}
	frame := update.Field120_1
	if uint32(frame) != update.MonsterDef.MissileAttackFrame216 {
		if frame != 1 {
			return frame
		}
		voice := rt.Voice(unit)
		if voice != nil {
			rt.Audio(*(*uint32)(unsafe.Add(voice, 56)), unit)
		}
		// Production callers observe this action's effects. Keep the incoming
		// C source's low pointer byte; a void audio call's physical EAX is not
		// a defined action result and is not an original-register claim.
		return byte(uintptr(voice))
	}
	var arg SpellAcceptArg
	if mode != 0 {
		if mode != 1 {
			return byte(uint32(mode) - 1)
		}
		arg.Pos = head.ArgPos(2)
	} else {
		arg.Obj = head.ArgObj(2)
		MonsterCastRandomRecoilNative541490(unit, arg.Obj, &arg.Pos, rt)
	}
	if result := rt.Buff(unit, 29); result != 0 {
		return byte(result)
	}
	return MonsterCastNative541300(int32(head.ArgU32(0)), unit, &arg, rt)
}

// MonsterCastRandomRecoilNative541490 preserves every binary32 spill and
// live load across RNG. Output may alias the target's position or velocity.
func MonsterCastRandomRecoilNative541490(unit, target *Object, out *types.Pointf, rt MonsterCastRuntime541300) {
	update := (*MonsterUpdateData)(unit.UpdateData)
	variance := experienceSpill32TowardZero(objectForceSub53_52DF80(1, objectForceMemory32_52DF80(update.Field330)))
	maximum := experienceSpill32TowardZero(objectForceAdd53_52DF80(float64(variance), 1))
	out.X = target.PosVec.X
	out.Y = target.PosVec.Y
	recoil := rt.RandomFloat(variance, maximum)
	x := objectForceMul53_52DF80(objectForceMul53_52DF80(recoil, objectForceMemory32_52DF80(target.VelVec.X)), 6)
	out.X = experienceSpill32TowardZero(objectForceSub53_52DF80(objectForceMemory32_52DF80(out.X), x))
	y := objectForceMul53_52DF80(objectForceMul53_52DF80(recoil, objectForceMemory32_52DF80(target.VelVec.Y)), 6)
	out.Y = experienceSpill32TowardZero(objectForceSub53_52DF80(objectForceMemory32_52DF80(out.Y), y))
	spread := experienceSpill32TowardZero(objectForceAdd53_52DF80(
		objectForceMul53_52DF80(float64(variance), objectForceMemory32_52DF80(float32(0.8))), objectForceMemory32_52DF80(float32(0.2))))
	noiseX := objectForceMul53_52DF80(rt.RandomFloat(-60, 60), float64(spread))
	out.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(noiseX, objectForceMemory32_52DF80(out.X)))
	noiseY := objectForceMul53_52DF80(rt.RandomFloat(-60, 60), float64(spread))
	out.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(noiseY, objectForceMemory32_52DF80(out.Y)))
}

// MonsterCastDirection533CC0 binds the existing direction service without
// passing a native Object through its still-retained PE32 C declaration.
func MonsterCastDirection533CC0(unit *Object, pos *types.Pointf) {
	if unit != nil {
		x := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(pos.X), objectForceMemory32_52DF80(unit.PosVec.X)))
		y := experienceSpill32TowardZero(objectForceSub53_52DF80(float64(pos.Y), objectForceMemory32_52DF80(unit.PosVec.Y)))
		// The shared DirFromVec models default nearest-even. This caller
		// reaches 509ED0/419A70 under gameplay ToZero, including the FISTP.
		angle := math.Atan2(float64(y), float64(x))
		angle = objectForceAdd53_52DF80(angle, float64(math.Float32frombits(boomDirectionTauBits509ED0)))
		angle = objectForceMul53_52DF80(angle, float64(math.Float32frombits(boomDirectionScaleBits509ED0)))
		angle = objectForceAdd53_52DF80(angle, float64(math.Float32frombits(boomDirectionHalfBits509ED0)))
		value := experienceSpill32TowardZero(angle)
		var direction Dir16
		if value >= -2147483648 && value < 2147483648 {
			direction = Dir16(uint8(int32(value)))
		}
		unit.Direction2 = direction
	}
}
