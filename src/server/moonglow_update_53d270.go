package server

import "github.com/opennox/libs/types"

// MoonglowUpdateRuntime53D270 supplies the world operations used by the
// thing.bin object update callback. Object and owner identities stay native-width.
type MoonglowUpdateRuntime53D270 struct {
	Frame         func() uint32
	FPS           func() uint32
	ValidPosition func(types.Pointf) bool
	Move          func(*Object, types.Pointf)
	DelayedDelete func(*Object)
	BuffOff       func(*Object, EnchantID)
}

// MoonglowUpdate53D270 follows GAME.EXE 0053D270 without reading the PE32
// ObjOwner and Field32 offsets from a native-width object.
//
//go:noinline
func MoonglowUpdate53D270(visual *Object, runtime MoonglowUpdateRuntime53D270) {
	LifetimeEffectUpdateRaw53CC90(LifetimeEffectMoonglow53D270, visual, LifetimeEffectRuntime53CC90{
		Frame: runtime.Frame, FPS: runtime.FPS, ValidPosition: runtime.ValidPosition, DelayedDelete: runtime.DelayedDelete, BuffOff: runtime.BuffOff,
		Move: func(o *Object, p *types.Pointf) { runtime.Move(o, *p) },
	})
}
