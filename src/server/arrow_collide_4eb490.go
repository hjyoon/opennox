package server

const (
	arrowQuestFlag4EB490        = uint32(0x1000)
	arrowPlayerClassBit4EB490   = uint8(0x04)
	arrowUntargetableFlag4EB490 = uint32(0x8000)
	arrowDamageType4EB490       = uint8(11)
	arrowDefaultStrength4EB490  = int32(30)
	arrowBoltTypeName4EB490     = "ArcherBolt"
)

// arrowAttack4EB490 is the initialized portion of the temporary attack
// record passed to the item and player attack-effect chains by GAME.EXE.
// Object references deliberately use the caller's native representation.
type arrowAttack4EB490[O any] struct {
	Damage     float32
	DamageType uint8
	Radius     float32
	Owner      O
	PosX       float32
	PosY       float32
	Field24    uint32
	Source     O
}

type arrowCollideHooks4EB490[O, D, M, H any] struct {
	loadTypeIndex         func(O) uint16
	loadCollideData       func(O) D
	lookupProjectileClass func(uint16) M
	loadOwner             func(O) O
	strength              func(O) int32
	gameFlag              func(uint32) bool
	findParentPlayer      func(O) O
	loadClassLo           func(O) uint8
	isEnemy               func(O, O) bool
	tracePoint            func() (int32, int32, bool)
	calcBoltDamage        func(int32, M) float64
	floatToInt            func(float64) int32
	damageMap             func(int32, int32, int32, uint32, O)
	delayedDelete         func(O)
	loadArcherBoltType    func() uint32
	lookupType            func(string) uint32
	storeArcherBoltType   func(uint32)
	loadFlags             func(O) uint32
	loadPosX              func(O) float32
	loadPosY              func(O) float32
	loadRadius            func(O) float32
	loadDataOwner         func(D) O
	applyAttackEffect     func(O, O, *arrowAttack4EB490[O])
	preAttackEffects      func(O, O, O, *arrowAttack4EB490[O])
	targetDamage          func(O, O, O, int32, uint32) int32
	loadHealth            func(O) H
	loadHealthCur         func(H) uint16
	loadHealthMax         func(H) uint16
}

// arrowCollide4EB490 preserves GAME.EXE 004EB490. Type index and collide-data
// are entry-cached. The source owner, target damage callback, source type and
// target health remain live at the same points where the original reloads
// them. Collision is part of the registered callback ABI but is never read.
func arrowCollide4EB490[O, D, M, H comparable](
	source, target O,
	_ any,
	hooks arrowCollideHooks4EB490[O, D, M, H],
) {
	typeIndex := hooks.loadTypeIndex(source)
	data := hooks.loadCollideData(source)
	projectileClass := hooks.lookupProjectileClass(typeIndex)
	var zeroModifier M
	if projectileClass == zeroModifier {
		return
	}

	var zeroObject O
	owner := hooks.loadOwner(source)
	if owner == target && target != zeroObject {
		return
	}
	strength := arrowDefaultStrength4EB490
	if owner != zeroObject {
		strength = hooks.strength(owner)
	}

	wallImpact := func() {
		x, y, ok := hooks.tracePoint()
		if ok {
			damage := hooks.floatToInt(hooks.calcBoltDamage(strength, projectileClass))
			hooks.damageMap(x, y, damage, uint32(arrowDamageType4EB490), source)
		}
		hooks.delayedDelete(source)
	}

	if hooks.gameFlag(arrowQuestFlag4EB490) {
		parent := hooks.findParentPlayer(source)
		if parent != zeroObject {
			if target == zeroObject {
				wallImpact()
				return
			}
			if hooks.loadClassLo(parent)&arrowPlayerClassBit4EB490 != 0 &&
				hooks.loadClassLo(target)&arrowPlayerClassBit4EB490 != 0 &&
				!hooks.isEnemy(parent, target) {
				return
			}
		}
	}
	if target == zeroObject {
		wallImpact()
		return
	}

	// GAME.EXE performs this second, otherwise-unused strength call before
	// touching the ArcherBolt cache or the target flags.
	hooks.strength(hooks.loadOwner(source))
	archerBoltType := hooks.loadArcherBoltType()
	if archerBoltType == 0 {
		archerBoltType = hooks.lookupType(arrowBoltTypeName4EB490)
		hooks.storeArcherBoltType(archerBoltType)
	}
	if hooks.loadFlags(target)&arrowUntargetableFlag4EB490 != 0 {
		return
	}

	posX := hooks.loadPosX(source)
	posY := hooks.loadPosY(source)
	damage := float32(hooks.calcBoltDamage(strength, projectileClass))
	radius := hooks.loadRadius(source)
	dataOwner := hooks.loadDataOwner(data)
	attack := arrowAttack4EB490[O]{
		Damage:     damage,
		DamageType: arrowDamageType4EB490,
		Radius:     radius,
		Owner:      dataOwner,
		PosX:       posX,
		PosY:       posY,
		Field24:    0,
		Source:     source,
	}
	hooks.applyAttackEffect(source, dataOwner, &attack)
	if hooks.loadOwner(source) != target {
		hooks.preAttackEffects(target, hooks.loadDataOwner(data), source, &attack)
	}

	damageInt := hooks.floatToInt(float64(attack.Damage) + 0.5)
	parent := hooks.findParentPlayer(source)
	result := hooks.targetDamage(target, parent, source, damageInt, uint32(attack.DamageType))

	// Both values are live after the damage callback. GAME.EXE loads the
	// global type cache first and observes only AL from non-Archer callbacks.
	archerBoltType = hooks.loadArcherBoltType()
	if uint32(hooks.loadTypeIndex(source)) == archerBoltType {
		health := hooks.loadHealth(target)
		var zeroHealth H
		if health != zeroHealth {
			if hooks.loadHealthCur(health) == 0 {
				if hooks.loadHealthMax(health) != 0 {
					return
				}
			}
		}
	} else if uint8(result) == 0 {
		return
	}
	hooks.delayedDelete(source)
}
