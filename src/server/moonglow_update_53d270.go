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
	owner := visual.ObjOwner
	if owner == nil {
		runtime.DelayedDelete(visual)
		return
	}
	if uint8(owner.ObjClass)&4 == 0 {
		return
	}
	if uint8(owner.ObjFlags)&0x20 != 0 || runtime.Frame()-visual.Field32 > 300*runtime.FPS() {
		runtime.DelayedDelete(visual)
		runtime.BuffOff(owner, ENCHANT_MOONGLOW)
		return
	}
	player := owner.ControllingPlayer()
	point := types.Pointf{X: float32(int32(player.CursorVec.X)), Y: float32(int32(player.CursorVec.Y))}
	if runtime.ValidPosition(point) {
		runtime.Move(visual, point)
	}
}
