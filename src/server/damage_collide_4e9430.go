package server

type damageCollideHooks4E9430[O, D, H comparable] struct {
	loadCollideData func(O) D
	loadHealth      func(O) H
	loadDamage      func(D) uint8
	loadFrameLow    func() uint8
	loadDamageType  func(D) int32
	findParent      func(O) O
	damage          func(O, O, O, int32, int32) int32
}

// damageCollide4E9430 preserves GAME.EXE 004E9430. The collide-data pointer
// is cached before the target and health gates. Damage one alternates between
// zero and one on the low frame bit; every other byte is divided by two with
// unsigned truncation. The registered third collision argument is not read.
func damageCollide4E9430[O, D, H comparable, C any](
	source, target O,
	collision C,
	hooks damageCollideHooks4E9430[O, D, H],
) {
	_ = collision
	collideData := hooks.loadCollideData(source)
	var zeroObject O
	if target == zeroObject {
		return
	}
	health := hooks.loadHealth(target)
	var zeroHealth H
	if health == zeroHealth {
		return
	}

	damageByte := hooks.loadDamage(collideData)
	damage := int32(damageByte >> 1)
	if damageByte == 1 && hooks.loadFrameLow()&damageByte != 0 {
		damage = 1
	}
	damageType := hooks.loadDamageType(collideData)
	parent := hooks.findParent(source)
	_ = hooks.damage(target, parent, source, damage, damageType)
}
