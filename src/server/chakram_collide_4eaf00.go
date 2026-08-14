package server

const (
	chakramDestroyedFlag4EAF00    = uint32(0x20)
	chakramInvalidOwnerMask4EAF00 = uint32(0x8020)
	chakramUntargetableFlag4EAF00 = uint32(0x8000)
	chakramMaterialFXMask4EAF00   = uint8(0x30)
	chakramPlayerClassBit4EAF00   = uint8(0x04)
	chakramImpactFX4EAF00         = uint32(150)
	chakramCatchSound4EAF00       = uint32(892)
	chakramDropSound4EAF00        = uint32(893)
	chakramMapDamage4EAF00        = int32(1)
	chakramDamageType4EAF00       = uint32(0)
	chakramAttackRadiusAdd4EAF00  = float32(30)
	chakramReturnStateHome4EAF00  = uint8(0)
	chakramReturnStateDrop4EAF00  = uint8(1)
	chakramReturnStateSeek4EAF00  = uint8(2)
)

// chakramAttack4EAF00 is the initialized portion of the temporary attack
// record passed to the item and player attack-effect chains by GAME.EXE.
// Object references deliberately use the caller's native representation.
type chakramAttack4EAF00[O any] struct {
	Damage     float32
	DamageType uint8
	Radius     float32
	Owner      O
	PosX       float32
	PosY       float32
	Field24    uint32
	Source     O
}

type chakramCollideHooks4EAF00[O comparable, U, P, V, C, M any] struct {
	loadUpdateData        func(O) U
	inventoryFirst        func(O) O
	loadFlags             func(O) uint32
	loadMaterialLo        func(O) uint8
	loadOwner             func(O) O
	loadClassLo           func(O) uint8
	ownerHasWeapon        func(O) bool
	loadTypeIndex         func(O) uint16
	loadPosX              func(O) float32
	loadPosY              func(O) float32
	loadRadius            func(O) float32
	position              func(O) P
	velocity              func(O) V
	loadReflections       func(U) uint8
	storeReflections      func(U, uint8)
	loadReturnState       func(U) uint8
	storeReturnState      func(U, uint8)
	storeReturnTarget     func(U, O)
	storeLastHit          func(U, O)
	pointFX               func(uint32, P)
	wallReflect           func(C, V)
	randomReflect         func(O)
	tracePoint            func() (int32, int32, bool)
	damageMap             func(int32, int32, int32, uint32, O)
	drop                  func(O, O, P)
	delayedDelete         func(O)
	retarget              func(O)
	detach                func(O, O)
	inventoryPut          func(O, O, uint32)
	equipWeapon           func(O, O, uint32, uint32)
	audio                 func(uint32, O)
	sameTeam              func(O, O) bool
	lookupProjectileClass func(uint16) M
	strength              func(O) int32
	calcBoltDamage        func(int32, M) float32
	applyAttackEffect     func(O, O, *chakramAttack4EAF00[O])
	preAttackEffects      func(O, O, O, *chakramAttack4EAF00[O])
	floatToInt            func(float64) int32
	targetDamage          func(O, O, O, int32, uint32)
	projectileReflect     func(O, O)
	createAt              func(O, O, P)
}

