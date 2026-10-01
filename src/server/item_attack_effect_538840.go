package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// ItemApplyAttackEffect538840 preserves GAME.EXE 00538840: cache the
// modifier array once, then reload each of its four entries after callbacks.
// The target argument is nil; context is the native temporary attack record.
func (s *Server) ItemApplyAttackEffect538840(item, owner *Object, context unsafe.Pointer) int32 {
	return itemApplyAttackEffect538840(item, owner, context, func(fn unsafe.Pointer, mod *ModifierEff, item, owner *Object, context unsafe.Pointer) {
		ccall.CallVoidPtr5(fn, mod.C(), item.CObj(), owner.CObj(), nil, context)
	})
}

func itemApplyAttackEffect538840(item, owner *Object, context unsafe.Pointer,
	invoke func(unsafe.Pointer, *ModifierEff, *Object, *Object, unsafe.Pointer),
) int32 {
	data := item.InitDataModifier()
	for i := 0; i < 4; i++ {
		mod := data.Modifiers[i]
		if mod != nil {
			fn := mod.Attack40.Fnc
			if fn != nil {
				invoke(fn, mod, item, owner, context)
			}
		}
	}
	return 0
}
