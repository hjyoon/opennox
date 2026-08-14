package legacy

/*
#include "chakram_collide_4eaf00.h"
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

//export nox_xxx_collideChakram_4EAF00
func nox_xxx_collideChakram_4EAF00(
	source, target *C.nox_object_t,
	collision *C.float,
) {
	srv := GetServer()
	srv.S().ChakramInMotionCollide4EAF00(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.ChakramCollideRuntime4EAF00{
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
			MoveUpdate:    Nox_xxx_moveUpdateSpecial_517970,
			DetachInventory: func(owner, item *server.Object) {
				Sub_4ED0C0(owner, item)
			},
			InventoryPut: func(owner, item *server.Object, mode uint32) {
				Nox_xxx_inventoryPutImpl_4F3070(owner, item, int(mode))
			},
			EquipWeapon: func(owner, item *server.Object, first, second uint32) {
				C.nox_xxx_playerEquipWeapon_53A420(
					(*C.uint32_t)(owner.CObj()),
					(*C.nox_object_t)(item.CObj()),
					C.int(first),
					C.int(second),
				)
			},
			ApplyAttackEffect: func(source, owner *server.Object, attack *server.ChakramAttackData) {
				ccall.CallIntUPtr3(
					C.nox_xxx_itemApplyAttackEffect_538840,
					uintptr(source.CObj()),
					uintptr(owner.CObj()),
					uintptr(unsafe.Pointer(attack)),
				)
			},
			PreAttackEffects: func(target, owner, source *server.Object, attack *server.ChakramAttackData) {
				ccall.CallIntUPtr4(
					C.nox_xxx_playerPreAttackEffects_538290,
					uintptr(target.CObj()),
					uintptr(owner.CObj()),
					uintptr(source.CObj()),
					uintptr(unsafe.Pointer(attack)),
				)
			},
			CreateAt: func(item, owner *server.Object, pos types.Pointf) {
				srv.CreateObjectAt(item, owner, pos)
			},
		},
	)
}
