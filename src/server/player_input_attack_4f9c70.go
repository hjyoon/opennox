package server

const playerInputAttackChargedStamina4F9C70 = int32(45)

// playerAimsAtEnemyHooks4F9DC0 exposes every observable pointer load and
// predicate in GAME.EXE 004F9DC0 without assuming a host pointer width.
type playerAimsAtEnemyHooks4F9DC0[O comparable, U any] struct {
	loadUpdate func(O) U
	loadCursor func(U) O
	isEnemy    func(O, O) bool
	questMode  func() bool
}

// playerAimsAtEnemy4F9DC0 preserves the original short-circuit order. A nil
// cursor always permits an attack; otherwise hostility is checked before the
// Quest-mode fallback.
func playerAimsAtEnemy4F9DC0[O comparable, U any](
	unit O,
	hooks playerAimsAtEnemyHooks4F9DC0[O, U],
) int32 {
	var zero O
	if unit == zero {
		return 0
	}
	update := hooks.loadUpdate(unit)
	cursor := hooks.loadCursor(update)
	if cursor == zero || hooks.isEnemy(unit, cursor) || hooks.questMode() {
		return 1
	}
	return 0
}

// playerInputAttackHooks4F9C70 exposes every observable load, store, and call
// in GAME.EXE 004F9C70. Object-like values retain their native identity while
// state, flags, frame, stamina, and spell identifiers keep their original
// fixed-width domains.
type playerInputAttackHooks4F9C70[O comparable, U, P, W, D any] struct {
	aimsAtEnemy       func(O) int32
	loadUpdate        func(O) U
	loadPlayer        func(U) P
	loadWeaponEquip   func(P) uint32
	actionState       func(O) int32
	loadWeapon        func(U) W
	loadWeaponUseData func(W) D
	loadCharge        func(D) uint8
	loadMaxCharge     func(D) uint8
	subStamina        func(O, int32) int32
	loadWeaponFlags   func(D) uint32
	storeWeaponFlags  func(D, uint32)
	loadFrame         func() uint32
	storeAttackFrame  func(O, uint32)
	storeAnimFrame    func(U, uint8)
	setState          func(O, PlayerState) bool
	useByNetCode      func(O, W) int32
	loadState         func(U) PlayerState
	weaponStamina     func(uint32) int32
	adjustStamina     func(O, uint8)
	buffOff           func(O, int32) int32
	cancelDurSpell    func(int32, O) int32
}

// playerInputAttack4F9C70 preserves GAME.EXE 004F9C70's exact branch and
// callback order. In particular, the ready ranged path reloads the equipped
// weapon after SetState before invoking UseByNetCode, and the ordinary path
// reloads Player and WeaponEquip after ActionState may have changed them.
func playerInputAttack4F9C70[O comparable, U, P, W, D any](
	unit O,
	hooks playerInputAttackHooks4F9C70[O, U, P, W, D],
) {
	var zero O
	if unit == zero || hooks.aimsAtEnemy(unit) == 0 {
		return
	}

	update := hooks.loadUpdate(unit)
	player := hooks.loadPlayer(update)
	equip := hooks.loadWeaponEquip(player)
	if equip == 0 {
		if hooks.loadState(update) != PlayerState1 {
			_ = hooks.setState(unit, PlayerState1)
		}
		return
	}

	if equip&playerRangedWeaponMask4FA2B0 != 0 && hooks.actionState(unit) != 29 {
		weapon := hooks.loadWeapon(update)
		useData := hooks.loadWeaponUseData(weapon)
		charge := hooks.loadCharge(useData)
		if charge != 0 || hooks.loadMaxCharge(useData) == 0 {
			frame := hooks.loadFrame()
			hooks.storeAttackFrame(unit, frame)
			hooks.storeAnimFrame(update, 0)
			_ = hooks.setState(unit, PlayerState1)
			weapon = hooks.loadWeapon(update)
			_ = hooks.useByNetCode(unit, weapon)
		} else if hooks.subStamina(unit, playerInputAttackChargedStamina4F9C70) != 0 {
			flags := hooks.loadWeaponFlags(useData)
			hooks.storeWeaponFlags(useData, flags|2)
			frame := hooks.loadFrame()
			hooks.storeAttackFrame(unit, frame)
			hooks.storeAnimFrame(update, 0)
			_ = hooks.setState(unit, PlayerState1)
		}
	} else if hooks.loadState(update) != PlayerState1 {
		player = hooks.loadPlayer(update)
		equip = hooks.loadWeaponEquip(player)
		stamina := hooks.weaponStamina(equip)
		if hooks.subStamina(unit, stamina) != 0 {
			frame := hooks.loadFrame()
			hooks.storeAttackFrame(unit, frame)
			hooks.storeAnimFrame(update, 0)
			if !hooks.setState(unit, PlayerState1) {
				hooks.adjustStamina(unit, uint8(-stamina))
			}
		}
	}

	_ = hooks.buffOff(unit, 0)
	_ = hooks.buffOff(unit, 23)
	_ = hooks.cancelDurSpell(67, unit)
}
