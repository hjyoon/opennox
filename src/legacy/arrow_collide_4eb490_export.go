package legacy

/*
#include "arrow_collide_4eb490.h"
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_collideArrow_4EB490
func nox_xxx_collideArrow_4EB490(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().ArrowCollide4EB490(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.ArrowCollideRuntime4EB490{
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
			ApplyAttackEffect: func(source, owner *server.Object, attack *server.ArrowAttackData) {
				ccall.CallIntUPtr3(
					C.nox_xxx_itemApplyAttackEffect_538840,
					uintptr(source.CObj()),
					uintptr(owner.CObj()),
					uintptr(unsafe.Pointer(attack)),
				)
			},
			PreAttackEffects: func(target, owner, source *server.Object, attack *server.ArrowAttackData) {
				ccall.CallIntUPtr4(
					C.nox_xxx_playerPreAttackEffects_538290,
					uintptr(target.CObj()),
					uintptr(owner.CObj()),
					uintptr(source.CObj()),
					uintptr(unsafe.Pointer(attack)),
				)
			},
		},
	)
}

//export nox_server_arrowCollideDataSetOwner_4EB490
func nox_server_arrowCollideDataSetOwner_4EB490(source, owner *C.nox_object_t) {
	obj := asObjectS((*nox_object_t)(source))
	data := (*server.ArrowCollideData)(obj.CollideData)
	data.Owner = asObjectS((*nox_object_t)(owner))
}
