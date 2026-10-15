package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// Stock scripts also send source-less CRUSH and BITE to NPC PlayerDamage.
// 004E19CE bypasses facing/Reflect/equipment defenses without an attacker.
// Case 2 (004E1EE8) uses half absorption; case 8 (004E1F84) uses full
// absorption. Both preserve the cached marker and reload the live carry.
func playerDamageNPCScriptHit4E17B0(
	target *Object, cached playerDamageGreatSwordContext4E17B0,
	absorption float32, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, nil, nil, damage, typ)
	}
	liveNPC := func() bool {
		return target.Class().Has(object.ClassMonster) && !target.Class().Has(object.ClassPlayer) && target.UpdateData != nil
	}
	*cached.marker = 0
	if !liveNPC() || !playerDamageArmorReady4E17B0(target, r) || !liveNPC() {
		return reject("unsupported NPC scripted hit armor record")
	}
	factor := float64(absorption)
	if typ == object.DamageCrush {
		factor *= .5
	}
	carry := &target.UpdateDataMonster().Field1
	scaled := float32((1 - factor) * float64(damage))
	accumulated := scaled + math.Float32frombits(*carry)
	effective := playerDamageRound4E17B0(accumulated)
	*carry = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, nil, nil, damage-effective, typ, r)
	if !liveNPC() {
		return reject("unsupported live NPC scripted hit marker record")
	}
	if *cached.marker == 0 {
		*cached.marker, *cached.markerType = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing NPC scripted hit Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing NPC scripted hit default damage")
	}
	return true, r.DefaultDamage(target, nil, nil, effective, typ)
}
