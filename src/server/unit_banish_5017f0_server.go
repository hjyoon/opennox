package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
)

// UnitBanishRuntime5017F0 supplies services owned by the legacy root. Object
// pointers remain native-width until the callbacks receive them.
type UnitBanishRuntime5017F0 struct {
	DelayedDelete func(*Object)
	SendPointFX   func(netmsg.Op, types.Pointf)
}

type unitBanishNativeDeps5017F0 struct {
	loadGlyphCache  func() uint32
	lookupType      func(string) uint32
	storeGlyphCache func(uint32)
	delayedDelete   func(*Object)
	sendPointFX     func(uint8, *types.Pointf)
}

func unitBanishNative5017F0(unit *Object, deps unitBanishNativeDeps5017F0) {
	unitBanish5017F0(unit, unitBanishHooks5017F0[*Object]{
		loadGlyphCache:  deps.loadGlyphCache,
		lookupType:      deps.lookupType,
		storeGlyphCache: deps.storeGlyphCache,
		loadFirst: func(unit *Object) *Object {
			return unit.InvFirstItem
		},
		loadNext: func(item *Object) *Object {
			return item.InvNextItem
		},
		loadTypeIndex: func(item *Object) uint16 {
			return item.TypeInd
		},
		delayedDelete: deps.delayedDelete,
		sendPointFX: func(code uint8, unit *Object) {
			deps.sendPointFX(code, &unit.PosVec)
		},
	})
}

func unitBanishServerDeps5017F0(s *Server, runtime UnitBanishRuntime5017F0) unitBanishNativeDeps5017F0 {
	return unitBanishNativeDeps5017F0{
		loadGlyphCache: s.Types.banishGlyphIDCached5017F0,
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeGlyphCache: s.Types.storeBanishGlyphID5017F0,
		delayedDelete:   runtime.DelayedDelete,
		sendPointFX: func(code uint8, position *types.Pointf) {
			runtime.SendPointFX(netmsg.Op(code), *position)
		},
	}
}

// BanishUnit5017F0 binds GAME.EXE 005017F0 to native-width Object inventory
// links and the function's dedicated fixed-width Glyph type cache.
func (s *Server) BanishUnit5017F0(unit *Object, runtime UnitBanishRuntime5017F0) {
	unitBanishNative5017F0(unit, unitBanishServerDeps5017F0(s, runtime))
}
