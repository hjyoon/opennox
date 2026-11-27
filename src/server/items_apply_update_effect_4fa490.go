package server

import "unsafe"

// ItemsApplyUpdateEffect preserves GAME.EXE 004FA490. Each callback receives
// the equipped item, not its owner. The next item and all modifier slots are
// live after callbacks; the current item's modifier-array pointer is cached.
func (s *Server) ItemsApplyUpdateEffect(owner *Object) {
	itemsApplyUpdateEffect4FA490(owner, callModifierUpdateEffect)
}

func itemsApplyUpdateEffect4FA490(owner *Object, invoke func(unsafe.Pointer, *ModifierEff, *Object)) {
	for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
		if uint32(item.ObjFlags)&0x100 == 0 || uint32(item.ObjClass)&0x13001000 == 0 {
			continue
		}
		data := item.InitDataModifier()
		for i := 0; i < 4; i++ {
			mod := data.Modifiers[i]
			if mod != nil {
				fn := mod.Update100.Fnc
				if fn != nil {
					invoke(fn, mod, item)
				}
			}
		}
	}
}
