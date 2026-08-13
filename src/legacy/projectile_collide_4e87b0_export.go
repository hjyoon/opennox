package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideProjectileGeneric_4E87B0
func nox_xxx_collideProjectileGeneric_4E87B0(
	projectile, other *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().ProjectileCollide4E87B0(
		asObjectS((*nox_object_t)(projectile)),
		asObjectS((*nox_object_t)(other)),
		unsafe.Pointer(collision),
		server.ProjectileCollideRuntime4E87B0{
			TraceHitPoint: func() *ntype.Point32 {
				if Get_dword_5d4594_2488620() == 0 {
					return nil
				}
				return memmap.PtrT[ntype.Point32](0x5D4594, 2488612)
			},
			DamageMap: func(x, y, damage int32, damageType object.DamageType, source *server.Object) {
				srv.Nox_xxx_damageToMap_534BC0(int(x), int(y), int(damage), damageType, source)
			},
			DelayedDelete: srv.DelayedDelete,
		},
	)
}
