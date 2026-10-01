package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

type preAttackEffectsRuntime538290 struct {
	gameplay func() bool
	enemy    func(*Object, *Object) bool
	buff     func(*Object, int32) int32
	invoke   func(unsafe.Pointer, *ModifierEff, *Object, *Object, *Object, unsafe.Pointer)
}

// PlayerPreAttackEffects538290 preserves the original friendly-fire and
// buff gates, then invokes only the two enchantment slots' pre-hit callbacks.
func (s *Server) PlayerPreAttackEffects538290(target, owner, item *Object, context unsafe.Pointer) int32 {
	return playerPreAttackEffects538290(target, owner, item, context, preAttackEffectsRuntime538290{
		gameplay: func() bool { return noxflags.HasGamePlay(noxflags.GameplayFlag1) },
		enemy:    s.IsEnemyTo,
		buff:     func(unit *Object, buff int32) int32 { return unit.UnitBuffTest4FF350(buff) },
		invoke: func(fn unsafe.Pointer, mod *ModifierEff, item, owner, target *Object, ctx unsafe.Pointer) {
			ccall.CallVoidPtr5(fn, mod.C(), item.CObj(), owner.CObj(), target.CObj(), ctx)
		},
	})
}

func playerPreAttackEffects538290(target, owner, item *Object, context unsafe.Pointer, r preAttackEffectsRuntime538290) int32 {
	if item == nil {
		return 0
	}
	data := item.InitDataModifier()
	if !r.gameplay() && owner != nil && uint8(owner.ObjClass)&6 != 0 && !r.enemy(owner, target) {
		return 0
	}
	if result := r.buff(target, 23); result != 0 {
		return result
	}
	if result := r.buff(target, 27); result != 0 {
		return result
	}
	for i := 2; i < 4; i++ {
		mod := data.Modifiers[i]
		if mod != nil {
			fn := mod.AttackPreHit52.Fnc
			if fn != nil {
				r.invoke(fn, mod, item, owner, target, context)
			}
		}
	}
	return 0
}
