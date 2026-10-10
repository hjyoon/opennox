package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// Source-less Damage(..., nil, amount, BLADE) is used by stock map scripts,
// including Wiz01a's ApprenticeStart. The caller preserves the entry flags
// and invulnerability gates and caches the NPC absorption/marker at 004E187E.
// No facing, Reflect, sword or shield prefix runs without a source (004E19CE).
func playerDamageNPCScriptBlade4E17B0(
	target *Object, cached playerDamageGreatSwordContext4E17B0,
	absorption float32, damage int32, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	const typ = object.DamageBlade
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, nil, nil, damage, typ)
	}
	liveNPC := func() bool {
		return target.Class().Has(object.ClassMonster) && !target.Class().Has(object.ClassPlayer) && target.UpdateData != nil
	}
	*cached.marker = 0
	if !liveNPC() || !playerDamageArmorReady4E17B0(target, r) || !liveNPC() {
		return reject("unsupported NPC scripted BLADE armor record")
	}
	// Case 0 enters 004E1F84, just like case 3, rather than the raw-wear
	// lava tail. Keep the f64 multiply/f32 spill, then reload LIVE carry.
	carry := &target.UpdateDataMonster().Field1
	scaled := float32((1 - float64(absorption)) * float64(damage))
	accumulated := scaled + math.Float32frombits(*carry)
	effective := playerDamageRound4E17B0(accumulated)
	*carry = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, nil, nil, damage-effective, typ, r)
	if !liveNPC() {
		return reject("unsupported live NPC scripted BLADE marker record")
	}
	if *cached.marker == 0 {
		*cached.marker, *cached.markerType = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	// 004E2025 queries GodMode after durability/marker; only a player
	// is immune. Do not turn that debug flag into NPC script immunity.
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing NPC scripted BLADE Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing NPC scripted BLADE default damage")
	}
	return true, r.DefaultDamage(target, nil, nil, effective, typ)
}
