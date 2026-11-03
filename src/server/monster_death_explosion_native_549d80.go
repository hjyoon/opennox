package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/sound"
)

// MonsterDeathExplosionNative549D80 implements the four original raw entries.
// The entry does not admit or dereference the object: the original world
// service receives its live position address, including on faulty inputs.
func (unit *Object) MonsterDeathExplosionNative549D80(entry uint32, runtime MonsterDeathExplosionRuntime549D80) int32 {
	switch entry {
	case 0x549d80:
		damage := uint32(96)
		if runtime.GameFlag(2048) != 0 {
			damage = 30
		}
		runtime.Push(unit, 96, 10, 100, unit, 0, 0)
		runtime.Damage(unit, 96, 10, damage, 7, unit, nil)
		runtime.Spark(unit, 128)
		runtime.Audio(42, unit, 0, 0)
		runtime.Delete(unit)
	case 0x549e00:
		runtime.Push(unit, 150, 10, 150, unit, 0, 0)
		runtime.Damage(unit, 150, 10, 148, 7, unit, nil)
		runtime.Spark(unit, 255)
		runtime.Audio(42, unit, 0, 0)
		runtime.Delete(unit)
	case 0x549e70, 0x54a250:
		runtime.Point(129, unit)
	default:
		panic("unsupported raw monster-death entry")
	}
	return 1
}

// Ordinary admission remains in MonsterDeadCallbackNative549D80. Each adapter
// reads the point at its own world call instead of keeping the pre-push value.
func monsterDeathOrdinaryExplosionRuntime549D80(runtime MonsterDeadCallbackRuntime549D80) MonsterDeathExplosionRuntime549D80 {
	return MonsterDeathExplosionRuntime549D80{
		GameFlag: func(uint32) uint32 {
			if runtime.CoopMode() {
				return 1
			}
			return 0
		},
		Push: func(pointOwner *Object, outer, inner, force float32, source *Object, _, _ uintptr) uint32 {
			runtime.PushUnits(pointOwner.PosVec, outer, inner, force, source)
			return 0
		},
		Damage: func(pointOwner *Object, outer, inner float32, damage, typ uint32, source, _ *Object) uint32 {
			runtime.DamageUnits(pointOwner.PosVec, outer, inner, int(int32(damage)), object.DamageType(typ), source)
			return 0
		},
		Spark: func(pointOwner *Object, code uint32) uint32 {
			runtime.SparkExplosion(pointOwner.PosVec, byte(code))
			return 0
		},
		Point: func(code uint32, pointOwner *Object) uint32 {
			runtime.PointFX(netmsg.Op(code), pointOwner.PosVec)
			return 0
		},
		Audio: func(id uint32, unit *Object, _, _ uint32) uint32 {
			runtime.Audio(sound.ID(id), unit)
			return 0
		},
		Delete: func(unit *Object) uint32 {
			runtime.DelayedDelete(unit)
			return 0
		},
	}
}
