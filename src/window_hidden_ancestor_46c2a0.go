package opennox

type windowHiddenAncestorHooks46C2A0[Window comparable] struct {
	loadFlagsLowByte func(Window) uint8
	loadParent       func(Window) Window
}

// windowHiddenAncestor46C2A0 preserves GAME.EXE 0046C2A0's load order.
// Pointer-bearing values stay native-width tokens. A nil input is treated as
// hidden, while a nonnil chain is visible only when every low-byte flags load
// lacks the hidden bit and the final parent load returns nil.
func windowHiddenAncestor46C2A0[Window comparable](
	win Window,
	hooks windowHiddenAncestorHooks46C2A0[Window],
) int32 {
	var nilWindow Window
	if win == nilWindow {
		return 1
	}
	if hooks.loadFlagsLowByte(win)&0x10 != 0 {
		return 1
	}
	for parent := hooks.loadParent(win); parent != nilWindow; parent = hooks.loadParent(parent) {
		if hooks.loadFlagsLowByte(parent)&0x10 != 0 {
			return 1
		}
	}
	return 0
}
