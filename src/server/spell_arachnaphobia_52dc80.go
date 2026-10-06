package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
)

type arachnaphobiaCastDeps52DC80 struct {
	loadCache  func() uint32
	storeCache func(uint32)
	lookupType func(string) uint32
	traceRay   func(types.Pointf, types.Pointf, MapTraceFlags) bool
	newObject  func(uint32) *Object
	createAt   func(*Object, *Object, types.Pointf)
	inform     func(uint8, byte, int32)
}

// arachnaphobiaCast52DC80 restores GAME.EXE 0052DC80..0052DD44. The
// type cache and notification retain DWORD/BYTE widths; caster, owner,
// argument and Player pointers must not pass through the five-int C entry.
// The original does not use spell ID, second object, argument Obj or level.
func arachnaphobiaCast52DC80(owner, caster *Object, arg *SpellAcceptArg, h arachnaphobiaCastDeps52DC80) int32 {
	if h.loadCache() == 0 {
		h.storeCache(h.lookupType("ArachnaphobiaFocus"))
	}
	from, to := caster.PosVec, arg.Pos
	if !h.traceRay(from, to, MapTraceFlags(9)) {
		// 0052DCD6 reads the live class AFTER trace, then follows the live
		// player-update/player pointers. There is no sound on either branch.
		if caster.ObjClass.Has(object.ClassPlayer) {
			player := (*PlayerUpdateData)(caster.UpdateData).Player
			h.inform(player.PlayerInd, 0, 2)
		}
		return 0
	}
	// 0052DD10 reloads the cache after trace; 0052DD22/25 read the live
	// argument coordinates after allocation. Even allocation failure is 1.
	if focus := h.newObject(h.loadCache()); focus != nil {
		h.createAt(focus, owner, arg.Pos)
	}
	return 1
}

type ArachnaphobiaCastRuntime52DC80 struct {
	TypeCache *uint32
	CreateAt  func(*Object, *Object, types.Pointf)
}

func (s *Server) CastArachnaphobia52DC80(_ int32, _ *Object, owner, caster *Object, arg *SpellAcceptArg, _ int32, runtime ArachnaphobiaCastRuntime52DC80) int32 {
	return arachnaphobiaCast52DC80(owner, caster, arg, arachnaphobiaCastDeps52DC80{
		loadCache:  func() uint32 { return *runtime.TypeCache },
		storeCache: func(kind uint32) { *runtime.TypeCache = kind },
		lookupType: func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		traceRay:   s.MapTraceRay,
		newObject:  func(kind uint32) *Object { return s.NewObjectByTypeInd(int(int32(kind))) },
		createAt:   runtime.CreateAt,
		inform: func(index uint8, code byte, value int32) {
			s.NetInformTextMsg(ntype.PlayerInd(index), code, int(value))
		},
	})
}
