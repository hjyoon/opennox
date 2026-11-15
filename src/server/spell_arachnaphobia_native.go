package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// ArachnaphobiaNativeRuntime52DC80 supplies the original type, map and world observation boundaries.
type ArachnaphobiaNativeRuntime52DC80 struct {
	LoadCache  func() uint32
	StoreCache func(uint32)
	LookupType func(string) uint32
	TraceRay   func(types.Pointf, types.Pointf, MapTraceFlags) bool
	NewObject  func(uint32) *Object
	CreateAt   func(*Object, *Object, types.Pointf)
	Inform     func(uint8, byte, int32)
}

// ArachnaphobiaNative52DC80 preserves live native object, argument and update/player links.
func ArachnaphobiaNative52DC80(owner, caster *Object, arg *SpellAcceptArg, rt ArachnaphobiaNativeRuntime52DC80) int32 {
	if rt.LoadCache() == 0 {
		rt.StoreCache(rt.LookupType("ArachnaphobiaFocus"))
	}
	from, to := caster.PosVec, arg.Pos
	if !rt.TraceRay(from, to, MapTraceFlags(9)) {
		// 0052DCD6 reads the live class AFTER trace, then follows the live
		// player-update/player pointers. There is no sound on either branch.
		if caster.ObjClass.Has(object.ClassPlayer) {
			player := (*PlayerUpdateData)(caster.UpdateData).Player
			rt.Inform(player.PlayerInd, 0, 2)
		}
		return 0
	}
	// 0052DD10 reloads the cache after trace; 0052DD22/25 read the live
	// argument coordinates after allocation. Even allocation failure is 1.
	if focus := rt.NewObject(rt.LoadCache()); focus != nil {
		rt.CreateAt(focus, owner, arg.Pos)
	}
	return 1
}
