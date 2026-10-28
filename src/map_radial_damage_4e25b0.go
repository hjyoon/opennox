package opennox

import (
	"image"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// MapDamageUnitsAroundNative4E25B0 is the pointer-preserving C/Go engine entry.
func (s *Server) MapDamageUnitsAroundNative4E25B0(center *types.Pointf, outer, inner float32, damage int32, typ object.DamageType, source, excluded *server.Object) {
	server.MapRadialDamageRaw4E25B0(center, outer, inner, damage, typ, source, excluded, server.MapRadialDamageRuntime4E25B0{
		DamageSource:      func() bool { return doDamageWalls },
		ResetDamageSource: func() { doDamageWalls = true },
		EachInRect: func(rect *types.Rectf, context *server.MapRadialDamageContext4E25B0, callback func(*server.Object, *server.MapRadialDamageContext4E25B0)) {
			s.Map.EachObjInRect(*rect, func(target *server.Object) bool {
				callback(target, context)
				return true
			})
		},
		Trace: func(ray *types.Rectf) int32 {
			if s.MapTraceRay(ray.Min, ray.Max, server.MapTraceFlag1) {
				return 1
			}
			return 0
		},
		Damage: func(target, source *server.Object, damage int32, typ object.DamageType) {
			target.CallDamage(source, nil, int(damage), typ)
		},
		DamageWalls: func(rect image.Rectangle, center *types.Pointf, radius float32, damage int32, typ object.DamageType, source *server.Object) {
			s.Nox_xxx_mapDamageToWalls_534FC0(rect, *center, radius, int(damage), typ, source)
		},
	})
}
