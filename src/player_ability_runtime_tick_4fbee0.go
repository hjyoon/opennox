package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

const (
	playerAbilityRuntimeWarrior4FBEE0      = uint8(0)
	playerAbilityRuntimeReadyState4FBEE0   = uint8(1)
	playerAbilityRuntimeEndingSound4FBEE0  = int32(2)
	playerAbilityRuntimeInactive4FBEE0     = uint8(0)
	playerAbilityRuntimeBerserkState4FBEE0 = uint8(13)
)

type playerAbilityRuntimeTickHooks4FBEE0[
	P comparable,
	U comparable,
	R comparable,
] struct {
	firstPlayer     func() P
	nextPlayer      func(P) P
	loadPlayerUnit  func(P) U
	loadPlayerClass func(P) uint8
	loadPlayerIndex func(P) uint8
	loadCooldown    func(uint8, server.Ability) int32
	storeCooldown   func(uint8, server.Ability, int32)
	reportState     func(U, server.Ability, uint8)
	loadExecHead    func() R
	storeExecHead   func(R)
	loadExecUnit    func(R) U
	loadExecAbility func(R) server.Ability
	loadExecFrame   func(R) uint32
	loadExecNext    func(R) R
	loadExecPrev    func(R) R
	storeExecNext   func(R, R)
	storeExecPrev   func(R, R)
	loadUnitFlags   func(U) object.Flags
	loadFrame       func() uint32
	loadEndingSound func(server.Ability, int32) int32
	audio           func(int32, U, int32, uint32)
	reportActive    func(U, server.Ability, uint8)
	setPlayerState  func(U, uint8)
	freeExec        func(R)
}

// playerAbilityRuntimeTick4FBEE0 preserves GAME.EXE 004FBEE0. Cooldowns are
// signed 32-bit cells indexed by the live PlayerInd byte. The executable
// reloads that index after each decrement before deciding whether to report a
// ready ability, and reloads PlayerUnit only for that report. The next Player
// is obtained after all callbacks for the current one.
//
// Active records form one global doubly linked list. Unit and successor are
// cached immediately on arrival. Expiration callbacks deliberately reload the
// record's ability and unit, unlinking deliberately reloads its live links,
// while traversal continues through the successor cached before any callback.
// Frame comparison is unsigned and removes a live record only when current
// frame is strictly greater than its deadline. No defensive nil checks are
// added beyond the PlayerUnit check present in the executable.
func playerAbilityRuntimeTick4FBEE0[
	P comparable,
	U comparable,
	R comparable,
](hooks playerAbilityRuntimeTickHooks4FBEE0[P, U, R]) {
	var zeroPlayer P
	var zeroUnit U
	for player := hooks.firstPlayer(); player != zeroPlayer; player = hooks.nextPlayer(player) {
		unit := hooks.loadPlayerUnit(player)
		if unit == zeroUnit || hooks.loadPlayerClass(player) != playerAbilityRuntimeWarrior4FBEE0 {
			continue
		}
		for ability := server.AbilityInvalid; ability < server.AbilityMax; ability++ {
			index := hooks.loadPlayerIndex(player)
			cooldown := hooks.loadCooldown(index, ability)
			if cooldown == 0 {
				continue
			}
			cooldown--
			hooks.storeCooldown(index, ability, cooldown)

			index = hooks.loadPlayerIndex(player)
			if hooks.loadCooldown(index, ability) == 0 {
				unit = hooks.loadPlayerUnit(player)
				hooks.reportState(unit, ability, playerAbilityRuntimeReadyState4FBEE0)
			}
		}
	}

	var zeroExec R
	for exec := hooks.loadExecHead(); exec != zeroExec; {
		unit := hooks.loadExecUnit(exec)
		cachedNext := hooks.loadExecNext(exec)
		remove := hooks.loadUnitFlags(unit).HasAny(object.FlagDestroyed | object.FlagDead)
		if !remove {
			frame := hooks.loadFrame()
			deadline := hooks.loadExecFrame(exec)
			if frame <= deadline {
				exec = cachedNext
				continue
			}

			ability := hooks.loadExecAbility(exec)
			soundID := hooks.loadEndingSound(ability, playerAbilityRuntimeEndingSound4FBEE0)
			hooks.audio(soundID, unit, 0, 0)

			ability = hooks.loadExecAbility(exec)
			unit = hooks.loadExecUnit(exec)
			hooks.reportActive(unit, ability, playerAbilityRuntimeInactive4FBEE0)

			ability = hooks.loadExecAbility(exec)
			if ability == server.AbilityBerserk {
				unit = hooks.loadExecUnit(exec)
				hooks.setPlayerState(unit, playerAbilityRuntimeBerserkState4FBEE0)
			}
		}

		liveNext := hooks.loadExecNext(exec)
		if liveNext != zeroExec {
			livePrev := hooks.loadExecPrev(exec)
			hooks.storeExecPrev(liveNext, livePrev)
		}
		livePrev := hooks.loadExecPrev(exec)
		if livePrev != zeroExec {
			liveNext = hooks.loadExecNext(exec)
			hooks.storeExecNext(livePrev, liveNext)
		} else {
			liveNext = hooks.loadExecNext(exec)
			hooks.storeExecHead(liveNext)
		}
		hooks.freeExec(exec)
		exec = cachedNext
	}
}
