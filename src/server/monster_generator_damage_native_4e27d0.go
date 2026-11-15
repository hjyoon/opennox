package server

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
)

// MonsterGeneratorDamageNativeRuntime4E27D0 retains the external services of
// the raw entry. The default damage result keeps its complete signed dword.
type MonsterGeneratorDamageNativeRuntime4E27D0 struct {
	Frame       func() uint32
	PointFX     func(byte, byte, *types.Pointf)
	Audio       func(int32, *Object)
	GetHP       func(*Object) uint16
	Default     func(*Object, *Object, *Object, int32, int32) int32
	Script      func(*ScriptCallback, *Object, *Object, ScriptEventType)
	SetStatus   func(*Object, uint32)
	UnsetStatus func(*Object, uint32)
}

// SendMonsterGeneratorDamagePointFX4E27D0 retains the existing public C point
// effect's signed-DWORD truncation. The older typed packet API is independent.
func (s *Server) SendMonsterGeneratorDamagePointFX4E27D0(op, subtype byte, point *types.Pointf) {
	var packet [6]byte
	packet[0], packet[1] = op, subtype
	binary.LittleEndian.PutUint16(packet[2:4], uint16(netFXTruncateSignedDWORD419A70(point.X)))
	binary.LittleEndian.PutUint16(packet[4:6], uint16(netFXTruncateSignedDWORD419A70(point.Y)))
	s.Nox_xxx_netSendFxAllCli_523030(*point, packet[:])
}

func monsterGeneratorNativeNormalize509F20(point *types.Pointf) {
	y, x := point.Y, point.X
	xSquared := objectForceMul53_52DF80(float64(x), float64(x))
	ySquared := objectForceMul53_52DF80(float64(y), float64(y))
	length := monsterMoveForceSqrtChop53_50D50C(objectForceAdd53_52DF80(ySquared, xSquared))
	length32 := experienceSpill32TowardZero(length)
	point.X = experienceSpill32TowardZero(objectForceDiv53_52DF80(float64(point.X), float64(length32)))
	point.Y = experienceSpill32TowardZero(objectForceDiv53_52DF80(float64(point.Y), float64(length32)))
}

func monsterGeneratorNativeThreshold4E27D0(maximum uint16, factor uint32) int32 {
	value := experienceSpill32TowardZero(objectForceMul53_52DF80(float64(maximum), float64(math.Float32frombits(factor))))
	return int32(value)
}

// MonsterGeneratorDamageNative4E27D0 caches the entry-time update record,
// preserves live health/status reloads across services, and uses gameplay
// PC53/ToZero for each original operation, binary32 spill and conversion.
func MonsterGeneratorDamageNative4E27D0(unit, source, weapon *Object, damage, typ int32, rt MonsterGeneratorDamageNativeRuntime4E27D0) int32 {
	flags := uint32(unit.ObjFlags)
	update := (*MonsterGenUpdateData)(unit.UpdateData)
	if flags&0x8020 != 0 {
		return 0
	}
	frame := rt.Frame()
	if frame-unit.Frame134 > 20 || frame%30 == 0 {
		if weapon == nil {
			rt.PointFX(0xf0, 26, &unit.PosVec)
		} else {
			epsilon := float64(math.Float32frombits(0x3c23d70a))
			var point types.Pointf
			point.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(
				objectForceSub53_52DF80(float64(weapon.PosVec.X), objectForceMemory32_52DF80(unit.PosVec.X)), epsilon))
			point.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(
				objectForceSub53_52DF80(float64(weapon.PosVec.Y), objectForceMemory32_52DF80(unit.PosVec.Y)), epsilon))
			monsterGeneratorNativeNormalize509F20(&point)
			point.X = experienceSpill32TowardZero(objectForceAdd53_52DF80(
				objectForceMul53_52DF80(float64(point.X), 22), objectForceMemory32_52DF80(unit.PosVec.X)))
			point.Y = experienceSpill32TowardZero(objectForceAdd53_52DF80(
				objectForceMul53_52DF80(float64(point.Y), 22), objectForceMemory32_52DF80(unit.PosVec.Y)))
			rt.PointFX(0xf0, 26, &point)
		}
		rt.Audio(1001, unit)
	}
	oldHP := rt.GetHP(unit)
	result := rt.Default(unit, source, weapon, damage, typ)
	if rt.GetHP(unit) < oldHP {
		// The original three-argument script service is bound to the existing
		// native event API, which explicitly supplies GeneratorDamage (23).
		rt.Script((*ScriptCallback)(unsafe.Pointer(&update.Field48)), source, unit, NoxEventGeneratorDamage)
	}
	if uint32(unit.ObjFlags)&0x8020 != 0 {
		return result
	}
	health := unit.HealthData
	maximum, current := health.Max, health.Cur
	if int32(current) <= monsterGeneratorNativeThreshold4E27D0(maximum, 0x3eaa7efa) {
		if unit.Field5&0x100 != 0 {
			rt.UnsetStatus(unit, 0x100)
		}
		if unit.Field5&0x200 == 0 {
			rt.SetStatus(unit, 0x200)
		}
	} else if int32(current) <= monsterGeneratorNativeThreshold4E27D0(unit.HealthData.Max, 0x3f2a7efa) && unit.Field5&0x100 == 0 {
		rt.SetStatus(unit, 0x100)
	}
	return result
}
