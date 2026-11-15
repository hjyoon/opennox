package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func fireballTypeName52C790(level int32) string {
	// Original slots stay four bytes apart; PtrPtr owns native side values.
	offset := uintptr(int64(258864) + 4*int64(level))
	return GoStringP(*memmap.PtrPtr(0x587000, offset))
}

var fireballRuntimeForCall52C790 = fireballDefaultRuntime52C790

func fireballDefaultRuntime52C790() server.SpellFireballRuntime52C790 {
	return server.SpellFireballRuntime52C790{
		NewObject: func(level int32) *server.Object {
			return GetServer().S().NewObjectByTypeID(fireballTypeName52C790(level))
		},
		Direction: func(direction int16) types.Pointf {
			offset := uintptr(int64(194136) + 8*int64(direction))
			return types.Pointf{X: *memmap.PtrFloat32(0x587000, offset), Y: *memmap.PtrFloat32(0x587000, offset+4)}
		},
		Trace: func(ray *types.Rectf) bool {
			return GetServer().S().MapTraceRay(ray.Min, ray.Max, server.MapTraceFlags(5))
		},
		Place: func(projectile, owner *server.Object, pos types.Pointf) {
			GetServer().CreateObjectAt(projectile, owner, pos)
		},
		Balance: func(index int32) float64 {
			return GetServer().S().Balance.FloatInd("FireballSpeedCoeff", int(index))
		},
		Sound: func(id int32) int32 { return nox_xxx_spellGetAud44_424800(id, 0) },
		Audio: func(id int32, owner *server.Object) {
			GetServer().S().Audio.EventObj(sound.ID(id), owner, 0, 0)
		},
	}
}

func fireballCall52C790(id int32, firing *server.Object, level int32) int32 {
	return server.CastFireball52C790(id, firing, level, fireballRuntimeForCall52C790())
}
