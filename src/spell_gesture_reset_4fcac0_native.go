package opennox

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellGestureResetNativeDeps4FCAC0 struct {
	resetDurations       func(int32)
	loadMagicClass       func() alloc.ClassT[server.MagicEntityClass]
	freeAllMagicObjects  func(alloc.ClassT[server.MagicEntityClass])
	clearMagicEntityHead func()
	firstPlayerUnit      func() *server.Object
	loadPlayerUpdate     func(*server.Object) *server.PlayerUpdateData
	nextPlayerUnit       func(*server.Object) *server.Object
	newObjectByTypeID    func(string) *server.Object
	storeImaginaryCaster func(*server.Object)
	createObjectAt       func(object, owner *server.Object, pos types.Pointf)
}

func spellGestureResetNative4FCAC0(
	a1, a2 int32,
	deps spellGestureResetNativeDeps4FCAC0,
) int32 {
	return spellGestureReset4FCAC0(
		a1,
		a2,
		spellGestureResetHooks4FCAC0[
			alloc.ClassT[server.MagicEntityClass],
			*server.Object,
			*server.PlayerUpdateData,
			*server.Object,
		]{
			resetDurations:       deps.resetDurations,
			loadMagicClass:       deps.loadMagicClass,
			freeAllMagicObjects:  deps.freeAllMagicObjects,
			clearMagicEntityHead: deps.clearMagicEntityHead,
			firstPlayerUnit:      deps.firstPlayerUnit,
			loadPlayerUpdate:     deps.loadPlayerUpdate,
			storeField47LowByte: func(update *server.PlayerUpdateData, value uint8) {
				update.Field47_0 = value
			},
			storeSpellCastStart: func(update *server.PlayerUpdateData, value uint32) {
				update.SpellCastStart = value
			},
			storeTrapSpell: func(update *server.PlayerUpdateData, index int, value uint32) {
				update.TrapSpells[index] = value
			},
			storeTrapSpellCountLowByte: func(update *server.PlayerUpdateData, value uint8) {
				update.TrapSpellsCnt = update.TrapSpellsCnt&^0xff | uint32(value)
			},
			nextPlayerUnit:       deps.nextPlayerUnit,
			newObjectByTypeID:    deps.newObjectByTypeID,
			storeImaginaryCaster: deps.storeImaginaryCaster,
			createObjectAt: func(object, owner *server.Object, x, y float32) {
				deps.createObjectAt(object, owner, types.Pointf{X: x, Y: y})
			},
		},
	)
}

// nox_xxx_Fn_4FCAC0 is the sole active native-width spell-gesture reset.
// Its only decoded GAME.EXE caller is the Go-owned map-switch cleanup, so the
// obsolete C body and CGo entrypoint are intentionally absent.
func nox_xxx_Fn_4FCAC0(a1, a2 int32) int32 {
	s := noxServer
	return spellGestureResetNative4FCAC0(a1, a2, spellGestureResetNativeDeps4FCAC0{
		resetDurations: func(value int32) {
			s.Spells.Dur.SpellResetDurations4FE8A0(value)
		},
		loadMagicClass: func() alloc.ClassT[server.MagicEntityClass] {
			return magicEntityAlloc
		},
		freeAllMagicObjects: func(value alloc.ClassT[server.MagicEntityClass]) {
			value.FreeAllObjects()
		},
		clearMagicEntityHead: func() {
			magicEntityHead = nil
		},
		firstPlayerUnit: s.Players.FirstUnit,
		loadPlayerUpdate: func(unit *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(unit.UpdateData)
		},
		nextPlayerUnit:    s.Players.NextUnit,
		newObjectByTypeID: s.NewObjectByTypeID,
		storeImaginaryCaster: func(value *server.Object) {
			nox_xxx_imagCasterUnit_1569664 = value
		},
		createObjectAt: func(object, owner *server.Object, pos types.Pointf) {
			s.CreateObjectAt(object, owner, pos)
		},
	})
}