// chakramCollide4EAF00 preserves GAME.EXE 004EAF00. The update-data and first
// inventory pointers are entry-cached. Owner reads in the catch path remain
// live, while the attack path caches its owner after the same-team check.
//
// Wall and unit impacts share the reflection/return state machine, but the
// original stores state and return-target in a different order on each path;
// both orders are kept here. Collision is read only for a nil-target impact.
func chakramCollide4EAF00[O comparable, U, P, V, C, M comparable](
	source, target O,
	collision C,
	hooks chakramCollideHooks4EAF00[O, U, P, V, C, M],
) {
	update := hooks.loadUpdateData(source)
	returnReady := true
	item := hooks.inventoryFirst(source)

	var zeroObject O
	if item == zeroObject || hooks.loadFlags(item)&chakramDestroyedFlag4EAF00 != 0 {
		hooks.delayedDelete(source)
		return
	}

	if target == zeroObject || hooks.loadMaterialLo(target)&chakramMaterialFXMask4EAF00 != 0 {
		hooks.pointFX(chakramImpactFX4EAF00, hooks.position(source))
	}
	if target == zeroObject {
		var zeroCollision C
		if collision == zeroCollision {
			return
		}
		if hooks.loadReflections(update) != 0 {
			hooks.wallReflect(collision, hooks.velocity(source))
		} else {
			hooks.randomReflect(source)
		}
		reflections := hooks.loadReflections(update)
		if reflections != 0 {
			hooks.storeReflections(update, reflections-1)
		} else if hooks.loadReturnState(update) == chakramReturnStateHome4EAF00 {
			hooks.storeReturnState(update, chakramReturnStateSeek4EAF00)
			returnReady = false
		}
		if x, y, ok := hooks.tracePoint(); ok {
			hooks.damageMap(x, y, chakramMapDamage4EAF00, chakramDamageType4EAF00, source)
		}
		if hooks.loadReturnState(update) == chakramReturnStateDrop4EAF00 {
			hooks.drop(source, item, hooks.position(source))
			hooks.delayedDelete(source)
			return
		}
		if hooks.loadReflections(update) == 0 && returnReady {
			hooks.storeReturnState(update, chakramReturnStateHome4EAF00)
			hooks.storeReturnTarget(update, hooks.loadOwner(source))
			return
		}
		hooks.retarget(source)
		return
	}

	if target == hooks.loadOwner(source) {
		hooks.detach(source, item)
		hooks.inventoryPut(hooks.loadOwner(source), item, 1)
		owner := hooks.loadOwner(source)
		if owner != zeroObject && hooks.loadClassLo(owner)&chakramPlayerClassBit4EAF00 != 0 &&
			!hooks.ownerHasWeapon(owner) {
			hooks.equipWeapon(owner, item, 1, 1)
		}
		hooks.audio(chakramCatchSound4EAF00, source)
		hooks.delayedDelete(source)
		return
	}
	if hooks.sameTeam(source, target) {
		return
	}

	owner := hooks.loadOwner(source)
	projectileClass := hooks.lookupProjectileClass(hooks.loadTypeIndex(source))
	if owner == zeroObject || hooks.loadFlags(owner)&chakramInvalidOwnerMask4EAF00 != 0 {
		hooks.drop(source, item, hooks.position(source))
		hooks.delayedDelete(source)
		return
	}
	var zeroProjectileClass M
	if hooks.loadFlags(target)&chakramUntargetableFlag4EAF00 != 0 || projectileClass == zeroProjectileClass {
		return
	}

	strength := hooks.strength(owner)
	posX := hooks.loadPosX(source)
	posY := hooks.loadPosY(source)
	attack := chakramAttack4EAF00[O]{
		Damage:     hooks.calcBoltDamage(strength, projectileClass),
		DamageType: uint8(chakramDamageType4EAF00),
		Radius:     hooks.loadRadius(source) + chakramAttackRadiusAdd4EAF00,
		Owner:      owner,
		PosX:       posX,
		PosY:       posY,
		Field24:    0,
		Source:     source,
	}
	hooks.applyAttackEffect(source, owner, &attack)
	hooks.preAttackEffects(target, owner, source, &attack)
	damage := hooks.floatToInt(float64(attack.Damage) + 0.5)
	hooks.targetDamage(target, owner, source, damage, chakramDamageType4EAF00)
	if hooks.loadFlags(target)&chakramInvalidOwnerMask4EAF00 == 0 {
		hooks.storeLastHit(update, target)
	}

	if hooks.loadReflections(update) != 0 {
		hooks.projectileReflect(source, target)
	} else {
		hooks.randomReflect(source)
	}
	reflections := hooks.loadReflections(update)
	if reflections == 0 && hooks.loadReturnState(update) == chakramReturnStateHome4EAF00 {
		hooks.storeReturnState(update, chakramReturnStateSeek4EAF00)
		returnReady = false
	}
	if reflections != 0 {
		hooks.storeReflections(update, reflections-1)
	}
	if hooks.loadReturnState(update) == chakramReturnStateDrop4EAF00 {
		liveItem := hooks.inventoryFirst(source)
		hooks.audio(chakramDropSound4EAF00, source)
		hooks.detach(source, liveItem)
		hooks.createAt(liveItem, zeroObject, hooks.position(source))
		hooks.delayedDelete(source)
		return
	}
	if hooks.loadReflections(update) == 0 && returnReady {
		hooks.storeReturnTarget(update, owner)
		hooks.storeReturnState(update, chakramReturnStateHome4EAF00)
		return
	}
	hooks.retarget(source)
}
