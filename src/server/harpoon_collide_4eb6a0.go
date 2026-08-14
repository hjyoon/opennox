package server

const (
	harpoonTargetRejectFlags4EB6A0 = uint32(0x8020)
	harpoonUnitClassMask4EB6A0     = uint8(0x06)
	harpoonNoCollideFlag4EB6A0     = uint32(0x40)
	harpoonGameplayFlag4EB6A0      = uint32(1)
	harpoonAbility4EB6A0           = int32(3)
	harpoonDamageType4EB6A0        = uint32(11)
	harpoonReelSound4EB6A0         = uint32(999)
	harpoonGridInverse4EB6A0       = float32(0.043478262)
)

type harpoonCollideHooks4EB6A0[O, D any] struct {
	loadDamage         func() int32
	loadOwner          func(O) O
	loadPlayerData     func(O) D
	loadBalanceDamage  func() float32
	floatToInt         func(float32) int32
	storeDamage        func(int32)
	loadTargetFlags    func(O) uint32
	findParentPlayer   func(O) O
	targetDamage       func(O, O, O, int32, uint32) int32
	isEnemy            func(O, O) bool
	gameplayFlag       func(uint32) bool
	loadClassLo        func(O) uint8
	loadNewPosY        func(O) float32
	loadNewPosX        func(O) float32
	damageMap          func(int32, int32, int32, uint32, O)
	defaultDamageSound func(O, O)
	storeTarget        func(D, O)
	disableAbility     func(O, int32)
	delayedDelete      func(O)
	storeBolt          func(D, O)
	loadPosX           func(O) float32
	loadPosY           func(O) float32
	loadFrame          func() uint32
	storeTargetX       func(D, float32)
	storeTargetY       func(D, float32)
	storeFrame         func(D, uint32)
	loadSourceFlags    func(O) uint32
	storeSourceFlags   func(O, uint32)
	markRelation       func(O, O)
	audio              func(uint32, O)
}

// harpoonCollide4EB6A0 preserves GAME.EXE 004EB6A0. The damage cache,
// source owner and owner's player update-data are entry-cached. Only the
// owner passed to markRelation is reloaded on the successful attach path.
// Collision belongs to the registered callback ABI but is never read.
func harpoonCollide4EB6A0[O, D comparable](
	source, target O,
	_ any,
	hooks harpoonCollideHooks4EB6A0[O, D],
) {
	damage := hooks.loadDamage()
	owner := hooks.loadOwner(source)
	data := hooks.loadPlayerData(owner)
	if damage == 0 {
		damage = hooks.floatToInt(hooks.loadBalanceDamage())
		hooks.storeDamage(damage)
	}

	var zeroObject O
	breakBolt := func() {
		hooks.storeTarget(data, zeroObject)
		hooks.disableAbility(owner, harpoonAbility4EB6A0)
		hooks.delayedDelete(source)
		hooks.storeBolt(data, zeroObject)
	}

	if target == zeroObject {
		y := hooks.floatToInt(hooks.loadNewPosY(source) * harpoonGridInverse4EB6A0)
		x := hooks.floatToInt(hooks.loadNewPosX(source) * harpoonGridInverse4EB6A0)
		hooks.damageMap(x, y, damage, harpoonDamageType4EB6A0, source)
		breakBolt()
		return
	}
	if hooks.loadTargetFlags(target)&harpoonTargetRejectFlags4EB6A0 != 0 || target == owner {
		return
	}

	parent := hooks.findParentPlayer(source)
	if hooks.targetDamage(target, parent, source, damage, harpoonDamageType4EB6A0) == 0 {
		hooks.defaultDamageSound(target, source)
		breakBolt()
		return
	}
	if !hooks.isEnemy(owner, target) {
		if !hooks.gameplayFlag(harpoonGameplayFlag4EB6A0) ||
			hooks.loadClassLo(target)&harpoonUnitClassMask4EB6A0 == 0 {
			hooks.defaultDamageSound(target, source)
			breakBolt()
			return
		}
	}

	hooks.storeTarget(data, target)
	hooks.storeTargetX(data, hooks.loadPosX(target))
	hooks.storeTargetY(data, hooks.loadPosY(target))
	hooks.storeFrame(data, hooks.loadFrame())
	flags := hooks.loadSourceFlags(source)
	liveOwner := hooks.loadOwner(source)
	hooks.storeSourceFlags(source, flags|harpoonNoCollideFlag4EB6A0)
	hooks.markRelation(liveOwner, target)
	hooks.audio(harpoonReelSound4EB6A0, owner)
}
