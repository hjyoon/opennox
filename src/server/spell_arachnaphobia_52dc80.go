package server

import (
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
	return ArachnaphobiaNative52DC80(owner, caster, arg, ArachnaphobiaNativeRuntime52DC80{
		LoadCache:  h.loadCache,
		StoreCache: h.storeCache,
		LookupType: h.lookupType,
		TraceRay:   h.traceRay,
		NewObject:  h.newObject,
		CreateAt:   h.createAt,
		Inform:     h.inform,
	})
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
