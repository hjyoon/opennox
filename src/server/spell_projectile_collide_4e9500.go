package server

const (
	spellProjectileRejectFlags4E9500     = uint32(0x00008020)
	spellProjectilePlayerClass4E9500     = uint8(0x04)
	spellProjectileBlockingState4E9500   = uint8(16)
	spellProjectileGreatSwordState4E9500 = uint8(13)
	spellProjectileGreatSwordMask4E9500  = uint32(0x00000400)
	spellProjectileReflectEnchant4E9500  = uint32(27)
)

type spellProjectileCollideHooks4E9500[O, U, PU, P comparable, C comparable] struct {
	loadProjectileUpdate func(O) U
	wallReflect          func(C, O)
	loadFlags            func(O) uint32
	loadClassLow         func(O) uint8
	loadPlayerUpdate     func(O) PU
	loadPlayerState      func(PU) uint8
	loadDirection        func(O) int16
	checkPrevious        func(O, int32, O) int32
	audio                func(uint32, O)
	projectileReflect    func(O, O)
	changeOwner          func(O, O)
	loadPlayer           func(PU) P
	loadWeaponEquip      func(P) uint32
	randomInt            func(int32, int32) int32
	setPlayerState       func(O, int32)
	mapPlayerAction      func(O) int32
	playerAnimFrames     func(int32) (int32, int32)
	storeAnimFrame       func(PU, uint8)
	checkInversion       func(O, O) int32
	hasEnchant           func(O, uint32) int32
	checkCurrent         func(O, int32, O) int32
	loadTarget           func(U) O
	loadLevel            func(U) int32
	loadOwner            func(U) O
	loadSource           func(U) O
	loadSpell            func(U) int32
	spellAccept          func(int32, O, O, O, O, int32) int32
	delayedDelete        func(O)
}

// spellProjectileCollide4E9500 preserves GAME.EXE 004E9500. In particular,
// the projectile update-data pointer is cached before the other-object test,
// the player state is reloaded for the great-sword branch, and the cached
// projectile record survives every reflect/state callback. The original
// initializes only SpellAcceptArg.Obj; adapters zero the two indeterminate
// position words instead of exposing undefined stack data.
func spellProjectileCollide4E9500[O, U, PU, P comparable, C comparable](
	projectile, other O,
	collision C,
	hooks spellProjectileCollideHooks4E9500[O, U, PU, P, C],
) {
	update := hooks.loadProjectileUpdate(projectile)
	var zeroObject O
	if other == zeroObject {
		var zeroCollision C
		if collision != zeroCollision {
			hooks.wallReflect(collision, projectile)
		}
		return
	}
	if hooks.loadFlags(other)&spellProjectileRejectFlags4E9500 != 0 {
		return
	}

	if hooks.loadClassLow(other)&spellProjectilePlayerClass4E9500 != 0 {
		playerUpdate := hooks.loadPlayerUpdate(other)
		if hooks.loadPlayerState(playerUpdate) == spellProjectileBlockingState4E9500 {
			direction := int32(hooks.loadDirection(other))
			if hooks.checkPrevious(other, direction, projectile)&1 != 0 {
				hooks.audio(878, other)
				hooks.projectileReflect(projectile, other)
				hooks.changeOwner(projectile, other)
				return
			}
		}

		if hooks.loadPlayerState(playerUpdate) == spellProjectileGreatSwordState4E9500 {
			player := hooks.loadPlayer(playerUpdate)
			if hooks.loadWeaponEquip(player)&spellProjectileGreatSwordMask4E9500 != 0 {
				direction := int32(hooks.loadDirection(other))
				if hooks.checkPrevious(other, direction, projectile)&1 != 0 {
					state := hooks.randomInt(18, 20)
					hooks.audio(890, other)
					hooks.projectileReflect(projectile, other)
					hooks.changeOwner(projectile, other)
					hooks.setPlayerState(other, state)
					action := hooks.mapPlayerAction(other)
					start, _ := hooks.playerAnimFrames(action)
					hooks.storeAnimFrame(playerUpdate, uint8(start-1))
				}
			}
		}

		if hooks.checkInversion(other, projectile) != 0 {
			hooks.changeOwner(projectile, other)
			return
		}
		if hooks.hasEnchant(other, spellProjectileReflectEnchant4E9500) != 0 {
			direction := int32(hooks.loadDirection(other))
			if hooks.checkCurrent(other, direction, projectile)&1 != 0 {
				hooks.changeOwner(projectile, other)
				hooks.audio(122, other)
				return
			}
		}
	}

	if hooks.loadTarget(update) != other {
		hooks.projectileReflect(projectile, other)
		return
	}
	level := hooks.loadLevel(update)
	owner := hooks.loadOwner(update)
	source := hooks.loadSource(update)
	spellID := hooks.loadSpell(update)
	_ = hooks.spellAccept(spellID, source, owner, projectile, other, level)
	hooks.delayedDelete(projectile)
}

type spellProjectileInversionHooks4FA4F0[O, I, D, M, F comparable] struct {
	firstItem             func(O) I
	loadFlags             func(I) uint32
	loadClass             func(I) uint32
	loadInitData          func(I) D
	loadModifier          func(D, int) M
	loadDefendCollide     func(M) F
	inversionEffect       F
	findParent            func(O) O
	loadInversionStrength func(M) int32
	nextItem              func(I) I
}

// spellProjectileInversion4FA4F0 preserves the part of GAME.EXE 004FA4F0
// reached by SpellProjectileCollide. Only equipped weapon/armor/wand-like
// inventory entries and modifier slots two and three participate. A matching
// InversionEffect still performs the otherwise-unused projectile parent lookup
// before testing the effect's signed integer strength.
func spellProjectileInversion4FA4F0[O, I, D, M, F comparable](
	target, projectile O,
	hooks spellProjectileInversionHooks4FA4F0[O, I, D, M, F],
) int32 {
	var zeroItem I
	var zeroModifier M
	var zeroFunction F
	for item := hooks.firstItem(target); item != zeroItem; item = hooks.nextItem(item) {
		if hooks.loadFlags(item)&0x00000100 == 0 {
			continue
		}
		if hooks.loadClass(item)&0x13001000 == 0 {
			continue
		}
		data := hooks.loadInitData(item)
		for slot := 2; slot < 4; slot++ {
			modifier := hooks.loadModifier(data, slot)
			if modifier == zeroModifier {
				continue
			}
			function := hooks.loadDefendCollide(modifier)
			if function == zeroFunction || function != hooks.inversionEffect {
				continue
			}
			_ = hooks.findParent(projectile)
			if hooks.loadInversionStrength(modifier) >= 1 {
				return 1
			}
		}
	}
	return 0
}
