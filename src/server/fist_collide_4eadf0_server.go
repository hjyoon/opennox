package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// FistUpdateData is the fixed-width four-byte update record read by
// FistCollide and initialized by the Fist spell creator.
type FistUpdateData struct {
	Damage int32
}

type fistCollideNativeDeps4EADF0 struct {
	findParentPlayer func(*Object) *Object
	callTargetDamage func(unsafe.Pointer, *Object, *Object, *Object, int32, object.DamageType) int32
}

func fistCollideNative4EADF0(
	source, target *Object,
	collision *types.Pointf,
	deps fistCollideNativeDeps4EADF0,
) {
	fistCollide4EADF0(
		source,
		target,
		collision,
		fistCollideHooks4EADF0[*Object, *FistUpdateData, unsafe.Pointer]{
			loadZ: func(obj *Object) float32 {
				return obj.ZVal
			},
			loadUpdateData: func(obj *Object) *FistUpdateData {
				return (*FistUpdateData)(obj.UpdateData)
			},
			loadDamage: func(data *FistUpdateData) int32 {
				return data.Damage
			},
			findParentPlayer: deps.findParentPlayer,
			loadTargetDamage: func(obj *Object) unsafe.Pointer {
				return obj.Damage
			},
			callTargetDamage: func(
				fn unsafe.Pointer,
				target, parent, source *Object,
				damage int32,
				damageType uint32,
			) int32 {
				return deps.callTargetDamage(
					fn,
					target,
					parent,
					source,
					damage,
					object.DamageType(damageType),
				)
			},
		},
	)
}

func fistCollideRuntimeDeps4EADF0() fistCollideNativeDeps4EADF0 {
	return fistCollideNativeDeps4EADF0{
		findParentPlayer: (*Object).FindOwnerChainPlayer,
		callTargetDamage: func(
			fn unsafe.Pointer,
			target, parent, source *Object,
			damage int32,
			damageType object.DamageType,
		) int32 {
			return int32(ccall.CallIntUPtr5(
				fn,
				uintptr(target.CObj()),
				uintptr(toObjectC(parent)),
				uintptr(source.CObj()),
				uintptr(uint32(damage)),
				uintptr(uint32(damageType)),
			))
		},
	}
}

// FistCollide4EADF0 binds the original registered callback to native-width
// Object and collision pointers. The collision point is retained only for the
// callback ABI; GAME.EXE does not read it.
func (*Server) FistCollide4EADF0(
	source, target *Object,
	collision *types.Pointf,
) {
	fistCollideNative4EADF0(source, target, collision, fistCollideRuntimeDeps4EADF0())
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(FistUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(FistUpdateData{}.Damage)]
)
