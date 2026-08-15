package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// UndeadKillerCollideData is the native-width form of the original four-byte
// record. TurnUndead stores the duration-spell pointer in this sole field.
type UndeadKillerCollideData struct {
	Spell *DurSpell
}

// UndeadKillerCollideRuntime4EBD40 supplies object deletion, which remains
// owned by the legacy/root runtime. All collision, spell, health, owner and
// damage fields are read through native-width server layouts.
type UndeadKillerCollideRuntime4EBD40 struct {
	DelayedDelete func(*Object)
}

type undeadKillerCollideNativeDeps4EBD40 struct {
	findParentPlayer func(*Object) *Object
	callTargetDamage func(unsafe.Pointer, *Object, *Object, *Object, int32, object.DamageType) int32
	delayedDelete    func(*Object)
}

func undeadKillerCollideNative4EBD40(
	source, target *Object,
	collision *types.Pointf,
	deps undeadKillerCollideNativeDeps4EBD40,
) {
	undeadKillerCollide4EBD40(
		source,
		target,
		collision,
		undeadKillerCollideHooks4EBD40[
			*Object,
			*UndeadKillerCollideData,
			*DurSpell,
			unsafe.Pointer,
		]{
			loadClassLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			loadSubclassLow: func(obj *Object) uint8 {
				return uint8(obj.ObjSubClass)
			},
			loadCollideData: func(obj *Object) *UndeadKillerCollideData {
				return (*UndeadKillerCollideData)(obj.CollideData)
			},
			loadHP: func(obj *Object) uint16 {
				if obj == nil || obj.HealthData == nil {
					return 0
				}
				return obj.HealthData.Cur
			},
			loadSpell: func(data *UndeadKillerCollideData) *DurSpell {
				return data.Spell
			},
			loadRemaining: func(spell *DurSpell) int32 {
				return spell.Field72
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
			delayedDelete: deps.delayedDelete,
			storeRemaining: func(spell *DurSpell, remaining int32) {
				spell.Field72 = remaining
			},
		},
	)
}

func undeadKillerCollideServerDeps4EBD40(
	runtime UndeadKillerCollideRuntime4EBD40,
) undeadKillerCollideNativeDeps4EBD40 {
	return undeadKillerCollideNativeDeps4EBD40{
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
		delayedDelete: runtime.DelayedDelete,
	}
}

// UndeadKillerCollide4EBD40 binds the registered callback to native-width
// Object, collision-data and duration-spell layouts. The collision point is
// only tested for nil when target is nil, exactly as in GAME.EXE.
func (*Server) UndeadKillerCollide4EBD40(
	source, target *Object,
	collision *types.Pointf,
	runtime UndeadKillerCollideRuntime4EBD40,
) {
	undeadKillerCollideNative4EBD40(
		source,
		target,
		collision,
		undeadKillerCollideServerDeps4EBD40(runtime),
	)
}
