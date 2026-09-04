package legacy

import (
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

const (
	weaponRollBlockedStatus10001EE0 = uint32(0x03)
	weaponRollBlockedState10001EE0  = server.PlayerState1
	weaponRollQuiverFlag10001EE0    = uint32(0x02)
)

type weaponRollHooks10001EE0[O, D, P comparable] struct {
	loadUpdate       func(O) D
	loadPlayer       func(D) P
	loadPlayerStatus func(P) uint32
	loadState        func(D) server.PlayerState
	loadEquipped     func(D) O
	loadFirstItem    func(O) O
	loadNextItem     func(O) O
	loadPrevItem     func(O) O
	loadWeaponFlags  func(O) uint32
	loadPlayerClass  func(P) uint8
	classCanUse      func(O, uint8) bool
	checkStrength    func(O, O) bool
	tryDequip        func(O, O) bool
	tryEquip         func(O, O) bool
}

// weaponRoll10001EE0 preserves GameEx.dll 10001EE0 as transcribed in the
// historical GameEx.c. This routine is not part of the GAME.EXE oracle.
//
// UpdateData is cached for the entry gates and equipped-weapon load. The
// original reloads Object.UpdateData and UpdateData.Player for every candidate
// that reaches the class check, so the model does too. A nonzero direction
// follows InvNextItem; zero follows Field125. If no weapon is equipped, the
// direction is ignored and traversal starts at InvFirstItem using InvNextItem.
//
// Flags zero and two are skipped. Once a class-compatible, strong-enough item
// is found, the routine returns immediately even if dequip/equip fails. Equip
// is short-circuited when dequip fails. Results are canonical zero or one.
func weaponRoll10001EE0[O, D, P comparable](
	playerObject O,
	direction int8,
	hooks weaponRollHooks10001EE0[O, D, P],
) int32 {
	update := hooks.loadUpdate(playerObject)
	playerData := hooks.loadPlayer(update)
	if hooks.loadPlayerStatus(playerData)&weaponRollBlockedStatus10001EE0 != 0 {
		return 0
	}
	if hooks.loadState(update) == weaponRollBlockedState10001EE0 {
		return 0
	}

	var zeroObject O
	current := hooks.loadEquipped(update)
	if current != zeroObject {
		candidate := current
		for {
			if direction != 0 {
				candidate = hooks.loadNextItem(candidate)
			} else {
				candidate = hooks.loadPrevItem(candidate)
			}
			if candidate == zeroObject {
				return 0
			}
			if weaponRollCandidate10001EE0(playerObject, candidate, hooks) {
				if hooks.tryDequip(playerObject, current) && hooks.tryEquip(playerObject, candidate) {
					return 1
				}
				return 0
			}
		}
	}

	for candidate := hooks.loadFirstItem(playerObject); candidate != zeroObject; candidate = hooks.loadNextItem(candidate) {
		if weaponRollCandidate10001EE0(playerObject, candidate, hooks) {
			if hooks.tryEquip(playerObject, candidate) {
				return 1
			}
			return 0
		}
	}
	return 0
}

func weaponRollCandidate10001EE0[O, D, P comparable](
	playerObject O,
	candidate O,
	hooks weaponRollHooks10001EE0[O, D, P],
) bool {
	flags := hooks.loadWeaponFlags(candidate)
	if flags == 0 || flags == weaponRollQuiverFlag10001EE0 {
		return false
	}
	update := hooks.loadUpdate(playerObject)
	playerData := hooks.loadPlayer(update)
	class := hooks.loadPlayerClass(playerData)
	return hooks.classCanUse(candidate, class) && hooks.checkStrength(playerObject, candidate)
}

type weaponRollNativeDeps10001EE0 struct {
	loadWeaponFlags func(*server.Object) uint32
	classCanUse     func(*server.Object, player.Class) bool
	checkStrength   func(*server.Object, *server.Object) bool
	tryDequip       func(*server.Object, *server.Object) bool
	tryEquip        func(*server.Object, *server.Object) bool
}

// weaponRollNative10001EE0 is the native-width binding for GameEx.dll
// 10001EE0. Object, update-data, player, and inventory links stay as Go
// pointers instead of being narrowed through the PE32 layout.
func weaponRollNative10001EE0(
	playerObject *server.Object,
	direction int8,
	deps weaponRollNativeDeps10001EE0,
) int32 {
	return weaponRoll10001EE0(playerObject, direction, weaponRollHooks10001EE0[
		*server.Object,
		*server.PlayerUpdateData,
		*server.Player,
	]{
		loadUpdate: func(object *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(object.UpdateData)
		},
		loadPlayer: func(update *server.PlayerUpdateData) *server.Player {
			return update.Player
		},
		loadPlayerStatus: func(playerData *server.Player) uint32 {
			return playerData.Field3680
		},
		loadState: func(update *server.PlayerUpdateData) server.PlayerState {
			return update.State
		},
		loadEquipped: func(update *server.PlayerUpdateData) *server.Object {
			return update.EquippedWeapon
		},
		loadFirstItem: func(object *server.Object) *server.Object {
			return object.InvFirstItem
		},
		loadNextItem: func(object *server.Object) *server.Object {
			return object.InvNextItem
		},
		loadPrevItem: func(object *server.Object) *server.Object {
			return object.Field125
		},
		loadWeaponFlags: deps.loadWeaponFlags,
		loadPlayerClass: func(playerData *server.Player) uint8 {
			// PlayerClass is nil-safe elsewhere in the port, unlike the direct
			// Player+2251 byte load in GameEx. Preserve that fault boundary.
			if playerData == nil {
				panic("10001EE0: nil Player")
			}
			return uint8(playerData.PlayerClass())
		},
		classCanUse: func(object *server.Object, class uint8) bool {
			return deps.classCanUse(object, player.Class(class))
		},
		checkStrength: deps.checkStrength,
		tryDequip:     deps.tryDequip,
		tryEquip:      deps.tryEquip,
	})
}
