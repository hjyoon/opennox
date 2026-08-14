package legacy

/*
#include "monster_arrow_collide_4eb800.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideMonsterArrow_4EB800
func nox_xxx_collideMonsterArrow_4EB800(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().MonsterArrowCollide4EB800(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.MonsterArrowCollideRuntime4EB800{
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
