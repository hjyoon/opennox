package opennox

import "github.com/opennox/opennox/v1/client/gui"

// nox_xxx_wnd_46C2A0 is the sole active native-width implementation of the
// GAME.EXE window hidden-ancestor predicate.
//
//go:noinline
func nox_xxx_wnd_46C2A0(win *gui.Window) int32 {
	return windowHiddenAncestor46C2A0(win, windowHiddenAncestorHooks46C2A0[*gui.Window]{
		loadFlagsLowByte: func(win *gui.Window) uint8 {
			return uint8(win.Flags)
		},
		loadParent: func(win *gui.Window) *gui.Window {
			return win.Parent()
		},
	})
}
