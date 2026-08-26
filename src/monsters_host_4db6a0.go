package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type monstersAllBelongToHostHooks4DB6A0 struct {
	playerByInd      func(int) *server.Player
	playerUnit       func(*server.Player) *server.Object
	saveLocationType func() uint32
	firstObject      func() *server.Object
	nextObject       func(*server.Object) *server.Object
	typeInd          func(*server.Object) uint16
	firstOwned       func(*server.Object) *server.Object
	nextOwned        func(*server.Object) *server.Object
	setOwner         func(*server.Object, *server.Object)
	class            func(*server.Object) object.Class
	isSummoned       func(*server.Object) bool
	markMonitor      func(*server.Object)
	playerIndex      func(*server.Player) byte
	reportAcquire    func(int, *server.Object)
	markMinimap      func(ntype.PlayerInd, *server.Object, uint32)
	clearScriptID    func(*server.Object)
	delayedDelete    func(*server.Object)
}

// monstersAllBelongToHostNative4DB6A0 mirrors GAME.EXE 004DB6A0 while
// resolving Player and Object links at the host pointer width. In particular,
// the owned-list successor is loaded before ObjSetOwner rewires that list, and
// PlayerUnit/PlayerInd are reloaded at the same callback boundaries as PE32.
func monstersAllBelongToHostNative4DB6A0(h monstersAllBelongToHostHooks4DB6A0) {
	pl := h.playerByInd(server.HostPlayerIndex)
	if pl == nil || h.playerUnit(pl) == nil {
		return
	}
	typeInd := h.saveLocationType()
	loc := h.firstObject()
	for loc != nil && uint32(h.typeInd(loc)) != typeInd {
		loc = h.nextObject(loc)
	}
	if loc == nil {
		return
	}
	for it := h.firstOwned(loc); it != nil; {
		next := h.nextOwned(it)
		h.setOwner(h.playerUnit(pl), it)
		if h.class(it).Has(object.ClassMonster) && h.isSummoned(it) {
			h.markMonitor(it)
			h.reportAcquire(int(h.playerIndex(pl)), it)
			h.markMinimap(ntype.PlayerInd(h.playerIndex(pl)), it, 1)
		}
		it = next
	}
	h.clearScriptID(loc)
	h.delayedDelete(loc)
}

func (s *Server) monstersAllBelongToHost4DB6A0() {
	monstersAllBelongToHostNative4DB6A0(monstersAllBelongToHostHooks4DB6A0{
		playerByInd: func(ind int) *server.Player {
			return s.Players.ByInd(ntype.PlayerInd(ind))
		},
		playerUnit: func(pl *server.Player) *server.Object {
			return pl.PlayerUnit
		},
		saveLocationType: func() uint32 {
			cached := memmap.PtrUint32(0x5D4594, 1563124)
			if *cached == 0 {
				*cached = uint32(s.Types.IndByID("SaveGameLocation"))
			}
			return *cached
		},
		firstObject: s.Objs.First,
		nextObject: func(obj *server.Object) *server.Object {
			return obj.Next()
		},
		typeInd: func(obj *server.Object) uint16 {
			return obj.TypeInd
		},
		firstOwned: func(obj *server.Object) *server.Object {
			return obj.FirstOwned516()
		},
		nextOwned: func(obj *server.Object) *server.Object {
			return obj.NextOwned512()
		},
		setOwner: s.ObjSetOwner,
		class: func(obj *server.Object) object.Class {
			return obj.Class()
		},
		isSummoned: func(obj *server.Object) bool {
			return obj.UpdateDataMonster().StatusFlags.Has(object.MonStatusSummoned)
		},
		markMonitor: func(obj *server.Object) {
			obj.ObjSubClass |= object.SubClass(object.MonsterMonitor)
		},
		playerIndex: func(pl *server.Player) byte {
			return pl.PlayerInd
		},
		reportAcquire: legacy.Nox_xxx_netReportAcquireCreature_4D91A0,
		markMinimap: func(ind ntype.PlayerInd, obj *server.Object, flags uint32) {
			s.Players.Nox_xxx_netMarkMinimapObject_417190(ind, obj, flags)
		},
		clearScriptID: func(obj *server.Object) {
			obj.ScriptIDVal = 0
		},
		delayedDelete: s.DelayedDelete,
	})
}
